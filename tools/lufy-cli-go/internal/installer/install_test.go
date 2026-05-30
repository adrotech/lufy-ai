package installer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/backup"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/merger"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/verify"
)

func TestBuildPlanClassifiesCopySkipConflictAndUpdateManaged(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	svc := NewService()
	target := t.TempDir()

	copyPlan, err := svc.BuildPlan(Options{Target: target, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan(copy) error = %v", err)
	}
	if !hasAction(copyPlan.Actions, "copy", "lufy-ia.harness.md") || !hasAction(copyPlan.Actions, "agents-reference-create", "AGENTS.md") || !hasAction(copyPlan.Actions, "mkdir", ".opencode") || !hasAction(copyPlan.Actions, "merge-json", "opencode.json") || !hasAction(copyPlan.Actions, "verify", copyPlan.TargetRoot) {
		t.Fatalf("copy plan missing expected actions: %#v", copyPlan.Actions)
	}

	if err := svc.Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(initial) error = %v", err)
	}
	skipPlan, err := svc.BuildPlan(Options{Target: target, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan(skip) error = %v", err)
	}
	if !hasAction(skipPlan.Actions, "agents-reference-skip", "AGENTS.md") || !hasAction(skipPlan.Actions, "skip", "lufy-ia.harness.md") || hasActionKind(skipPlan.Actions, "copy") {
		t.Fatalf("skip plan unexpected actions: %#v", skipPlan.Actions)
	}

	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), []byte("local drift\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	conflictPlan, err := svc.BuildPlan(Options{Target: target, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan(conflict) error = %v", err)
	}
	if len(conflictPlan.Conflicts) != 0 || !hasAction(conflictPlan.Actions, "backup", "AGENTS.md") || !hasAction(conflictPlan.Actions, "agents-reference-insert", "AGENTS.md") {
		t.Fatalf("expected AGENTS.md reference insertion, actions=%#v conflicts=%#v", conflictPlan.Actions, conflictPlan.Conflicts)
	}

	updatedTarget := t.TempDir()
	if err := svc.Run(Options{Target: updatedTarget, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(updated initial) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, "lufy-ia.harness.md"), []byte("upstream changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	updatePlan, err := svc.BuildPlan(Options{Target: updatedTarget, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan(update) error = %v", err)
	}
	if !hasAction(updatePlan.Actions, "backup", "lufy-ia.harness.md") || !hasAction(updatePlan.Actions, "update-managed", "lufy-ia.harness.md") {
		t.Fatalf("update plan missing backup/update-managed: %#v", updatePlan.Actions)
	}
}

func TestBuildPlanWritesLufyNewForNoReplaceDriftWithSourceChange(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	svc := NewService()
	if err := svc.Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(initial) error = %v", err)
	}
	writeInstallerFile(t, filepath.Join(target, "tui.json"), "{\"user\":true}\n")
	writeInstallerFile(t, filepath.Join(source, "tui.json"), "{\"upstream\":true}\n")

	plan, err := svc.BuildPlan(Options{Target: target, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	if len(plan.Conflicts) != 0 || !hasAction(plan.Actions, "write-lufy-new", "tui.json.lufy-new") {
		t.Fatalf("expected no-replace lufy-new action without conflicts, actions=%#v conflicts=%#v", plan.Actions, plan.Conflicts)
	}
	if err := svc.Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(lufy-new) error = %v", err)
	}
	if got := string(readFileForTest(t, filepath.Join(target, "tui.json"))); got != "{\"user\":true}\n" {
		t.Fatalf("no-replace original was overwritten: %q", got)
	}
	if got := string(readFileForTest(t, filepath.Join(target, "tui.json.lufy-new"))); got != "{\"upstream\":true}\n" {
		t.Fatalf("lufy-new content mismatch: %q", got)
	}
}

