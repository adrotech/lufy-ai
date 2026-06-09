package merger

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
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

func TestRunRejectsUnmanagedAssetBeforeReadingConflictFiles(t *testing.T) {
	target := t.TempDir()
	if err := state.WriteAtomic(target, state.New(target, nil, []state.AssetState{{TargetRel: "other.json"}}, "test")); err != nil {
		t.Fatal(err)
	}

	err := NewService().Run(Options{Target: target, Path: "tui.json", AcceptTheirs: true}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "asset gestionado") {
		t.Fatalf("expected unmanaged asset error, got %v", err)
	}
}

func TestRunRejectsUnsafeOrMissingConflictFiles(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, target, ancestorRel string)
		want  string
	}{
		{
			name: "missing lufy-new",
			setup: func(t *testing.T, target, ancestorRel string) {
				write(t, filepath.Join(target, "tui.json"), "user\n")
				write(t, filepath.Join(target, ancestorRel), "old\n")
			},
			want: "lufy-new existente y seguro",
		},
		{
			name: "target is directory",
			setup: func(t *testing.T, target, ancestorRel string) {
				if err := os.Mkdir(filepath.Join(target, "tui.json"), 0o755); err != nil {
					t.Fatal(err)
				}
				write(t, filepath.Join(target, ancestorRel), "old\n")
				write(t, filepath.Join(target, "tui.json.lufy-new"), "new\n")
			},
			want: "target existente y seguro",
		},
		{
			name: "ancestor is symlink",
			setup: func(t *testing.T, target, ancestorRel string) {
				write(t, filepath.Join(target, "tui.json"), "user\n")
				write(t, filepath.Join(target, "tui.json.lufy-new"), "new\n")
				outside := filepath.Join(t.TempDir(), "ancestor.json")
				write(t, outside, "old\n")
				if err := os.MkdirAll(filepath.Dir(filepath.Join(target, ancestorRel)), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, filepath.Join(target, ancestorRel)); err != nil {
					t.Skipf("symlink no soportado en este entorno: %v", err)
				}
			},
			want: "symlink no permitido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := t.TempDir()
			ancestorRel := filepath.Join(".lufy-ai", "ancestors", "tui.json")
			tt.setup(t, target, ancestorRel)
			ancestorHash := ""
			if info, err := os.Lstat(filepath.Join(target, ancestorRel)); err == nil && info.Mode().IsRegular() {
				ancestorHash = hash(t, filepath.Join(target, ancestorRel))
			}
			if err := state.WriteAtomic(target, state.New(target, nil, []state.AssetState{{TargetRel: "tui.json", AncestorRel: ancestorRel, AncestorHash: ancestorHash}}, "test")); err != nil {
				t.Fatal(err)
			}

			err := NewService().Run(Options{Target: target, Path: "tui.json", AcceptTheirs: true}, &bytes.Buffer{})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q error, got %v", tt.want, err)
			}
		})
	}
}

