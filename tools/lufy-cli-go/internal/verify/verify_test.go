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
	writeVerifyDirs(t, target)
	for _, rel := range requiredManagedFiles {
		content := verifyFileContent(rel)
		writeVerifyFile(t, filepath.Join(target, rel), content)
	}
	var states []state.AssetState
	for _, rel := range requiredManagedFiles {
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

	if err := os.Remove(filepath.Join(target, filepath.Join(".opencode", "plugins", "agent-observatory.tsx"))); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(missing) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "fail: falta archivo crítico") {
		t.Fatalf("missing output unexpected: %s", out.String())
	}
}

func TestVerifyDetectsMissingCriticalDirectoryAndManifestEntry(t *testing.T) {
	target := t.TempDir()
	writeVerifyDirs(t, target)
	for _, rel := range requiredManagedFiles {
		writeVerifyFile(t, filepath.Join(target, rel), verifyFileContent(rel))
	}
	var states []state.AssetState
	for _, rel := range requiredManagedFiles {
		if rel == "tui.json" {
			continue
		}
		hash, err := assets.FileSHA256(filepath.Join(target, rel))
		if err != nil {
			t.Fatal(err)
		}
		states = append(states, state.AssetState{ID: rel, SourceRel: rel, TargetRel: rel, SourceSHA256: hash, TargetSHA256: hash, LastAction: "copy"})
	}
	if err := state.WriteAtomic(target, state.New(target, nil, states)); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(target, filepath.Join(".opencode", "skills"))); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(invalid structure) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "fail: falta directorio crítico: "+filepath.Join(".opencode", "skills")) {
		t.Fatalf("missing directory output unexpected: %s", out.String())
	}
	if !strings.Contains(out.String(), "fail: asset clave no está en manifest: tui.json") {
		t.Fatalf("missing manifest output unexpected: %s", out.String())
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
	writeVerifyDirs(t, actual)
	for _, rel := range requiredManagedFiles {
		content := verifyFileContent(rel)
		writeVerifyFile(t, filepath.Join(actual, rel), content)
	}
	var states []state.AssetState
	for _, rel := range requiredManagedFiles {
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

func TestVerifyFailsInvalidTUIJSON(t *testing.T) {
	target := t.TempDir()
	writeVerifyDirs(t, target)
	for _, rel := range requiredManagedFiles {
		writeVerifyFile(t, filepath.Join(target, rel), verifyFileContent(rel))
	}
	writeVerifyFile(t, filepath.Join(target, "tui.json"), "tui.json\n")

	var states []state.AssetState
	for _, rel := range requiredManagedFiles {
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
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(invalid tui.json) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "fail: JSON inválido en tui.json") {
		t.Fatalf("invalid tui.json output unexpected: %s", out.String())
	}
}

func TestVerifyFailsInvalidOrIncompleteMergeManagedOpenCode(t *testing.T) {
	target := validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, "opencode.json"), `{bad-json`)
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(invalid opencode.json) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "fail: JSON inválido en opencode.json") || !strings.Contains(out.String(), "estructura gestionada inválida") {
		t.Fatalf("invalid opencode.json output unexpected: %s", out.String())
	}

	target = validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, "opencode.json"), `{"$schema":"https://opencode.ai/config.json"}`)
	out.Reset()
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(incomplete opencode.json) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "falta clave gestionada mínima plugin") {
		t.Fatalf("incomplete opencode.json output unexpected: %s", out.String())
	}
}

func TestVerifyFailsUnsafeOrInvalidManagedOpenCodeStructure(t *testing.T) {
	target := validVerifyTarget(t)
	outside := filepath.Join(t.TempDir(), "opencode.json")
	writeVerifyFile(t, outside, `{"$schema":"https://opencode.ai/config.json","plugin":[]}`)
	if err := os.Remove(filepath.Join(target, "opencode.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(target, "opencode.json")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(symlink opencode.json) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "opencode.json") || !strings.Contains(out.String(), "symlink no permitido") {
		t.Fatalf("symlink opencode.json output unexpected: %s", out.String())
	}

	target = validVerifyTarget(t)
	if err := os.Remove(filepath.Join(target, "opencode.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(target, "opencode.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(directory opencode.json) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "archivo regular seguro") {
		t.Fatalf("directory opencode.json output unexpected: %s", out.String())
	}

	target = validVerifyTarget(t)
	writeVerifyFile(t, filepath.Join(target, "opencode.json"), `{"$schema":123,"plugin":[]}`)
	out.Reset()
	if err := NewService().Run(Options{Target: target, NoEngram: true}, &out); err == nil {
		t.Fatalf("Run(invalid managed type opencode.json) expected error, output=%s", out.String())
	}
	if !strings.Contains(out.String(), "$schema debe ser string") {
		t.Fatalf("invalid managed type output unexpected: %s", out.String())
	}
}

func validVerifyTarget(t *testing.T) string {
	t.Helper()
	target := t.TempDir()
	writeVerifyDirs(t, target)
	writeVerifyFile(t, filepath.Join(target, "opencode.json"), `{"$schema":"https://opencode.ai/config.json","plugin":[]}`)
	for _, rel := range requiredManagedFiles {
		writeVerifyFile(t, filepath.Join(target, rel), verifyFileContent(rel))
	}
	var states []state.AssetState
	for _, rel := range requiredManagedFiles {
		hash, err := assets.FileSHA256(filepath.Join(target, rel))
		if err != nil {
			t.Fatal(err)
		}
		states = append(states, state.AssetState{ID: rel, SourceRel: rel, TargetRel: rel, SourceSHA256: hash, TargetSHA256: hash, LastAction: "copy"})
	}
	if err := state.WriteAtomic(target, state.New(target, nil, states)); err != nil {
		t.Fatal(err)
	}
	return target
}

func writeVerifyDirs(t *testing.T, target string) {
	t.Helper()
	for _, rel := range requiredDirs {
		if err := os.MkdirAll(filepath.Join(target, rel), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeVerifyFile(t, filepath.Join(target, "opencode.json"), `{"$schema":"https://opencode.ai/config.json","plugin":[]}`)
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

func verifyFileContent(rel string) string {
	if rel == "tui.json" {
		return "{}\n"
	}
	return rel + "\n"
}
