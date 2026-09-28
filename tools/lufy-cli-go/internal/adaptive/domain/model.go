package domain

const (
	EvaluationSchemaVersion        = "lufy-adaptive-evaluation/v1"
	DemandSignalSchemaVersion      = "lufy-demand-signal/v1"
	CapabilityProfileSchemaVersion = "lufy-capability-profile/v1"
	DecisionSchemaVersion          = "lufy-adaptive-decision/v1"
	PolicyDeterministicV1          = "deterministic-v1"
)

type Mode string

const (
	ModeShadow   Mode = "shadow"
	ModeAdvisory Mode = "advisory"
)

type Action string

const (
	ActionDisabled    Action = "disabled"
	ActionRecommend   Action = "recommend"
	ActionEscalate    Action = "escalate"
	ActionNoCandidate Action = "no_candidate"
)

type RequiredCapability struct {
	Name   string `json:"name" yaml:"name"`
	Level  int    `json:"level" yaml:"level"`
	Weight int    `json:"weight" yaml:"weight"`
}

type Capability struct {
	Name  string `json:"name" yaml:"name"`
	Level int    `json:"level" yaml:"level"`
}

type DemandSignal struct {
	SchemaVersion        string               `json:"schema_version" yaml:"schema_version"`
	DemandID             string               `json:"demand_id" yaml:"demand_id"`
	TaskRef              string               `json:"task_ref" yaml:"task_ref"`
	SnapshotVersion      uint64               `json:"snapshot_version" yaml:"snapshot_version"`
	Priority             int                  `json:"priority" yaml:"priority"`
	RequiredBudget       int                  `json:"required_budget" yaml:"required_budget"`
	Risk                 int                  `json:"risk" yaml:"risk"`
	CoordinationCost     int                  `json:"coordination_cost" yaml:"coordination_cost"`
	RequiredCapabilities []RequiredCapability `json:"required_capabilities" yaml:"required_capabilities"`
	ProtectedBoundaries  []string             `json:"protected_boundaries,omitempty" yaml:"protected_boundaries,omitempty"`
	CausedByEventID      string               `json:"caused_by_event_id,omitempty" yaml:"caused_by_event_id,omitempty"`
}

type CapabilityProfile struct {
	SchemaVersion     string       `json:"schema_version" yaml:"schema_version"`
	ActorRef          string       `json:"actor_ref" yaml:"actor_ref"`
	Capabilities      []Capability `json:"capabilities" yaml:"capabilities"`
	AvailableBudget   int          `json:"available_budget" yaml:"available_budget"`
	RiskTolerance     int          `json:"risk_tolerance" yaml:"risk_tolerance"`
	CoordinationCost  int          `json:"coordination_cost" yaml:"coordination_cost"`
	ActiveAssignments []string     `json:"active_assignments,omitempty" yaml:"active_assignments,omitempty"`
	RoleHints         []string     `json:"role_hints,omitempty" yaml:"role_hints,omitempty"`
}

type EvaluationRequest struct {
	SchemaVersion string              `json:"schema_version" yaml:"schema_version"`
	Demand        DemandSignal        `json:"demand" yaml:"demand"`
	Profiles      []CapabilityProfile `json:"profiles" yaml:"profiles"`
}

type Policy struct {
	Enabled       bool   `json:"enabled"`
	Mode          Mode   `json:"mode"`
	Version       string `json:"version"`
	MaxCandidates int    `json:"max_candidates"`
}

type ScoreBreakdown struct {
	CapabilityMatch  int `json:"capability_match"`
	DemandPriority   int `json:"demand_priority"`
	CapacityFit      int `json:"capacity_fit"`
	RiskGap          int `json:"risk_gap"`
	CoordinationCost int `json:"coordination_cost"`
	Total            int `json:"total"`
}

type CandidateDecision struct {
	ActorRef  string         `json:"actor_ref"`
	RoleHint  string         `json:"role_hint,omitempty"`
	Eligible  bool           `json:"eligible"`
	Reason    string         `json:"reason,omitempty"`
	Breakdown ScoreBreakdown `json:"breakdown"`
}

type Decision struct {
	SchemaVersion    string              `json:"schema_version"`
	Action           Action              `json:"action"`
	Mode             Mode                `json:"mode"`
	PolicyVersion    string              `json:"policy_version"`
	Reason           string              `json:"reason,omitempty"`
	NextOwner        string              `json:"next_owner,omitempty"`
	SelectedActorRef string              `json:"selected_actor_ref,omitempty"`
	SelectedRoleHint string              `json:"selected_role_hint,omitempty"`
	Ranking          []CandidateDecision `json:"ranking"`
	GateAdvanced     bool                `json:"gate_advanced"`
	Fingerprint      string              `json:"fingerprint"`
}

func DefaultPolicy(mode Mode) Policy {
	return Policy{Enabled: true, Mode: mode, Version: PolicyDeterministicV1, MaxCandidates: 32}
}