func TestRunRequiresMergeToolWhenNoAcceptFlag(t *testing.T) {
	target, _ := writeManagedMergeFixture(t, "user\n", "old\n", "new\n")

	err := NewService().Run(Options{Target: target, Path: "tui.json"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "LUFY_MERGE_TOOL") {
		t.Fatalf("expected missing merge tool error, got %v", err)
	}
}

func TestAcceptResolutionRejectsStateWithoutManagedAsset(t *testing.T) {
	target, ancestorRel := writeManagedMergeFixture(t, "user\n", "old\n", "new\n")
	emptyState := state.New(target, nil, nil, "test")

	err := acceptResolution(target, "tui.json", filepath.Join(target, "tui.json"), filepath.Join(target, ancestorRel), filepath.Join(target, "tui.json.lufy-new"), true, &emptyState, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "asset gestionado") {
		t.Fatalf("expected missing managed asset error, got %v", err)
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

func TestRunAcceptTheirsUpdatesTargetStateAncestorAndRemovesSidecar(t *testing.T) {
	target := t.TempDir()
	write(t, filepath.Join(target, "tui.json"), "user\n")
	write(t, filepath.Join(target, "tui.json.lufy-new"), "new\n")
	ancestorRel := filepath.Join(".lufy-ai", "ancestors", "tui.json")
	write(t, filepath.Join(target, ancestorRel), "old\n")
	oldHash := hash(t, filepath.Join(target, ancestorRel))
	if err := state.WriteAtomic(target, state.New(target, nil, []state.AssetState{{ID: "tui.json", SourceRel: "tui.json", TargetRel: "tui.json", SourceSHA256: oldHash, TargetSHA256: hash(t, filepath.Join(target, "tui.json")), Policy: "no-replace", Scope: "project", AncestorRel: ancestorRel, AncestorHash: oldHash, LastAction: "write-lufy-new"}}, "test")); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Path: "tui.json", AcceptTheirs: true}, &out); err != nil {
		t.Fatal(err)
	}

	if got := string(read(t, filepath.Join(target, "tui.json"))); got != "new\n" {
		t.Fatalf("target = %q", got)
	}
	if got := string(read(t, filepath.Join(target, ancestorRel))); got != "new\n" {
		t.Fatalf("ancestor = %q", got)
	}
	if _, err := os.Stat(filepath.Join(target, "tui.json.lufy-new")); !os.IsNotExist(err) {
		t.Fatalf("sidecar still exists or unexpected stat error: %v", err)
	}
	after := loadState(t, target).AssetMap()["tui.json"]
	newHash := hash(t, filepath.Join(target, "tui.json"))
	if after.SourceSHA256 != newHash || after.TargetSHA256 != newHash || after.AncestorHash != newHash || after.LastAction != "merge-accept-theirs" {
		t.Fatalf("state not refreshed: %#v hash=%s", after, newHash)
	}
}

func TestRunAcceptOursPreservesTargetUpdatesStateAncestorAndRemovesSidecar(t *testing.T) {
	target := t.TempDir()
	write(t, filepath.Join(target, "tui.json"), "user\n")
	write(t, filepath.Join(target, "tui.json.lufy-new"), "new\n")
	ancestorRel := filepath.Join(".lufy-ai", "ancestors", "tui.json")
	write(t, filepath.Join(target, ancestorRel), "old\n")
	oldHash := hash(t, filepath.Join(target, ancestorRel))
	if err := state.WriteAtomic(target, state.New(target, nil, []state.AssetState{{ID: "tui.json", SourceRel: "tui.json", TargetRel: "tui.json", SourceSHA256: oldHash, TargetSHA256: oldHash, Policy: "no-replace", Scope: "project", AncestorRel: ancestorRel, AncestorHash: oldHash, LastAction: "write-lufy-new"}}, "test")); err != nil {
		t.Fatal(err)
	}

	if err := NewService().Run(Options{Target: target, Path: "tui.json", AcceptOurs: true}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	if got := string(read(t, filepath.Join(target, "tui.json"))); got != "user\n" {
		t.Fatalf("target = %q", got)
	}
	if got := string(read(t, filepath.Join(target, ancestorRel))); got != "user\n" {
		t.Fatalf("ancestor = %q", got)
	}
	if _, err := os.Stat(filepath.Join(target, "tui.json.lufy-new")); !os.IsNotExist(err) {
		t.Fatalf("sidecar still exists or unexpected stat error: %v", err)
	}
	after := loadState(t, target).AssetMap()["tui.json"]
	targetHash := hash(t, filepath.Join(target, "tui.json"))
	lufyNewHash := hashBytes([]byte("new\n"))
	if after.SourceSHA256 != lufyNewHash || after.TargetSHA256 != targetHash || after.AncestorHash != targetHash || after.LastAction != "merge-accept-ours" {
		t.Fatalf("state not refreshed: %#v targetHash=%s lufyNewHash=%s", after, targetHash, lufyNewHash)
	}
}

func TestRunAcceptFlagsAreRejectedTogetherBeforeMutatingFiles(t *testing.T) {
	target := t.TempDir()
	write(t, filepath.Join(target, "tui.json"), "user\n")
	write(t, filepath.Join(target, "tui.json.lufy-new"), "new\n")
	ancestorRel := filepath.Join(".lufy-ai", "ancestors", "tui.json")
	write(t, filepath.Join(target, ancestorRel), "old\n")
	oldHash := hash(t, filepath.Join(target, ancestorRel))
	if err := state.WriteAtomic(target, state.New(target, nil, []state.AssetState{{ID: "tui.json", SourceRel: "tui.json", TargetRel: "tui.json", SourceSHA256: oldHash, TargetSHA256: oldHash, Policy: "no-replace", Scope: "project", AncestorRel: ancestorRel, AncestorHash: oldHash, LastAction: "write-lufy-new"}}, "test")); err != nil {
		t.Fatal(err)
	}

	err := NewService().Run(Options{Target: target, Path: "tui.json", AcceptTheirs: true, AcceptOurs: true}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "no permite combinar") {
		t.Fatalf("expected conflicting accept flags error, got %v", err)
	}
	if got := string(read(t, filepath.Join(target, "tui.json"))); got != "user\n" {
		t.Fatalf("target mutated: %q", got)
	}
	if got := string(read(t, filepath.Join(target, ancestorRel))); got != "old\n" {
		t.Fatalf("ancestor mutated: %q", got)
	}
	if got := string(read(t, filepath.Join(target, "tui.json.lufy-new"))); got != "new\n" {
		t.Fatalf("sidecar mutated: %q", got)
	}
}

func TestRunAcceptTheirsUsesDefaultAncestorRelWhenStateOmitsIt(t *testing.T) {
	target := t.TempDir()
	write(t, filepath.Join(target, "tui.json"), "user\n")
	write(t, filepath.Join(target, "tui.json.lufy-new"), "new\n")
	ancestorRel := filepath.Join(".lufy-ai", "ancestors", "tui.json")
	write(t, filepath.Join(target, ancestorRel), "old\n")
	oldHash := hash(t, filepath.Join(target, ancestorRel))
	if err := state.WriteAtomic(target, state.New(target, nil, []state.AssetState{{ID: "tui.json", SourceRel: "tui.json", TargetRel: "tui.json", SourceSHA256: oldHash, TargetSHA256: oldHash, Policy: "no-replace", Scope: "project", AncestorHash: oldHash, LastAction: "write-lufy-new"}}, "test")); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Path: "tui.json", AcceptTheirs: true}, &out); err != nil {
		t.Fatal(err)
	}

	if got := string(read(t, filepath.Join(target, "tui.json"))); got != "new\n" {
		t.Fatalf("target = %q", got)
	}
	if got := string(read(t, filepath.Join(target, ancestorRel))); got != "new\n" {
		t.Fatalf("ancestor = %q", got)
	}
	if _, err := os.Stat(filepath.Join(target, "tui.json.lufy-new")); !os.IsNotExist(err) {
		t.Fatalf("sidecar still exists or unexpected stat error: %v", err)
	}
	if !strings.Contains(out.String(), "accept-theirs aplicado") {
		t.Fatalf("accept-theirs output unexpected: %s", out.String())
	}
}

