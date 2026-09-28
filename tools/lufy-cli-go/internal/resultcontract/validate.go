package resultcontract

import (
	"reflect"
	"strings"
)

const (
	maxStringBytes = 8 * 1024
	maxListItems   = 64
)

func Validate(contract Contract) error {
	if err := validateBounds(reflect.ValueOf(contract), ""); err != nil {
		return err
	}
	if contract.SchemaVersion != SchemaVersion {
		return invalidEnum("schema_version", "usar result-contract/v1")
	}
	if !oneOf(string(contract.Status), "ready", "implemented", "validated", "delivery_pending", "sync_pending", "blocked", "escalated", "delivered", "closed") {
		return invalidEnum("status", "usar un status documentado")
	}
	if strings.TrimSpace(contract.ExecutiveSummary) == "" {
		return required("executive_summary")
	}
	if err := requireList(contract.Artifacts.Changed, "artifacts.changed"); err != nil {
		return err
	}
	if err := requireList(contract.Artifacts.Referenced, "artifacts.referenced"); err != nil {
		return err
	}
	if contract.Ledger != nil {
		if contract.Ledger.RunID == "" {
			return required("ledger.run_id")
		}
		if contract.Ledger.EventID == "" {
			return required("ledger.event_id")
		}
		if !oneOf(contract.Ledger.Status, "recorded", "duplicate_noop", "conflict", "unavailable", "disabled", "not_applicable") {
			return invalidEnum("ledger.status", "usar un estado de ledger documentado")
		}
	}
	if contract.Overview != nil {
		if !oneOf(contract.Overview.Policy, "automatic", "optional", "none", "not_available", "not_applicable") {
			return invalidEnum("overview.policy", "usar una policy documentada")
		}
		if !oneOf(contract.Overview.Status, "generated", "refreshed", "preserved", "offered_pending", "skipped_by_user", "not_available", "not_applicable") {
			return invalidEnum("overview.status", "usar un status documentado")
		}
		if !oneOf(contract.Overview.Trigger, "new", "validate", "sync", "archive", "status_read_only", "user_choice", "not_applicable") {
			return invalidEnum("overview.trigger", "usar un trigger documentado")
		}
		if contract.Overview.Path == "" {
			return required("overview.path")
		}
	}
	if err := validateEvidence(contract.Evidence); err != nil {
		return err
	}
	if err := validateDiagnostics(contract.Diagnostics); err != nil {
		return err
	}
	if err := validateSurface(contract.SurfaceExecution); err != nil {
		return err
	}
	if err := validateStructural(contract.StructuralAcceptance); err != nil {
		return err
	}
	if err := validateWorkflow(contract.WorkflowDecision); err != nil {
		return err
	}
	if err := requireList(contract.Risks, "risks"); err != nil {
		return err
	}
	if !oneOf(contract.NextRecommended.Owner, "orchestrator", "explorer", "implementer", "validator", "reviewer", "delivery", "user", "none") {
		return invalidEnum("next_recommended.owner", "usar un owner documentado")
	}
	if contract.NextRecommended.Action == "" {
		return required("next_recommended.action")
	}
	if err := requireList(contract.SkillResolution.LocalSkillsUsed, "skill_resolution.local_skills_used"); err != nil {
		return err
	}
	if err := requireList(contract.SkillResolution.SelectedSkillPaths, "skill_resolution.selected_skill_paths"); err != nil {
		return err
	}
	if contract.SkillResolution.Notes == "" {
		return required("skill_resolution.notes")
	}
	return nil
}

func validateEvidence(evidence Evidence) error {
	if len(evidence.Commands) == 0 {
		return required("evidence.commands")
	}
	for _, command := range evidence.Commands {
		if command.Command == "" {
			return required("evidence.commands.command")
		}
		if !oneOf(string(command.Result), "passed", "failed", "blocked", "not_run") {
			return invalidEnum("evidence.commands.result", "usar passed, failed, blocked o not_run")
		}
		if command.Notes == "" {
			return required("evidence.commands.notes")
		}
	}
	return requireList(evidence.Static, "evidence.static")
}

func validateDiagnostics(value *Diagnostics) error {
	if value == nil {
		return nil
	}
	if value.MemoryProviderUsed == "" {
		return required("diagnostics.memory_provider_used")
	}
	if value.MemoryProviderUsed != "obsidian" && value.MemoryProviderUsed != "obsidian:not_available" && value.MemoryProviderUsed != "not_applicable" && value.MemoryProviderUsed != "not_available" && !strings.HasPrefix(value.MemoryProviderUsed, "external_fallback:") {
		return invalidEnum("diagnostics.memory_provider_used", "usar un provider documentado")
	}
	if !oneOf(value.ContextGraphStatus, "ready", "stale", "not_available", "disabled", "not_applicable", "not_run") {
		return invalidEnum("diagnostics.context_graph_status", "usar un status documentado")
	}
	if err := requireList(value.ContextGraphQueries, "diagnostics.context_graph_queries"); err != nil {
		return err
	}
	if value.FallbackReason == "" {
		return required("diagnostics.fallback_reason")
	}
	return validateBoolValue(value.GenericDiscoveryBeforeGraph, "diagnostics.generic_discovery_before_graph", true)
}

