package managedio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderMergeBlockPreservesLocalTextAndUpdatesManagedBlock(t *testing.T) {
	source := t.TempDir()
	target := t.TempDir()
	writeTestFile(t, filepath.Join(source, "AGENTS.md.template"), "<!-- LUFY:BEGIN project-guide -->\nnew\n<!-- LUFY:END project-guide -->\n")
	writeTestFile(t, filepath.Join(target, "AGENTS.md"), "local\n<!-- LUFY:BEGIN project-guide -->\nold\n<!-- LUFY:END project-guide -->\n")

	merged, err := RenderMergeBlock(source, "AGENTS.md.template", target, "AGENTS.md")
	if err != nil {
		t.Fatalf("RenderMergeBlock() error = %v", err)
	}

	got := string(merged)
	if !strings.Contains(got, "local") || !strings.Contains(got, "new") || strings.Contains(got, "old") {
		t.Fatalf("unexpected merge result: %s", got)
	}
}

func TestReadSourceAndWriteTargetRejectUnsafeFiles(t *testing.T) {
	source := t.TempDir()
	if err := os.Mkdir(filepath.Join(source, "source.md"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := ReadSourceContent(source, "source.md"); err == nil {
		t.Fatal("expected directory source to fail")
	}

	target := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.md")
	writeTestFile(t, outside, "outside\n")
	if err := os.Symlink(outside, filepath.Join(target, "dest.md")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	if err := WriteTargetFile(target, "dest.md", []byte("new\n")); err == nil {
		t.Fatal("expected symlink target to fail")
	}
}

func TestPathHelpers(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "file.txt"), "content\n")

	if got := TrimLufyNewSuffix("tui.json.lufy-new"); got != "tui.json" {
		t.Fatalf("TrimLufyNewSuffix() = %q", got)
	}
	if got := UniqueTargets([]string{"a", "b", "a"}); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("UniqueTargets() = %#v", got)
	}
	if !FileExists(filepath.Join(root, "file.txt")) || FileExists(filepath.Join(root, "missing.txt")) {
		t.Fatal("FileExists() returned unexpected result")
	}
}

func writeTestFile(t *testing.T, path string, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
