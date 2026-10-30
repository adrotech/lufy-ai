package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
)

func TestWriteAtomicAndLoad(t *testing.T) {
	target := t.TempDir()
	want := New(target, nil, []AssetState{{ID: "AGENTS.md", SourceRel: "AGENTS.md.template", TargetRel: "AGENTS.md", SourceSHA256: "abc", TargetSHA256: "abc", LastAction: "copy"}}, "test-fingerprint")
	if err := WriteAtomic(target, want); err != nil {
		t.Fatalf("WriteAtomic() error = %v", err)
	}
	got, err := Load(target)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got == nil || got.SchemaVersion != SchemaVersion || got.ToolVersion == "" || got.SourceRootFingerprint != "test-fingerprint" || len(got.Assets) != 1 || got.Assets[0].TargetRel != "AGENTS.md" {
		t.Fatalf("Load() unexpected state: %#v", got)
	}
	if got.Assets[0].Policy != "managed" || got.Assets[0].Scope != "project" {
		t.Fatalf("Load() did not default policy/scope: %#v", got.Assets[0])
	}
	if got.Tool != domain.ToolCodex || got.Execution != domain.DefaultExecutionConfig() {
		t.Fatalf("Load() did not default harness metadata: %#v", got)
	}
	if got.Assets[0].Tool != string(domain.ToolCodex) || got.Assets[0].Component != "managed" {
		t.Fatalf("Load() did not default asset ownership: %#v", got.Assets[0])
	}
}

func TestLoadRejectsUnsupportedSchema(t *testing.T) {
	target := t.TempDir()
	writeStateFixture(t, target, `{"schemaVersion":99}`)
	if _, err := Load(target); err == nil {
		t.Fatal("Load() expected unsupported schema error")
	}
}

func TestLoadRejectsOldOrMixedStateWithoutWriting(t *testing.T) {
	for _, body := range []string{`{"schemaVersion":1}`, `{"schemaVersion":2}`, `{"schemaVersion":3,"tool":"codex","execution":{"mode":"inline","delegation":"auto"},"methodologyByTier":{}}`, `{"schemaVersion":3,"tool":"codex","execution":{"mode":"inline","delegation":"off","extra":true}}`} {
		target := t.TempDir()
		writeStateFixture(t, target, body)
		if _, err := Load(target); err == nil {
			t.Fatalf("accepted %s", body)
		}
		after, err := os.ReadFile(Path(target))
		if err != nil || string(after) != body {
			t.Fatal("state mutated")
		}
	}
}

func TestLoadRejectsUnsupportedOwnership(t *testing.T) {
	target := t.TempDir()
	body := `{"schemaVersion":4,"tool":"codex","execution":{"mode":"inline","delegation":"auto"},"assets":[{"targetRel":"AGENTS.md","policy":"managed","scope":"project","tool":"other"}]}`
	writeStateFixture(t, target, body)

	_, err := Load(target)
	if err == nil || !strings.Contains(err.Error(), "tool") {
		t.Fatalf("Load() error = %v, want tool error", err)
	}
}

func TestLoadRejectsUnknownPolicy(t *testing.T) {
	target := t.TempDir()
	body := `{"schemaVersion":4,"tool":"codex","execution":{"mode":"inline","delegation":"auto"},"assets":[{"targetRel":"AGENTS.md","policy":"surprise","scope":"project"}]}`
	writeStateFixture(t, target, body)

	_, err := Load(target)
	if err == nil || !strings.Contains(err.Error(), "policy") {
		t.Fatalf("Load() error = %v, want policy error", err)
	}
}

func TestLoadRejectsUnknownScope(t *testing.T) {
	target := t.TempDir()
	body := `{"schemaVersion":4,"tool":"codex","execution":{"mode":"inline","delegation":"auto"},"assets":[{"targetRel":"AGENTS.md","policy":"managed","scope":"elsewhere"}]}`
	writeStateFixture(t, target, body)

	_, err := Load(target)
	if err == nil || !strings.Contains(err.Error(), "scope") {
		t.Fatalf("Load() error = %v, want scope error", err)
	}
}

func writeStateFixture(t *testing.T, target, body string) {
	t.Helper()
	path := Path(target)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
