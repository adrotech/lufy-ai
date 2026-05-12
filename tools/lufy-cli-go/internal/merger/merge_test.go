package merger

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

func TestRunValidatesPrerequisitesBeforeTool(t *testing.T) {
	target := t.TempDir()
	if err := state.WriteAtomic(target, state.New(target, nil, []state.AssetState{{TargetRel: "tui.json", AncestorRel: filepath.Join(".lufy-ai", "ancestors", "tui.json")}}, "test")); err != nil {
		t.Fatal(err)
	}
	err := NewService().Run(Options{Target: target, Path: "tui.json"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "merge requiere") {
		t.Fatalf("expected prerequisite error, got %v", err)
	}
}

func TestRunInvokesConfiguredToolAndPreservesOnFailure(t *testing.T) {
	target := t.TempDir()
	write(t, filepath.Join(target, "tui.json"), "user\n")
	write(t, filepath.Join(target, "tui.json.lufy-new"), "new\n")
	write(t, filepath.Join(target, ".lufy-ai", "ancestors", "tui.json"), "old\n")
	ancestorRel := filepath.Join(".lufy-ai", "ancestors", "tui.json")
	ancestorHash := hash(t, filepath.Join(target, ancestorRel))
	if err := state.WriteAtomic(target, state.New(target, nil, []state.AssetState{{TargetRel: "tui.json", AncestorRel: ancestorRel, AncestorHash: ancestorHash}}, "test")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LUFY_MERGE_TOOL", "false")
	err := NewService().Run(Options{Target: target, Path: "tui.json"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "archivos preservados") {
		t.Fatalf("expected tool failure, got %v", err)
	}
	if got := string(read(t, filepath.Join(target, "tui.json"))); got != "user\n" {
		t.Fatalf("target mutated after tool failure: %q", got)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) []byte {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func hash(t *testing.T, path string) string {
	t.Helper()
	h := sha256.Sum256(read(t, path))
	return hex.EncodeToString(h[:])
}
