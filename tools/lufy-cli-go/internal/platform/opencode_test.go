package platform

import (
	"path/filepath"
	"testing"
)

func TestResolveOpenCodeConfigRootUsesXDGConfigHome(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	t.Setenv("HOME", t.TempDir())
	got, err := ResolveOpenCodeConfigRoot()
	if err != nil {
		t.Fatal(err)
	}
	want, err := ResolveTargetPath(filepath.Join(xdg, "opencode"))
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("ResolveOpenCodeConfigRoot() = %q, want %q", got, want)
	}
}

func TestResolveOpenCodeConfigRootFallsBackToHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", home)
	got, err := ResolveOpenCodeConfigRoot()
	if err != nil {
		t.Fatal(err)
	}
	want, err := ResolveTargetPath(filepath.Join(home, ".config", "opencode"))
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("ResolveOpenCodeConfigRoot() = %q, want %q", got, want)
	}
}
