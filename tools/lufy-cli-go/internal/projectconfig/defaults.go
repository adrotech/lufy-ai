package projectconfig

func defaultTDD() TDDConfig {
	return TDDConfig{Strict: true, TriangulateRequired: true, EdgeCaseCategories: []string{"boundary", "error_path", "concurrency", "data_shape", "time_sensitive"}}
}

func defaultWorkflowLimits() WorkflowLimits {
	return WorkflowLimits{Sizing: WorkflowSizing{LOCBudget: 400}, Routing: WorkflowRouting{Strategy: "proportional-sdd"}, ProposalSlicingStrategy: "review-slices-on-multi-risk", DeliveryBatchStrategy: "ask-on-risk", StopRules: []string{"pause_on_scope_growth", "escalate_on_security_or_delivery_risk", "stop_before_unauthorized_git_or_gh"}, Preflight: []string{"read_project_config", "confirm_applicable_toolchain", "plan_grouped_validation"}}
}

func DefaultMemoryConfig() MemoryConfig {
	return MemoryConfig{
		Provider:       "obsidian",
		Root:           ".lufy/memory",
		GitPolicy:      "ignored",
		SchemaVersion:  1,
		Search:         "rg",
		BacklinksIndex: ".lufy/memory/index/backlinks.json",
	}
}

func DefaultParallelExecutionConfig() ParallelExecutionConfig {
	return ParallelExecutionConfig{
		Enabled:                  true,
		Strategy:                 "independent_review_slices",
		MaxParallelAgents:        3,
		RequiresIndependentFiles: true,
		RequiresMergePlan:        true,
		ValidationMode:           "grouped_after_join",
	}
}

func DefaultAgentLens(surfaceType string) AgentLens {
	switch surfaceType {
	case "frontend":
		return AgentLens{
			PrimaryConcerns:        []string{"ux_states", "accessibility", "responsive_layout", "design_system", "client_state", "api_consumption", "perceived_performance", "feature_driven_structure", "feature_colocation", "feature_public_barrels_index_ts", "pages_as_routing_only"},
			ValidationExpectations: []string{"typecheck", "lint", "unit_tests", "build", "browser_check_when_ui_changes", "feature_boundary_review", "structural_acceptance_audit"},
			StructuralExpectations: []string{"src/features/<feature>/components", "src/features/<feature>/hooks", "src/features/<feature>/services_when_applicable", "src/features/<feature>/types.ts_when_applicable", "src/features/<feature>/index.ts_public_barrel", "src/pages_for_routing_and_layouts_only"},
		}
	case "backend":
		return AgentLens{
			PrimaryConcerns:        []string{"domain_invariants", "api_contracts", "auth", "persistence", "transactions", "idempotency", "observability", "resilience", "architecture_consistency"},
			ValidationExpectations: []string{"unit_tests", "integration_tests_when_contract_changes", "static_analysis", "coverage", "structural_acceptance_audit"},
			StructuralExpectations: []string{"follow_project_profile_surface_architecture", "audit_existing_architecture_before_new_layers", "block_validation_on_requested_structure_drift"},
		}
	case "mobile":
		return AgentLens{PrimaryConcerns: []string{"navigation_flows", "offline_and_network_states", "accessibility", "device_constraints", "platform_differences", "release_channels"}, ValidationExpectations: []string{"typecheck", "lint", "unit_tests", "build_or_bundle_check", "device_flow_check_when_ui_changes"}}
	case "cli":
		return AgentLens{PrimaryConcerns: []string{"command_contracts", "flags_and_exit_codes", "filesystem_safety", "idempotency", "scriptability", "error_messages"}, ValidationExpectations: []string{"unit_tests", "command_smokes", "static_analysis", "build"}}
	case "infra":
		return AgentLens{PrimaryConcerns: []string{"plan_drift", "secrets", "least_privilege", "rollback", "environment_parity", "supply_chain"}, ValidationExpectations: []string{"format", "validate", "plan_review", "policy_check_when_available"}}
	case "fullstack":
		return AgentLens{
			PrimaryConcerns:        []string{"frontend_backend_contract", "error_state_mapping", "e2e_critical_paths", "rollout_and_rollback", "api_version_compatibility", "feature_driven_frontend_structure", "feature_colocation", "feature_public_barrels_index_ts", "pages_as_routing_only"},
			ValidationExpectations: []string{"contract_tests_when_available", "frontend_validation", "backend_validation", "e2e_smoke_when_flow_changes", "feature_boundary_review", "structural_acceptance_audit"},
			StructuralExpectations: []string{"frontend_feature_dirs_under_src_features", "backend_structure_from_connected_backend_surface", "block_validation_on_requested_structure_drift"},
		}
	default:
		return AgentLens{PrimaryConcerns: []string{"public_contracts", "api_shape", "compatibility", "maintainability", "consumer_usage"}, ValidationExpectations: []string{"unit_tests", "static_analysis", "build_or_package_check"}}
	}
}

