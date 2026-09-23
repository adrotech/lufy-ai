package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	adaptivedomain "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adaptive/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

func TestLedgerAdapterAssignYieldReleasesOnlyAfterDurableReceipt(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	store := newAdaptiveStore(t, now)
	adapter := NewLedgerAdapter(store, func() time.Time { return now })
	ctx := context.Background()

	demandResult, err := adapter.RecordDemand(ctx, RecordDemandRequest{
		RunID: "run-adaptive-e2e", Demand: testDemand(), IdempotencyKey: "demand:1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if demandResult.Status != runledger.AppendRecorded || len(demandResult.State.Waiting) != 1 || demandResult.State.Version != 1 {
		t.Fatalf("demand result = %#v", demandResult)
	}

	recommended := recordTestRecommendation(t, adapter, ctx, "run-adaptive-e2e", demandResult, "recommend:1")
	assignment := testAssignment(recommended.State.Version, now)
	assignResult, err := adapter.Assign(ctx, AssignRequest{
		RunID: "run-adaptive-e2e", Assignment: assignment, IdempotencyKey: "assign:1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if assignResult.Status != runledger.AppendRecorded || len(assignResult.State.ActiveAssignments) != 1 ||
		len(assignResult.State.Waiting) != 0 || len(assignResult.State.ConsumedBudget) != 1 ||
		assignResult.State.ConsumedBudget[0].Consumed != assignment.RequiredBudget {
		t.Fatalf("assign result = %#v", assignResult)
	}

	checkpoint := testYield(assignResult.State.Version, assignment)
	yieldResult, err := adapter.Yield(ctx, YieldRequest{
		RunID: "run-adaptive-e2e", Checkpoint: checkpoint, IdempotencyKey: "yield:1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if yieldResult.Status != runledger.AppendRecorded || len(yieldResult.State.ActiveAssignments) != 0 ||
		len(yieldResult.State.ConsumedBudget) != 0 || len(yieldResult.State.Waiting) != 1 ||
		yieldResult.State.Waiting[0].WaitingCycle != 1 {
		t.Fatalf("yield result = %#v", yieldResult)
	}

	retry, err := adapter.Yield(ctx, YieldRequest{
		RunID: "run-adaptive-e2e", Checkpoint: checkpoint, IdempotencyKey: "yield:1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if retry.Status != runledger.AppendDuplicateNoop || retry.EventID != yieldResult.EventID ||
		retry.State.Waiting[0].WaitingCycle != 1 {
		t.Fatalf("yield retry = %#v", retry)
	}
	conflicting := checkpoint
	conflicting.NextStatus = "completed"
	_, err = adapter.Yield(ctx, YieldRequest{
		RunID: "run-adaptive-e2e", Checkpoint: conflicting, IdempotencyKey: "yield:1",
	})
	if !errors.Is(err, runledger.ErrIdempotencyConflict) {
		t.Fatalf("conflicting yield error = %v", err)
	}
	status, err := adapter.Status(ctx, "run-adaptive-e2e", 128, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.ActiveAssignments) != 0 || len(status.Waiting) != 1 || status.Waiting[0].WaitingCycle != 1 {
		t.Fatalf("conflicting retry mutated state: %#v", status)
	}
	events, err := store.LoadRun(ctx, "run-adaptive-e2e")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 4 || events[1].CausedByEventID != events[0].EventID ||
		events[2].CausedByEventID != events[1].EventID || events[3].CausedByEventID != events[2].EventID {
		t.Fatalf("adaptive causal chain = %#v", events)
	}
	report, err := store.VerifyRun(ctx, "run-adaptive-e2e")
	if err != nil {
		t.Fatal(err)
	}
	if !report.Healthy || report.EventCount != 4 {
		t.Fatalf("adaptive causal verification = %#v", report)
	}
}

func TestLedgerAdapterStaleYieldDoesNotReleaseAssignment(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	store := newAdaptiveStore(t, now)
	adapter := NewLedgerAdapter(store, func() time.Time { return now })
	ctx := context.Background()
	demand, err := adapter.RecordDemand(ctx, RecordDemandRequest{
		RunID: "run-stale-yield", Demand: testDemand(), IdempotencyKey: "demand:stale",
	})
	if err != nil {
		t.Fatal(err)
	}
	recommended := recordTestRecommendation(t, adapter, ctx, "run-stale-yield", demand, "recommend:stale")
	assignment := testAssignment(recommended.State.Version, now)
	assigned, err := adapter.Assign(ctx, AssignRequest{
		RunID: "run-stale-yield", Assignment: assignment, IdempotencyKey: "assign:stale",
	})
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := testYield(assigned.State.Version-1, assignment)
	_, err = adapter.Yield(ctx, YieldRequest{
		RunID: "run-stale-yield", Checkpoint: checkpoint, IdempotencyKey: "yield:stale",
	})
	if !errors.Is(err, runledger.ErrVersionConflict) {
		t.Fatalf("stale yield error = %v", err)
	}
	status, err := adapter.Status(ctx, "run-stale-yield", 128, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.ActiveAssignments) != 1 || len(status.ConsumedBudget) != 1 {
		t.Fatalf("stale yield released state: %#v", status)
	}
}

func TestLedgerAdapterConcurrentAssignmentsFenceOneWriter(t *testing.T) {
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	store := newAdaptiveStore(t, now)
	adapter := NewLedgerAdapter(store, func() time.Time { return now })
	ctx := context.Background()
	demand, err := adapter.RecordDemand(ctx, RecordDemandRequest{
		RunID: "run-assignment-race", Demand: testDemand(), IdempotencyKey: "demand:race",
	})
	if err != nil {
		t.Fatal(err)
	}
	recommended := recordTestRecommendation(t, adapter, ctx, "run-assignment-race", demand, "recommend:race")
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for index := 0; index < 2; index++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			assignment := testAssignment(recommended.State.Version, now)
			assignment.AssignmentID = fmt.Sprintf("assignment-%d", index+1)
			_, assignErr := adapter.Assign(ctx, AssignRequest{
				RunID: "run-assignment-race", Assignment: assignment,
				IdempotencyKey: fmt.Sprintf("assign:race:%d", index),
			})
			results <- assignErr
		}(index)
	}
	wg.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, runledger.ErrVersionConflict), errors.Is(err, ErrAssignmentUnavailable):
			conflicts++
		default:
			t.Fatalf("unexpected assignment error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
	status, err := adapter.Status(ctx, "run-assignment-race", 128, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.ActiveAssignments) != 1 {
		t.Fatalf("active assignments = %#v", status.ActiveAssignments)
	}
}

func TestLedgerAdapterConcurrentYieldsReleaseOnlyOnce(t *testing.T) {
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	store := newAdaptiveStore(t, now)
	adapter := NewLedgerAdapter(store, func() time.Time { return now })
	ctx := context.Background()
	demand, err := adapter.RecordDemand(ctx, RecordDemandRequest{
		RunID: "run-yield-race", Demand: testDemand(), IdempotencyKey: "demand:yield-race",
	})
	if err != nil {
		t.Fatal(err)
	}
	recommended := recordTestRecommendation(t, adapter, ctx, "run-yield-race", demand, "recommend:yield-race")
	assignment := testAssignment(recommended.State.Version, now)
	assigned, err := adapter.Assign(ctx, AssignRequest{
		RunID: "run-yield-race", Assignment: assignment, IdempotencyKey: "assign:yield-race",
	})
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	results := make(chan error, 2)
	for index := 0; index < 2; index++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			checkpoint := testYield(assigned.State.Version, assignment)
			checkpoint.NextStatus = []string{"waiting", "completed"}[index]
			_, yieldErr := adapter.Yield(ctx, YieldRequest{
				RunID: "run-yield-race", Checkpoint: checkpoint,
				IdempotencyKey: fmt.Sprintf("yield:race:%d", index),
			})
			results <- yieldErr
		}(index)
	}
	wg.Wait()
	close(results)
	successes, conflicts := 0, 0
	for resultErr := range results {
		switch {
		case resultErr == nil:
			successes++
		case errors.Is(resultErr, runledger.ErrVersionConflict), errors.Is(resultErr, ErrAssignmentUnavailable):
			conflicts++
		default:
			t.Fatalf("unexpected yield error: %v", resultErr)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
	status, err := adapter.Status(ctx, "run-yield-race", 128, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.ActiveAssignments) != 0 || len(status.ConsumedBudget) != 0 || status.Version != 4 {
		t.Fatalf("yield race status = %#v", status)
	}
}

func TestLedgerAdapterYieldRejectsOwnerLeaseAndExpiry(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		mutate  func(*adaptivedomain.YieldCheckpoint, *time.Time)
		wantErr error
	}{
		{
			name: "owner mismatch",
			mutate: func(value *adaptivedomain.YieldCheckpoint, _ *time.Time) {
				value.ActorRef = strings.Repeat("9", 64)
			},
			wantErr: ErrAssignmentUnavailable,
		},
		{
			name: "lease digest mismatch",
			mutate: func(value *adaptivedomain.YieldCheckpoint, _ *time.Time) {
				value.Lease.TokenDigest = strings.Repeat("8", 64)
			},
			wantErr: ErrLeaseMismatch,
		},
		{
			name: "lease expiry mismatch",
			mutate: func(value *adaptivedomain.YieldCheckpoint, _ *time.Time) {
				value.Lease.ExpiresAt = value.Lease.ExpiresAt.Add(time.Second)
			},
			wantErr: ErrLeaseMismatch,
		},
		{
			name: "expired lease",
			mutate: func(value *adaptivedomain.YieldCheckpoint, clock *time.Time) {
				*clock = value.Lease.ExpiresAt
			},
			wantErr: ErrLeaseExpired,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
			clock := now
			store := newAdaptiveStore(t, now)
			adapter := NewLedgerAdapter(store, func() time.Time { return clock })
			ctx := context.Background()
			demand, err := adapter.RecordDemand(ctx, RecordDemandRequest{
				RunID: "run-lease-check", Demand: testDemand(), IdempotencyKey: "demand:lease-check",
			})
			if err != nil {
				t.Fatal(err)
			}
			recommended := recordTestRecommendation(t, adapter, ctx, "run-lease-check", demand, "recommend:lease-check")
			assignment := testAssignment(recommended.State.Version, now)
			assigned, err := adapter.Assign(ctx, AssignRequest{
				RunID: "run-lease-check", Assignment: assignment, IdempotencyKey: "assign:lease-check",
			})
			if err != nil {
				t.Fatal(err)
			}
			checkpoint := testYield(assigned.State.Version, assignment)
			test.mutate(&checkpoint, &clock)
			_, err = adapter.Yield(ctx, YieldRequest{
				RunID: "run-lease-check", Checkpoint: checkpoint, IdempotencyKey: "yield:lease-check",
			})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("yield error = %v, want %v", err, test.wantErr)
			}
			status, statusErr := adapter.Status(ctx, "run-lease-check", 128, 5)
			if statusErr != nil {
				t.Fatal(statusErr)
			}
			if len(status.ActiveAssignments) != 1 || len(status.ConsumedBudget) != 1 || len(status.Waiting) != 0 {
				t.Fatalf("rejected yield mutated state: %#v", status)
			}
		})
	}
}

func TestLedgerAdapterStorageFailurePreservesAssignment(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	store := newAdaptiveStore(t, now)
	adapter := NewLedgerAdapter(store, func() time.Time { return now })
	ctx := context.Background()
	demand, err := adapter.RecordDemand(ctx, RecordDemandRequest{
		RunID: "run-storage-failure", Demand: testDemand(), IdempotencyKey: "demand:storage-failure",
	})
	if err != nil {
		t.Fatal(err)
	}
	recommended := recordTestRecommendation(t, adapter, ctx, "run-storage-failure", demand, "recommend:storage-failure")
	assignment := testAssignment(recommended.State.Version, now)
	assigned, err := adapter.Assign(ctx, AssignRequest{
		RunID: "run-storage-failure", Assignment: assignment, IdempotencyKey: "assign:storage-failure",
	})
	if err != nil {
		t.Fatal(err)
	}
	injected := errors.New("storage unavailable")
	failing := NewLedgerAdapter(&appendFailingStore{Store: store, err: injected}, func() time.Time { return now })
	_, err = failing.Yield(ctx, YieldRequest{
		RunID: "run-storage-failure", Checkpoint: testYield(assigned.State.Version, assignment),
		IdempotencyKey: "yield:storage-failure",
	})
	if !errors.Is(err, injected) {
		t.Fatalf("yield error = %v", err)
	}
	status, err := adapter.Status(ctx, "run-storage-failure", 128, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.ActiveAssignments) != 1 || len(status.ConsumedBudget) != 1 || status.Version != 3 {
		t.Fatalf("storage failure released assignment: %#v", status)
	}
}

func TestLedgerAdapterRejectsHistoricalAssignmentIDReuse(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	store := newAdaptiveStore(t, now)
	adapter := NewLedgerAdapter(store, func() time.Time { return now })
	ctx := context.Background()
	firstDemand, err := adapter.RecordDemand(ctx, RecordDemandRequest{
		RunID: "run-assignment-reuse", Demand: testDemand(), IdempotencyKey: "demand:reuse:1",
	})
	if err != nil {
		t.Fatal(err)
	}
	firstRecommended := recordTestRecommendation(t, adapter, ctx, "run-assignment-reuse", firstDemand, "recommend:reuse:1")
	firstAssignment := testAssignment(firstRecommended.State.Version, now)
	assigned, err := adapter.Assign(ctx, AssignRequest{
		RunID: "run-assignment-reuse", Assignment: firstAssignment, IdempotencyKey: "assign:reuse:1",
	})
	if err != nil {
		t.Fatal(err)
	}
	completed := testYield(assigned.State.Version, firstAssignment)
	completed.NextStatus = "completed"
	if _, err := adapter.Yield(ctx, YieldRequest{
		RunID: "run-assignment-reuse", Checkpoint: completed, IdempotencyKey: "yield:reuse:1",
	}); err != nil {
		t.Fatal(err)
	}
	secondDemand := testDemand()
	secondDemand.DemandID = "demand-2"
	secondDemand.TaskRef = "task-2"
	recorded, err := adapter.RecordDemand(ctx, RecordDemandRequest{
		RunID: "run-assignment-reuse", Demand: secondDemand, IdempotencyKey: "demand:reuse:2",
	})
	if err != nil {
		t.Fatal(err)
	}
	secondRecommended := recordTestRecommendation(t, adapter, ctx, "run-assignment-reuse", recorded, "recommend:reuse:2")
	reused := testAssignment(secondRecommended.State.Version, now)
	reused.DemandID = secondDemand.DemandID
	reused.TaskRef = secondDemand.TaskRef
	_, err = adapter.Assign(ctx, AssignRequest{
		RunID: "run-assignment-reuse", Assignment: reused, IdempotencyKey: "assign:reuse:2",
	})
	if !errors.Is(err, ErrAssignmentUnavailable) {
		t.Fatalf("assignment id reuse error = %v", err)
	}
}

func TestProjectIgnoresHistoricalNonAdaptiveEventsAndOrdersWaiting(t *testing.T) {
	t.Parallel()
	events := []runledger.Event{{
		SchemaVersion: runledger.SchemaVersion, EventID: "evt-old", RunID: "run-history",
		LamportClock: 1, LocalSequence: 1, ObservedAt: time.Now().UTC(),
		Kind: runledger.KindStart, Source: runledger.Source{Adapter: "codex", EventName: "SessionStart"},
		IdempotencyKeyHash: strings.Repeat("a", 64), Fingerprint: strings.Repeat("b", 64),
	}}
	status := Project("run-history", events, 1, 1)
	if status.Version != 1 || len(status.ActiveAssignments) != 0 || len(status.Waiting) != 0 {
		t.Fatalf("historical status = %#v", status)
	}
}

func TestProjectOrdersBoundsAndRebuildsWaitingPoolDeterministically(t *testing.T) {
	t.Parallel()
	events := []runledger.Event{
		adaptiveDemandEvent("evt-001", "demand-a", 1, 90, 1),
		adaptiveDemandEvent("evt-002", "demand-b", 2, 90, 5),
		adaptiveDemandEvent("evt-003", "demand-c", 3, 80, 9),
	}
	first := Project("run-waiting", events, 2, 5)
	second := Project("run-waiting", events, 2, 5)
	if first.Version != 3 || !first.Truncated || len(first.Waiting) != 2 {
		t.Fatalf("waiting status = %#v", first)
	}
	if first.Waiting[0].DemandID != "demand-b" || !first.Waiting[0].StarvationRisk ||
		first.Waiting[1].DemandID != "demand-a" || first.Waiting[1].StarvationRisk {
		t.Fatalf("waiting order/starvation = %#v", first.Waiting)
	}
	if first.Waiting[0].Priority != 90 || first.Waiting[1].Priority != 90 {
		t.Fatalf("projection silently changed priority: %#v", first.Waiting)
	}
	firstJSON, err := json.MarshalIndent(first, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.MarshalIndent(second, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("projection replay changed:\n%s\n%s", firstJSON, secondJSON)
	}
	golden, err := os.ReadFile(filepath.Join("testdata", "adaptive-status.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	var compactGolden, compactActual bytes.Buffer
	if err := json.Compact(&compactGolden, golden); err != nil {
		t.Fatal(err)
	}
	if err := json.Compact(&compactActual, firstJSON); err != nil {
		t.Fatal(err)
	}
	if compactGolden.String() != compactActual.String() {
		t.Fatalf("adaptive status differs from golden:\n%s", firstJSON)
	}
}

func TestAdaptiveEventsPersistOnlyContentFreeReferences(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	store := newAdaptiveStore(t, now)
	adapter := NewLedgerAdapter(store, func() time.Time { return now })
	ctx := context.Background()
	demand, err := adapter.RecordDemand(ctx, RecordDemandRequest{
		RunID: "run-privacy", Demand: testDemand(), IdempotencyKey: "demand:privacy",
	})
	if err != nil {
		t.Fatal(err)
	}
	recommended := recordTestRecommendation(t, adapter, ctx, "run-privacy", demand, "recommend:privacy")
	assignment := testAssignment(recommended.State.Version, now)
	assigned, err := adapter.Assign(ctx, AssignRequest{
		RunID: "run-privacy", Assignment: assignment, IdempotencyKey: "assign:privacy",
	})
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := testYield(assigned.State.Version, assignment)
	if _, err := adapter.Yield(ctx, YieldRequest{
		RunID: "run-privacy", Checkpoint: checkpoint, IdempotencyKey: "yield:privacy",
	}); err != nil {
		t.Fatal(err)
	}
	events, err := store.LoadRun(ctx, "run-privacy")
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"super-secret", "prompt", "response", "summary", "hypothesis_text", "failed_attempt_text"} {
		if strings.Contains(strings.ToLower(string(body)), forbidden) {
			t.Fatalf("durable events contain %q: %s", forbidden, body)
		}
	}
}

func newAdaptiveStore(t *testing.T, now time.Time) *runledger.FileStore {
	t.Helper()
	var mu sync.Mutex
	next := 0
	store, err := runledger.NewFileStore(t.TempDir(), runledger.Options{
		Now: func() time.Time { return now },
		NewID: func() (string, error) {
			mu.Lock()
			defer mu.Unlock()
			next++
			return fmt.Sprintf("evt-adaptive-%03d", next), nil
		},
		LockTimeout: 3 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	return store
}

type appendFailingStore struct {
	runledger.Store
	err error
}

func (s *appendFailingStore) Append(context.Context, runledger.AppendRequest) (runledger.AppendResult, error) {
	return runledger.AppendResult{}, s.err
}

func adaptiveDemandEvent(eventID, demandID string, sequence uint64, priority, waitingCycle int) runledger.Event {
	return runledger.Event{
		SchemaVersion: runledger.SchemaVersion, EventID: eventID, RunID: "run-waiting",
		LamportClock: sequence, LocalSequence: sequence,
		ObservedAt: time.Date(2026, 9, 22, 12, 0, int(sequence), 0, time.UTC),
		Kind:       runledger.KindDemand, Source: runledger.Source{Adapter: "lufy", EventName: "adaptive_demand"},
		TaskRef: "task-" + demandID,
		Adaptive: &runledger.AdaptiveMetadata{
			SchemaVersion: runledger.AdaptiveMetadataSchemaVersion,
			DemandID:      demandID, Priority: priority, RequiredBudget: 40, WaitingCycle: waitingCycle,
		},
		IdempotencyKeyHash: strings.Repeat(string(rune('a'+sequence)), 64),
		Fingerprint:        strings.Repeat(string(rune('d'+sequence)), 64),
	}
}

func testDemand() adaptivedomain.DemandSignal {
	return adaptivedomain.DemandSignal{
		SchemaVersion: adaptivedomain.DemandSignalSchemaVersion,
		DemandID:      "demand-1", TaskRef: "task-1", SnapshotVersion: 1,
		Priority: 80, RequiredBudget: 40, Risk: 30, CoordinationCost: 10,
		RequiredCapabilities: []adaptivedomain.RequiredCapability{{Name: "implementation", Level: 80, Weight: 5}},
	}
}

func recordTestRecommendation(
	t *testing.T,
	adapter *LedgerAdapter,
	ctx context.Context,
	runID string,
	demand OperationResult,
	idempotencyKey string,
) OperationResult {
	t.Helper()
	if len(demand.State.Waiting) == 0 {
		t.Fatal("demand state has no waiting item")
	}
	waiting := demand.State.Waiting[0]
	result, err := adapter.RecordRecommendation(ctx, RecordRecommendationRequest{
		RunID: runID,
		Recommendation: adaptivedomain.Recommendation{
			SchemaVersion: adaptivedomain.RecommendationSchemaVersion,
			DemandID:      waiting.DemandID, TaskRef: waiting.TaskRef,
			ActorRef: strings.Repeat("a", 64), RoleHint: "implementer",
			PolicyVersion: adaptivedomain.PolicyDeterministicV1,
			Fingerprint:   strings.Repeat("b", 64),
			Priority:      waiting.Priority, RequiredBudget: waiting.RequiredBudget, Score: 900,
			ExpectedProjectionVersion: demand.State.Version,
		},
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func testAssignment(version uint64, now time.Time) adaptivedomain.Assignment {
	return adaptivedomain.Assignment{
		SchemaVersion: adaptivedomain.AssignmentSchemaVersion,
		AssignmentID:  "assignment-1", DemandID: "demand-1", TaskRef: "task-1",
		ActorRef: strings.Repeat("a", 64), RoleHint: "implementer",
		PolicyVersion:             adaptivedomain.PolicyDeterministicV1,
		RecommendationFingerprint: strings.Repeat("b", 64),
		Priority:                  80, RequiredBudget: 40, Score: 900, ExpectedProjectionVersion: version,
		Lease: adaptivedomain.AssignmentLease{TokenDigest: strings.Repeat("c", 64), ExpiresAt: now.Add(15 * time.Minute)},
	}
}

func testYield(version uint64, assignment adaptivedomain.Assignment) adaptivedomain.YieldCheckpoint {
	return adaptivedomain.YieldCheckpoint{
		SchemaVersion: adaptivedomain.YieldSchemaVersion,
		AssignmentID:  assignment.AssignmentID, DemandID: assignment.DemandID, ActorRef: assignment.ActorRef,
		ExpectedProjectionVersion: version, Lease: assignment.Lease,
		Reason: "blocked", NextStatus: "waiting",
		ArtifactRefs: []adaptivedomain.ArtifactReference{{
			Kind: "source", PathSHA256: strings.Repeat("d", 64), ContentSHA256: strings.Repeat("e", 64),
		}},
		EvidenceRefs: []adaptivedomain.EvidenceReference{{
			Category: "test", Result: "failed", PathSHA256: strings.Repeat("f", 64),
		}},
		HypothesisRefs:        []string{strings.Repeat("1", 64)},
		FailedAttemptRefs:     []string{strings.Repeat("2", 64)},
		SuccessorCapabilities: []string{"diagnosis"},
	}
}