func validateSurface(value SurfaceExecution) error {
	if !oneOf(value.SchemaVersion, "surface-execution-plan/v1", "not_available", "not_applicable") {
		return invalidEnum("surface_execution.schema_version", "usar una versión documentada")
	}
	if !oneOf(value.Source, "explicit", "files", "git_diff", "carried_handoff", "not_available", "not_applicable") {
		return invalidEnum("surface_execution.source", "usar un source documentado")
	}
	if value.PrimarySurface == "" {
		return required("surface_execution.primary_surface")
	}
	if !oneOf(value.Mode, "single", "composed", "not_available", "not_applicable") {
		return invalidEnum("surface_execution.mode", "usar un mode documentado")
	}
	if err := requireList(value.ActiveSurfaces, "surface_execution.active_surfaces"); err != nil {
		return err
	}
	return requireList(value.ValidationRuleIDs, "surface_execution.validation_rule_ids")
}

func validateStructural(value *StructuralAcceptance) error {
	if value == nil {
		return nil
	}
	if !oneOf(value.Source, "user_prompt", "project_profile", "spec", "mixed", "not_available") {
		return invalidEnum("structural_acceptance.source", "usar un source documentado")
	}
	for path, list := range map[string][]string{
		"structural_acceptance.expected_directories":    value.ExpectedDirectories,
		"structural_acceptance.expected_architecture":   value.ExpectedArchitecture,
		"structural_acceptance.forbidden_root_patterns": value.ForbiddenRootPatterns,
	} {
		if err := requireList(list, path); err != nil {
			return err
		}
	}
	if value.Normalization == "" {
		return required("structural_acceptance.normalization")
	}
	if len(value.Audit) == 0 {
		return required("structural_acceptance.audit")
	}
	for _, audit := range value.Audit {
		if audit.FeatureOrSurface == "" {
			return required("structural_acceptance.audit.feature_or_surface")
		}
		if !oneOf(audit.Status, "satisfied", "missing", "blocked", "not_applicable") {
			return invalidEnum("structural_acceptance.audit.status", "usar un status documentado")
		}
		if audit.Notes == "" {
			return required("structural_acceptance.audit.notes")
		}
	}
	return nil
}

func validateWorkflow(value WorkflowDecision) error {
	for path, tier := range map[string]string{
		"workflow_decision.tier":         value.Tier,
		"workflow_decision.program_tier": value.ProgramTier,
		"workflow_decision.slice_tier":   value.SliceTier,
	} {
		if !oneOf(tier, "T1", "T2", "T3", "not_applicable") {
			return invalidEnum(path, "usar T1, T2, T3 o not_applicable")
		}
	}
	if err := validateBoolValue(value.FastPathAllowed, "workflow_decision.fast_path_allowed", false); err != nil {
		return err
	}
	if !oneOf(value.AdapterContext.ToolID, "opencode", "codex", "not_applicable") {
		return invalidEnum("workflow_decision.adapter_context.tool_id", "usar un tool_id documentado")
	}
	if !oneOf(value.AdapterContext.MethodologyID, "openspec", "lufy-sdd", "none", "not_applicable") {
		return invalidEnum("workflow_decision.adapter_context.methodology_id", "usar una methodology documentada")
	}
	if !oneOf(value.AdapterContext.MethodologyMode, "full", "lite", "none", "not_applicable") {
		return invalidEnum("workflow_decision.adapter_context.methodology_mode", "usar un mode documentado")
	}
	if err := validateBoolValue(value.AdapterContext.MethodologyRequired, "workflow_decision.adapter_context.methodology_required", false); err != nil {
		return err
	}
	if !oneOf(value.AdapterContext.ExecutionMode, "full-sdd", "sdd-lite", "express", "not_applicable") {
		return invalidEnum("workflow_decision.adapter_context.execution_mode", "usar un execution mode documentado")
	}
	if !oneOf(value.WorkflowLimitsSource, "workflow_limits", "not_available") {
		return invalidEnum("workflow_decision.workflow_limits_source", "usar workflow_limits o not_available")
	}
	limits := map[string]string{
		"workflow_decision.workflow_limits_paths.sizing":           value.WorkflowLimitsPaths.Sizing,
		"workflow_decision.workflow_limits_paths.routing":          value.WorkflowLimitsPaths.Routing,
		"workflow_decision.workflow_limits_paths.proposal_slicing": value.WorkflowLimitsPaths.ProposalSlicing,
		"workflow_decision.workflow_limits_paths.preflight":        value.WorkflowLimitsPaths.Preflight,
		"workflow_decision.workflow_limits_paths.stop_rules":       value.WorkflowLimitsPaths.StopRules,
	}
	for path, field := range limits {
		if field == "" {
			return required(path)
		}
	}
	if value.WorkflowLimitsPaths.DeliveryBatching == "" {
		return required("workflow_decision.workflow_limits_paths.delivery_batching")
	}
	if err := validateBoolValue(value.WorkloadDecisionNeeded, "workflow_decision.workload_decision_needed", false); err != nil {
		return err
	}
	if err := requireList(value.ReviewSlices, "workflow_decision.review_slices"); err != nil {
		return err
	}
	if !oneOf(value.PreflightStatus, "passed", "blocked", "not_applicable", "not_available") {
		return invalidEnum("workflow_decision.preflight_status", "usar un status documentado")
	}
	if !oneOf(value.StopRuleStatus, "clear", "triggered", "not_applicable", "not_available") {
		return invalidEnum("workflow_decision.stop_rule_status", "usar un status documentado")
	}
	if value.DeliveryBatchingGuidance == "" {
		return required("workflow_decision.delivery_batching_guidance")
	}
	return validateArtifactBranching(value.ArtifactBranching)
}