func TestRunLufyNewCanBeConsumedByMergeAcceptTheirs(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(initial) error = %v", err)
	}
	writeInstallerFile(t, filepath.Join(target, "tui.json"), "{\"user\":true}\n")
	writeInstallerFile(t, filepath.Join(source, "tui.json"), "{\"upstream\":true}\n")
	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(lufy-new) error = %v", err)
	}

	var out bytes.Buffer
	if err := merger.NewService().Run(merger.Options{Target: target, Path: "tui.json", AcceptTheirs: true}, &out); err != nil {
		t.Fatalf("merge accept-theirs error = %v", err)
	}
	if got := string(readFileForTest(t, filepath.Join(target, "tui.json"))); got != "{\"upstream\":true}\n" {
		t.Fatalf("merge did not accept upstream .lufy-new: %q", got)
	}
	if _, err := os.Stat(filepath.Join(target, "tui.json.lufy-new")); !os.IsNotExist(err) {
		t.Fatalf("merge did not remove lufy-new, stat err=%v", err)
	}
	if got := stateMustLoadForTest(t, target).AssetMap()["tui.json"].LastAction; got != "merge-accept-theirs" {
		t.Fatalf("merge did not update asset action, got %q", got)
	}
}

func TestRunAdoptsExistingUnmanagedMergeBlockFile(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	customAgents := "# AGENTS.md\n\nCustom project guide\n"
	writeInstallerFile(t, filepath.Join(target, "AGENTS.md"), customAgents)

	plan, err := NewService().BuildPlan(Options{Target: target, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	if len(plan.Conflicts) != 0 || !hasAction(plan.Actions, "backup", "AGENTS.md") || !hasAction(plan.Actions, "agents-reference-insert", "AGENTS.md") {
		t.Fatalf("expected agents reference insertion without conflicts, actions=%#v conflicts=%#v", plan.Actions, plan.Conflicts)
	}

	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := string(readFileForTest(t, filepath.Join(target, "AGENTS.md"))); got != customAgents+"\n@lufy-ia.harness.md\n" {
		t.Fatalf("custom AGENTS.md was rewritten: %q", got)
	}
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := st.AssetMap()["AGENTS.md"]; ok {
		t.Fatal("AGENTS.md no debe quedar registrado como asset completo")
	}
}

func TestRunRecordsAncestorsForSuccessfulWrites(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(initial) error = %v", err)
	}
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	harness := st.AssetMap()["lufy-ia.harness.md"]
	if harness.AncestorRel != ".lufy-ai/ancestors/lufy-ia.harness.md" || harness.AncestorHash != harness.SourceSHA256 {
		t.Fatalf("ancestor metadata not recorded: %#v", harness)
	}
	if got := string(readFileForTest(t, filepath.Join(target, harness.AncestorRel))); got != "harness template\n" {
		t.Fatalf("ancestor content mismatch: %q", got)
	}
}

func TestRenderMergeBlockPreservesLocalTextAndUpdatesBlocks(t *testing.T) {
	source := t.TempDir()
	target := t.TempDir()
	writeInstallerFile(t, filepath.Join(source, "AGENTS.md.template"), "<!-- LUFY:BEGIN project-guide -->\nnew lufy\n<!-- LUFY:END project-guide -->\n")
	writeInstallerFile(t, filepath.Join(target, "AGENTS.md"), "local intro\n<!-- LUFY:BEGIN project-guide -->\nold lufy\n<!-- LUFY:END project-guide -->\nlocal outro\n")

	merged, err := renderMergeBlock(source, "AGENTS.md.template", target, "AGENTS.md")
	if err != nil {
		t.Fatalf("renderMergeBlock() error = %v", err)
	}
	got := string(merged)
	for _, want := range []string{"local intro", "new lufy", "local outro"} {
		if !strings.Contains(got, want) {
			t.Fatalf("merged output missing %q: %s", want, got)
		}
	}
	if strings.Contains(got, "old lufy") {
		t.Fatalf("old block was not replaced: %s", got)
	}
}

func TestReadSourceAndWriteTargetRejectUnsafeFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "source.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := readSourceContent(root, "source.md"); err == nil {
		t.Fatalf("expected directory source to fail")
	}

	target := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.md")
	writeInstallerFile(t, outside, "outside\n")
	if err := os.Symlink(outside, filepath.Join(target, "dest.md")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	if err := writeTargetFile(target, "dest.md", []byte("new\n")); err == nil {
		t.Fatalf("expected symlink target to fail")
	}
}

