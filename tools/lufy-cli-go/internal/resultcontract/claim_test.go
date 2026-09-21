package resultcontract

import (
	"slices"
	"strings"
	"testing"
)

func TestEvaluateClaimHonorsCurrentRoleAllowedStatusMatrix(t *testing.T) {
	t.Parallel()

	roles := currentRolePolicy()
	allStatuses := []Status{
		StatusReady, StatusImplemented, StatusValidated, StatusDeliveryPending,
		StatusSyncPending, StatusBlocked, StatusEscalated, StatusDelivered, StatusClosed,
	}
	evidence := evidencePolicyAllowingAdvancedClaims()

	for role, allowed := range roles {
		role, allowed := role, allowed
		t.Run(role, func(t *testing.T) {
			t.Parallel()
			for _, status := range allStatuses {
				contract := decodeFixture(t, "testdata/valid-minimal.yaml")
				contract.Status = status
				decision := EvaluateClaim(contract, role, roles, evidence)
				wantAccepted := slices.Contains(allowed, status)
				if decision.Accepted != wantAccepted {
					t.Fatalf("EvaluateClaim(role=%q,status=%q).Accepted = %t, want %t; decision=%#v",
						role, status, decision.Accepted, wantAccepted, decision)
				}
				if !wantAccepted && decision.Reason != "role_status_not_allowed" {
					t.Fatalf("rejected reason = %q, want role_status_not_allowed", decision.Reason)
				}
			}
		})
	}
}

func TestEvaluateClaimRejectsUnknownRole(t *testing.T) {
	t.Parallel()

	contract := decodeFixture(t, "testdata/valid-minimal.yaml")
	decision := EvaluateClaim(contract, "unknown-role", currentRolePolicy(), evidencePolicyAllowingAdvancedClaims())
	if decision.Accepted {
		t.Fatalf("unknown role was accepted: %#v", decision)
	}
	if decision.Reason != "unknown_role" || decision.Recovery == "" {
		t.Fatalf("decision = %#v, want unknown_role with recovery", decision)
	}
}

func TestEvaluateClaimUsesInjectableEvidenceRequirements(t *testing.T) {
	t.Parallel()

	policy := EvidencePolicy{ByStatus: map[Status]EvidenceRequirement{
		StatusValidated: {
			Categories: []EvidenceCategory{EvidenceCategoryPassedCommand},
			NextOwner:  "validator",
		},
		StatusDelivered: {
			Categories: []EvidenceCategory{EvidenceCategoryPassedCommand, EvidenceCategoryArtifactReference},
			NextOwner:  "delivery",
		},
		StatusClosed: {
			Categories: []EvidenceCategory{EvidenceCategoryPassedCommand, EvidenceCategoryArtifactReference, EvidenceCategoryStatic},
			NextOwner:  "orchestrator",
		},
	}}
	cases := []struct {
		name          string
		status        Status
		role          string
		mutate        func(*Contract)
		wantMissing   EvidenceCategory
		wantNextOwner string
	}{
		{
			name: "validated needs passed command", status: StatusValidated, role: "validator",
			mutate:      func(contract *Contract) { contract.Evidence.Commands[0].Result = EvidenceNotRun },
			wantMissing: EvidenceCategoryPassedCommand, wantNextOwner: "validator",
		},
		{
			name: "delivered needs artifact reference", status: StatusDelivered, role: "delivery",
			mutate: func(contract *Contract) {
				contract.Evidence.Commands[0].Result = EvidencePassed
				contract.Artifacts.Referenced = []string{"none"}
			},
			wantMissing: EvidenceCategoryArtifactReference, wantNextOwner: "delivery",
		},
		{
			name: "closed needs static evidence", status: StatusClosed, role: "orchestrator",
			mutate: func(contract *Contract) {
				contract.Evidence.Commands[0].Result = EvidencePassed
				contract.Evidence.Static = []string{"not_applicable"}
			},
			wantMissing: EvidenceCategoryStatic, wantNextOwner: "orchestrator",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			contract := decodeFixture(t, "testdata/valid-minimal.yaml")
			contract.Status = tc.status
			tc.mutate(&contract)
			decision := EvaluateClaim(contract, tc.role, currentRolePolicy(), policy)
			if decision.Accepted {
				t.Fatalf("claim without required evidence accepted: %#v", decision)
			}
			if decision.Reason != "missing_evidence" || !slices.Contains(decision.MissingEvidence, tc.wantMissing) {
				t.Fatalf("decision = %#v, want missing %q", decision, tc.wantMissing)
			}
			if decision.NextOwner != tc.wantNextOwner || decision.Recovery == "" {
				t.Fatalf("decision = %#v, want next owner %q and recovery", decision, tc.wantNextOwner)
			}
		})
	}
}

