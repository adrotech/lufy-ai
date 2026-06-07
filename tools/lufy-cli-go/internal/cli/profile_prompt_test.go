package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
)

func TestPromptPrimarySurfaceUsesDefaultAndSelection(t *testing.T) {
	cfg := projectconfig.ProjectConfig{ProjectProfile: projectconfig.ProjectProfile{Surfaces: []projectconfig.ProjectSurface{{ID: "api", Type: "backend"}}}}
	var out bytes.Buffer
	selected, err := promptPrimarySurface(strings.NewReader("\n"), &out, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if selected != "backend" {
		t.Fatalf("default selected = %s", selected)
	}
	if !strings.Contains(out.String(), "Backend/API") {
		t.Fatalf("prompt output missing choices: %s", out.String())
	}

	out.Reset()
	selected, err = promptPrimarySurface(strings.NewReader("4\n"), &out, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if selected != "mobile" {
		t.Fatalf("explicit selected = %s", selected)
	}
}

func TestPromptPrimarySurfaceRejectsInvalidSelection(t *testing.T) {
	_, err := promptPrimarySurface(strings.NewReader("99\n"), &bytes.Buffer{}, projectconfig.ProjectConfig{})
	if err == nil {
		t.Fatal("expected invalid selection error")
	}
}

func TestApplyPrimarySurfaceTypeCreatesAndUpdatesLens(t *testing.T) {
	cfg := projectconfig.ProjectConfig{Stacks: []projectconfig.Stack{{ID: "typescript", Frameworks: []string{"react"}}}}
	profile := applyPrimarySurfaceType(cfg, "frontend")
	if len(profile.Surfaces) != 1 || profile.Surfaces[0].Type != "frontend" || !containsCLI(profile.Surfaces[0].AgentLens.PrimaryConcerns, "accessibility") {
		t.Fatalf("created profile unexpected: %#v", profile)
	}

	cfg.ProjectProfile = projectconfig.ProjectProfile{Surfaces: []projectconfig.ProjectSurface{{ID: "main", Type: "frontend"}}}
	profile = applyPrimarySurfaceType(cfg, "cli")
	if profile.Surfaces[0].Type != "cli" || !containsCLI(profile.Surfaces[0].AgentLens.PrimaryConcerns, "command_contracts") {
		t.Fatalf("updated profile unexpected: %#v", profile)
	}
}

func containsCLI(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
