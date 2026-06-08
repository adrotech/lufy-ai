package syncer

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/backup"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/installer"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/managedio"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/merger"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/verify"
)

func TestBuildPlanClassifiesSkipUpdateDriftAndUntracked(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	svc := NewService()

	skipTarget := installedTarget(t)
	skipPlan, err := svc.BuildPlan(Options{Target: skipTarget, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan(skip) error = %v", err)
	}
	if !hasSyncAction(skipPlan.Actions, "skip", "lufy-ia.harness.md") || hasSyncActionKind(skipPlan.Actions, "update-managed") {
		t.Fatalf("skip plan unexpected: actions=%#v conflicts=%#v", skipPlan.Actions, skipPlan.Conflicts)
	}

	updateTarget := installedTarget(t)
	writeFile(t, filepath.Join(source, "lufy-ia.harness.md"), "upstream changed\n")
	updatePlan, err := svc.BuildPlan(Options{Target: updateTarget, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan(update) error = %v", err)
	}
	if !hasSyncAction(updatePlan.Actions, "backup", "lufy-ia.harness.md") || !hasSyncAction(updatePlan.Actions, "update-managed", "lufy-ia.harness.md") {
		t.Fatalf("update plan missing backup/merge-block: actions=%#v conflicts=%#v", updatePlan.Actions, updatePlan.Conflicts)
	}

	driftTarget := installedTarget(t)
	writeFile(t, filepath.Join(driftTarget, "AGENTS.md"), "local drift\n")
	driftPlan, err := svc.BuildPlan(Options{Target: driftTarget, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan(drift) error = %v", err)
	}
	if len(driftPlan.Conflicts) != 0 || !hasSyncAction(driftPlan.Actions, "warn-agents-reference", "AGENTS.md") {
		t.Fatalf("expected agents reference warning, actions=%#v conflicts=%#v", driftPlan.Actions, driftPlan.Conflicts)
	}

	untrackedTarget := t.TempDir()
	writeFile(t, filepath.Join(untrackedTarget, "AGENTS.md"), "untracked\n")
	if err := state.WriteAtomic(untrackedTarget, state.New(untrackedTarget, nil, nil, "test-fingerprint")); err != nil {
		t.Fatal(err)
	}
	untrackedPlan, err := svc.BuildPlan(Options{Target: untrackedTarget, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan(untracked) error = %v", err)
	}
	if !hasSyncAction(untrackedPlan.Actions, "warn-agents-reference", "AGENTS.md") {
		t.Fatalf("expected untracked AGENTS warning, got actions=%#v conflicts=%#v", untrackedPlan.Actions, untrackedPlan.Conflicts)
	}
}

func TestPlanBuilderBuildsSyncPlanWithoutRewritingState(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	before := readFile(t, state.Path(target))

	plan, err := PlanBuilder{}.Build(Options{Target: target, NoEngram: true})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !hasSyncAction(plan.Actions, "skip", "lufy-ia.harness.md") {
		t.Fatalf("plan missing expected skip action: %#v", plan.Actions)
	}
	if !bytes.Equal(before, readFile(t, state.Path(target))) {
		t.Fatal("planner rewrote install-state")
	}
}

func TestRunSkipsPinnedAssetWithoutAdvancingManifest(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	before := st.AssetMap()["lufy-ia.harness.md"]
	before.Pinned = true
	before.PinnedReason = "local override"
	for i := range st.Assets {
		if st.Assets[i].TargetRel == before.TargetRel {
			st.Assets[i] = before
			break
		}
	}
	if err := state.WriteAtomic(target, *st); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(source, "lufy-ia.harness.md"), "upstream changed while pinned\n")
	targetBefore := readFile(t, filepath.Join(target, "lufy-ia.harness.md"))

	plan, err := NewService().BuildPlan(Options{Target: target, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	if !hasSyncAction(plan.Actions, "pinned-skip", "lufy-ia.harness.md") || hasSyncAction(plan.Actions, "update-managed", "lufy-ia.harness.md") {
		t.Fatalf("expected pinned skip without update: actions=%#v conflicts=%#v", plan.Actions, plan.Conflicts)
	}

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &out); err != nil {
		t.Fatalf("Run(sync pinned) error = %v, output=%s", err, out.String())
	}
	if got := readFile(t, filepath.Join(target, "lufy-ia.harness.md")); !bytes.Equal(got, targetBefore) {
		t.Fatalf("pinned target was mutated: before=%q after=%q", targetBefore, got)
	}
	after := stateMustLoad(t, target).AssetMap()["lufy-ia.harness.md"]
	if !after.Pinned || after.SourceSHA256 != before.SourceSHA256 || after.TargetSHA256 != before.TargetSHA256 || after.PinnedReason != "local override" {
		t.Fatalf("pinned manifest state advanced unexpectedly: before=%#v after=%#v", before, after)
	}
}

func TestBuildPlanWritesLufyNewForNoReplaceDriftWithSourceChange(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	writeFile(t, filepath.Join(target, "tui.json"), "{\"user\":true}\n")
	writeFile(t, filepath.Join(source, "tui.json"), "{\"upstream\":true}\n")

	plan, err := NewService().BuildPlan(Options{Target: target, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	if len(plan.Conflicts) != 0 || !hasSyncAction(plan.Actions, "write-lufy-new", "tui.json.lufy-new") {
		t.Fatalf("expected no-replace lufy-new action without conflicts, actions=%#v conflicts=%#v", plan.Actions, plan.Conflicts)
	}
	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(lufy-new) error = %v", err)
	}
	if got := string(readFile(t, filepath.Join(target, "tui.json"))); got != "{\"user\":true}\n" {
		t.Fatalf("no-replace original was overwritten: %q", got)
	}
	if got := string(readFile(t, filepath.Join(target, "tui.json.lufy-new"))); got != "{\"upstream\":true}\n" {
		t.Fatalf("lufy-new content mismatch: %q", got)
	}
}

func TestRunLufyNewCanBeResolvedByMergerAcceptTheirs(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	writeFile(t, filepath.Join(target, "tui.json"), "{\"user\":true}\n")
	writeFile(t, filepath.Join(source, "tui.json"), "{\"upstream\":true}\n")

	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(sync lufy-new) error = %v", err)
	}
	if got := string(readFile(t, filepath.Join(target, "tui.json.lufy-new"))); got != "{\"upstream\":true}\n" {
		t.Fatalf("sync lufy-new content mismatch: %q", got)
	}

	var out bytes.Buffer
	if err := merger.NewService().Run(merger.Options{Target: target, Path: "tui.json", AcceptTheirs: true}, &out); err != nil {
		t.Fatalf("merge accept-theirs error = %v", err)
	}
	if got := string(readFile(t, filepath.Join(target, "tui.json"))); got != "{\"upstream\":true}\n" {
		t.Fatalf("merge did not accept sync .lufy-new: %q", got)
	}
	if _, err := os.Stat(filepath.Join(target, "tui.json.lufy-new")); !os.IsNotExist(err) {
		t.Fatalf("merge did not remove sync .lufy-new, stat err=%v", err)
	}
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	asset := st.AssetMap()["tui.json"]
	if asset.LastAction != "merge-accept-theirs" || asset.TargetSHA256 != asset.SourceSHA256 || asset.AncestorHash != asset.TargetSHA256 {
		t.Fatalf("merge did not refresh synced asset state: %#v", asset)
	}
}

func TestRunRefreshesAncestorForSuccessfulUpdate(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	writeFile(t, filepath.Join(source, "lufy-ia.harness.md"), "upstream changed\n")

	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(sync) error = %v", err)
	}
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	harness := st.AssetMap()["lufy-ia.harness.md"]
	if harness.AncestorRel != ".lufy-ai/ancestors/lufy-ia.harness.md" || harness.AncestorHash != harness.SourceSHA256 {
		t.Fatalf("ancestor metadata not refreshed: %#v", harness)
	}
	if got := string(readFile(t, filepath.Join(target, harness.AncestorRel))); got != "upstream changed\n" {
		t.Fatalf("ancestor content mismatch: %q", got)
	}
}

func TestRenderMergeBlockForSync(t *testing.T) {
	source := t.TempDir()
	target := t.TempDir()
	writeFile(t, filepath.Join(source, "AGENTS.md.template"), "<!-- LUFY:BEGIN project-guide -->\nnew sync\n<!-- LUFY:END project-guide -->\n")
	writeFile(t, filepath.Join(target, "AGENTS.md"), "local\n<!-- LUFY:BEGIN project-guide -->\nold sync\n<!-- LUFY:END project-guide -->\n")

	merged, err := managedio.RenderMergeBlock(source, "AGENTS.md.template", target, "AGENTS.md")
	if err != nil {
		t.Fatalf("managedio.RenderMergeBlock() error = %v", err)
	}
	if got := string(merged); !strings.Contains(got, "local") || !strings.Contains(got, "new sync") || strings.Contains(got, "old sync") {
		t.Fatalf("unexpected merge-block output: %s", got)
	}
}

func TestSyncReadSourceAndWriteTargetRejectUnsafeFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "source.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := managedio.ReadSourceContent(root, "source.md"); err == nil {
		t.Fatalf("expected directory source to fail")
	}

	target := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.md")
	writeFile(t, outside, "outside\n")
	if err := os.Symlink(outside, filepath.Join(target, "dest.md")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	if err := managedio.WriteTargetFile(target, "dest.md", []byte("new\n")); err == nil {
		t.Fatalf("expected symlink target to fail")
	}
}

func TestDryRunPerformsNoMutations(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	projectConfigPath := filepath.Join(target, projectconfig.ProjectConfigPath)
	if err := os.Remove(projectConfigPath); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(source, "outside-source.txt"), "must not sync\n")
	writeFile(t, filepath.Join(target, "user-note.txt"), "preserve me\n")
	writeFile(t, filepath.Join(source, "lufy-ia.harness.md"), "upstream changed\n")

	stateBefore := readFile(t, state.Path(target))
	targetBefore := readFile(t, filepath.Join(target, "AGENTS.md"))
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, DryRun: true, Yes: true, NoEngram: true}, &out); err != nil {
		t.Fatalf("Run(dry-run) error = %v", err)
	}
	if !strings.Contains(out.String(), "Modo dry-run") || !strings.Contains(out.String(), "update-managed") {
		t.Fatalf("dry-run output unexpected: %s", out.String())
	}
	if !bytes.Equal(stateBefore, readFile(t, state.Path(target))) {
		t.Fatal("dry-run rewrote install-state")
	}
	if !bytes.Equal(targetBefore, readFile(t, filepath.Join(target, "AGENTS.md"))) {
		t.Fatal("dry-run rewrote managed file")
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy-ai", "backups")); !os.IsNotExist(err) {
		t.Fatalf("dry-run created backups dir, stat err=%v", err)
	}
	if _, err := os.Stat(projectConfigPath); !os.IsNotExist(err) {
		t.Fatalf("dry-run recreated project config, stat err=%v", err)
	}
}

