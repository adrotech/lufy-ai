package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adaptive/domain"
)

func TestRecommendModesProtectedBoundaryAndCapacity(t *testing.T) {
	t.Parallel()
	base := testRuntimeConfig()
	tests := []struct {
		name       string
		config     RuntimeConfig
		request    domain.EvaluationRequest
		status     domain.AdaptiveStatus
		wantAction domain.Action
		wantGate   bool
	}{
		{name: "disabled", config: withEnabled(base, false), request: testEvaluation(), wantAction: domain.ActionDisabled},
		{name: "shadow", config: base, request: testEvaluation(), wantAction: domain.ActionRecommend},
		{name: "protected", config: base, request: withProtected(testEvaluation()), wantAction: domain.ActionEscalate},
		{
			name: "capacity exhausted", config: base, request: testEvaluation(),
			status:     domain.AdaptiveStatus{ActiveAssignments: []domain.ActiveAssignment{{AssignmentID: "a"}, {AssignmentID: "b"}}},
			wantAction: domain.ActionNoCandidate,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(LedgerPort{Status: func(context.Context, string, int, int) (domain.AdaptiveStatus, error) {
				return test.status, nil
			}})
			result, err := service.Recommend(context.Background(), RecommendRequest{
				RunID: "run-application", Evaluation: test.request, Config: test.config,
			})
			if err != nil {
				t.Fatal(err)
			}
			if result.Decision.Action != test.wantAction || result.Decision.GateAdvanced != test.wantGate || result.Durability != "not_recorded" {
				t.Fatalf("result = %#v", result)
			}
			if test.name == "capacity exhausted" && (!result.Capacity.Exhausted || result.Capacity.Limit != 2) {
				t.Fatalf("capacity = %#v", result.Capacity)
			}
		})
	}
}

func TestAdvisoryRecommendRecordsDemandAndCurrentRecommendation(t *testing.T) {
	t.Parallel()
	evaluation := testEvaluation()
	var gotRecommendation domain.Recommendation
	service := NewServiceWithClock(LedgerPort{
		Status: func(context.Context, string, int, int) (domain.AdaptiveStatus, error) {
			return domain.AdaptiveStatus{}, errors.New("status unavailable")
		},
		RecordDemand: func(_ context.Context, runID string, demand domain.DemandSignal, key string) (domain.LedgerOperation, error) {
			if runID != "run-advisory" || demand.DemandID != "demand-1" || key != "request-1:demand" {
				t.Fatalf("demand call run=%s demand=%#v key=%s", runID, demand, key)
			}
			return domain.LedgerOperation{Status: "recorded", EventID: "evt-demand", EventSequence: 7}, nil
		},
		RecordRecommendation: func(_ context.Context, runID string, recommendation domain.Recommendation, key string) (domain.LedgerOperation, error) {
			gotRecommendation = recommendation
			if runID != "run-advisory" || key != "request-1:recommendation" {
				t.Fatalf("recommendation call run=%s key=%s", runID, key)
			}
			return domain.LedgerOperation{Status: "recorded", EventID: "evt-recommendation", EventSequence: 8}, nil
		},
	}, func() time.Time { return time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC) })
	config := testRuntimeConfig()
	config.Mode = domain.ModeAdvisory
	result, err := service.Recommend(context.Background(), RecommendRequest{
		RunID: "run-advisory", Evaluation: evaluation, Config: config,
		Record: true, IdempotencyKey: "request-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Durability != "recorded" || result.DemandReceipt == nil || result.RecommendationReceipt == nil {
		t.Fatalf("result = %#v", result)
	}
	if gotRecommendation.ExpectedProjectionVersion != 7 || gotRecommendation.Fingerprint != result.Decision.Fingerprint ||
		gotRecommendation.ActorRef != result.Decision.SelectedActorRef || gotRecommendation.Score == 0 {
		t.Fatalf("recommendation = %#v", gotRecommendation)
	}
}

