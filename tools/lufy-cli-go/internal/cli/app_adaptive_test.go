package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	adaptiveapp "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adaptive/application"
	adaptivedomain "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adaptive/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
)

func TestAdaptiveRecommendIsReadOnlyAndDisabledByDefault(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	input := mustJSON(t, cliAdaptiveEvaluation())
	var stdout, stderr bytes.Buffer
	code := Run([]string{"adaptive", "recommend", "--target", target, "--json"}, Dependencies{
		Stdin: bytes.NewReader(input), Stdout: &stdout, Stderr: &stderr,
	})
	if code != ExitOK {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var result adaptiveapp.RecommendationResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Decision.Action != adaptivedomain.ActionDisabled || result.Durability != "not_recorded" || result.Decision.GateAdvanced {
		t.Fatalf("result = %#v", result)
	}
}

func TestAdaptiveCLIAdvisoryRecommendAssignYieldStatus(t *testing.T) {
	target := t.TempDir()
	writeAdaptiveCLIConfig(t, target, true, "advisory")
	evaluation := cliAdaptiveEvaluation()

	recommendation := adaptiveapp.RecommendationResult{}
	runAdaptiveJSONCommand(t, target, []string{
		"adaptive", "recommend", "--run", "run-cli-adaptive", "--record", "--idempotency-key", "request-1", "--json",
	}, mustJSON(t, evaluation), &recommendation)
	if recommendation.Durability != "recorded" || recommendation.RecommendationReceipt == nil ||
		recommendation.RecommendationReceipt.EventSequence != 2 || recommendation.Decision.GateAdvanced {
		t.Fatalf("recommendation = %#v", recommendation)
	}
	selected := recommendation.Decision.Ranking[0]
	assignment := adaptivedomain.Assignment{
		SchemaVersion: adaptivedomain.AssignmentSchemaVersion,
		AssignmentID:  "assignment-cli-1", DemandID: evaluation.Demand.DemandID, TaskRef: evaluation.Demand.TaskRef,
		ActorRef: recommendation.Decision.SelectedActorRef, RoleHint: recommendation.Decision.SelectedRoleHint,
		PolicyVersion:             recommendation.Decision.PolicyVersion,
		RecommendationFingerprint: recommendation.Decision.Fingerprint,
		Priority:                  evaluation.Demand.Priority, RequiredBudget: evaluation.Demand.RequiredBudget,
		Score: selected.Breakdown.Total, ExpectedProjectionVersion: recommendation.RecommendationReceipt.EventSequence,
		Lease: adaptivedomain.AssignmentLease{
			TokenDigest: strings.Repeat("c", 64), ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		},
	}
	assigned := adaptivedomain.LedgerOperation{}
	runAdaptiveJSONCommand(t, target, []string{
		"adaptive", "assign", "--run", "run-cli-adaptive", "--record", "--idempotency-key", "assign-1", "--json",
	}, mustJSON(t, assignment), &assigned)
	if assigned.Status != "recorded" || assigned.EventSequence != 3 || len(assigned.State.ActiveAssignments) != 1 {
		t.Fatalf("assigned = %#v", assigned)
	}

	checkpoint := adaptivedomain.YieldCheckpoint{
		SchemaVersion: adaptivedomain.YieldSchemaVersion,
		AssignmentID:  assignment.AssignmentID, DemandID: assignment.DemandID, ActorRef: assignment.ActorRef,
		ExpectedProjectionVersion: assigned.EventSequence, Lease: assignment.Lease,
		Reason: "manual", NextStatus: "waiting", SuccessorCapabilities: []string{"diagnosis"},
	}
	yielded := adaptivedomain.LedgerOperation{}
	runAdaptiveJSONCommand(t, target, []string{
		"adaptive", "yield", "--run", "run-cli-adaptive", "--record", "--idempotency-key", "yield-1", "--json",
	}, mustJSON(t, checkpoint), &yielded)
	if yielded.Status != "recorded" || yielded.EventSequence != 4 || len(yielded.State.ActiveAssignments) != 0 ||
		len(yielded.State.Waiting) != 1 || yielded.State.Waiting[0].WaitingCycle != 1 {
		t.Fatalf("yielded = %#v", yielded)
	}

	status := adaptivedomain.AdaptiveStatus{}
	runAdaptiveJSONCommand(t, target, []string{
		"adaptive", "status", "--run", "run-cli-adaptive", "--json",
	}, nil, &status)
	if status.Version != 4 || len(status.ActiveAssignments) != 0 || len(status.Waiting) != 1 {
		t.Fatalf("status = %#v", status)
	}
}

func TestAdaptiveCLIShadowFileInputHasNoDurableEffects(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	writeAdaptiveCLIConfig(t, target, true, "shadow")
	inputPath := filepath.Join(target, "evaluation.json")
	if err := os.WriteFile(inputPath, mustJSON(t, cliAdaptiveEvaluation()), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"adaptive", "recommend", "--target", target, "--run", "run-cli-shadow", "--file", inputPath, "--json",
	}, Dependencies{Stdout: &stdout, Stderr: &stderr})
	if code != ExitOK {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var recommendation adaptiveapp.RecommendationResult
	if err := json.Unmarshal(stdout.Bytes(), &recommendation); err != nil {
		t.Fatal(err)
	}
	if recommendation.Decision.Action != adaptivedomain.ActionRecommend || recommendation.Durability != "not_recorded" {
		t.Fatalf("shadow recommendation = %#v", recommendation)
	}
	status := adaptivedomain.AdaptiveStatus{}
	runAdaptiveJSONCommand(t, target, []string{
		"adaptive", "status", "--run", "run-cli-shadow", "--json",
	}, nil, &status)
	if status.Version != 0 || len(status.ActiveAssignments) != 0 || len(status.Waiting) != 0 {
		t.Fatalf("shadow created durable state: %#v", status)
	}
}

func TestAdaptiveMutationsRequireExplicitRecordAndStableExit(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	writeAdaptiveCLIConfig(t, target, true, "advisory")
	assignment := adaptivedomain.Assignment{
		SchemaVersion: adaptivedomain.AssignmentSchemaVersion,
		AssignmentID:  "assignment-rejected", DemandID: "demand-rejected", TaskRef: "task-rejected",
		ActorRef: strings.Repeat("a", 64), RoleHint: "implementer",
		PolicyVersion:             adaptivedomain.PolicyDeterministicV1,
		RecommendationFingerprint: strings.Repeat("b", 64),
		Priority:                  80, RequiredBudget: 40, Score: 900, ExpectedProjectionVersion: 1,
		Lease: adaptivedomain.AssignmentLease{TokenDigest: strings.Repeat("c", 64), ExpiresAt: time.Now().UTC().Add(10 * time.Minute)},
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"adaptive", "assign", "--run", "run-cli-rejected", "--idempotency-key", "assign-rejected",
	}, Dependencies{Stdin: bytes.NewReader(mustJSON(t, assignment)), Stdout: &stdout, Stderr: &stderr})
	if code != ExitResultRejected || !strings.Contains(stderr.String(), "recording explícito") {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"adaptive", "--help"}, Dependencies{Stdout: &stdout, Stderr: &stderr}); code != ExitOK || !strings.Contains(stdout.String(), "4=rejected 5=conflict 6=unavailable") {
		t.Fatalf("help code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func writeAdaptiveCLIConfig(t *testing.T, target string, enabled bool, mode string) {
	t.Helper()
	config, err := projectconfig.Scan(target, time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	config.AdaptiveRouting.Enabled = enabled
	config.AdaptiveRouting.Mode = mode
	path := filepath.Join(target, projectconfig.ProjectConfigPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := (projectconfig.ConfigStore{}).Write(path, config); err != nil {
		t.Fatal(err)
	}
}

func runAdaptiveJSONCommand(t *testing.T, target string, args []string, input []byte, destination any) {
	t.Helper()
	args = append(args, "--target", target)
	var stdout, stderr bytes.Buffer
	code := Run(args, Dependencies{Stdin: bytes.NewReader(input), Stdout: &stdout, Stderr: &stderr})
	if code != ExitOK {
		t.Fatalf("Run(%v) code=%d stderr=%s stdout=%s", args, code, stderr.String(), stdout.String())
	}
	if err := json.Unmarshal(stdout.Bytes(), destination); err != nil {
		t.Fatalf("Run(%v) invalid JSON %q: %v", args, stdout.String(), err)
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func cliAdaptiveEvaluation() adaptivedomain.EvaluationRequest {
	return adaptivedomain.EvaluationRequest{
		SchemaVersion: adaptivedomain.EvaluationSchemaVersion,
		Demand: adaptivedomain.DemandSignal{
			SchemaVersion: adaptivedomain.DemandSignalSchemaVersion,
			DemandID:      "demand-cli-1", TaskRef: "task-cli-1", SnapshotVersion: 1,
			Priority: 80, RequiredBudget: 40, Risk: 20, CoordinationCost: 10,
			RequiredCapabilities: []adaptivedomain.RequiredCapability{{Name: "implementation", Level: 80, Weight: 5}},
		},
		Profiles: []adaptivedomain.CapabilityProfile{{
			SchemaVersion: adaptivedomain.CapabilityProfileSchemaVersion,
			ActorRef:      strings.Repeat("a", 64), AvailableBudget: 80, RiskTolerance: 80, CoordinationCost: 5,
			Capabilities: []adaptivedomain.Capability{{Name: "implementation", Level: 90}},
			RoleHints:    []string{"implementer"},
		}},
	}
}
