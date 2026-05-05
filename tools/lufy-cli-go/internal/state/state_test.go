package state

import "testing"

func TestWriteAtomicAndLoad(t *testing.T) {
	target := t.TempDir()
	want := New(target, nil, []AssetState{{ID: "AGENTS.md", SourceRel: "AGENTS.md.template", TargetRel: "AGENTS.md", SourceSHA256: "abc", TargetSHA256: "abc", LastAction: "copy"}})
	if err := WriteAtomic(target, want); err != nil {
		t.Fatalf("WriteAtomic() error = %v", err)
	}
	got, err := Load(target)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got == nil || got.SchemaVersion != SchemaVersion || len(got.Assets) != 1 || got.Assets[0].TargetRel != "AGENTS.md" {
		t.Fatalf("Load() unexpected state: %#v", got)
	}
}

func TestLoadRejectsUnsupportedSchema(t *testing.T) {
	target := t.TempDir()
	if err := WriteAtomic(target, InstallState{SchemaVersion: 99}); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(target); err == nil {
		t.Fatal("Load() expected unsupported schema error")
	}
}
