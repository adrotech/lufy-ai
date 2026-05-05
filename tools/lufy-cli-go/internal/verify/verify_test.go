package verify

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

func TestVerifyDetectsMissingAndHashMismatch(t *testing.T) {
	target := t.TempDir()
	for _, rel := range requiredAssets {
		writeVerifyFile(t, filepath.Join(target, rel), rel+"\n")
	}
	var states []state.AssetState
	for _, rel := range requiredAssets {
		hash, err := assets.FileSHA256(filepath.Join(target, rel))
		if err != nil {
			t.Fatal(err)
		}
		states = append(states, state.AssetState{ID: rel, SourceRel: rel, TargetRel: rel, SourceSHA256: hash, TargetSHA256: hash, LastAction: "copy"})
	}
	if err := state.WriteAtomic(target, state.New(target, nil, states)); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err != nil {
		t.Fatalf("Run(valid) error = %v, output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "ok: verify estructural completo") {
		t.Fatalf("valid verify output unexpected: %s", out.String())
	}

	writeVerifyFile(t, filepath.Join(target, "AGENTS.md"), "drift\n")
	out.Reset()
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(drift) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "fail: drift en AGENTS.md") {
		t.Fatalf("drift output unexpected: %s", out.String())
	}

	if err := os.Remove(filepath.Join(target, filepath.Join(".opencode", "commands", "opsx-apply.md"))); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(missing) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "fail: falta asset clave") {
		t.Fatalf("missing output unexpected: %s", out.String())
	}
}

func TestVerifyFailsCorruptManifestAndMovedTarget(t *testing.T) {
	target := t.TempDir()
	if err := os.MkdirAll(filepath.Join(target, ".lufy-ai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(state.Path(target), []byte("{bad-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "install-state.json inválido") {
		t.Fatalf("expected corrupt manifest error, got %v", err)
	}

	actual := t.TempDir()
	for _, rel := range requiredAssets {
		writeVerifyFile(t, filepath.Join(actual, rel), rel+"\n")
	}
	var states []state.AssetState
	for _, rel := range requiredAssets {
		hash, err := assets.FileSHA256(filepath.Join(actual, rel))
		if err != nil {
			t.Fatal(err)
		}
		states = append(states, state.AssetState{ID: rel, SourceRel: rel, TargetRel: rel, SourceSHA256: hash, TargetSHA256: hash, LastAction: "copy"})
	}
	if err := state.WriteAtomic(actual, state.New(t.TempDir(), nil, states)); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := NewService().Run(Options{Target: actual, NoEngram: true}, &out)
	if err == nil || !strings.Contains(out.String(), "targetRoot del manifest no coincide") {
		t.Fatalf("expected moved target failure, err=%v output=%s", err, out.String())
	}
}

func writeVerifyFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
