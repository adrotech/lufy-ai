package surfaceplan

import (
	"reflect"
	"testing"
)

type fakeProjectSource struct{ project ProjectDefinition }

func (f fakeProjectSource) Load(string) (ProjectDefinition, error) { return f.project, nil }

type fakeChangeSource struct{ files []string }

func (f fakeChangeSource) ChangedFiles(_, _ string) ([]string, error) { return f.files, nil }

func TestServiceBuildsStablePlanFromChangePort(t *testing.T) {
	service := NewServiceWith(fakeProjectSource{project: sampleProject()}, fakeChangeSource{files: []string{"web/src/App.tsx", "web/src/App.tsx"}})

	plan, err := service.Build(Options{Target: t.TempDir(), Base: "origin/develop"})

	if err != nil {
		t.Fatal(err)
	}
	if plan.SchemaVersion != SchemaVersion || plan.Source != "git_diff" || plan.PrimarySurface != "web" || !reflect.DeepEqual(plan.ChangedFiles, []string{"web/src/App.tsx"}) {
		t.Fatalf("unexpected plan: %#v", plan)
	}
}

func TestServiceMarksExplicitScopeSource(t *testing.T) {
	service := NewServiceWith(fakeProjectSource{project: sampleProject()}, fakeChangeSource{})

	plan, err := service.Build(Options{Target: t.TempDir(), RequestedSurface: "web", Files: []string{"web/src/App.tsx"}})

	if err != nil {
		t.Fatal(err)
	}
	if plan.Source != "explicit" {
		t.Fatalf("source = %q", plan.Source)
	}
}

func TestServiceRejectsGitOptionAsBase(t *testing.T) {
	service := NewServiceWith(fakeProjectSource{project: sampleProject()}, fakeChangeSource{})

	if _, err := service.Build(Options{Target: t.TempDir(), Base: "--output=/tmp/x"}); err == nil {
		t.Fatal("expected git option base to fail")
	}
}

func TestServiceRejectsFilesOutsideTarget(t *testing.T) {
	service := NewServiceWith(fakeProjectSource{project: sampleProject()}, fakeChangeSource{})

	if _, err := service.Build(Options{Target: t.TempDir(), Files: []string{"../secret"}}); err == nil {
		t.Fatal("expected path traversal to fail")
	}
}
