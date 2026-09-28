package resultcontract

const SchemaVersion = "result-contract/v1"

type Status string

const (
	StatusReady           Status = "ready"
	StatusImplemented     Status = "implemented"
	StatusValidated       Status = "validated"
	StatusDeliveryPending Status = "delivery_pending"
	StatusSyncPending     Status = "sync_pending"
	StatusBlocked         Status = "blocked"
	StatusEscalated       Status = "escalated"
	StatusDelivered       Status = "delivered"
	StatusClosed          Status = "closed"
)

type EvidenceResult string

const (
	EvidencePassed  EvidenceResult = "passed"
	EvidenceFailed  EvidenceResult = "failed"
	EvidenceBlocked EvidenceResult = "blocked"
	EvidenceNotRun  EvidenceResult = "not_run"
)

type Contract struct {
	SchemaVersion        string                `json:"schema_version" yaml:"schema_version"`
	Status               Status                `json:"status" yaml:"status"`
	LegacyFallback       bool                  `json:"legacy_fallback" yaml:"legacy_fallback"`
	ExecutiveSummary     string                `json:"executive_summary" yaml:"executive_summary"`
	Artifacts            Artifacts             `json:"artifacts" yaml:"artifacts"`
	Ledger               *Ledger               `json:"ledger,omitempty" yaml:"ledger,omitempty"`
	Overview             *Overview             `json:"overview,omitempty" yaml:"overview,omitempty"`
	Evidence             Evidence              `json:"evidence" yaml:"evidence"`
	Diagnostics          *Diagnostics          `json:"diagnostics,omitempty" yaml:"diagnostics,omitempty"`
	SurfaceExecution     SurfaceExecution      `json:"surface_execution" yaml:"surface_execution"`
	StructuralAcceptance *StructuralAcceptance `json:"structural_acceptance,omitempty" yaml:"structural_acceptance,omitempty"`
	WorkflowDecision     WorkflowDecision      `json:"workflow_decision" yaml:"workflow_decision"`
	Risks                []string              `json:"risks" yaml:"risks"`
	NextRecommended      NextRecommended       `json:"next_recommended" yaml:"next_recommended"`
	SkillResolution      SkillResolution       `json:"skill_resolution" yaml:"skill_resolution"`
}

type Artifacts struct {
	Changed    []string `json:"changed" yaml:"changed"`
	Referenced []string `json:"referenced" yaml:"referenced"`
}

type Ledger struct {
	RunID   string `json:"run_id" yaml:"run_id"`
	EventID string `json:"event_id" yaml:"event_id"`
	Status  string `json:"status" yaml:"status"`
}

type Overview struct {
	Policy  string `json:"policy" yaml:"policy"`
	Status  string `json:"status" yaml:"status"`
	Trigger string `json:"trigger" yaml:"trigger"`
	Path    string `json:"path" yaml:"path"`
}

type Evidence struct {
	Commands []CommandEvidence `json:"commands" yaml:"commands"`
	Static   []string          `json:"static" yaml:"static"`
}

type CommandEvidence struct {
	Command string         `json:"command" yaml:"command"`
	Result  EvidenceResult `json:"result" yaml:"result"`
	Notes   string         `json:"notes" yaml:"notes"`
}

type Diagnostics struct {
	MemoryProviderUsed          string    `json:"memory_provider_used" yaml:"memory_provider_used"`
	ContextGraphStatus          string    `json:"context_graph_status" yaml:"context_graph_status"`
	ContextGraphQueries         []string  `json:"context_graph_queries" yaml:"context_graph_queries"`
	FallbackReason              string    `json:"fallback_reason" yaml:"fallback_reason"`
	GenericDiscoveryBeforeGraph BoolValue `json:"generic_discovery_before_graph" yaml:"generic_discovery_before_graph"`
}

type SurfaceExecution struct {
	SchemaVersion     string   `json:"schema_version" yaml:"schema_version"`
	Source            string   `json:"source" yaml:"source"`
	PrimarySurface    string   `json:"primary_surface" yaml:"primary_surface"`
	Mode              string   `json:"mode" yaml:"mode"`
	ActiveSurfaces    []string `json:"active_surfaces" yaml:"active_surfaces"`
	ValidationRuleIDs []string `json:"validation_rule_ids" yaml:"validation_rule_ids"`
}

type StructuralAcceptance struct {
	Source                string            `json:"source" yaml:"source"`
	ExpectedDirectories   []string          `json:"expected_directories" yaml:"expected_directories"`
	ExpectedArchitecture  []string          `json:"expected_architecture" yaml:"expected_architecture"`
	ForbiddenRootPatterns []string          `json:"forbidden_root_patterns" yaml:"forbidden_root_patterns"`
	Normalization         string            `json:"normalization" yaml:"normalization"`
	Audit                 []StructuralAudit `json:"audit" yaml:"audit"`
}

