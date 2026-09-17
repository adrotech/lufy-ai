package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/surfaceplan"
)

func TestRunPlanJSONResolvesFullstackContract(t *testing.T) {
	target := writePlanFixture(t)
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"plan", "--target", target, "--files", "web/src/App.tsx,api/openapi.yaml", "--json"}, Dependencies{Stdout: &out, Stderr: &errOut})

	if code != ExitOK {
		t.Fatalf("expected success, code=%d stderr=%s", code, errOut.String())
	}
	var plan surfaceplan.ExecutionPlan
	if err := json.Unmarshal(out.Bytes(), &plan); err != nil {
		t.Fatalf("invalid json %q: %v", out.String(), err)
	}
	if plan.SchemaVersion != surfaceplan.SchemaVersion || plan.Mode != surfaceplan.ModeComposed || plan.PrimarySurface != "fullstack" {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	if plan.Source != "files" {
		t.Fatalf("source = %q", plan.Source)
	}
}

func TestRunPlanHumanIncludesCapabilitiesAndEvidence(t *testing.T) {
	target := writePlanFixture(t)
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"plan", "--target", target, "--surface", "web", "--files", "web/src/App.tsx", "--capabilities", "realtime,offline"}, Dependencies{Stdout: &out, Stderr: &errOut})

	if code != ExitOK {
		t.Fatalf("expected success, code=%d stderr=%s", code, errOut.String())
	}
	for _, expected := range []string{"Superficie primaria: web", "realtime_frame_timing", "offline_operation", "evidencia:"} {
		if !strings.Contains(out.String(), expected) {
			t.Fatalf("output missing %q: %s", expected, out.String())
		}
	}
}

func TestRunPlanBackendContractAutomaticallyComposesFullstack(t *testing.T) {
	target := writePlanFixture(t)
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"plan", "--target", target, "--files", "api/openapi.yaml", "--json"}, Dependencies{Stdout: &out, Stderr: &errOut})

	if code != ExitOK {
		t.Fatalf("expected success, code=%d stderr=%s", code, errOut.String())
	}
	var plan surfaceplan.ExecutionPlan
	if err := json.Unmarshal(out.Bytes(), &plan); err != nil {
		t.Fatal(err)
	}
	if plan.Mode != surfaceplan.ModeComposed || plan.PrimarySurface != "fullstack" {
		t.Fatalf("backend contract should compose fullstack: %#v", plan)
	}
}

func TestRunPlanRejectsUnknownSurface(t *testing.T) {
	target := writePlanFixture(t)
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"plan", "--target", target, "--surface", "mobile", "--files", "web/src/App.tsx"}, Dependencies{Stdout: &out, Stderr: &errOut})

	if code != ExitRuntimeErr || !strings.Contains(errOut.String(), "no configurada") {
		t.Fatalf("expected actionable error, code=%d stderr=%s", code, errOut.String())
	}
}

func writePlanFixture(t *testing.T) string {
	t.Helper()
	target := t.TempDir()
	cfg := projectconfig.ProjectConfig{
		SchemaVersion: projectconfig.SchemaVersion,
		ProjectProfile: projectconfig.ProjectProfile{Surfaces: []projectconfig.ProjectSurface{
			{ID: "web", Type: "frontend", Roots: []string{"web", "."}, Stacks: []string{"typescript"}, AgentLens: projectconfig.AgentLens{ValidationExpectations: []string{"typecheck", "browser_check_when_ui_changes"}}},
			{ID: "api", Type: "backend", Roots: []string{"api", "."}, Stacks: []string{"go"}, AgentLens: projectconfig.AgentLens{ValidationExpectations: []string{"unit_tests", "integration_tests_when_contract_changes"}}},
			{ID: "fullstack", Type: "fullstack", Roots: []string{"."}, Connects: []string{"web", "api"}, AgentLens: projectconfig.AgentLens{ValidationExpectations: []string{"contract_tests_when_available", "e2e_smoke_when_flow_changes"}}},
		}},
		Stacks: []projectconfig.Stack{{ID: "typescript", StaticAnalysis: projectconfig.CommandConfig{Command: "tsc --noEmit"}}, {ID: "go", TestRunner: projectconfig.CommandConfig{Command: "go test ./..."}}},
	}
	content, err := projectconfig.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(target, projectconfig.ProjectConfigPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	return target
}
