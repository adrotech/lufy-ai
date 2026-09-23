package domain

import (
	"strings"
	"testing"
)

func TestEvaluateRanksDeterministicallyAndExplainsTerms(t *testing.T) {
	t.Parallel()
	request := validEvaluationRequest()
	request.Profiles = []CapabilityProfile{
		{
			SchemaVersion:   CapabilityProfileSchemaVersion,
			ActorRef:        strings.Repeat("b", 64),
			Capabilities:    []Capability{{Name: "implementation", Level: 90}},
			AvailableBudget: 80, RiskTolerance: 70, CoordinationCost: 10,
			RoleHints: []string{"implementer"},
		},
		{
			SchemaVersion:   CapabilityProfileSchemaVersion,
			ActorRef:        strings.Repeat("a", 64),
			Capabilities:    []Capability{{Name: "implementation", Level: 90}},
			AvailableBudget: 80, RiskTolerance: 70, CoordinationCost: 10,
			RoleHints: []string{"implementer"},
		},
	}

	first, err := Evaluate(request, DefaultPolicy(ModeShadow))
	if err != nil {
		t.Fatal(err)
	}
	second, err := Evaluate(request, DefaultPolicy(ModeShadow))
	if err != nil {
		t.Fatal(err)
	}
	if first.Fingerprint == "" || first.Fingerprint != second.Fingerprint {
		t.Fatalf("fingerprint no determinista: %q != %q", first.Fingerprint, second.Fingerprint)
	}
	if first.Action != ActionRecommend || first.SelectedActorRef != strings.Repeat("a", 64) {
		t.Fatalf("decision = %#v, se esperaba tie-break por actor_ref", first)
	}
	if first.GateAdvanced || len(first.Ranking) != 2 {
		t.Fatalf("decision debe ser explicable y no avanzar gates: %#v", first)
	}
	got := first.Ranking[0].Breakdown
	if got.CapabilityMatch != 100 || got.DemandPriority != 80 || got.CapacityFit != 100 || got.RiskGap != 0 || got.CoordinationCost != 30 {
		t.Fatalf("breakdown inesperado: %#v", got)
	}
	if got.Total != 930 {
		t.Fatalf("total=%d, want 930", got.Total)
	}
}

func TestEvaluateProtectedBoundaryPreemptsScoring(t *testing.T) {
	t.Parallel()
	request := validEvaluationRequest()
	request.Demand.ProtectedBoundaries = []string{"delivery"}

	decision, err := Evaluate(request, DefaultPolicy(ModeAdvisory))
	if err != nil {
		t.Fatal(err)
	}
	if decision.Action != ActionEscalate || decision.SelectedActorRef != "" || len(decision.Ranking) != 0 {
		t.Fatalf("protected decision = %#v", decision)
	}
	if decision.Reason != "protected_boundary" || decision.NextOwner != "user" || decision.GateAdvanced {
		t.Fatalf("protected metadata = %#v", decision)
	}
}

func TestEvaluateExcludesMissingCapabilityAndInsufficientBudget(t *testing.T) {
	t.Parallel()
	request := validEvaluationRequest()
	request.Profiles = append(request.Profiles,
		CapabilityProfile{
			SchemaVersion:   CapabilityProfileSchemaVersion,
			ActorRef:        strings.Repeat("b", 64),
			Capabilities:    []Capability{{Name: "review", Level: 100}},
			AvailableBudget: 100, RiskTolerance: 100, RoleHints: []string{"reviewer"},
		},
		CapabilityProfile{
			SchemaVersion:   CapabilityProfileSchemaVersion,
			ActorRef:        strings.Repeat("c", 64),
			Capabilities:    []Capability{{Name: "implementation", Level: 100}},
			AvailableBudget: 10, RiskTolerance: 100, RoleHints: []string{"implementer"},
		},
	)

	decision, err := Evaluate(request, DefaultPolicy(ModeShadow))
	if err != nil {
		t.Fatal(err)
	}
	if decision.Action != ActionRecommend || len(decision.Ranking) != 3 {
		t.Fatalf("decision = %#v", decision)
	}
	reasons := map[string]string{}
	for _, candidate := range decision.Ranking {
		if !candidate.Eligible {
			reasons[candidate.ActorRef] = candidate.Reason
		}
	}
	if reasons[strings.Repeat("b", 64)] != "missing_capability" || reasons[strings.Repeat("c", 64)] != "insufficient_budget" {
		t.Fatalf("ineligible reasons = %#v", reasons)
	}
}

func TestEvaluateDisabledNeverRecommends(t *testing.T) {
	t.Parallel()
	decision, err := Evaluate(validEvaluationRequest(), Policy{
		Enabled: false, Mode: ModeShadow, Version: PolicyDeterministicV1, MaxCandidates: 32,
	})
	if err != nil {
		t.Fatal(err)
	}
	if decision.Action != ActionDisabled || len(decision.Ranking) != 0 || decision.GateAdvanced {
		t.Fatalf("disabled decision = %#v", decision)
	}
}

func validEvaluationRequest() EvaluationRequest {
	return EvaluationRequest{
		SchemaVersion: EvaluationSchemaVersion,
		Demand: DemandSignal{
			SchemaVersion: DemandSignalSchemaVersion,
			DemandID:      "demand-phase5", TaskRef: "task-1", SnapshotVersion: 1,
			Priority: 80, RequiredBudget: 40, Risk: 40, CoordinationCost: 20,
			RequiredCapabilities: []RequiredCapability{{Name: "implementation", Level: 80, Weight: 5}},
		},
		Profiles: []CapabilityProfile{{
			SchemaVersion:   CapabilityProfileSchemaVersion,
			ActorRef:        strings.Repeat("a", 64),
			Capabilities:    []Capability{{Name: "implementation", Level: 80}},
			AvailableBudget: 80, RiskTolerance: 60, CoordinationCost: 10,
			RoleHints: []string{"implementer"},
		}},
	}
}