func TestRunRequiresYesBeforeCreatingMissingProjectConfig(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	projectConfigPath := filepath.Join(target, projectconfig.ProjectConfigPath)
	if err := os.Remove(projectConfigPath); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(source, "lufy-ia.harness.md"), "upstream changed\n")

	err := NewService().Run(Options{Target: target, NoEngram: true}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "requiere --yes") {
		t.Fatalf("expected --yes error, got %v", err)
	}
	if _, err := os.Stat(projectConfigPath); !os.IsNotExist(err) {
		t.Fatalf("sync without --yes recreated project config, stat err=%v", err)
	}
}

func TestRunCreatesMissingProjectConfigAfterConfirmation(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	projectConfigPath := filepath.Join(target, projectconfig.ProjectConfigPath)
	if err := os.Remove(projectConfigPath); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer

	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &out); err != nil {
		t.Fatalf("Run(sync) error = %v, output=%s", err, out.String())
	}

	if _, err := os.Stat(projectConfigPath); err != nil {
		t.Fatalf("confirmed sync did not recreate project config: %v", err)
	}
	if !strings.Contains(out.String(), "- [project-config] "+projectconfig.ProjectConfigPath) {
		t.Fatalf("sync output missing project config action: %s", out.String())
	}
}

func TestRunCreatesSyncBackupManifestAndUpdatesState(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	writeFile(t, filepath.Join(source, "outside-source.txt"), "must not sync\n")
	writeFile(t, filepath.Join(target, "user-note.txt"), "preserve me\n")
	writeFile(t, filepath.Join(source, "lufy-ia.harness.md"), "upstream changed\n")

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &out); err != nil {
		t.Fatalf("Run(sync) error = %v, output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "[backup]") || !strings.Contains(out.String(), "[verify]") {
		t.Fatalf("sync output missing backup/verify: %s", out.String())
	}

	manifests, err := filepath.Glob(filepath.Join(target, ".lufy-ai", "backups", "*", "manifest.json"))
	if err != nil || len(manifests) != 1 {
		t.Fatalf("expected one backup manifest, manifests=%v err=%v", manifests, err)
	}
	var manifest backup.Manifest
	if err := json.Unmarshal(readFile(t, manifests[0]), &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Cause != "sync" || len(manifest.Files) == 0 || manifest.Files[0].Path != "lufy-ia.harness.md" || manifest.Files[0].Status != "captured" || manifest.Files[0].Cause != "sync" {
		t.Fatalf("manifest unexpected: %#v", manifest)
	}

	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	harness := st.AssetMap()["lufy-ia.harness.md"]
	current, err := assets.FileSHA256(filepath.Join(target, "lufy-ia.harness.md"))
	if err != nil {
		t.Fatal(err)
	}
	if harness.TargetSHA256 != current || harness.LastAction != "sync-update-managed" {
		t.Fatalf("state was not refreshed for harness: %#v current=%s", harness, current)
	}
	if st.ToolVersion == "" || st.SourceRootFingerprint == "" || st.SourceRootFingerprint == "dev-checkout" || st.SourceChangeID == "install-managed-assets-with-hash-idempotency" {
		t.Fatalf("sync state metadata not populated from runtime/catalog: %#v", st)
	}
	if _, ok := st.AssetMap()["outside-source.txt"]; ok {
		t.Fatal("sync registered source file outside managed catalog")
	}
	if _, err := os.Stat(filepath.Join(target, "outside-source.txt")); !os.IsNotExist(err) {
		t.Fatalf("sync copied source file outside managed catalog, stat err=%v", err)
	}
	if got := string(readFile(t, filepath.Join(target, "user-note.txt"))); got != "preserve me\n" {
		t.Fatalf("sync mutated target file outside managed catalog: %q", got)
	}

	infoBefore, err := os.Stat(filepath.Join(target, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(second sync) error = %v", err)
	}
	infoAfter, err := os.Stat(filepath.Join(target, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !infoBefore.ModTime().Equal(infoAfter.ModTime()) {
		t.Fatal("second sync rewrote unchanged managed file")
	}
}

func TestSyncWarnsWhenAgentsReferenceMissingWithoutMutatingAgents(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	agentsPath := filepath.Join(target, "AGENTS.md")
	writeFile(t, agentsPath, "# Proyecto\n\nConvenciones locales\n")
	before := readFile(t, agentsPath)

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &out); err != nil {
		t.Fatalf("Run(sync missing reference) error = %v, output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "warn-agents-reference") || !strings.Contains(out.String(), "@lufy-ia.harness.md") {
		t.Fatalf("sync output missing agents reference warning: %s", out.String())
	}
	if got := readFile(t, agentsPath); !bytes.Equal(got, before) {
		t.Fatalf("sync mutated AGENTS.md: before=%q after=%q", before, got)
	}
}

func TestSyncMigratesLegacyManagedAgentsStateNonDestructively(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	agentsPath := filepath.Join(target, "AGENTS.md")
	legacyAgents := []byte("legacy managed AGENTS\n")
	writeFile(t, agentsPath, string(legacyAgents))
	agentsHash, err := assets.FileSHA256(agentsPath)
	if err != nil {
		t.Fatal(err)
	}
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	st.Assets = append(st.Assets, state.AssetState{ID: "AGENTS.md", SourceRel: "AGENTS.md.template", TargetRel: "AGENTS.md", SourceSHA256: agentsHash, TargetSHA256: agentsHash, Policy: "merge-block", Scope: "project", LastAction: "legacy-managed"})
	if err := state.WriteAtomic(target, *st); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &out); err != nil {
		t.Fatalf("Run(sync legacy) error = %v, output=%s", err, out.String())
	}
	if got := readFile(t, agentsPath); !bytes.Equal(got, legacyAgents) {
		t.Fatalf("legacy AGENTS.md was mutated: %q", got)
	}
	after, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := after.AssetMap()["AGENTS.md"]; ok {
		t.Fatalf("legacy AGENTS.md entry remained in manifest: %#v", after.AssetMap()["AGENTS.md"])
	}
	if _, ok := after.AssetMap()["lufy-ia.harness.md"]; !ok {
		t.Fatal("harness missing from manifest after legacy migration")
	}
}

func TestSyncBlocksLegacyManagedAgentsDriftWithoutMutations(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	agentsPath := filepath.Join(target, "AGENTS.md")
	legacyAgents := []byte("legacy managed AGENTS\n")
	writeFile(t, agentsPath, string(legacyAgents))
	agentsHash, err := assets.FileSHA256(agentsPath)
	if err != nil {
		t.Fatal(err)
	}
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	st.Assets = append(st.Assets, state.AssetState{ID: "AGENTS.md", SourceRel: "AGENTS.md.template", TargetRel: "AGENTS.md", SourceSHA256: agentsHash, TargetSHA256: agentsHash, Policy: "merge-block", Scope: "project", LastAction: "legacy-managed"})
	if err := state.WriteAtomic(target, *st); err != nil {
		t.Fatal(err)
	}
	stateBefore := readFile(t, state.Path(target))
	writeFile(t, agentsPath, "legacy managed AGENTS\nlocal drift\n")
	agentsBefore := readFile(t, agentsPath)

	plan, err := NewService().BuildPlan(Options{Target: target, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	if !hasSyncConflict(plan.Conflicts, "AGENTS.md", "legacy gestionado tiene drift local") || hasSyncAction(plan.Actions, "warn-agents-reference", "AGENTS.md") {
		t.Fatalf("expected actionable legacy drift conflict, actions=%#v conflicts=%#v", plan.Actions, plan.Conflicts)
	}

	var out bytes.Buffer
	err = NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &out)
	if err == nil || !strings.Contains(err.Error(), "sync bloqueado") {
		t.Fatalf("Run(sync legacy drift) expected blocked conflict, err=%v output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "agrega la referencia manualmente") || !strings.Contains(out.String(), "recordedTarget") {
		t.Fatalf("sync output missing actionable legacy drift details: %s", out.String())
	}
	if got := readFile(t, agentsPath); !bytes.Equal(got, agentsBefore) {
		t.Fatalf("legacy drift AGENTS.md was mutated: before=%q after=%q", agentsBefore, got)
	}
	if got := readFile(t, state.Path(target)); !bytes.Equal(got, stateBefore) {
		t.Fatalf("blocked sync rewrote install-state: before=%q after=%q", stateBefore, got)
	}
}

func TestSyncRecoveryErrorRestoresBackup(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, "AGENTS.md"), "before\n")
	backupDir, err := backup.BackupFiles(target, []string{"AGENTS.md"}, "test-sync-rollback", &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(backupDir, "manifest.json")
	writeFile(t, filepath.Join(target, "AGENTS.md"), "after\n")

	err = syncRecoveryError(errors.New("boom"), target, manifestPath, 1)
	if err == nil || !strings.Contains(err.Error(), "rollback automático restauró 1") {
		t.Fatalf("unexpected recovery error: %v", err)
	}
	if got := string(readFile(t, filepath.Join(target, "AGENTS.md"))); got != "before\n" {
		t.Fatalf("rollback did not restore file: %q", got)
	}
}

func TestApplyRejectsUnknownActionKind(t *testing.T) {
	target := t.TempDir()
	plan := Plan{
		TargetRoot: target,
		Previous:   &state.InstallState{SchemaVersion: state.SchemaVersion},
		Actions:    []Action{{Kind: ActionKind("unknown-action"), Target: "x"}},
	}

	err := NewService().apply(plan, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "acción sync no soportada: unknown-action") {
		t.Fatalf("expected unknown action error, got %v", err)
	}
}

func TestRunKeepsRetiredManagedAssetsTracked(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	obsoleteRel := filepath.Join(".opencode", "commands", "opsx-apply.md")
	if err := os.Remove(filepath.Join(source, obsoleteRel)); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(source, "lufy-ia.harness.md"), "upstream changed\n")

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &out); err != nil {
		t.Fatalf("Run(sync) error = %v, output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "[retired] "+obsoleteRel) {
		t.Fatalf("sync output missing retired asset: %s", out.String())
	}
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	retired, ok := st.AssetMap()[obsoleteRel]
	if !ok {
		t.Fatalf("retired asset disappeared from install-state")
	}
	if retired.LastAction != "retired" {
		t.Fatalf("retired asset state unexpected: %#v", retired)
	}
	if _, err := os.Stat(filepath.Join(target, obsoleteRel)); err != nil {
		t.Fatalf("retired asset should remain in target: %v", err)
	}
}

func TestRunCreatesNewCatalogAssetsWhileUpdatingExistingManagedAssets(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	newRel := filepath.Join(".opencode", "commands", "new-command.md")
	writeFile(t, filepath.Join(source, newRel), "new command\n")
	writeFile(t, filepath.Join(source, "lufy-ia.harness.md"), "upstream changed\n")

	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(sync) error = %v", err)
	}
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := st.AssetMap()[newRel]; !ok || got.LastAction != "sync-update-managed" {
		t.Fatalf("new catalog asset was not registered as managed: %#v", got)
	}
	if got := string(readFile(t, filepath.Join(target, newRel))); got != "new command\n" {
		t.Fatalf("sync did not copy new catalog asset, got %q", got)
	}
	if st.AssetMap()["lufy-ia.harness.md"].LastAction != "sync-update-managed" {
		t.Fatalf("existing managed asset was not updated: %#v", st.AssetMap()["lufy-ia.harness.md"])
	}
}

func TestRunDefaultUpgradeDoesNotIntroduceLufySDDAndRemainsVerifiable(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	writeFile(t, filepath.Join(source, "lufy-ia.harness.md"), "upstream changed\n")
	writeFile(t, filepath.Join(source, ".lufy", "sdd", "README.md"), "new lufy-sdd should stay unselected\n")

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &out); err != nil {
		t.Fatalf("Run(default upgrade sync) error = %v, output=%s", err, out.String())
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy", "sdd")); !os.IsNotExist(err) {
		t.Fatalf("default sync should not create .lufy/sdd, stat err=%v", err)
	}

	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	if st.Tool != domain.ToolInitialDefault {
		t.Fatalf("synced tool = %s", st.Tool)
	}
	if got := st.MethodologyByTier[domain.TierT1]; got.ID != domain.MethodologySpecWorkflow || got.Mode != domain.MethodologyModeFull {
		t.Fatalf("synced T1 methodology = %#v", got)
	}
	if st.AssetMap()["lufy-ia.harness.md"].LastAction != "sync-update-managed" {
		t.Fatalf("existing harness asset was not updated: %#v", st.AssetMap()["lufy-ia.harness.md"])
	}
	if hasSyncedTargetPrefix(st, filepath.Join(".lufy", "sdd")) {
		t.Fatalf("default sync should not register lufy-sdd assets: %#v", st.Assets)
	}

	var verifyOut bytes.Buffer
	if err := verify.NewService().Run(verify.Options{Target: target, NoEngram: true}, &verifyOut); err != nil {
		t.Fatalf("verify after default sync error = %v, output=%s", err, verifyOut.String())
	}
	if !strings.Contains(verifyOut.String(), "ok: verify estructural completo") {
		t.Fatalf("verify after default sync output unexpected: %s", verifyOut.String())
	}
}

func TestRunMergeManagedOpenCodePreservesUnknownKeysAndStateExcludesHash(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	writeFile(t, filepath.Join(target, "opencode.json"), `{"custom":{"keep":true}}`)

	plan, err := NewService().BuildPlan(Options{Target: target, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	if !hasSyncAction(plan.Actions, "merge-json", "opencode.json") || hasSyncAction(plan.Actions, "update-managed", "opencode.json") {
		t.Fatalf("expected merge-json without update-managed for opencode.json: %#v", plan.Actions)
	}

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &out); err != nil {
		t.Fatalf("Run(sync) error = %v, output=%s", err, out.String())
	}
	decoded := readOpenCodeForSyncTest(t, target)
	if decoded["custom"] == nil || decoded["$schema"] == nil || decoded["plugin"] == nil {
		t.Fatalf("opencode.json merge-managed unexpected: %#v", decoded)
	}
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := st.AssetMap()["opencode.json"]; ok {
		t.Fatal("opencode.json no debe registrarse como asset gestionado completo por hash")
	}
}

func TestBuildPlanRejectsInvalidOpenCodeWithoutOverwrite(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	stateBefore := readFile(t, state.Path(target))
	writeFile(t, filepath.Join(target, "opencode.json"), `{bad-json`)

	_, err := NewService().BuildPlan(Options{Target: target, NoEngram: true})
	if err == nil || !strings.Contains(err.Error(), "opencode.json inválido") {
		t.Fatalf("expected invalid opencode.json error, got %v", err)
	}
	if got := string(readFile(t, filepath.Join(target, "opencode.json"))); got != `{bad-json` {
		t.Fatalf("invalid opencode.json was overwritten: %q", got)
	}
	if got := readFile(t, state.Path(target)); string(got) != string(stateBefore) {
		t.Fatal("sync BuildPlan rewrote state after invalid opencode.json")
	}
}

func TestBuildPlanRejectsInvalidOpenCodeManagedKeyTypesWithoutAssetUpdates(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	stateBefore := readFile(t, state.Path(target))
	writeFile(t, filepath.Join(source, "lufy-ia.harness.md"), "upstream changed but must not apply\n")
	writeFile(t, filepath.Join(target, "opencode.json"), `{"$schema":123,"plugin":{}}`)

	_, err := NewService().BuildPlan(Options{Target: target, NoEngram: true})
	if err == nil || !strings.Contains(err.Error(), "$schema debe ser string") {
		t.Fatalf("expected managed key type error, got %v", err)
	}
	if got := string(readFile(t, filepath.Join(target, "opencode.json"))); got != `{"$schema":123,"plugin":{}}` {
		t.Fatalf("invalid opencode.json was overwritten: %q", got)
	}
	if got := readFile(t, state.Path(target)); string(got) != string(stateBefore) {
		t.Fatal("sync BuildPlan rewrote state after invalid opencode.json structure")
	}
}

func TestBuildPlanRejectsTargetSymlinkEscape(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	if err := os.Remove(filepath.Join(target, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.md")
	writeFile(t, outside, "outside\n")
	if err := os.Symlink(outside, filepath.Join(target, "AGENTS.md")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	plan, err := NewService().BuildPlan(Options{Target: target, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	if !hasSyncConflict(plan.Conflicts, "AGENTS.md", "symlink no permitido") {
		t.Fatalf("expected symlink conflict, got %#v", plan.Conflicts)
	}
}

func TestRunRejectsMissingOrCorruptState(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	missingErr := NewService().Run(Options{Target: t.TempDir(), Yes: true, NoEngram: true}, &bytes.Buffer{})
	if missingErr == nil || !strings.Contains(missingErr.Error(), "sync requiere") {
		t.Fatalf("expected missing state error, got %v", missingErr)
	}

	target := t.TempDir()
	if err := os.MkdirAll(filepath.Join(target, ".lufy-ai"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, state.Path(target), "{not-json")
	corruptErr := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{})
	if corruptErr == nil || !strings.Contains(corruptErr.Error(), "install-state.json inválido") {
		t.Fatalf("expected corrupt state error, got %v", corruptErr)
	}
}

func installedTarget(t *testing.T) string {
	t.Helper()
	target := t.TempDir()
	if err := installer.NewService().Run(installer.Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("install fixture error = %v", err)
	}
	return target
}

func hasSyncAction(actions []Action, kind ActionKind, target string) bool {
	for _, action := range actions {
		if action.Kind == kind && action.Target == target {
			return true
		}
	}
	return false
}

func hasSyncActionKind(actions []Action, kind ActionKind) bool {
	for _, action := range actions {
		if action.Kind == kind {
			return true
		}
	}
	return false
}

func hasSyncConflict(conflicts []Conflict, path, reasonPart string) bool {
	for _, conflict := range conflicts {
		if conflict.Path == path && strings.Contains(conflict.Reason, reasonPart) {
			return true
		}
	}
	return false
}

func hasSyncedTargetPrefix(st *state.InstallState, prefix string) bool {
	normalizedPrefix := filepath.ToSlash(prefix)
	for _, asset := range st.Assets {
		target := filepath.ToSlash(asset.TargetRel)
		if target == normalizedPrefix || strings.HasPrefix(target, normalizedPrefix+"/") {
			return true
		}
	}
	return false
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

func minimalSource(t *testing.T) string {
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
		writeFile(t, filepath.Join(root, rel), content)
	}
	return root
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
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

func readOpenCodeForSyncTest(t *testing.T, target string) map[string]any {
	t.Helper()
	var decoded map[string]any
	if err := json.Unmarshal(readFile(t, filepath.Join(target, "opencode.json")), &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

func stateMustLoad(t *testing.T, target string) *state.InstallState {
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
