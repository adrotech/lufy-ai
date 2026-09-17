package surfaceplan

import (
	"errors"
	"reflect"
	"testing"
)

func TestResolverSelectsMostSpecificRoot(t *testing.T) {
	project := sampleProject()

	resolution, err := (SurfaceResolverStrategy{}).Resolve(project, "auto", []string{"web/src/App.tsx"})

	if err != nil {
		t.Fatal(err)
	}
	if resolution.Mode != ModeSingle || resolution.PrimarySurface != "web" || !reflect.DeepEqual(ids(resolution.Surfaces), []string{"web"}) {
		t.Fatalf("unexpected resolution: %#v", resolution)
	}
}

func TestResolverComposesConnectedSurfaces(t *testing.T) {
	project := sampleProject()

	resolution, err := (SurfaceResolverStrategy{}).Resolve(project, "auto", []string{"web/src/App.tsx", "api/openapi.yaml"})

	if err != nil {
		t.Fatal(err)
	}
	if resolution.Mode != ModeComposed || resolution.PrimarySurface != "fullstack" {
		t.Fatalf("unexpected composition: %#v", resolution)
	}
	if len(resolution.Contracts) != 1 || resolution.Contracts[0].From != "api" || resolution.Contracts[0].To != "web" {
		t.Fatalf("unexpected contracts: %#v", resolution.Contracts)
	}
}

func TestResolverElevatesBackendContractChangeToFullstack(t *testing.T) {
	resolution, err := (SurfaceResolverStrategy{}).Resolve(sampleProject(), "auto", []string{"api/openapi.yaml"})

	if err != nil {
		t.Fatal(err)
	}
	if resolution.Mode != ModeComposed || resolution.PrimarySurface != "fullstack" {
		t.Fatalf("contract change should compose fullstack: %#v", resolution)
	}
	found := false
	for _, decision := range resolution.Decisions {
		found = found || decision.Code == "contract_surface_composition"
	}
	if !found {
		t.Fatalf("missing contract composition decision: %#v", resolution.Decisions)
	}
}

func TestResolverHonorsExplicitSurface(t *testing.T) {
	resolution, err := (SurfaceResolverStrategy{}).Resolve(sampleProject(), "backend", []string{"web/src/App.tsx"})

	if err != nil {
		t.Fatal(err)
	}
	if resolution.PrimarySurface != "api" || resolution.Decisions[0].Code != "explicit_surface" {
		t.Fatalf("unexpected explicit resolution: %#v", resolution)
	}
}

func TestResolverReturnsActionableAmbiguityWithoutComposition(t *testing.T) {
	project := sampleProject()
	project.Surfaces = project.Surfaces[:2]

	_, err := (SurfaceResolverStrategy{}).Resolve(project, "auto", []string{"README.md"})

	var ambiguous AmbiguousSurfaceError
	if !errors.As(err, &ambiguous) || !reflect.DeepEqual(ambiguous.Choices, []string{"api", "web"}) {
		t.Fatalf("expected ambiguity, got %#v", err)
	}
}

func sampleProject() ProjectDefinition {
	return ProjectDefinition{Surfaces: []SurfaceDefinition{
		{ID: "web", Type: "frontend", Roots: []string{"web", "."}, Stacks: []string{"typescript"}, ValidationExpectations: []string{"typecheck", "browser_check_when_ui_changes"}},
		{ID: "api", Type: "backend", Roots: []string{"api", "."}, Stacks: []string{"go"}, ValidationExpectations: []string{"unit_tests", "integration_tests_when_contract_changes"}},
		{ID: "fullstack", Type: "fullstack", Roots: []string{"."}, Connects: []string{"web", "api"}, ValidationExpectations: []string{"contract_tests_when_available", "e2e_smoke_when_flow_changes"}},
	}, Stacks: map[string]StackDefinition{
		"typescript": {ID: "typescript", StaticCommand: "tsc --noEmit", TestCommand: "vitest run"},
		"go":         {ID: "go", TestCommand: "go test ./..."},
	}}
}