func validateArtifactBranching(value ArtifactBranching) error {
	if !oneOf(value.Status, "recommended", "not_needed", "disabled", "not_applicable") {
		return invalidEnum("workflow_decision.artifact_branching.status", "usar un status documentado")
	}
	if !oneOf(value.Stage, "proposal", "design", "tasks", "not_applicable") {
		return invalidEnum("workflow_decision.artifact_branching.stage", "usar un stage documentado")
	}
	if !oneOf(string(value.CandidateCount), "1", "2", "not_applicable") {
		return invalidEnum("workflow_decision.artifact_branching.candidate_count", "usar 1, 2 o not_applicable")
	}
	if value.Reason == "" {
		return required("workflow_decision.artifact_branching.reason")
	}
	for path, boolean := range map[string]BoolValue{
		"workflow_decision.artifact_branching.parallel_allowed":    value.ParallelAllowed,
		"workflow_decision.artifact_branching.requires_join":       value.RequiresJoin,
		"workflow_decision.artifact_branching.merge_plan_required": value.MergePlanRequired,
	} {
		if err := validateBoolValue(boolean, path, false); err != nil {
			return err
		}
	}
	if value.CandidateIsolation == "" {
		return required("workflow_decision.artifact_branching.candidate_isolation")
	}
	for path, list := range map[string][]string{
		"workflow_decision.artifact_branching.human_escalation_triggers": value.HumanEscalationTriggers,
		"workflow_decision.artifact_branching.candidate_paths":           value.CandidatePaths,
		"workflow_decision.artifact_branching.canonical_artifact_set":    value.CanonicalArtifactSet,
	} {
		if err := requireList(list, path); err != nil {
			return err
		}
	}
	for _, trigger := range value.HumanEscalationTriggers {
		if !oneOf(trigger, "public_contract", "security", "product_direction", "significant_ux", "non_objective_tradeoff", "not_applicable") {
			return invalidEnum("workflow_decision.artifact_branching.human_escalation_triggers", "usar triggers documentados")
		}
	}
	return nil
}

func validateBoolValue(value BoolValue, path string, allowUnavailable bool) error {
	allowed := oneOf(string(value), "true", "false", "not_applicable")
	if allowUnavailable {
		allowed = allowed || value == BoolNotAvailable
	}
	if !allowed {
		return invalidEnum(path, "usar true, false o un marcador documentado")
	}
	return nil
}

func validateBounds(value reflect.Value, path string) error {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		return validateBounds(value.Elem(), path)
	}
	switch value.Kind() {
	case reflect.Struct:
		typ := value.Type()
		for index := 0; index < value.NumField(); index++ {
			name := strings.Split(typ.Field(index).Tag.Get("yaml"), ",")[0]
			if err := validateBounds(value.Field(index), joinPath(path, name)); err != nil {
				return err
			}
		}
	case reflect.Slice:
		if value.Len() > maxListItems {
			return diagnostic("list_too_large", displayPath(path), "reducir la lista a 64 elementos o menos")
		}
		for index := 0; index < value.Len(); index++ {
			if err := validateBounds(value.Index(index), path); err != nil {
				return err
			}
		}
	case reflect.String:
		if len(value.String()) > maxStringBytes {
			return diagnostic("string_too_large", displayPath(path), "reducir el valor a 8 KiB o menos")
		}
	}
	return nil
}

func invalidEnum(path, recovery string) error {
	return diagnostic("invalid_enum", path, recovery)
}

func requireList(values []string, path string) error {
	if len(values) == 0 {
		return required(path)
	}
	for _, value := range values {
		if value == "" {
			return required(path)
		}
	}
	return nil
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
