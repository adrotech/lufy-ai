package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveTargetPathResolvesExistingParent(t *testing.T) {
	root := t.TempDir()
	link := filepath.Join(t.TempDir(), "target-link")
	if err := os.Symlink(root, link); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}

	got, err := ResolveTargetPath(filepath.Join(link, "child"))
	if err != nil {
		t.Fatalf("ResolveTargetPath() error = %v", err)
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(resolvedRoot, "child")
	if got != want {
		t.Fatalf("ResolveTargetPath() = %q, want %q", got, want)
	}
}

func TestResolveTargetPathKeepsMissingPath(t *testing.T) {
	raw := filepath.Join(t.TempDir(), "missing", "child")
	got, err := ResolveTargetPath(raw)
	if err != nil {
		t.Fatalf("ResolveTargetPath() error = %v", err)
	}
	want, err := filepath.Abs(raw)
	if err != nil {
		t.Fatal(err)
	}
	want = filepath.Clean(want)
	if got != want {
		t.Fatalf("ResolveTargetPath() = %q, want %q", got, want)
	}
}
