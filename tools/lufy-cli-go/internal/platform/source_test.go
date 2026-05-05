package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureRelativeSafeRejectsTraversal(t *testing.T) {
	bad := []string{"", ".", "..", "../x", "/tmp/x"}
	for _, path := range bad {
		if _, err := EnsureRelativeSafe(path); err == nil {
			t.Fatalf("EnsureRelativeSafe(%q) expected error", path)
		}
	}
	if got, err := EnsureRelativeSafe(".opencode/agents"); err != nil || got != ".opencode/agents" {
		t.Fatalf("EnsureRelativeSafe() = %q, %v", got, err)
	}
}

func TestSafeJoinRejectsSymlinkParent(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "managed")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	if _, err := SafeJoin(root, filepath.Join("managed", "file.txt")); err == nil {
		t.Fatal("SafeJoin() expected symlink parent error")
	}
}
