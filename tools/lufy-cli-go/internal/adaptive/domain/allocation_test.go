package domain

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

func TestDecodeAssignmentAndYieldCheckpointAcceptStrictDocuments(t *testing.T) {
	t.Parallel()
	assignmentBody := `schema_version: lufy-adaptive-assignment/v1
assignment_id: assignment-1
demand_id: demand-1
task_ref: task-1
actor_ref: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
role_hint: implementer
policy_version: deterministic-v1
recommendation_fingerprint: bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
priority: 80
required_budget: 40
score: 900
expected_projection_version: 3
lease:
  token_digest: cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc
  expires_at: "2026-09-22T12:15:00Z"
`
	assignment, err := DecodeAssignment(strings.NewReader(assignmentBody))
	if err != nil {
		t.Fatal(err)
	}
	if assignment.AssignmentID != "assignment-1" || assignment.Lease.ExpiresAt.IsZero() {
		t.Fatalf("assignment = %#v", assignment)
	}

	yieldBody := `schema_version: lufy-yield-checkpoint/v1
assignment_id: assignment-1
demand_id: demand-1
actor_ref: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
expected_projection_version: 4
lease:
  token_digest: cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc
  expires_at: "2026-09-22T12:15:00Z"
reason: blocked
next_status: waiting
artifact_refs:
  - kind: spec
    path_sha256: dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd
    content_sha256: eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee
evidence_refs:
  - category: test
    result: passed
    path_sha256: ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
hypothesis_refs: ["1111111111111111111111111111111111111111111111111111111111111111"]
failed_attempt_refs: ["2222222222222222222222222222222222222222222222222222222222222222"]
successor_capabilities: [validation]
`
	checkpoint, err := DecodeYieldCheckpoint(strings.NewReader(yieldBody))
	if err != nil {
		t.Fatal(err)
	}
	if checkpoint.NextStatus != "waiting" || len(checkpoint.ArtifactRefs) != 1 || len(checkpoint.SuccessorCapabilities) != 1 {
		t.Fatalf("checkpoint = %#v", checkpoint)
	}
}

func TestDecodeAllocationDocumentsRejectUnsafeInput(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		body []byte
	}{
		{name: "nil", body: nil},
		{name: "unknown field", body: []byte("schema_version: lufy-adaptive-assignment/v1\nprompt: secret-canary\n")},
		{name: "invalid utf8", body: []byte{0xff, 0xfe}},
		{name: "oversized", body: bytes.Repeat([]byte("x"), maxInputBytes+1)},
		{name: "trailing document", body: []byte("schema_version: lufy-adaptive-assignment/v1\n---\nschema_version: second\n")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var reader io.Reader
			if test.body != nil {
				reader = bytes.NewReader(test.body)
			}
			if _, err := DecodeAssignment(reader); err == nil {
				t.Fatal("se esperaba rechazo")
			} else if strings.Contains(err.Error(), "secret-canary") {
				t.Fatalf("diagnóstico filtró contenido: %v", err)
			}
		})
	}
}

func TestValidateAllocationContractsRejectInvalidFields(t *testing.T) {
	t.Parallel()
	assignment := validAllocationAssignment()
	assignmentCases := []func(*Assignment){
		func(v *Assignment) { v.SchemaVersion = "v0" },
		func(v *Assignment) { v.AssignmentID = "contains space" },
		func(v *Assignment) { v.ActorRef = "raw-actor" },
		func(v *Assignment) { v.RoleHint = "invalid role" },
		func(v *Assignment) { v.Priority = 101 },
		func(v *Assignment) { v.RequiredBudget = 0 },
		func(v *Assignment) { v.Score = 2001 },
		func(v *Assignment) { v.ExpectedProjectionVersion = 0 },
		func(v *Assignment) { v.Lease.TokenDigest = "raw-token" },
	}
	for index, mutate := range assignmentCases {
		value := assignment
		mutate(&value)
		if err := ValidateAssignment(value); err == nil {
			t.Fatalf("assignment case %d: se esperaba error", index)
		}
	}

	recommendation := Recommendation{
		SchemaVersion: RecommendationSchemaVersion, DemandID: "demand-1", TaskRef: "task-1",
		ActorRef: strings.Repeat("a", 64), RoleHint: "implementer", PolicyVersion: PolicyDeterministicV1,
		Fingerprint: strings.Repeat("b", 64), Priority: 80, RequiredBudget: 40, Score: 900, ExpectedProjectionVersion: 3,
	}
	recommendationCases := []func(*Recommendation){
		func(v *Recommendation) { v.SchemaVersion = "v0" },
		func(v *Recommendation) { v.DemandID = "contains space" },
		func(v *Recommendation) { v.ActorRef = "raw-actor" },
		func(v *Recommendation) { v.RoleHint = "invalid role" },
		func(v *Recommendation) { v.PolicyVersion = "other" },
		func(v *Recommendation) { v.Priority = -1 },
		func(v *Recommendation) { v.RequiredBudget = 101 },
		func(v *Recommendation) { v.Score = -5001 },
		func(v *Recommendation) { v.ExpectedProjectionVersion = 0 },
	}
	for index, mutate := range recommendationCases {
		value := recommendation
		mutate(&value)
		if err := ValidateRecommendation(value); err == nil {
			t.Fatalf("recommendation case %d: se esperaba error", index)
		}
	}
}