func TestRunConfiguredToolSuccessFinalizesResolution(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("este smoke usa un script POSIX como LUFY_MERGE_TOOL")
	}
	target := t.TempDir()
	write(t, filepath.Join(target, "tui.json"), "user\n")
	write(t, filepath.Join(target, "tui.json.lufy-new"), "new\n")
	ancestorRel := filepath.Join(".lufy-ai", "ancestors", "tui.json")
	write(t, filepath.Join(target, ancestorRel), "old\n")
	ancestorHash := hash(t, filepath.Join(target, ancestorRel))
	if err := state.WriteAtomic(target, state.New(target, nil, []state.AssetState{{TargetRel: "tui.json", AncestorRel: ancestorRel, AncestorHash: ancestorHash, TargetSHA256: hash(t, filepath.Join(target, "tui.json"))}}, "test")); err != nil {
		t.Fatal(err)
	}
	toolPath := filepath.Join(t.TempDir(), "merge-tool.sh")
	if err := os.WriteFile(toolPath, []byte("#!/bin/sh\nprintf 'merged\\n' > \"$2\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LUFY_MERGE_TOOL", toolPath)

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Path: "tui.json"}, &out); err != nil {
		t.Fatal(err)
	}

	if got := string(read(t, filepath.Join(target, "tui.json"))); got != "merged\n" {
		t.Fatalf("target not finalized from external tool: %q", got)
	}
	if got := string(read(t, filepath.Join(target, ancestorRel))); got != "merged\n" {
		t.Fatalf("ancestor not finalized from external tool: %q", got)
	}
	if _, err := os.Stat(filepath.Join(target, "tui.json.lufy-new")); !os.IsNotExist(err) {
		t.Fatalf("sidecar still exists or unexpected stat error: %v", err)
	}
	after := loadState(t, target).AssetMap()["tui.json"]
	if after.SourceSHA256 != hashBytes([]byte("new\n")) || after.TargetSHA256 != hashBytes([]byte("merged\n")) || after.AncestorHash != after.TargetSHA256 || after.LastAction != "merge-tool" {
		t.Fatalf("state not finalized from external tool: %#v", after)
	}
	if !strings.Contains(out.String(), "Merge tool aplicado") {
		t.Fatalf("tool success output unexpected: %s", out.String())
	}
}

func writeManagedMergeFixture(t *testing.T, targetContent, ancestorContent, lufyNewContent string) (string, string) {
	t.Helper()
	target := t.TempDir()
	write(t, filepath.Join(target, "tui.json"), targetContent)
	write(t, filepath.Join(target, "tui.json.lufy-new"), lufyNewContent)
	ancestorRel := filepath.Join(".lufy-ai", "ancestors", "tui.json")
	write(t, filepath.Join(target, ancestorRel), ancestorContent)
	ancestorHash := hash(t, filepath.Join(target, ancestorRel))
	if err := state.WriteAtomic(target, state.New(target, nil, []state.AssetState{{ID: "tui.json", SourceRel: "tui.json", TargetRel: "tui.json", SourceSHA256: ancestorHash, TargetSHA256: hash(t, filepath.Join(target, "tui.json")), Policy: "no-replace", Scope: "project", AncestorRel: ancestorRel, AncestorHash: ancestorHash, LastAction: "write-lufy-new"}}, "test")); err != nil {
		t.Fatal(err)
	}
	return target, ancestorRel
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
	return hashBytes(read(t, path))
}

func hashBytes(body []byte) string {
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:])
}

func loadState(t *testing.T, target string) *state.InstallState {
	t.Helper()
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	if st == nil {
		t.Fatal("missing install-state")
	}
	return st
}
