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
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/installer"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
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
	if !hasSyncAction(skipPlan.Actions, "skip", "AGENTS.md") || hasSyncActionKind(skipPlan.Actions, "update-managed") {
		t.Fatalf("skip plan unexpected: actions=%#v conflicts=%#v", skipPlan.Actions, skipPlan.Conflicts)
	}

	updateTarget := installedTarget(t)
	writeFile(t, filepath.Join(source, "AGENTS.md.template"), "upstream changed\n")
	updatePlan, err := svc.BuildPlan(Options{Target: updateTarget, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan(update) error = %v", err)
	}
	if !hasSyncAction(updatePlan.Actions, "backup", "AGENTS.md") || !hasSyncAction(updatePlan.Actions, "update-managed", "AGENTS.md") {
		t.Fatalf("update plan missing backup/update-managed: actions=%#v conflicts=%#v", updatePlan.Actions, updatePlan.Conflicts)
	}

	driftTarget := installedTarget(t)
	writeFile(t, filepath.Join(driftTarget, "AGENTS.md"), "local drift\n")
	driftPlan, err := svc.BuildPlan(Options{Target: driftTarget, NoEngram: true})
	if err != nil {
		t.Fatalf("BuildPlan(drift) error = %v", err)
	}
	if !hasSyncConflict(driftPlan.Conflicts, "AGENTS.md", "drift local") {
		t.Fatalf("expected drift conflict, got %#v", driftPlan.Conflicts)
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
	if !hasSyncConflict(untrackedPlan.Conflicts, "AGENTS.md", "no gestionado") {
		t.Fatalf("expected untracked conflict, got %#v", untrackedPlan.Conflicts)
	}
}

func TestDryRunPerformsNoMutations(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	writeFile(t, filepath.Join(source, "outside-source.txt"), "must not sync\n")
	writeFile(t, filepath.Join(target, "user-note.txt"), "preserve me\n")
	writeFile(t, filepath.Join(source, "AGENTS.md.template"), "upstream changed\n")

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
}

func TestRunCreatesSyncBackupManifestAndUpdatesState(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	writeFile(t, filepath.Join(source, "outside-source.txt"), "must not sync\n")
	writeFile(t, filepath.Join(target, "user-note.txt"), "preserve me\n")
	writeFile(t, filepath.Join(source, "AGENTS.md.template"), "upstream changed\n")

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
	if manifest.Cause != "sync" || len(manifest.Files) == 0 || manifest.Files[0].Path != "AGENTS.md" || manifest.Files[0].Status != "captured" || manifest.Files[0].Cause != "sync" {
		t.Fatalf("manifest unexpected: %#v", manifest)
	}

	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	agents := st.AssetMap()["AGENTS.md"]
	current, err := assets.FileSHA256(filepath.Join(target, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if agents.SourceSHA256 != current || agents.TargetSHA256 != current || agents.LastAction != "sync-update-managed" {
		t.Fatalf("state was not refreshed for AGENTS.md: %#v current=%s", agents, current)
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

func TestRunKeepsRetiredManagedAssetsTracked(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	obsoleteRel := filepath.Join(".opencode", "commands", "opsx-apply.md")
	if err := os.Remove(filepath.Join(source, obsoleteRel)); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(source, "AGENTS.md.template"), "upstream changed\n")

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

func TestRunSkipsNewCatalogAssetsWhileUpdatingExistingManagedAssets(t *testing.T) {
	source := minimalSource(t)
	chdirForTest(t, source)
	target := installedTarget(t)
	newRel := filepath.Join(".opencode", "commands", "new-command.md")
	writeFile(t, filepath.Join(source, newRel), "new command\n")
	writeFile(t, filepath.Join(source, "AGENTS.md.template"), "upstream changed\n")

	if err := NewService().Run(Options{Target: target, Yes: true, NoEngram: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(sync) error = %v", err)
	}
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := st.AssetMap()[newRel]; ok {
		t.Fatalf("new catalog asset absent from target should not be registered")
	}
	if _, err := os.Stat(filepath.Join(target, newRel)); !os.IsNotExist(err) {
		t.Fatalf("sync copied new catalog asset unexpectedly, stat err=%v", err)
	}
	if st.AssetMap()["AGENTS.md"].LastAction != "sync-update-managed" {
		t.Fatalf("existing managed asset was not updated: %#v", st.AssetMap()["AGENTS.md"])
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
	writeFile(t, filepath.Join(source, "AGENTS.md.template"), "upstream changed but must not apply\n")
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

func hasSyncAction(actions []Action, kind, target string) bool {
	for _, action := range actions {
		if action.Kind == kind && action.Target == target {
			return true
		}
	}
	return false
}

func hasSyncActionKind(actions []Action, kind string) bool {
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
		"AGENTS.md.template":                                           "agents template\n",
		"tui.json":                                                     "{}\n",
		filepath.Join(".opencode", ".gitignore"):                       "node_modules\n",
		filepath.Join(".opencode", "README.md"):                        "readme\n",
		filepath.Join(".opencode", "package.json"):                     "{}\n",
		filepath.Join(".opencode", "package-lock.json"):                "{}\n",
		filepath.Join(".opencode", "agents", "orchestrator.md"):        "orchestrator\n",
		filepath.Join(".opencode", "commands", "opsx-apply.md"):        "apply\n",
		filepath.Join(".opencode", "skills", "sdd-workflow", "x.md"):   "skill\n",
		filepath.Join(".opencode", "policies", "delivery.md"):          "delivery\n",
		filepath.Join(".opencode", "plugins", "agent-observatory.tsx"): "plugin\n",
		filepath.Join(".opencode", "agent-observatory", "state.ts"):    "state\n",
		filepath.Join("openspec", "config.yaml"):                       "config\n",
		filepath.Join("openspec", "README.md"):                         "openspec\n",
		filepath.Join("openspec", "specs", ".gitkeep"):                 "",
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