func TestValidateYieldCheckpointRejectsUnsafeReferences(t *testing.T) {
	t.Parallel()
	base := validAllocationYield()
	tests := []func(*YieldCheckpoint){
		func(v *YieldCheckpoint) { v.SchemaVersion = "v0" },
		func(v *YieldCheckpoint) { v.AssignmentID = "contains space" },
		func(v *YieldCheckpoint) { v.ActorRef = "raw-actor" },
		func(v *YieldCheckpoint) { v.ExpectedProjectionVersion = 0 },
		func(v *YieldCheckpoint) { v.Reason = "unknown" },
		func(v *YieldCheckpoint) { v.NextStatus = "unknown" },
		func(v *YieldCheckpoint) { v.SuccessorCapabilities = make([]string, maxCapabilities+1) },
		func(v *YieldCheckpoint) { v.ArtifactRefs[0].PathSHA256 = "raw-path" },
		func(v *YieldCheckpoint) { v.EvidenceRefs[0].Category = "raw category" },
		func(v *YieldCheckpoint) { v.HypothesisRefs[0] = "raw hypothesis" },
		func(v *YieldCheckpoint) { v.SuccessorCapabilities[0] = "raw capability" },
	}
	for index, mutate := range tests {
		value := base
		value.ArtifactRefs = append([]ArtifactReference(nil), base.ArtifactRefs...)
		value.EvidenceRefs = append([]EvidenceReference(nil), base.EvidenceRefs...)
		value.HypothesisRefs = append([]string(nil), base.HypothesisRefs...)
		value.SuccessorCapabilities = append([]string(nil), base.SuccessorCapabilities...)
		mutate(&value)
		if err := ValidateYieldCheckpoint(value); err == nil {
			t.Fatalf("case %d: se esperaba error", index)
		}
	}
}

func validAllocationAssignment() Assignment {
	return Assignment{
		SchemaVersion: AssignmentSchemaVersion, AssignmentID: "assignment-1", DemandID: "demand-1", TaskRef: "task-1",
		ActorRef: strings.Repeat("a", 64), RoleHint: "implementer", PolicyVersion: PolicyDeterministicV1,
		RecommendationFingerprint: strings.Repeat("b", 64), Priority: 80, RequiredBudget: 40, Score: 900,
		ExpectedProjectionVersion: 3,
		Lease:                     AssignmentLease{TokenDigest: strings.Repeat("c", 64), ExpiresAt: time.Date(2026, 9, 22, 12, 15, 0, 0, time.UTC)},
	}
}

func validAllocationYield() YieldCheckpoint {
	assignment := validAllocationAssignment()
	return YieldCheckpoint{
		SchemaVersion: YieldSchemaVersion, AssignmentID: assignment.AssignmentID, DemandID: assignment.DemandID,
		ActorRef: assignment.ActorRef, ExpectedProjectionVersion: 4, Lease: assignment.Lease,
		Reason: "manual", NextStatus: "waiting",
		ArtifactRefs:          []ArtifactReference{{Kind: "spec", PathSHA256: strings.Repeat("d", 64)}},
		EvidenceRefs:          []EvidenceReference{{Category: "test", Result: "passed"}},
		HypothesisRefs:        []string{strings.Repeat("e", 64)},
		SuccessorCapabilities: []string{"validation"},
	}
}
