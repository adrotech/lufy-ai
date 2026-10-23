package assets

import (
	"path/filepath"
	"strings"
	"testing"
)

// Estas pruebas protegen el texto distribuido y la matriz de decisiones.
// No ejecutan modelos ni prueban conducta, latencia o consumo de tokens.
func TestProportionalWorkflowContractS1(t *testing.T) {
	root := repoRoot(t)
	rel := filepath.Join(".lufy", "contracts", "delivery.md")
	body := readAssetTestFile(t, filepath.Join(root, rel))
	mirror := readAssetTestFile(t, filepath.Join(root, "tools", "lufy-cli-go", "internal", "assets", "embedded", rel))
	if body != mirror {
		t.Error("delivery contract and embedded mirror differ")
	}
	cases := []struct {
		name string
		want []string
	}{
		{"AC1_risk_and_workload", []string{"riesgo, incertidumbre e impacto", "recuento de archivos por sí solo no eleva el tier", "workload"}},
		{"AC2_small_sensitive_change", []string{"una línea", "auth runtime", "configuración sensible", "al menos T2", "fast_path_allowed: false"}},
		{"AC3_significant_file", []string{"archivo significativo", "comportamiento, aceptación, contrato, configuración operacional o instrucciones del harness", "documental no lo excluye"}},
		{"AC3_coverage_checkpoint", []string{"cuatro o más archivos significativos", "plan/slice vigente", "alcance, aceptación y validación", "frontera segura"}},
		{"AC4_workload_checkpoint", []string{"cinco o más archivos", "decisión de workload", "única unidad justificada", "no fragmentar T3 artificialmente"}},
		{"AC5_constraint_provenance", []string{"workflow_limits", "not_available", "no inventar budgets", "checkpoints del harness", "configuración efectivamente leída"}},
		{"AC6_independent_gates", []string{"aprobación post-plan", "seguridad", "validación", "delivery", "stop rule independiente", "blocked", "escalated", "recovery"}},
		{"AC7_visible_mirrors", []string{"misma decisión lógica", "churn real", "assessment ejecutable", "paridad"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, fragment := range tc.want {
				if !strings.Contains(strings.ToLower(body), strings.ToLower(fragment)) {
					t.Errorf("delivery contract missing documented rule %q", fragment)
				}
			}
		})
	}
}

func TestProportionalWorkflowDecisionMatrixS1(t *testing.T) {
	body := readAssetTestFile(t, filepath.Join(repoRoot(t), "docs", "harness-workflow.md"))
	cases := []struct {
		name      string
		caseText  string
		allowed   []string
		forbidden []string
	}{
		{"AC1_mechanical", "ocho archivos", []string{"tier", "workload"}, []string{"recuento"}},
		{"AC2_security", "una línea de auth", []string{"router", "T2", "false"}, []string{"T3"}},
		{"AC3_covered", "cuatro significativos con plan", []string{"cobertura", "gates"}, []string{"aprobación", "gate"}},
		{"AC3_uncovered", "cuatro significativos sin plan", []string{"pausa", "plan"}, []string{"escritura", "silenciosa"}},
		{"AC4_estimated", "cinco estimados", []string{"workload", "unidad", "justifica"}, []string{"slices", "artificiales"}},
		{"AC5_missing_config", "config ausente", []string{"not_available", "guardrails"}, []string{"budget", "inventado"}},
		{"AC6_stop_rule", "stop rule activa", []string{"blocked", "escalated", "recovery"}, []string{"ignorar", "stop rule"}},
		{"AC6_approval", "auto-chain sin aprobación", []string{"esperar", "aprobación"}, []string{"metadata", "permiso"}},
		{"AC7_mirrors", "fuente y mirror", []string{"decisión lógica", "diff real"}, []string{"excluir", "review ejecutable"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var cells []string
			for _, line := range strings.Split(body, "\n") {
				line = strings.TrimSpace(line)
				if !strings.HasPrefix(line, "|") {
					continue
				}
				row := strings.Split(strings.Trim(line, "|"), "|")
				if len(row) == 3 && strings.Contains(strings.ToLower(row[0]), tc.caseText) {
					if cells != nil {
						t.Fatalf("duplicate decision-matrix case %q", tc.caseText)
					}
					cells = row
				}
			}
			if cells == nil {
				t.Fatalf("missing case/expected/prohibited matrix row for %q", tc.caseText)
			}
			for i, fragments := range [][]string{tc.allowed, tc.forbidden} {
				for _, fragment := range fragments {
					if !strings.Contains(strings.ToLower(cells[i+1]), strings.ToLower(fragment)) {
						t.Errorf("matrix cell %d for %q missing %q", i+1, tc.caseText, fragment)
					}
				}
			}
		})
	}
	for _, fragment := range []string{"pruebas textuales y de paridad", "no evalúan la conducta", "latencia", "tokens"} {
		if !strings.Contains(strings.ToLower(body), fragment) {
			t.Errorf("matrix must document evidence limitation %q", fragment)
		}
	}
}