func TestMutationsRequireIntentRespectModeAndAllowDisabledYieldCleanup(t *testing.T) {
	t.Parallel()
	config := testRuntimeConfig()
	config.Mode = domain.ModeAdvisory
	assignment := testAssignment()
	checkpoint := testCheckpoint(assignment)
	assignCalls, yieldCalls := 0, 0
	service := NewServiceWithClock(LedgerPort{
		Assign: func(context.Context, string, domain.Assignment, string) (domain.LedgerOperation, error) {
			assignCalls++
			return domain.LedgerOperation{Status: "recorded"}, nil
		},
		Yield: func(context.Context, string, domain.YieldCheckpoint, string) (domain.LedgerOperation, error) {
			yieldCalls++
			return domain.LedgerOperation{Status: "recorded"}, nil
		},
	}, func() time.Time { return time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC) })
	_, err := service.Assign(context.Background(), MutationRequest[domain.Assignment]{Config: config, Value: assignment})
	if !errors.Is(err, ErrRecordingRequired) {
		t.Fatalf("assign without record = %v", err)
	}
	shadow := config
	shadow.Mode = domain.ModeShadow
	_, err = service.Assign(context.Background(), MutationRequest[domain.Assignment]{Config: shadow, Value: assignment, Record: true})
	if !errors.Is(err, ErrMutationDisabled) {
		t.Fatalf("shadow assign = %v", err)
	}
	if _, err := service.Assign(context.Background(), MutationRequest[domain.Assignment]{
		RunID: "run", Config: config, Value: assignment, Record: true, IdempotencyKey: "assign",
	}); err != nil {
		t.Fatal(err)
	}
	disabled := config
	disabled.Enabled = false
	if _, err := service.Yield(context.Background(), MutationRequest[domain.YieldCheckpoint]{
		RunID: "run", Config: disabled, Value: checkpoint, Record: true, IdempotencyKey: "yield",
	}); err != nil {
		t.Fatalf("disabled cleanup yield = %v", err)
	}
	if assignCalls != 1 || yieldCalls != 1 {
		t.Fatalf("assign calls=%d yield calls=%d", assignCalls, yieldCalls)
	}
}

func TestRecommendLedgerUnavailableIsNonDurableUntilRecordingRequested(t *testing.T) {
	t.Parallel()
	config := testRuntimeConfig()
	config.Mode = domain.ModeAdvisory
	service := NewService(LedgerPort{})
	result, err := service.Recommend(context.Background(), RecommendRequest{Evaluation: testEvaluation(), Config: config})
	if err != nil || result.Decision.Action != domain.ActionRecommend || result.Capacity.Availability != "not_available" {
		t.Fatalf("read-only recommendation = %#v err=%v", result, err)
	}
	_, err = service.Recommend(context.Background(), RecommendRequest{
		RunID: "run", Evaluation: testEvaluation(), Config: config, Record: true, IdempotencyKey: "request",
	})
	if !errors.Is(err, ErrLedgerUnavailable) {
		t.Fatalf("recording without ledger = %v", err)
	}
}

func testRuntimeConfig() RuntimeConfig {
	return RuntimeConfig{
		Enabled: true, Mode: domain.ModeShadow, PolicyVersion: domain.PolicyDeterministicV1,
		LeaseTTLSeconds: 900, MaxCandidates: 32, MaxWaitingItems: 128, StarvationAfterCycles: 5,
		ParallelEnabled: true, MaxParallelAgents: 3, MaxConcurrentSlices: 2,
	}
}

func withEnabled(config RuntimeConfig, enabled bool) RuntimeConfig {
	config.Enabled = enabled
	return config
}

func withProtected(request domain.EvaluationRequest) domain.EvaluationRequest {
	request.Demand.ProtectedBoundaries = []string{"delivery"}
	return request
}

func testEvaluation() domain.EvaluationRequest {
	return domain.EvaluationRequest{
		SchemaVersion: domain.EvaluationSchemaVersion,
		Demand: domain.DemandSignal{
			SchemaVersion: domain.DemandSignalSchemaVersion, DemandID: "demand-1", TaskRef: "task-1",
			SnapshotVersion: 1, Priority: 80, RequiredBudget: 40, Risk: 20, CoordinationCost: 10,
			RequiredCapabilities: []domain.RequiredCapability{{Name: "implementation", Level: 80, Weight: 5}},
		},
		Profiles: []domain.CapabilityProfile{{
			SchemaVersion: domain.CapabilityProfileSchemaVersion,
			ActorRef:      strings.Repeat("a", 64), AvailableBudget: 80, RiskTolerance: 80, CoordinationCost: 5,
			Capabilities: []domain.Capability{{Name: "implementation", Level: 90}}, RoleHints: []string{"implementer"},
		}},
	}
}

func testAssignment() domain.Assignment {
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	return domain.Assignment{
		SchemaVersion: domain.AssignmentSchemaVersion,
		AssignmentID:  "assignment-1", DemandID: "demand-1", TaskRef: "task-1",
		ActorRef: strings.Repeat("a", 64), RoleHint: "implementer",
		PolicyVersion: domain.PolicyDeterministicV1, RecommendationFingerprint: strings.Repeat("b", 64),
		Priority: 80, RequiredBudget: 40, Score: 900, ExpectedProjectionVersion: 3,
		Lease: domain.AssignmentLease{TokenDigest: strings.Repeat("c", 64), ExpiresAt: now.Add(15 * time.Minute)},
	}
}

func testCheckpoint(assignment domain.Assignment) domain.YieldCheckpoint {
	return domain.YieldCheckpoint{
		SchemaVersion: domain.YieldSchemaVersion, AssignmentID: assignment.AssignmentID,
		DemandID: assignment.DemandID, ActorRef: assignment.ActorRef,
		ExpectedProjectionVersion: 4, Lease: assignment.Lease, Reason: "manual", NextStatus: "completed",
	}
}
