package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAncestorPathAndAssetMap(t *testing.T) {
	target := t.TempDir()
	rel, err := AncestorRel(".opencode/agents/test-writer.md")
	if err != nil {
		t.Fatal(err)
	}
	if rel != ".lufy/managed-state/ancestors/.opencode/agents/test-writer.md" {
		t.Fatalf("unexpected ancestor rel: %s", rel)
	}
	path, err := AncestorPath(target, ".opencode/agents/test-writer.md")
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(target, rel) {
		t.Fatalf("unexpected ancestor path: %s", path)
	}

	st := New(target, nil, []AssetState{{ID: "a", TargetRel: "a.txt"}, {ID: "b", TargetRel: "b.txt"}}, "fingerprint")
	assets := st.AssetMap()
	if len(assets) != 2 || assets["a.txt"].ID != "a" || assets["b.txt"].ID != "b" {
		t.Fatalf("unexpected asset map: %#v", assets)
	}
}

func TestLoadMissingAndInvalidState(t *testing.T) {
	if st, err := Load(t.TempDir()); err != nil || st != nil {
		t.Fatalf("missing state = %#v, %v", st, err)
	}

	target := t.TempDir()
	if err := os.MkdirAll(filepath.Join(target, ".lufy", "managed-state"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(target), []byte("{bad"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(target); err == nil {
		t.Fatalf("invalid state should fail")
	}
}