func TestRunLufyNewDoesNotBackupOrRefreshAncestor(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(initial) error = %v", err)
	}
	before, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	writeInstallerFile(t, filepath.Join(target, "tui.json"), "{\"user\":true}\n")
	writeInstallerFile(t, filepath.Join(source, "tui.json"), "{\"upstream\":true}\n")
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &out); err != nil {
		t.Fatalf("Run(lufy-new) error = %v", err)
	}
	if strings.Contains(out.String(), "- [backup]") {
		t.Fatalf("lufy-new should not backup original target: %s", out.String())
	}
	after, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	if after.AssetMap()["tui.json"].AncestorHash != before.AssetMap()["tui.json"].AncestorHash {
		t.Fatalf("lufy-new refreshed ancestor unexpectedly: before=%#v after=%#v", before.AssetMap()["tui.json"], after.AssetMap()["tui.json"])
	}
}

func TestRunIsIdempotentAndDoesNotRewriteState(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	svc := NewService()
	if err := svc.Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(initial) error = %v", err)
	}
	before, err := os.ReadFile(state.Path(target))
	if err != nil {
		t.Fatal(err)
	}
	infoBefore, err := os.Stat(filepath.Join(target, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(second) error = %v", err)
	}
	after, err := os.ReadFile(state.Path(target))
	if err != nil {
		t.Fatal(err)
	}
	infoAfter, err := os.Stat(filepath.Join(target, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("install-state changed on idempotent run")
	}
	if !infoBefore.ModTime().Equal(infoAfter.ModTime()) {
		t.Fatal("managed file mtime changed on idempotent run")
	}
}

func TestRunRequiresYesForRealMutation(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	err := NewService().Run(Options{Target: target, NoEngram: true}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "requiere --yes") {
		t.Fatalf("expected --yes error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("install without --yes mutated target, stat err=%v", err)
	}
}

func TestInstallDryRunPlanOutputRegression(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	var out bytes.Buffer
	resolvedTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		t.Fatal(err)
	}

	if err := NewService().Run(Options{Target: target, DryRun: true, Yes: true, NoEngram: true}, &out); err != nil {
		t.Fatalf("Run(dry-run) error = %v", err)
	}

	for _, want := range []string{
		"Plan de instalación para " + resolvedTarget,
		"Source root: ",
		"- [mkdir] .opencode (directorio padre requerido)",
		"- [agents-reference-create] AGENTS.md (AGENTS.md user-owned ausente; se crea referencia mínima al harness)",
		"- [copy] lufy-ia.harness.md (archivo gestionado ausente)",
		"- [merge-json] opencode.json (configuración OpenCode gestionada con merge conservador)",
		"Engram: omitido por --no-engram",
		"Modo dry-run: sin mutaciones en filesystem",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("dry-run output missing %q:\n%s", want, out.String())
		}
	}
}

func TestInstallMergeManagedOpenCodePreservesUnknownKeysAndStateExcludesHash(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	writeInstallerFile(t, filepath.Join(target, "opencode.json"), `{"custom":{"keep":true},"provider":{"local":{"models":{"keep-me":{}}}}}`)

	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(install) error = %v", err)
	}
	decoded := readOpenCodeForTest(t, target)
	if decoded["custom"] == nil || decoded["provider"] == nil {
		t.Fatalf("opencode.json perdió claves desconocidas: %#v", decoded)
	}
	if decoded["$schema"] == nil || decoded["plugin"] == nil {
		t.Fatalf("opencode.json no contiene estructura gestionada mínima: %#v", decoded)
	}
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	if st.ToolVersion == "" || st.SourceRootFingerprint == "" || st.SourceRootFingerprint == "dev-checkout" || st.SourceChangeID == "install-managed-assets-with-hash-idempotency" {
		t.Fatalf("install state metadata not populated from runtime/catalog: %#v", st)
	}
	if _, ok := st.AssetMap()["opencode.json"]; ok {
		t.Fatal("opencode.json no debe registrarse como asset gestionado completo por hash")
	}

	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(reinstall) error = %v", err)
	}
	decoded = readOpenCodeForTest(t, target)
	if decoded["custom"] == nil || decoded["provider"] == nil {
		t.Fatalf("reinstall perdió claves desconocidas: %#v", decoded)
	}
}

func TestInstallRejectsInvalidOpenCodeWithoutOverwrite(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	writeInstallerFile(t, filepath.Join(target, "opencode.json"), `{bad-json`)

	err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "opencode.json inválido") {
		t.Fatalf("expected invalid opencode.json error, got %v", err)
	}
	if got := string(readFileForTest(t, filepath.Join(target, "opencode.json"))); got != `{bad-json` {
		t.Fatalf("invalid opencode.json was overwritten: %q", got)
	}
	if _, err := os.Stat(state.Path(target)); !os.IsNotExist(err) {
		t.Fatalf("install wrote state after invalid opencode.json, stat err=%v", err)
	}
}

