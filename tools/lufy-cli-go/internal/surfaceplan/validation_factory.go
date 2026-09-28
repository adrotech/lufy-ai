package surfaceplan

import (
	"sort"
	"strings"
)

type ValidationPlanFactory struct{}

func (ValidationPlanFactory) Build(project ProjectDefinition, resolution Resolution, files, requestedCapabilities []string) ([]ValidationRule, []string, []string) {
	signals := detectSignals(files)
	capabilitySurfaces := collectCapabilitySurfaces(resolution.Surfaces, requestedCapabilities)
	capabilities := sortedMapKeys(capabilitySurfaces)
	builder := newRuleBuilder()

	for _, surface := range resolution.Surfaces {
		for _, expectation := range surface.ValidationExpectations {
			if !expectationApplies(expectation, signals) {
				continue
			}
			rule := ruleForExpectation(expectation, surface.ID)
			rule.Commands = commandsForExpectation(project, surface, expectation)
			builder.add(rule)
		}
	}

	if frontend := surfaceIDsByType(resolution.Surfaces, "frontend"); signals["ui_change"] && len(frontend) > 0 {
		builder.add(ValidationRule{ID: "browser_evidence", Category: "frontend", Trigger: "ui_change", Required: true, Surfaces: frontend, Evidence: []string{"recorrido visual e interacción principal verificados"}, Rationale: "un cambio visual requiere evidencia del comportamiento renderizado"})
	}
	if signals["contract_change"] && resolution.Mode == ModeComposed {
		builder.add(ValidationRule{ID: "contract_compatibility", Category: "contract", Trigger: "contract_change", Required: true, Surfaces: ids(resolution.Surfaces), Evidence: []string{"productor y consumidores compatibles", "errores del backend mapeados por el frontend"}, Rationale: "el contrato cruza superficies conectadas"})
		builder.add(ValidationRule{ID: "e2e_cross_surface", Category: "end_to_end", Trigger: "contract_change", Required: true, Surfaces: ids(resolution.Surfaces), Evidence: []string{"flujo crítico completo verificado"}, Rationale: "las pruebas aisladas no demuestran la integración del contrato"})
	}
	if signals["persistence_change"] {
		builder.add(ValidationRule{ID: "persistence_compatibility", Category: "persistence", Trigger: "persistence_change", Required: true, Surfaces: ids(resolution.Surfaces), Evidence: []string{"migración hacia adelante", "lectura de estado anterior", "rollback o recuperación documentada"}, Rationale: "el estado persistido necesita compatibilidad y recuperación"})
	}

	for _, capability := range capabilities {
		if rule, ok := ruleForCapability(capability, capabilitySurfaces[capability]); ok {
			builder.add(rule)
		}
	}
	return builder.rules, sortedTrueKeys(signals), capabilities
}

type ruleBuilder struct {
	rules []ValidationRule
	index map[string]int
}

func newRuleBuilder() *ruleBuilder {
	return &ruleBuilder{index: map[string]int{}}
}

func (b *ruleBuilder) add(rule ValidationRule) {
	if rule.ID == "" {
		return
	}
	if index, ok := b.index[rule.ID]; ok {
		current := b.rules[index]
		current.Required = current.Required || rule.Required
		current.Surfaces = uniqueStrings(append(current.Surfaces, rule.Surfaces...))
		current.Commands = uniqueStrings(append(current.Commands, rule.Commands...))
		current.Evidence = uniqueStrings(append(current.Evidence, rule.Evidence...))
		b.rules[index] = current
		return
	}
	rule.Surfaces = uniqueStrings(rule.Surfaces)
	rule.Commands = uniqueStrings(rule.Commands)
	rule.Evidence = uniqueStrings(rule.Evidence)
	b.index[rule.ID] = len(b.rules)
	b.rules = append(b.rules, rule)
}

func ruleForExpectation(expectation, surfaceID string) ValidationRule {
	category := "quality"
	evidence := []string{expectation + " satisfecho"}
	switch expectation {
	case "typecheck", "static_analysis":
		category = "static_analysis"
	case "lint", "format":
		category = "style"
	case "unit_tests", "integration_tests_when_contract_changes", "coverage":
		category = "tests"
	case "build", "build_or_bundle_check", "build_or_package_check":
		category = "build"
	case "browser_check_when_ui_changes", "device_flow_check_when_ui_changes":
		category = "runtime"
	case "contract_tests_when_available":
		category = "contract"
	case "e2e_smoke_when_flow_changes":
		category = "end_to_end"
	case "feature_boundary_review", "structural_acceptance_audit":
		category = "architecture"
	}
	return ValidationRule{ID: expectation, Category: category, Trigger: expectation, Required: true, Surfaces: []string{surfaceID}, Evidence: evidence, Rationale: "expectativa declarada por project_profile"}
}

