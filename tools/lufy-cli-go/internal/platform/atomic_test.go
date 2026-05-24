package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomicCreatesParentsAndSetsContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "file.txt")
	if err := WriteFileAtomic(path, []byte("hello"), 0o640); err != nil {
		t.Fatalf("write atomic: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "hello" {
		t.Fatalf("unexpected body: %q", string(body))
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o640 {
		t.Fatalf("unexpected permissions: %o", got)
	}
}

func TestWriteFileAtomicFailsWhenParentIsFile(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	if err := os.WriteFile(parent, []byte("not dir"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(filepath.Join(parent, "child.txt"), []byte("x"), 0o644); err == nil {
		t.Fatalf("expected parent file to fail")
	}
}
