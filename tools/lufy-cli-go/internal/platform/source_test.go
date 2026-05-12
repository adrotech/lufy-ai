package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureRelativeSafeRejectsTraversal(t *testing.T) {
	bad := []string{"", ".", "..", "../x", "..\\x", "safe/../../x", "safe\\..\\..\\x", "/tmp/x", "C:\\tmp\\x", "C:/tmp/x"}
	for _, path := range bad {
		if _, err := EnsureRelativeSafe(path); err == nil {
			t.Fatalf("EnsureRelativeSafe(%q) expected error", path)
		}
	}
	if got, err := EnsureRelativeSafe(".opencode/agents"); err != nil || got != filepath.Join(".opencode", "agents") {
		t.Fatalf("EnsureRelativeSafe() = %q, %v", got, err)
	}
	if got, err := EnsureRelativeSafe(".opencode\\agents"); err != nil || got != filepath.Join(".opencode", "agents") {
		t.Fatalf("EnsureRelativeSafe(backslash) = %q, %v", got, err)
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

func TestResolveSourceRootRequiresGoModuleMarker(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "AGENTS.md"), "# test\n")
	mkdir(t, filepath.Join(root, ".opencode"))
	mkdir(t, filepath.Join(root, "openspec"))
	writeFile(t, filepath.Join(root, "openspec", "config.yaml"), "specs: []\n")

	if _, err := ResolveSourceRoot(root); err == nil {
		t.Fatal("ResolveSourceRoot() expected error without tools/lufy-cli-go/go.mod marker")
	}

	writeFile(t, filepath.Join(root, "tools", "lufy-cli-go", "go.mod"), "module example.test/lufy\n")
	nested := filepath.Join(root, "tools", "lufy-cli-go", "cmd")
	mkdir(t, nested)
	got, err := ResolveSourceRoot(nested)
	if err != nil {
		t.Fatalf("ResolveSourceRoot() error = %v", err)
	}
	if got != root {
		t.Fatalf("ResolveSourceRoot() = %q, want %q", got, root)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
}