func ruleForCapability(capability string, surfaces []string) (ValidationRule, bool) {
	base := ValidationRule{Trigger: "capability:" + capability, Required: true, Surfaces: surfaces}
	switch capability {
	case "realtime":
		base.ID, base.Category, base.Evidence, base.Rationale = "realtime_frame_timing", "performance", []string{"cadencia y latencia medidas bajo carga representativa"}, "los loops interactivos necesitan presupuesto temporal observable"
	case "rendering":
		base.ID, base.Category, base.Evidence, base.Rationale = "visual_rendering", "runtime", []string{"render real inspeccionado", "fallback de renderer verificado cuando aplica"}, "la salida renderizada no se demuestra sólo con tests unitarios"
	case "offline":
		base.ID, base.Category, base.Evidence, base.Rationale = "offline_operation", "resilience", []string{"flujo principal ejecutado sin red"}, "la capacidad offline requiere una prueba sin conectividad"
	case "persistent-state":
		base.ID, base.Category, base.Evidence, base.Rationale = "persistent_state_compatibility", "persistence", []string{"guardar, cerrar y restaurar estado", "compatibilidad de versión verificada"}, "el estado durable debe sobrevivir reinicios y evolución de formato"
	case "desktop-shell":
		base.ID, base.Category, base.Evidence, base.Rationale = "desktop_runtime_parity", "runtime", []string{"flujo validado en shell de escritorio y navegador cuando ambos aplican"}, "el webview puede comportarse distinto del navegador"
	default:
		return ValidationRule{}, false
	}
	return base, true
}

func expectationApplies(expectation string, signals map[string]bool) bool {
	switch expectation {
	case "browser_check_when_ui_changes", "device_flow_check_when_ui_changes":
		return signals["ui_change"]
	case "integration_tests_when_contract_changes", "contract_tests_when_available":
		return signals["contract_change"]
	case "e2e_smoke_when_flow_changes":
		return signals["contract_change"] || signals["ui_change"]
	default:
		return true
	}
}

func commandsForExpectation(project ProjectDefinition, surface SurfaceDefinition, expectation string) []string {
	var out []string
	for _, stackID := range surface.Stacks {
		stack, ok := project.Stacks[stackID]
		if !ok {
			continue
		}
		command := ""
		switch expectation {
		case "typecheck", "static_analysis":
			command = stack.StaticCommand
		case "lint":
			command = stack.LintCommand
		case "unit_tests", "integration_tests_when_contract_changes", "contract_tests_when_available":
			command = stack.TestCommand
		case "coverage":
			command = stack.CoverageCommand
		}
		if strings.TrimSpace(command) != "" && command != "TODO" {
			out = append(out, command)
		}
	}
	return uniqueStrings(out)
}

func detectSignals(files []string) map[string]bool {
	signals := map[string]bool{}
	for _, file := range files {
		path := strings.ToLower(normalizeSlash(file))
		if hasAnySuffix(path, ".tsx", ".jsx", ".css", ".scss", ".sass", ".vue", ".svelte") || containsAny(path, "/components/", "/pages/", "/views/") {
			signals["ui_change"] = true
		}
		if containsAny(path, "openapi", "/api/", "contract", "/dto", "schema.graphql", "/shared/types", "/types/api") {
			signals["contract_change"] = true
		}
		if containsAny(path, "migration", "/database/", "/db/", ".sql", "/repositories/", "/persistence/", "schema.prisma") {
			signals["persistence_change"] = true
		}
	}
	return signals
}

func collectCapabilitySurfaces(surfaces []SurfaceDefinition, requested []string) map[string][]string {
	out := map[string][]string{}
	for _, surface := range surfaces {
		for _, capability := range surface.Capabilities {
			capability = strings.ToLower(strings.TrimSpace(capability))
			if capability != "" {
				out[capability] = uniqueStrings(append(out[capability], surface.ID))
			}
		}
	}
	all := ids(surfaces)
	for _, capability := range requested {
		capability = strings.ToLower(strings.TrimSpace(capability))
		if capability != "" {
			out[capability] = uniqueStrings(append(out[capability], all...))
		}
	}
	return out
}

func surfaceIDsByType(surfaces []SurfaceDefinition, surfaceType string) []string {
	var out []string
	for _, surface := range surfaces {
		if surface.Type == surfaceType {
			out = append(out, surface.ID)
		}
	}
	return out
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func hasAnySuffix(value string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(value, suffix) {
			return true
		}
	}
	return false
}

func sortedTrueKeys(values map[string]bool) []string {
	var out []string
	for key, value := range values {
		if value {
			out = append(out, key)
		}
	}
	sort.Strings(out)
	return out
}

func sortedMapKeys[T any](values map[string]T) []string {
	var out []string
	for key := range values {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
