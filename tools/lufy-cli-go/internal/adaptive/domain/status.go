package domain

import "time"

const AdaptiveStatusSchemaVersion = "lufy-adaptive-status/v1"

type ActiveAssignment struct {
	AssignmentID              string    `json:"assignment_id"`
	DemandID                  string    `json:"demand_id"`
	TaskRef                   string    `json:"task_ref"`
	ActorRef                  string    `json:"actor_ref"`
	RoleHint                  string    `json:"role_hint"`
	PolicyVersion             string    `json:"policy_version"`
	RecommendationFingerprint string    `json:"recommendation_fingerprint"`
	Priority                  int       `json:"priority"`
	RequiredBudget            int       `json:"required_budget"`
	Score                     int       `json:"score"`
	WaitingCycle              int       `json:"waiting_cycle"`
	LeaseTokenDigest          string    `json:"lease_token_digest"`
	LeaseExpiresAt            time.Time `json:"lease_expires_at"`
}

type WaitingItem struct {
	DemandID       string `json:"demand_id"`
	TaskRef        string `json:"task_ref"`
	Priority       int    `json:"priority"`
	RequiredBudget int    `json:"required_budget"`
	WaitingCycle   int    `json:"waiting_cycle"`
	State          string `json:"state"`
	StarvationRisk bool   `json:"starvation_risk"`
}

type ActorBudget struct {
	ActorRef string `json:"actor_ref"`
	Consumed int    `json:"consumed"`
}

type AdaptiveStatus struct {
	SchemaVersion     string             `json:"schema_version"`
	RunID             string             `json:"run_id"`
	Version           uint64             `json:"version"`
	ActiveAssignments []ActiveAssignment `json:"active_assignments"`
	Waiting           []WaitingItem      `json:"waiting"`
	ConsumedBudget    []ActorBudget      `json:"consumed_budget"`
	Truncated         bool               `json:"truncated"`
	SourceDigest      string             `json:"source_digest"`
}

type LedgerOperation struct {
	Status        string         `json:"status"`
	EventID       string         `json:"event_id"`
	EventSequence uint64         `json:"event_sequence"`
	State         AdaptiveStatus `json:"state"`
}
