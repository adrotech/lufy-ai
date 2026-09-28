package domain

import "time"

const (
	RecommendationSchemaVersion = "lufy-adaptive-recommendation/v1"
	AssignmentSchemaVersion     = "lufy-adaptive-assignment/v1"
	YieldSchemaVersion          = "lufy-yield-checkpoint/v1"
)

type Recommendation struct {
	SchemaVersion             string `json:"schema_version" yaml:"schema_version"`
	DemandID                  string `json:"demand_id" yaml:"demand_id"`
	TaskRef                   string `json:"task_ref" yaml:"task_ref"`
	ActorRef                  string `json:"actor_ref" yaml:"actor_ref"`
	RoleHint                  string `json:"role_hint,omitempty" yaml:"role_hint,omitempty"`
	PolicyVersion             string `json:"policy_version" yaml:"policy_version"`
	Fingerprint               string `json:"fingerprint" yaml:"fingerprint"`
	Priority                  int    `json:"priority" yaml:"priority"`
	RequiredBudget            int    `json:"required_budget" yaml:"required_budget"`
	Score                     int    `json:"score" yaml:"score"`
	ExpectedProjectionVersion uint64 `json:"expected_projection_version" yaml:"expected_projection_version"`
}

type AssignmentLease struct {
	TokenDigest string    `json:"token_digest" yaml:"token_digest"`
	ExpiresAt   time.Time `json:"expires_at" yaml:"expires_at"`
}

type Assignment struct {
	SchemaVersion             string          `json:"schema_version" yaml:"schema_version"`
	AssignmentID              string          `json:"assignment_id" yaml:"assignment_id"`
	DemandID                  string          `json:"demand_id" yaml:"demand_id"`
	TaskRef                   string          `json:"task_ref" yaml:"task_ref"`
	ActorRef                  string          `json:"actor_ref" yaml:"actor_ref"`
	RoleHint                  string          `json:"role_hint" yaml:"role_hint"`
	PolicyVersion             string          `json:"policy_version" yaml:"policy_version"`
	RecommendationFingerprint string          `json:"recommendation_fingerprint" yaml:"recommendation_fingerprint"`
	Priority                  int             `json:"priority" yaml:"priority"`
	RequiredBudget            int             `json:"required_budget" yaml:"required_budget"`
	Score                     int             `json:"score" yaml:"score"`
	ExpectedProjectionVersion uint64          `json:"expected_projection_version" yaml:"expected_projection_version"`
	Lease                     AssignmentLease `json:"lease" yaml:"lease"`
}

type ArtifactReference struct {
	Kind          string `json:"kind" yaml:"kind"`
	PathSHA256    string `json:"path_sha256" yaml:"path_sha256"`
	ContentSHA256 string `json:"content_sha256,omitempty" yaml:"content_sha256,omitempty"`
}

type EvidenceReference struct {
	Category   string `json:"category" yaml:"category"`
	Result     string `json:"result" yaml:"result"`
	PathSHA256 string `json:"path_sha256,omitempty" yaml:"path_sha256,omitempty"`
}

type YieldCheckpoint struct {
	SchemaVersion             string              `json:"schema_version" yaml:"schema_version"`
	AssignmentID              string              `json:"assignment_id" yaml:"assignment_id"`
	DemandID                  string              `json:"demand_id" yaml:"demand_id"`
	ActorRef                  string              `json:"actor_ref" yaml:"actor_ref"`
	ExpectedProjectionVersion uint64              `json:"expected_projection_version" yaml:"expected_projection_version"`
	Lease                     AssignmentLease     `json:"lease" yaml:"lease"`
	Reason                    string              `json:"reason" yaml:"reason"`
	NextStatus                string              `json:"next_status" yaml:"next_status"`
	ArtifactRefs              []ArtifactReference `json:"artifact_refs,omitempty" yaml:"artifact_refs,omitempty"`
	EvidenceRefs              []EvidenceReference `json:"evidence_refs,omitempty" yaml:"evidence_refs,omitempty"`
	HypothesisRefs            []string            `json:"hypothesis_refs,omitempty" yaml:"hypothesis_refs,omitempty"`
	FailedAttemptRefs         []string            `json:"failed_attempt_refs,omitempty" yaml:"failed_attempt_refs,omitempty"`
	SuccessorCapabilities     []string            `json:"successor_capabilities,omitempty" yaml:"successor_capabilities,omitempty"`
}

func ValidateDemandSignal(demand DemandSignal) error {
	return validateDemand(demand)
}