func TestInstallRejectsUnsafeOpenCodePath(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	outside := filepath.Join(t.TempDir(), "opencode.json")
	writeInstallerFile(t, outside, `{"$schema":"https://opencode.ai/config.json","plugin":[]}`)
	if err := os.Symlink(outside, filepath.Join(target, "opencode.json")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}

	err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{})
	if err == nil || !unsafeOpenCodeInstallError(err) {
		t.Fatalf("expected unsafe opencode.json error, got %v", err)
	}
	if _, err := os.Stat(state.Path(target)); !os.IsNotExist(err) {
		t.Fatalf("install wrote state after unsafe opencode.json, stat err=%v", err)
	}
}

func TestInstallRejectsInvalidOpenCodeManagedKeyTypes(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	writeInstallerFile(t, filepath.Join(target, "opencode.json"), `{"$schema":false,"plugin":{}}`)

	err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "$schema debe ser string") {
		t.Fatalf("expected managed key type error, got %v", err)
	}
	if got := string(readFileForTest(t, filepath.Join(target, "opencode.json"))); got != `{"$schema":false,"plugin":{}}` {
		t.Fatalf("invalid opencode.json was overwritten: %q", got)
	}
	if _, err := os.Stat(state.Path(target)); !os.IsNotExist(err) {
		t.Fatalf("install wrote state after invalid opencode.json structure, stat err=%v", err)
	}
}

func unsafeOpenCodeInstallError(err error) bool {
	return strings.Contains(err.Error(), "archivo regular seguro") || strings.Contains(err.Error(), "symlink no permitido")
}

func TestBackupFlagCreatesExplicitBackupOnInstalledTarget(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	svc := NewService()
	if err := svc.Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(initial) error = %v", err)
	}
	var out bytes.Buffer
	if err := svc.Run(Options{Target: target, Yes: true, NoEngram: true, Backup: true}, &out); err != nil {
		t.Fatalf("Run(backup) error = %v", err)
	}
	if !strings.Contains(out.String(), "- [backup]") {
		t.Fatalf("backup flag did not create backup action: %s", out.String())
	}
}

