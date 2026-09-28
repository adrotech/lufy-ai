package surfaceplan

import "testing"

func TestValidationFactoryBuildsCrossSurfaceRulesAndCommands(t *testing.T) {
	project := sampleProject()
	resolution, err := (SurfaceResolverStrategy{}).Resolve(project, "auto", []string{"web/src/App.tsx", "api/openapi.yaml"})
	if err != nil {
		t.Fatal(err)
	}

	rules, signals, _ := (ValidationPlanFactory{}).Build(project, resolution, []string{"web/src/App.tsx", "api/openapi.yaml"}, nil)

	for _, signal := range []string{"contract_change", "ui_change"} {
		if !hasString(signals, signal) {
			t.Fatalf("missing signal %s: %#v", signal, signals)
		}
	}
	for _, id := range []string{"typecheck", "unit_tests", "browser_evidence", "contract_compatibility", "e2e_cross_surface"} {
		if !hasRule(rules, id) {
			t.Fatalf("missing rule %s: %#v", id, rules)
		}
	}
	typecheck := findRule(rules, "typecheck")
	if !hasString(typecheck.Commands, "tsc --noEmit") {
		t.Fatalf("missing discovered command: %#v", typecheck)
	}
	unit := findRule(rules, "unit_tests")
	if !hasString(unit.Commands, "go test ./...") || hasString(unit.Commands, "vitest run") {
		t.Fatalf("unit test commands leaked across surfaces: %#v", unit)
	}
}

func TestValidationFactoryAddsGenericInteractiveCapabilities(t *testing.T) {
	project := sampleProject()
	project.Surfaces[0].Capabilities = []string{"rendering", "offline", "persistent-state"}
	resolution, err := (SurfaceResolverStrategy{}).Resolve(project, "web", nil)
	if err != nil {
		t.Fatal(err)
	}

	rules, _, capabilities := (ValidationPlanFactory{}).Build(project, resolution, nil, []string{"realtime", "desktop-shell"})

	for _, capability := range []string{"desktop-shell", "offline", "persistent-state", "realtime", "rendering"} {
		if !hasString(capabilities, capability) {
			t.Fatalf("missing capability %s: %#v", capability, capabilities)
		}
	}
	for _, id := range []string{"realtime_frame_timing", "visual_rendering", "offline_operation", "persistent_state_compatibility", "desktop_runtime_parity"} {
		if !hasRule(rules, id) {
			t.Fatalf("missing capability rule %s", id)
		}
	}
	if rendering := findRule(rules, "visual_rendering"); !hasString(rendering.Surfaces, "web") || hasString(rendering.Surfaces, "api") {
		t.Fatalf("declared capability leaked to another surface: %#v", rendering)
	}
}

func TestValidationFactoryAddsPersistenceCompatibility(t *testing.T) {
	project := sampleProject()
	resolution, err := (SurfaceResolverStrategy{}).Resolve(project, "api", []string{"api/migrations/001.sql"})
	if err != nil {
		t.Fatal(err)
	}

	rules, signals, _ := (ValidationPlanFactory{}).Build(project, resolution, []string{"api/migrations/001.sql"}, nil)

	if !hasString(signals, "persistence_change") || !hasRule(rules, "persistence_compatibility") {
		t.Fatalf("missing persistence policy: signals=%v rules=%#v", signals, rules)
	}
}

func hasRule(rules []ValidationRule, id string) bool { return findRule(rules, id).ID != "" }

func findRule(rules []ValidationRule, id string) ValidationRule {
	for _, rule := range rules {
		if rule.ID == id {
			return rule
		}
	}
	return ValidationRule{}
}

func hasString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
