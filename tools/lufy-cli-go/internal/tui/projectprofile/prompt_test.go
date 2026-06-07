package projectprofile

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
)

func TestPromptFallbackPreservesDetectedProfileWithoutTerminal(t *testing.T) {
	cfg := projectconfig.ProjectConfig{ProjectProfile: projectconfig.ProjectProfile{Surfaces: []projectconfig.ProjectSurface{{ID: "web", Type: "frontend"}}}}
	var out bytes.Buffer
	prompt := NewPrompt(Options{
		Input:      strings.NewReader(""),
		Output:     &out,
		IsTerminal: func(io.Reader, io.Writer) bool { return false },
	})

	profile, err := prompt(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if profile.Surfaces[0].Type != "frontend" {
		t.Fatalf("profile changed in fallback: %#v", profile)
	}
	if !strings.Contains(out.String(), "modo no interactivo") {
		t.Fatalf("fallback message missing: %s", out.String())
	}
}

func TestPromptUsesProgramRunnerWhenTerminal(t *testing.T) {
	cfg := projectconfig.ProjectConfig{ProjectProfile: projectconfig.ProjectProfile{Surfaces: []projectconfig.ProjectSurface{{ID: "web", Type: "frontend"}}}}
	var called bool
	prompt := NewPrompt(Options{
		Input:      strings.NewReader(""),
		Output:     &bytes.Buffer{},
		IsTerminal: func(io.Reader, io.Writer) bool { return true },
		RunProgram: func(model Model, _ io.Reader, _ io.Writer) (projectconfig.ProjectProfile, error) {
			called = true
			model.changeSelectedType(1)
			model.confirmed = true
			return model.Result()
		},
	})

	profile, err := prompt(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("program runner was not called")
	}
	if profile.Surfaces[0].Type != "backend" {
		t.Fatalf("runner result not returned: %#v", profile)
	}
}