func DefaultArchitectureProfile(surfaceType string) ArchitectureProfile {
	switch surfaceType {
	case "frontend":
		return ArchitectureProfile{
			Preferred:              "feature_driven",
			Options:                []string{"feature_driven"},
			ReviewRequired:         true,
			StructuralExpectations: defaultArchitectureStructuralExpectations("frontend", "feature_driven"),
			Notes:                  "Usar src/features/<feature> con colocation y index.ts como frontera publica; pages queda para routing/layouts.",
		}
	case "backend":
		return ArchitectureProfile{
			Preferred:              "controller_service_repository",
			Options:                []string{"controller_service_repository", "clean_architecture", "hexagonal"},
			ReviewRequired:         true,
			StructuralExpectations: defaultArchitectureStructuralExpectations("backend", "controller_service_repository"),
			Notes:                  "Revisar arquitectura existente antes de crear capas nuevas; minimo controller/service/repository.",
		}
	case "fullstack":
		return ArchitectureProfile{
			Preferred:              "feature_driven_frontend",
			Options:                []string{"feature_driven_frontend"},
			ReviewRequired:         true,
			StructuralExpectations: defaultArchitectureStructuralExpectations("fullstack", "feature_driven_frontend"),
			Notes:                  "Fullstack combina frontend feature-driven; la arquitectura backend se define en la surface backend conectada.",
		}
	default:
		return ArchitectureProfile{}
	}
}

func DefaultArchitectureStructuralExpectations(surfaceType, preferred string) []string {
	return defaultArchitectureStructuralExpectations(surfaceType, preferred)
}

func ApplySurfaceDefaults(surface ProjectSurface) ProjectSurface {
	if len(surface.AgentLens.PrimaryConcerns) == 0 && len(surface.AgentLens.ValidationExpectations) == 0 {
		surface.AgentLens = DefaultAgentLens(surface.Type)
	}
	if len(surface.AgentLens.StructuralExpectations) == 0 {
		surface.AgentLens.StructuralExpectations = DefaultAgentLens(surface.Type).StructuralExpectations
	}
	surface.Architecture = completeArchitectureProfile(surface.Type, surface.Architecture)
	return surface
}

func completeArchitectureProfile(surfaceType string, profile ArchitectureProfile) ArchitectureProfile {
	defaults := DefaultArchitectureProfile(surfaceType)
	if isZeroArchitecture(profile) {
		return defaults
	}
	if profile.Preferred == "" {
		profile.Preferred = defaults.Preferred
	}
	if len(profile.Options) == 0 {
		profile.Options = defaults.Options
	}
	if len(profile.StructuralExpectations) == 0 {
		profile.StructuralExpectations = defaultArchitectureStructuralExpectations(surfaceType, profile.Preferred)
	}
	if profile.Notes == "" {
		profile.Notes = defaults.Notes
	}
	return profile
}

func isZeroArchitecture(profile ArchitectureProfile) bool {
	return len(profile.Detected) == 0 &&
		profile.Preferred == "" &&
		len(profile.Options) == 0 &&
		len(profile.StructuralExpectations) == 0 &&
		!profile.ReviewRequired &&
		profile.Notes == ""
}

func defaultArchitectureStructuralExpectations(surfaceType, preferred string) []string {
	switch surfaceType {
	case "frontend":
		return []string{"src/features/<feature>/components", "src/features/<feature>/hooks", "src/features/<feature>/services_when_applicable", "src/features/<feature>/types.ts_when_applicable", "src/features/<feature>/index.ts_public_barrel", "src/pages_for_routing_and_layouts_only"}
	case "backend":
		switch preferred {
		case "clean_architecture":
			return []string{"domain_layer_has_entities_and_invariants", "application_or_usecase_layer_has_business_flows", "infrastructure_layer_contains_external_adapters", "controllers_or_handlers_do_not_import_persistence_entities"}
		case "hexagonal":
			return []string{"domain_core_has_no_adapter_dependencies", "ports_define_inbound_and_outbound_contracts", "adapters_implement_ports_at_boundaries", "composition_root_wires_dependencies"}
		default:
			return []string{"controllers_or_handlers_expose_transport_contracts_only", "services_own_business_rules", "repositories_isolate_persistence", "dtos_or_contracts_do_not_expose_persistence_entities"}
		}
	case "fullstack":
		return []string{"frontend_feature_dirs_under_src_features", "backend_structure_from_connected_backend_surface"}
	default:
		return nil
	}
}