type StructuralAudit struct {
	FeatureOrSurface string `json:"feature_or_surface" yaml:"feature_or_surface"`
	Status           string `json:"status" yaml:"status"`
	Notes            string `json:"notes" yaml:"notes"`
}

type WorkflowDecision struct {
	Tier                     string              `json:"tier" yaml:"tier"`
	ProgramTier              string              `json:"program_tier" yaml:"program_tier"`
	SliceTier                string              `json:"slice_tier" yaml:"slice_tier"`
	FastPathAllowed          BoolValue           `json:"fast_path_allowed" yaml:"fast_path_allowed"`
	AdapterContext           AdapterContext      `json:"adapter_context" yaml:"adapter_context"`
	WorkflowLimitsSource     string              `json:"workflow_limits_source" yaml:"workflow_limits_source"`
	WorkflowLimitsPaths      WorkflowLimitsPaths `json:"workflow_limits_paths" yaml:"workflow_limits_paths"`
	WorkloadDecisionNeeded   BoolValue           `json:"workload_decision_needed" yaml:"workload_decision_needed"`
	ReviewSlices             []string            `json:"review_slices" yaml:"review_slices"`
	PreflightStatus          string              `json:"preflight_status" yaml:"preflight_status"`
	StopRuleStatus           string              `json:"stop_rule_status" yaml:"stop_rule_status"`
	DeliveryBatchingGuidance string              `json:"delivery_batching_guidance" yaml:"delivery_batching_guidance"`
	ArtifactBranching        ArtifactBranching   `json:"artifact_branching" yaml:"artifact_branching"`
}

type AdapterContext struct {
	ToolID              string    `json:"tool_id" yaml:"tool_id"`
	MethodologyID       string    `json:"methodology_id" yaml:"methodology_id"`
	MethodologyMode     string    `json:"methodology_mode" yaml:"methodology_mode"`
	MethodologyRequired BoolValue `json:"methodology_required" yaml:"methodology_required"`
	ExecutionMode       string    `json:"execution_mode" yaml:"execution_mode"`
}

type WorkflowLimitsPaths struct {
	Sizing           string `json:"sizing" yaml:"sizing"`
	Routing          string `json:"routing" yaml:"routing"`
	ProposalSlicing  string `json:"proposal_slicing" yaml:"proposal_slicing"`
	DeliveryBatching string `json:"delivery_batching" yaml:"delivery_batching"`
	Preflight        string `json:"preflight" yaml:"preflight"`
	StopRules        string `json:"stop_rules" yaml:"stop_rules"`
}

type ArtifactBranching struct {
	Status                  string     `json:"status" yaml:"status"`
	Stage                   string     `json:"stage" yaml:"stage"`
	CandidateCount          CountValue `json:"candidate_count" yaml:"candidate_count"`
	Reason                  string     `json:"reason" yaml:"reason"`
	ParallelAllowed         BoolValue  `json:"parallel_allowed" yaml:"parallel_allowed"`
	RequiresJoin            BoolValue  `json:"requires_join" yaml:"requires_join"`
	CandidateIsolation      string     `json:"candidate_isolation" yaml:"candidate_isolation"`
	MergePlanRequired       BoolValue  `json:"merge_plan_required" yaml:"merge_plan_required"`
	HumanEscalationTriggers []string   `json:"human_escalation_triggers" yaml:"human_escalation_triggers"`
	CandidatePaths          []string   `json:"candidate_paths" yaml:"candidate_paths"`
	CanonicalArtifactSet    []string   `json:"canonical_artifact_set" yaml:"canonical_artifact_set"`
}

type NextRecommended struct {
	Owner  string `json:"owner" yaml:"owner"`
	Action string `json:"action" yaml:"action"`
}

type SkillResolution struct {
	LocalSkillsUsed      []string `json:"local_skills_used" yaml:"local_skills_used"`
	SelectedSkillPaths   []string `json:"selected_skill_paths" yaml:"selected_skill_paths"`
	BootstrapRecommended bool     `json:"bootstrap_recommended" yaml:"bootstrap_recommended"`
	Notes                string   `json:"notes" yaml:"notes"`
}

// BoolValue preserves fields whose documented wire representation is either a
// JSON boolean or an explicit availability marker.
type BoolValue string

const (
	BoolFalse         BoolValue = "false"
	BoolTrue          BoolValue = "true"
	BoolNotApplicable BoolValue = "not_applicable"
	BoolNotAvailable  BoolValue = "not_available"
)

type CountValue string

const (
	CountOne           CountValue = "1"
	CountTwo           CountValue = "2"
	CountNotApplicable CountValue = "not_applicable"
)
