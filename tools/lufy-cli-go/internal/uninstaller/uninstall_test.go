package uninstaller

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/agentsref"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/installer"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/verify"
)

func TestRunDryRunAndRealUninstallThenReinstall(t *testing.T) {
	target := t.TempDir()
	if err := installer.NewService().Run(installer.Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("install fixture error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), []byte("# Project\n\nKeep me\n\n"+agentsref.Reference+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, DryRun: true}, &out); err != nil {
		t.Fatalf("dry-run uninstall error = %v", err)
	}
	if !strings.Contains(out.String(), "Modo dry-run") {
		t.Fatalf("dry-run output unexpected: %s", out.String())
	}
	if _, err := os.Stat(state.Path(target)); err != nil {
		t.Fatalf("dry-run removed install-state: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, agentsref.HarnessFile)); err != nil {
		t.Fatalf("dry-run removed harness: %v", err)
	}

	err := NewService().Run(Options{Target: target}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "requiere --yes") {
		t.Fatalf("uninstall without --yes expected confirmation error, got %v", err)
	}

	out.Reset()
	if err := NewService().Run(Options{Target: target, Yes: true}, &out); err != nil {
		t.Fatalf("real uninstall error = %v, output=%s", err, out.String())
	}
	if _, err := os.Stat(state.Path(target)); !os.IsNotExist(err) {
		t.Fatalf("install-state should be removed, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(target, agentsref.HarnessFile)); !os.IsNotExist(err) {
		t.Fatalf("harness should be removed, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "opencode.json")); err != nil {
		t.Fatalf("opencode.json should be preserved: %v", err)
	}
	agentsBody, err := os.ReadFile(filepath.Join(target, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(agentsBody), agentsref.Reference) || !strings.Contains(string(agentsBody), "Keep me") {
		t.Fatalf("AGENTS.md reference removal unexpected: %q", string(agentsBody))
	}
	if matches, err := filepath.Glob(filepath.Join(target, ".lufy", "managed-state", "backups", "*", "manifest.json")); err != nil || len(matches) == 0 {
		t.Fatalf("expected uninstall backup manifest, matches=%v err=%v", matches, err)
	}

	if err := installer.NewService().Run(installer.Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("reinstall after uninstall error = %v", err)
	}
	if err := verify.NewService().Run(verify.Options{Target: target, NoEngram: true, Quiet: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("verify after reinstall error = %v", err)
	}
}

func TestRunBlocksOnManagedAssetDrift(t *testing.T) {
	target := t.TempDir()
	if err := installer.NewService().Run(installer.Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("install fixture error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, agentsref.HarnessFile), []byte("local drift\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := NewService().Run(Options{Target: target, Yes: true}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "conflicto") {
		t.Fatalf("expected drift conflict, got %v", err)
	}
	if _, statErr := os.Stat(state.Path(target)); statErr != nil {
		t.Fatalf("state should remain after blocked uninstall: %v", statErr)
	}
	if got := string(readFile(t, filepath.Join(target, agentsref.HarnessFile))); got != "local drift\n" {
		t.Fatalf("drifted harness was changed: %q", got)
	}
}

func TestRunKeepStatePreservesInstallState(t *testing.T) {
	target := t.TempDir()
	if err := installer.NewService().Run(installer.Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("install fixture error = %v", err)
	}

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Yes: true, KeepState: true}, &out); err != nil {
		t.Fatalf("uninstall keep-state error = %v, output=%s", err, out.String())
	}
	if _, err := os.Stat(state.Path(target)); err != nil {
		t.Fatalf("install-state should be preserved by --keep-state: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, agentsref.HarnessFile)); !os.IsNotExist(err) {
		t.Fatalf("harness should still be removed by --keep-state, stat err=%v", err)
	}
}

func TestBuildPlanRequiresInstallState(t *testing.T) {
	_, err := NewService().BuildPlan(Options{Target: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "no hay instalación gestionada") {
		t.Fatalf("expected missing install-state error, got %v", err)
	}
}

func TestHelpersCoverRecoveryAndCleanupBranches(t *testing.T) {
	target := t.TempDir()
	if err := removeFile(target, "missing.txt"); err != nil {
		t.Fatalf("removeFile missing error = %v", err)
	}
	if err := removeEmptyDir(target, "missing-dir"); err != nil {
		t.Fatalf("removeEmptyDir missing error = %v", err)
	}
	if err := os.Mkdir(filepath.Join(target, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := removeEmptyDir(target, "empty"); err != nil {
		t.Fatalf("removeEmptyDir empty error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "empty")); !os.IsNotExist(err) {
		t.Fatalf("empty dir should be removed, stat err=%v", err)
	}

	baseErr := errors.New("boom")
	if got := uninstallRecoveryError(baseErr, ""); got != baseErr {
		t.Fatalf("uninstallRecoveryError without backup = %v", got)
	}
	withBackup := uninstallRecoveryError(baseErr, "/tmp/backup")
	if withBackup == nil || !strings.Contains(withBackup.Error(), "restore") {
		t.Fatalf("uninstallRecoveryError with backup unexpected: %v", withBackup)
	}
	if shortHash("abc") != "abc" || shortHash("1234567890abcdef") != "1234567890ab" {
		t.Fatalf("shortHash unexpected")
	}
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return body
}