func TestEvaluateClaimAcceptsAdvancedClaimWhenConfiguredCategoriesExist(t *testing.T) {
	t.Parallel()

	policy := EvidencePolicy{ByStatus: map[Status]EvidenceRequirement{
		StatusValidated: {Categories: []EvidenceCategory{EvidenceCategoryPassedCommand}, NextOwner: "validator"},
		StatusDelivered: {Categories: []EvidenceCategory{EvidenceCategoryPassedCommand, EvidenceCategoryArtifactReference}, NextOwner: "delivery"},
		StatusClosed:    {Categories: []EvidenceCategory{EvidenceCategoryPassedCommand, EvidenceCategoryArtifactReference, EvidenceCategoryStatic}, NextOwner: "orchestrator"},
	}}
	for _, tc := range []struct {
		status Status
		role   string
	}{
		{StatusValidated, "validator"},
		{StatusDelivered, "delivery"},
		{StatusClosed, "orchestrator"},
	} {
		contract := decodeFixture(t, "testdata/valid-minimal.yaml")
		contract.Status = tc.status
		contract.Evidence.Commands[0].Result = EvidencePassed
		decision := EvaluateClaim(contract, tc.role, currentRolePolicy(), policy)
		if !decision.Accepted || len(decision.MissingEvidence) != 0 {
			t.Fatalf("EvaluateClaim(status=%q) = %#v, want accepted", tc.status, decision)
		}
	}
}

func TestAdvancedClaimFailsClosedWithoutConfiguredEvidencePolicy(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		status Status
		role   string
	}{
		{StatusValidated, "validator"},
		{StatusDelivered, "delivery"},
		{StatusClosed, "orchestrator"},
	} {
		contract := decodeFixture(t, "testdata/valid-minimal.yaml")
		contract.Status = tc.status
		contract.Evidence.Commands[0].Result = EvidencePassed
		decision := EvaluateClaim(contract, tc.role, currentRolePolicy(), EvidencePolicy{})
		if decision.Accepted || decision.Reason != "evidence_policy_unavailable" {
			t.Fatalf("EvaluateClaim(status=%q) = %#v, want fail-closed policy rejection", tc.status, decision)
		}
	}
}

func TestReflectionFinalTextAndSchemaSubstringDoNotAdvanceClaim(t *testing.T) {
	t.Parallel()

	prose := `Terminé todo. schema_version: result-contract/v1. Los tests pasaron y el gate está validado.`
	if _, err := Decode(strings.NewReader(prose)); err == nil {
		t.Fatal("schema substring in prose decoded as a contract")
	}

	contract := decodeFixture(t, "testdata/valid-minimal.yaml")
	contract.Status = StatusValidated
	contract.ExecutiveSummary = "Reflexioné y confirmo que todos los tests pasaron."
	contract.Evidence.Static = []string{
		"schema_version: result-contract/v1",
		"Texto final: validación exitosa.",
	}
	contract.Evidence.Commands[0].Result = EvidenceNotRun
	policy := EvidencePolicy{ByStatus: map[Status]EvidenceRequirement{
		StatusValidated: {Categories: []EvidenceCategory{EvidenceCategoryPassedCommand}, NextOwner: "validator"},
	}}
	decision := EvaluateClaim(contract, "validator", currentRolePolicy(), policy)
	if decision.Accepted || decision.Reason != "missing_evidence" {
		t.Fatalf("textual assertion advanced validated gate: %#v", decision)
	}
	if !slices.Contains(decision.MissingEvidence, EvidenceCategoryPassedCommand) {
		t.Fatalf("missing evidence = %#v, want passed command", decision.MissingEvidence)
	}
}

func currentRolePolicy() RolePolicy {
	return RolePolicy{
		"orchestrator": {StatusReady, StatusBlocked, StatusEscalated, StatusDeliveryPending, StatusSyncPending, StatusClosed},
		"router":       {StatusReady, StatusBlocked, StatusEscalated, StatusDeliveryPending},
		"explorer":     {StatusReady, StatusBlocked, StatusEscalated},
		"implementer":  {StatusImplemented, StatusValidated, StatusBlocked, StatusEscalated, StatusDeliveryPending},
		"test-writer":  {StatusImplemented, StatusValidated, StatusBlocked, StatusEscalated},
		"validator":    {StatusValidated, StatusBlocked, StatusEscalated, StatusDeliveryPending},
		"reviewer":     {StatusReady, StatusBlocked, StatusEscalated},
		"delivery":     {StatusDeliveryPending, StatusSyncPending, StatusBlocked, StatusDelivered, StatusClosed},
	}
}

func evidencePolicyAllowingAdvancedClaims() EvidencePolicy {
	return EvidencePolicy{ByStatus: map[Status]EvidenceRequirement{
		StatusValidated: {},
		StatusDelivered: {},
		StatusClosed:    {},
	}}
}
