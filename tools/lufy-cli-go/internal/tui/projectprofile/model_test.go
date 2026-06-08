package projectprofile

import (
	"errors"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	tea "github.com/charmbracelet/bubbletea"
)

func TestModelChangesSelectedSurfaceTypeAndLens(t *testing.T) {
	model := NewModel(projectconfig.ProjectConfig{
		ProjectProfile: projectconfig.ProjectProfile{Surfaces: []projectconfig.ProjectSurface{{ID: "web", Type: "frontend", Roots: []string{"."}, Stacks: []string{"typescript"}}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result, err := updated.(Model).Result()
	if err != nil {
		t.Fatal(err)
	}
	if result.Surfaces[0].Type != "backend" {
		t.Fatalf("surface type = %s", result.Surfaces[0].Type)
	}
	if !contains(result.Surfaces[0].AgentLens.PrimaryConcerns, "domain_invariants") {
		t.Fatalf("backend lens not applied: %#v", result.Surfaces[0].AgentLens)
	}
	if result.Surfaces[0].Architecture.Preferred != "controller_service_repository" {
		t.Fatalf("backend architecture not applied: %#v", result.Surfaces[0].Architecture)
	}
	if !contains(result.Surfaces[0].Architecture.StructuralExpectations, "services_own_business_rules") {
		t.Fatalf("backend structural expectations not applied: %#v", result.Surfaces[0].Architecture)
	}
}

func TestModelCyclesSelectedArchitecture(t *testing.T) {
	model := NewModel(projectconfig.ProjectConfig{
		ProjectProfile: projectconfig.ProjectProfile{Surfaces: []projectconfig.ProjectSurface{{ID: "api", Type: "backend", Roots: []string{"api"}, Stacks: []string{"go"}}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result, err := updated.(Model).Result()
	if err != nil {
		t.Fatal(err)
	}
	if result.Surfaces[0].Architecture.Preferred != "clean_architecture" {
		t.Fatalf("architecture preferred = %#v", result.Surfaces[0].Architecture)
	}
	if !contains(result.Surfaces[0].Architecture.StructuralExpectations, "application_or_usecase_layer_has_business_flows") {
		t.Fatalf("architecture structural expectations not cycled = %#v", result.Surfaces[0].Architecture)
	}
}

func TestModelNavigatesMultipleSurfaces(t *testing.T) {
	model := NewModel(projectconfig.ProjectConfig{
		ProjectProfile: projectconfig.ProjectProfile{Surfaces: []projectconfig.ProjectSurface{
			{ID: "web", Type: "frontend", Roots: []string{"web"}, Stacks: []string{"typescript"}},
			{ID: "api", Type: "backend", Roots: []string{"api"}, Stacks: []string{"go"}},
		}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRight})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result, err := updated.(Model).Result()
	if err != nil {
		t.Fatal(err)
	}
	if result.Surfaces[0].Type != "frontend" {
		t.Fatalf("first surface should stay unchanged: %#v", result.Surfaces[0])
	}
	if result.Surfaces[1].Type != "fullstack" {
		t.Fatalf("second surface type = %s", result.Surfaces[1].Type)
	}
}

func TestModelToggleRejectsNoActiveSurface(t *testing.T) {
	model := NewModel(projectconfig.ProjectConfig{
		ProjectProfile: projectconfig.ProjectProfile{Surfaces: []projectconfig.ProjectSurface{{ID: "web", Type: "frontend"}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeySpace})
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_, err := updated.(Model).Result()
	if !errors.Is(err, ErrNoActiveSurface) {
		t.Fatalf("expected ErrNoActiveSurface, got %v", err)
	}
}

func TestModelCancelReturnsCancellationError(t *testing.T) {
	model := NewModel(projectconfig.ProjectConfig{
		ProjectProfile: projectconfig.ProjectProfile{Surfaces: []projectconfig.ProjectSurface{{ID: "web", Type: "frontend"}}},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	_, err := updated.(Model).Result()
	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("expected ErrCancelled, got %v", err)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