func ValidateRecommendation(value Recommendation) error {
	if value.SchemaVersion != RecommendationSchemaVersion {
		return fieldError("recommendation.schema_version", "usar lufy-adaptive-recommendation/v1")
	}
	if !idPattern.MatchString(value.DemandID) || !idPattern.MatchString(value.TaskRef) {
		return fieldError("recommendation.identity", "usar identificadores bounded")
	}
	if !digestPattern.MatchString(value.ActorRef) || !digestPattern.MatchString(value.Fingerprint) {
		return fieldError("recommendation.references", "usar referencias SHA-256")
	}
	if value.RoleHint != "" && !tokenPattern.MatchString(value.RoleHint) {
		return fieldError("recommendation.role_hint", "usar un token bounded")
	}
	if value.PolicyVersion != PolicyDeterministicV1 {
		return fieldError("recommendation.policy_version", "usar deterministic-v1")
	}
	if err := rangeValue("recommendation.priority", value.Priority, 0, 100); err != nil {
		return err
	}
	if err := rangeValue("recommendation.required_budget", value.RequiredBudget, 1, 100); err != nil {
		return err
	}
	if value.Score < -5000 || value.Score > 2000 {
		return fieldError("recommendation.score", "usar un score bounded")
	}
	if value.ExpectedProjectionVersion == 0 {
		return fieldError("recommendation.expected_projection_version", "usar una versión positiva")
	}
	return nil
}

func ValidateAssignment(value Assignment) error {
	if value.SchemaVersion != AssignmentSchemaVersion {
		return fieldError("assignment.schema_version", "usar lufy-adaptive-assignment/v1")
	}
	if !idPattern.MatchString(value.AssignmentID) || !idPattern.MatchString(value.DemandID) || !idPattern.MatchString(value.TaskRef) {
		return fieldError("assignment.identity", "usar identificadores bounded")
	}
	if !digestPattern.MatchString(value.ActorRef) || !digestPattern.MatchString(value.RecommendationFingerprint) {
		return fieldError("assignment.references", "usar referencias SHA-256")
	}
	if !tokenPattern.MatchString(value.RoleHint) || value.PolicyVersion != PolicyDeterministicV1 {
		return fieldError("assignment.policy", "usar role hint y policy documentados")
	}
	if err := rangeValue("assignment.priority", value.Priority, 0, 100); err != nil {
		return err
	}
	if err := rangeValue("assignment.required_budget", value.RequiredBudget, 1, 100); err != nil {
		return err
	}
	if value.Score < -5000 || value.Score > 2000 {
		return fieldError("assignment.score", "usar un score bounded")
	}
	if value.ExpectedProjectionVersion == 0 {
		return fieldError("assignment.expected_projection_version", "usar una versión positiva")
	}
	if !digestPattern.MatchString(value.Lease.TokenDigest) || value.Lease.ExpiresAt.IsZero() {
		return fieldError("assignment.lease", "usar digest SHA-256 y expiración")
	}
	return nil
}

func ValidateYieldCheckpoint(value YieldCheckpoint) error {
	if value.SchemaVersion != YieldSchemaVersion {
		return fieldError("yield.schema_version", "usar lufy-yield-checkpoint/v1")
	}
	if !idPattern.MatchString(value.AssignmentID) || !idPattern.MatchString(value.DemandID) {
		return fieldError("yield.identity", "usar identificadores bounded")
	}
	if !digestPattern.MatchString(value.ActorRef) || !digestPattern.MatchString(value.Lease.TokenDigest) || value.Lease.ExpiresAt.IsZero() {
		return fieldError("yield.lease", "usar owner/lease pseudonimizados")
	}
	if value.ExpectedProjectionVersion == 0 {
		return fieldError("yield.expected_projection_version", "usar una versión positiva")
	}
	if !oneOf(value.Reason, "blocked", "capacity_change", "higher_value_successor", "lease_expiring", "manual") {
		return fieldError("yield.reason", "usar una razón documentada")
	}
	if !oneOf(value.NextStatus, "waiting", "blocked", "escalated", "completed") {
		return fieldError("yield.next_status", "usar un estado documentado")
	}
	if len(value.ArtifactRefs) > maxCapabilities || len(value.EvidenceRefs) > maxCapabilities ||
		len(value.HypothesisRefs) > maxCapabilities || len(value.FailedAttemptRefs) > maxCapabilities ||
		len(value.SuccessorCapabilities) > maxCapabilities {
		return fieldError("yield.references", "reducir referencias a 32 o menos")
	}
	for _, ref := range value.ArtifactRefs {
		if !tokenPattern.MatchString(ref.Kind) || !digestPattern.MatchString(ref.PathSHA256) ||
			ref.ContentSHA256 != "" && !digestPattern.MatchString(ref.ContentSHA256) {
			return fieldError("yield.artifact_refs", "usar kind y digests documentados")
		}
	}
	for _, ref := range value.EvidenceRefs {
		if !tokenPattern.MatchString(ref.Category) || !tokenPattern.MatchString(ref.Result) ||
			ref.PathSHA256 != "" && !digestPattern.MatchString(ref.PathSHA256) {
			return fieldError("yield.evidence_refs", "usar metadata content-free")
		}
	}
	for _, refs := range [][]string{value.HypothesisRefs, value.FailedAttemptRefs} {
		for _, ref := range refs {
			if !digestPattern.MatchString(ref) {
				return fieldError("yield.digest_refs", "usar SHA-256")
			}
		}
	}
	for _, capability := range value.SuccessorCapabilities {
		if !tokenPattern.MatchString(capability) {
			return fieldError("yield.successor_capabilities", "usar tokens bounded")
		}
	}
	return nil
}