// La equivalencia comprobada es textual: no simula las decisiones de un agente.
func TestProportionalWorkflowCodexAndSharedSurfacesS2(t *testing.T) {
	root := repoRoot(t)
	paths := []string{
		".codex/agents/orchestrator.toml", ".codex/agents/sdd-router.toml", ".codex/agents/implementer.toml",
		"AGENTS.md", "AGENTS.md.template", "lufy-ia.harness.md",
	}
	rules := []struct {
		acceptance string
		fragments  []string
	}{
		{"AC1", []string{"riesgo, incertidumbre e impacto", "recuento de archivos por sí solo no eleva el tier"}},
		{"AC2", []string{"auth runtime", "configuración sensible", "al menos T2", "fast_path_allowed: false"}},
		{"AC3", []string{"archivo significativo", "cuatro o más archivos significativos", "plan/slice vigente", "frontera segura"}},
		{"AC4", []string{"cinco o más archivos", "decisión de workload", "única unidad justificada", "no fragmentar T3 artificialmente"}},
		{"AC5", []string{"workflow_limits", "not_available", "no inventar budgets", "checkpoints del harness"}},
		{"AC6", []string{"aprobación post-plan", "stop rule independiente", "seguridad", "validación", "delivery"}},
		{"AC7", []string{"misma decisión lógica", "churn real", "assessment ejecutable", "paridad"}},
	}
	for _, rel := range paths {
		t.Run(rel, func(t *testing.T) {
			body := readAssetTestFile(t, filepath.Join(root, filepath.FromSlash(rel)))
			for _, rule := range rules {
				t.Run(rule.acceptance, func(t *testing.T) {
					for _, fragment := range rule.fragments {
						if !strings.Contains(strings.ToLower(body), strings.ToLower(fragment)) {
							t.Errorf("%s missing proportional rule %q", rel, fragment)
						}
					}
				})
			}
			// AGENTS.md es local; template y agentes se distribuyen como assets.
			if rel != "AGENTS.md" {
				mirror := readAssetTestFile(t, filepath.Join(root, "tools", "lufy-cli-go", "internal", "assets", "embedded", filepath.FromSlash(rel)))
				if body != mirror {
					t.Error("proportional instructions differ from embedded mirror")
				}
			}
		})
	}
}

// El checkpoint de cinco archivos expresa workload en cada agente Codex y mirror.
func TestProportionalWorkflowCodexCountIsWorkloadS2(t *testing.T) {
	root := repoRoot(t)
	for _, name := range []string{"orchestrator", "sdd-router", "implementer"} {
		for _, prefix := range []string{"", "tools/lufy-cli-go/internal/assets/embedded"} {
			rel := filepath.Join(prefix, ".codex", "agents", name+".toml")
			t.Run(rel, func(t *testing.T) {
				body := strings.ToLower(readAssetTestFile(t, filepath.Join(root, rel)))
				found := false
				for _, line := range strings.Split(body, "\n") {
					if !strings.Contains(line, "cinco o más archivos") {
						continue
					}
					found = true
					for _, fragment := range []string{"decisión de workload", "única unidad justificada", "no fragmentar t3 artificialmente"} {
						if !strings.Contains(line, fragment) {
							t.Errorf("Codex file-count checkpoint missing workload decision %q", fragment)
						}
					}
					for _, forbidden := range []string{"escalar tier", "elevar el tier", "escalate tier"} {
						if strings.Contains(line, forbidden) {
							t.Errorf("Codex file-count checkpoint must not order tier escalation: %q", forbidden)
						}
					}
				}
				if !found {
					t.Error("Codex five-file workload checkpoint missing")
				}
			})
		}
	}
}
