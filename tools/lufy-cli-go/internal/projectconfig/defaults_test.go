package projectconfig

import "testing"

func TestDefaultArchitectureStructuralExpectationsByBackendStyle(t *testing.T) {
	cases := map[string]string{
		"controller_service_repository": "repositories_isolate_persistence",
		"clean_architecture":            "application_or_usecase_layer_has_business_flows",
		"hexagonal":                     "adapters_implement_ports_at_boundaries",
	}

	for preferred, want := range cases {
		t.Run(preferred, func(t *testing.T) {
			got := DefaultArchitectureStructuralExpectations("backend", preferred)
			if !contains(got, want) {
				t.Fatalf("DefaultArchitectureStructuralExpectations(%q) = %#v, missing %q", preferred, got, want)
			}
		})
	}
}

func TestApplySurfaceDefaultsCompletesStructuralExpectations(t *testing.T) {
	surface := ApplySurfaceDefaults(ProjectSurface{
		ID:   "api",
		Type: "backend",
		AgentLens: AgentLens{
			PrimaryConcerns:        []string{"custom-domain"},
			ValidationExpectations: []string{"custom-validation"},
		},
		Architecture: ArchitectureProfile{
			Preferred: "hexagonal",
		},
	})

	if !contains(surface.AgentLens.StructuralExpectations, "follow_project_profile_surface_architecture") {
		t.Fatalf("agent lens structural expectations were not completed: %#v", surface.AgentLens)
	}
	if !contains(surface.Architecture.Options, "controller_service_repository") {
		t.Fatalf("architecture options were not completed: %#v", surface.Architecture)
	}
	if !contains(surface.Architecture.StructuralExpectations, "ports_define_inbound_and_outbound_contracts") {
		t.Fatalf("hexagonal structural expectations were not completed: %#v", surface.Architecture)
	}
}