func TestInstallRecoveryErrorRestoresBackup(t *testing.T) {
	target := t.TempDir()
	writeInstallerFile(t, filepath.Join(target, "AGENTS.md"), "before\n")
	backupDir, err := backup.BackupFiles(target, []string{"AGENTS.md"}, "test-install-rollback", &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	writeInstallerFile(t, filepath.Join(target, "AGENTS.md"), "after\n")

	err = installRecoveryError(errors.New("boom"), target, backupDir, 1)
	if err == nil || !strings.Contains(err.Error(), "rollback automático restauró 1") {
		t.Fatalf("unexpected recovery error: %v", err)
	}
	if got := string(readFileForTest(t, filepath.Join(target, "AGENTS.md"))); got != "before\n" {
		t.Fatalf("rollback did not restore file: %q", got)
	}
}

func TestBuildPlanConflictsOnTargetSymlinkParent(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(target, ".opencode")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	plan, err := NewService().BuildPlan(Options{Target: target, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	if len(plan.Conflicts) == 0 {
		t.Fatalf("expected symlink conflict, actions=%#v", plan.Actions)
	}
}

func TestRunRejectsCorruptState(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	if err := os.MkdirAll(filepath.Join(target, ".lufy-ai"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(state.Path(target), []byte("{not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "install-state.json inválido") {
		t.Fatalf("expected corrupt state error, got %v", err)
	}
}

func TestApplyInstallReportsRecoveryBackupOnPartialError(t *testing.T) {
	target := t.TempDir()
	writeInstallerFile(t, filepath.Join(target, "AGENTS.md"), "current\n")
	hash := fileHashForTest(t, filepath.Join(target, "AGENTS.md"))
	plan := Plan{
		SourceRoot: t.TempDir(),
		TargetRoot: target,
		Previous:   &state.InstallState{SchemaVersion: state.SchemaVersion, Assets: []state.AssetState{{ID: "AGENTS.md", TargetRel: "AGENTS.md", TargetSHA256: hash}}},
		Actions: []Action{
			{Kind: "backup", Target: "AGENTS.md"},
			{Kind: "update-managed", Source: "missing-template", Target: "AGENTS.md"},
		},
	}
	err := NewService().applyInstall(plan, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "backup de recovery") || !strings.Contains(err.Error(), "acciones aplicadas=0") {
		t.Fatalf("expected recovery error, got %v", err)
	}
}

func TestInstallAndVerifyIntegration(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("install error = %v", err)
	}
	var out bytes.Buffer
	if err := verify.NewService().Run(verify.Options{Target: target, NoEngram: true}, &out); err != nil {
		t.Fatalf("verify error = %v, output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "ok: verify estructural completo") {
		t.Fatalf("verify output unexpected: %s", out.String())
	}
}

func TestApplyInstallRemovesNewStateWhenPostVerifyFails(t *testing.T) {
	target := t.TempDir()
	plan := Plan{TargetRoot: target}
	err := NewService().applyInstall(plan, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "verify falló") {
		t.Fatalf("expected verify failure, got %v", err)
	}
	if _, statErr := os.Stat(state.Path(target)); !os.IsNotExist(statErr) {
		t.Fatalf("expected state rollback after verify failure, stat err=%v", statErr)
	}
}

func hasAction(actions []Action, kind, target string) bool {
	for _, action := range actions {
		if action.Kind == kind && action.Target == target {
			return true
		}
	}
	return false
}

func hasActionKind(actions []Action, kind string) bool {
	for _, action := range actions {
		if action.Kind == kind {
			return true
		}
	}
	return false
}

func writeInstallerFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fileHashForTest(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return hashBytesForTest(body)
}

func readOpenCodeForTest(t *testing.T, target string) map[string]any {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(target, "opencode.json"))
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

func readFileForTest(t *testing.T, path string) []byte {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func stateMustLoadForTest(t *testing.T, target string) *state.InstallState {
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

func hashBytesForTest(body []byte) string {
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:])
}

func chdirForTest(t *testing.T, dir string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
}

func minimalInstallerSource(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"AGENTS.md":                                                    "agents root\n",
		"AGENTS.md.template":                                           "<!-- LUFY:BEGIN project-guide -->\nagents template\n<!-- LUFY:END project-guide -->\n",
		"lufy-ia.harness.md":                                           "harness template\n",
		"tui.json":                                                     "{}\n",
		filepath.Join(".opencode", ".gitignore"):                       "node_modules\n",
		filepath.Join(".opencode", "README.md"):                        "readme\n",
		filepath.Join(".opencode", "package.json"):                     "{}\n",
		filepath.Join(".opencode", "package-lock.json"):                "{}\n",
		filepath.Join(".opencode", "agents", "orchestrator.md"):        "orchestrator\n",
		filepath.Join(".opencode", "commands", "opsx-apply.md"):        "apply\n",
		filepath.Join(".opencode", "hooks", "format-dispatch.sh"):      "hook\n",
		filepath.Join(".opencode", "skills", "sdd-workflow", "x.md"):   "skill\n",
		filepath.Join(".opencode", "templates", "sdd-lite.md"):         "lite\n",
		filepath.Join(".opencode", "templates", "result-contract.md"):  "result\n",
		filepath.Join(".opencode", "policies", "delivery.md"):          "delivery\n",
		filepath.Join(".opencode", "plugins", "agent-observatory.tsx"): "plugin\n",
		filepath.Join(".opencode", "agent-observatory", "state.ts"):    "state\n",
		filepath.Join("openspec", "config.yaml"):                       "config\n",
		filepath.Join("openspec", "UPSTREAM.json"):                     "{}\n",
		filepath.Join("openspec", "README.md"):                         "openspec\n",
		filepath.Join("openspec", "specs", ".gitkeep"):                 "",
		filepath.Join(".lufy", "sdd", "README.md"):                     "lufy-sdd\n",
		filepath.Join(".lufy", "sdd", "changes", ".gitkeep"):           "",
		filepath.Join(".lufy", "sdd", "decisions", ".gitkeep"):         "",
		filepath.Join(".lufy", "sdd", "specs", ".gitkeep"):             "",
		filepath.Join(".lufy", "sdd", "verification", ".gitkeep"):      "",
		filepath.Join("tools", "lufy-cli-go", "go.mod"):                "module github.com/adrianrojas/lufy-ai/tools/lufy-cli-go\n",
	}
	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
