package projectconfig

func defaultTDD() TDDConfig {
	return TDDConfig{Strict: true, TriangulateRequired: true, EdgeCaseCategories: []string{"boundary", "error_path", "concurrency", "data_shape", "time_sensitive"}}
}

func defaultWorkflowLimits() WorkflowLimits {
	return WorkflowLimits{Sizing: WorkflowSizing{LOCBudget: 400}, Routing: WorkflowRouting{Strategy: "proportional-sdd"}, ProposalSlicingStrategy: "review-slices-on-multi-risk", DeliveryBatchStrategy: "ask-on-risk", StopRules: []string{"pause_on_scope_growth", "escalate_on_security_or_delivery_risk", "stop_before_unauthorized_git_or_gh"}, Preflight: []string{"read_project_config", "confirm_applicable_toolchain", "plan_grouped_validation"}}
}

func DefaultAgentLens(surfaceType string) AgentLens {
	switch surfaceType {
	case "frontend":
		return AgentLens{PrimaryConcerns: []string{"ux_states", "accessibility", "responsive_layout", "design_system", "client_state", "api_consumption", "perceived_performance"}, ValidationExpectations: []string{"typecheck", "lint", "unit_tests", "build", "browser_check_when_ui_changes"}}
	case "backend":
		return AgentLens{PrimaryConcerns: []string{"domain_invariants", "api_contracts", "auth", "persistence", "transactions", "idempotency", "observability", "resilience"}, ValidationExpectations: []string{"unit_tests", "integration_tests_when_contract_changes", "static_analysis", "coverage"}}
	case "mobile":
		return AgentLens{PrimaryConcerns: []string{"navigation_flows", "offline_and_network_states", "accessibility", "device_constraints", "platform_differences", "release_channels"}, ValidationExpectations: []string{"typecheck", "lint", "unit_tests", "build_or_bundle_check", "device_flow_check_when_ui_changes"}}
	case "cli":
		return AgentLens{PrimaryConcerns: []string{"command_contracts", "flags_and_exit_codes", "filesystem_safety", "idempotency", "scriptability", "error_messages"}, ValidationExpectations: []string{"unit_tests", "command_smokes", "static_analysis", "build"}}
	case "infra":
		return AgentLens{PrimaryConcerns: []string{"plan_drift", "secrets", "least_privilege", "rollback", "environment_parity", "supply_chain"}, ValidationExpectations: []string{"format", "validate", "plan_review", "policy_check_when_available"}}
	case "fullstack":
		return AgentLens{PrimaryConcerns: []string{"frontend_backend_contract", "error_state_mapping", "e2e_critical_paths", "rollout_and_rollback", "api_version_compatibility"}, ValidationExpectations: []string{"contract_tests_when_available", "frontend_validation", "backend_validation", "e2e_smoke_when_flow_changes"}}
	default:
		return AgentLens{PrimaryConcerns: []string{"public_contracts", "api_shape", "compatibility", "maintainability", "consumer_usage"}, ValidationExpectations: []string{"unit_tests", "static_analysis", "build_or_package_check"}}
	}
}
