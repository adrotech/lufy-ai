package installer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/agentsref"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/backup"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/managedio"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/skillregistry"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/verify"
)

func TestBuildPlanClassifiesCopySkipConflictAndUpdateManaged(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	svc := NewService()
	target := t.TempDir()

	copyPlan, err := svc.BuildPlan(Options{Target: target})
	if err != nil {
		t.Fatalf("BuildPlan(copy) error = %v", err)
	}
	if !hasAction(copyPlan.Actions, "copy", "lufy-ia.harness.md") || !hasAction(copyPlan.Actions, "agents-reference-create", "AGENTS.md") || !hasAction(copyPlan.Actions, "mkdir", ".codex") || !hasAction(copyPlan.Actions, "verify", copyPlan.TargetRoot) {
		t.Fatalf("copy plan missing expected actions: %#v", copyPlan.Actions)
	}

	if err := svc.Run(Options{Target: target, Yes: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(initial) error = %v", err)
	}
	skipPlan, err := svc.BuildPlan(Options{Target: target})
	if err != nil {
		t.Fatalf("BuildPlan(skip) error = %v", err)
	}
	if !hasAction(skipPlan.Actions, "agents-reference-skip", "AGENTS.md") || !hasAction(skipPlan.Actions, "skip", "lufy-ia.harness.md") || hasActionKind(skipPlan.Actions, "copy") {
		t.Fatalf("skip plan unexpected actions: %#v", skipPlan.Actions)
	}

	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), []byte("local drift\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	conflictPlan, err := svc.BuildPlan(Options{Target: target})
	if err != nil {
		t.Fatalf("BuildPlan(conflict) error = %v", err)
	}
	if len(conflictPlan.Conflicts) != 0 || !hasAction(conflictPlan.Actions, "backup", "AGENTS.md") || !hasAction(conflictPlan.Actions, "agents-reference-insert", "AGENTS.md") {
		t.Fatalf("expected AGENTS.md integration insertion, actions=%#v conflicts=%#v", conflictPlan.Actions, conflictPlan.Conflicts)
	}

	updatedTarget := t.TempDir()
	if err := svc.Run(Options{Target: updatedTarget, Yes: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(updated initial) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, "lufy-ia.harness.md"), []byte("upstream changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	updatePlan, err := svc.BuildPlan(Options{Target: updatedTarget})
	if err != nil {
		t.Fatalf("BuildPlan(update) error = %v", err)
	}
	if !hasAction(updatePlan.Actions, "backup", "lufy-ia.harness.md") || !hasAction(updatePlan.Actions, "update-managed", "lufy-ia.harness.md") {
		t.Fatalf("update plan missing backup/update-managed: %#v", updatePlan.Actions)
	}
}

func TestPlanBuilderBuildsPlanWithoutApplyingMutations(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()

	plan, err := PlanBuilder{}.Build(Options{Target: target})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if !hasAction(plan.Actions, "copy", "lufy-ia.harness.md") {
		t.Fatalf("plan missing expected copy action: %#v", plan.Actions)
	}
	if _, err := os.Stat(filepath.Join(target, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("planner mutated AGENTS.md, stat err=%v", err)
	}
	if _, err := os.Stat(state.Path(target)); !os.IsNotExist(err) {
		t.Fatalf("planner wrote install-state, stat err=%v", err)
	}
}

func TestInlineInstallPreservesValidationWithoutRoleInjection(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	cfg, err := projectconfig.Scan(target, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	cfg.Validation.AllowedCommands.Implementer = []string{"go test ./..."}
	if err := (projectconfig.ConfigStore{}).Write(projectconfig.Path(target), cfg); err != nil {
		t.Fatal(err)
	}
	if err := NewService().Run(Options{Target: target, Yes: true}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	got, err := projectconfig.Load(projectconfig.Path(target))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Validation.AllowedCommands.Implementer, cfg.Validation.AllowedCommands.Implementer) {
		t.Fatal("project validation changed")
	}
	if _, err := os.Stat(filepath.Join(target, ".codex", "agents")); !os.IsNotExist(err) {
		t.Fatal("role chain installed")
	}
}

func TestRunAdoptsExistingUnmanagedMergeBlockFile(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	customAgents := "# AGENTS.md\n\nCustom project guide\n"
	writeInstallerFile(t, filepath.Join(target, "AGENTS.md"), customAgents)

	plan, err := NewService().BuildPlan(Options{Target: target})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	if len(plan.Conflicts) != 0 || !hasAction(plan.Actions, "backup", "AGENTS.md") || !hasAction(plan.Actions, "agents-reference-insert", "AGENTS.md") {
		t.Fatalf("expected agents integration insertion without conflicts, actions=%#v conflicts=%#v", plan.Actions, plan.Conflicts)
	}

	if err := NewService().Run(Options{Target: target, Yes: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := string(readFileForTest(t, filepath.Join(target, "AGENTS.md"))); got != customAgents+"\n"+agentsref.ManagedBlock() {
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

func TestRunInstallsCodexManagedSurface(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()

	harness := domain.DefaultHarnessConfig()
	harness.Tool = domain.ToolCodex
	if err := NewService().Run(Options{Target: target, Yes: true, Harness: harness}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(codex) error = %v", err)
	}

	for _, rel := range []string{
		filepath.Join(".agents", "skills", "lufy-mem-search", "SKILL.md"),
		filepath.Join(".lufy", "contracts", "work.md"),
		filepath.Join(".codex", "config.toml"),
		filepath.Join(".codex", "lufy-agent-mapping.md"),
		filepath.Join(".codex", "hooks.json"),
	} {
		if _, err := os.Stat(filepath.Join(target, rel)); err != nil {
			t.Fatalf("codex install missing %s: %v", rel, err)
		}
	}
	for _, rel := range []string{".opencode", "opencode.json", "tui.json"} {
		if _, err := os.Stat(filepath.Join(target, rel)); !os.IsNotExist(err) {
			t.Fatalf("codex install should not create %s, stat err=%v", rel, err)
		}
	}
	if body := string(readFileForTest(t, filepath.Join(target, "AGENTS.md"))); !strings.Contains(body, agentsref.BeginMarker) {
		t.Fatalf("codex install missing managed AGENTS block: %q", body)
	}
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	if st.Tool != domain.ToolCodex {
		t.Fatalf("install state tool = %s", st.Tool)
	}
	cfg, err := projectconfig.Load(projectconfig.Path(target))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Tool != st.Tool || !reflect.DeepEqual(cfg.Execution, st.Execution) {
		t.Fatalf("project config and install state diverged: config=%#v state=%#v", cfg, st)
	}
	if _, ok := st.AssetMap()[filepath.Join(".codex", "config.toml")]; !ok {
		t.Fatalf("install state missing codex config asset")
	}
}

func TestInstallRollsBackHarnessConfigWhenApplyFails(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	if _, err := projectconfig.NewService().MergeHarnessSelection(target, domain.DefaultHarnessConfig()); err != nil {
		t.Fatal(err)
	}
	before := readFileForTest(t, projectconfig.Path(target))

	svc := NewService()
	svc.actionExecutor = failingInstallActionExecutor{err: errors.New("apply failed")}
	err := svc.Run(Options{Target: target, Yes: true, Harness: domain.HarnessConfig{Tool: domain.ToolCodex, Execution: domain.DefaultExecutionConfig()}}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "apply failed") {
		t.Fatalf("expected apply failure, got %v", err)
	}
	after := readFileForTest(t, projectconfig.Path(target))
	if !bytes.Equal(before, after) {
		t.Fatalf("project config was not rolled back\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

type failingInstallActionExecutor struct{ err error }

func (f failingInstallActionExecutor) Apply(Plan, io.Writer) error { return f.err }

func TestRunRecordsAncestorsForSuccessfulWrites(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	if err := NewService().Run(Options{Target: target, Yes: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(initial) error = %v", err)
	}
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	harness := st.AssetMap()["lufy-ia.harness.md"]
	if harness.AncestorRel != ".lufy/managed-state/ancestors/lufy-ia.harness.md" || harness.AncestorHash != harness.SourceSHA256 {
		t.Fatalf("ancestor metadata not recorded: %#v", harness)
	}
	if got := string(readFileForTest(t, filepath.Join(target, harness.AncestorRel))); got != "harness template\n" {
		t.Fatalf("ancestor content mismatch: %q", got)
	}
}

func TestRunDefaultInstallSelectsInlineOnly(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	if err := NewService().Run(Options{Target: target, Yes: true}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{".opencode", "openspec", ".lufy/sdd", ".codex/agents"} {
		if _, err := os.Stat(filepath.Join(target, rel)); !os.IsNotExist(err) {
			t.Fatalf("installed retired asset %s", rel)
		}
	}
	st := stateMustLoadForTest(t, target)
	if st.Tool != domain.ToolCodex || st.Execution != domain.DefaultExecutionConfig() {
		t.Fatalf("inline not selected: %#v", st)
	}
}

func TestRenderMergeBlockPreservesLocalTextAndUpdatesBlocks(t *testing.T) {
	source := t.TempDir()
	target := t.TempDir()
	writeInstallerFile(t, filepath.Join(source, "AGENTS.md.template"), "<!-- LUFY:BEGIN project-guide -->\nnew lufy\n<!-- LUFY:END project-guide -->\n")
	writeInstallerFile(t, filepath.Join(target, "AGENTS.md"), "local intro\n<!-- LUFY:BEGIN project-guide -->\nold lufy\n<!-- LUFY:END project-guide -->\nlocal outro\n")

	merged, err := managedio.RenderMergeBlock(source, "AGENTS.md.template", target, "AGENTS.md")
	if err != nil {
		t.Fatalf("managedio.RenderMergeBlock() error = %v", err)
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
	if _, err := managedio.ReadSourceContent(root, "source.md"); err == nil {
		t.Fatalf("expected directory source to fail")
	}

	target := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.md")
	writeInstallerFile(t, outside, "outside\n")
	if err := os.Symlink(outside, filepath.Join(target, "dest.md")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	if err := managedio.WriteTargetFile(target, "dest.md", []byte("new\n")); err == nil {
		t.Fatalf("expected symlink target to fail")
	}
}

func TestRunLufyNewDoesNotBackupOrRefreshAncestor(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	if err := NewService().Run(Options{Target: target, Yes: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(initial) error = %v", err)
	}
	before, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	writeInstallerFile(t, filepath.Join(target, "tui.json"), "{\"user\":true}\n")
	writeInstallerFile(t, filepath.Join(source, "tui.json"), "{\"upstream\":true}\n")
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Yes: true}, &out); err != nil {
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
	if err := svc.Run(Options{Target: target, Yes: true}, &bytes.Buffer{}); err != nil {
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
	if err := svc.Run(Options{Target: target, Yes: true}, &bytes.Buffer{}); err != nil {
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
	err := NewService().Run(Options{Target: target}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "requiere --yes") {
		t.Fatalf("expected --yes error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("install without --yes mutated target, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(target, projectconfig.ProjectConfigPath)); !os.IsNotExist(err) {
		t.Fatalf("install without --yes created project config, stat err=%v", err)
	}
}

func TestRunCreatesProjectConfigAfterConfirmation(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	var out bytes.Buffer

	if err := NewService().Run(Options{Target: target, Yes: true}, &out); err != nil {
		t.Fatalf("Run(confirmed install) error = %v, output=%s", err, out.String())
	}

	if _, err := os.Stat(filepath.Join(target, projectconfig.ProjectConfigPath)); err != nil {
		t.Fatalf("confirmed install did not create project config: %v", err)
	}
	if !strings.Contains(out.String(), "- [project-config] "+projectconfig.ProjectConfigPath) {
		t.Fatalf("install output missing project config action: %s", out.String())
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

	if err := NewService().Run(Options{Target: target, DryRun: true, Yes: true}, &out); err != nil {
		t.Fatalf("Run(dry-run) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, projectconfig.ProjectConfigPath)); !os.IsNotExist(err) {
		t.Fatalf("dry-run created project config, stat err=%v", err)
	}

	for _, want := range []string{
		"Plan de instalación para " + resolvedTarget,
		"Source root: ",
		"- [mkdir] .codex (directorio gestionado ausente)",
		"- [agents-reference-create] AGENTS.md (AGENTS.md user-owned ausente; se crea integración mínima al harness)",
		"- [copy] lufy-ia.harness.md (archivo gestionado ausente)",

		"Modo dry-run: sin mutaciones en filesystem",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("dry-run output missing %q:\n%s", want, out.String())
		}
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
	if err := svc.Run(Options{Target: target, Yes: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("Run(initial) error = %v", err)
	}
	var out bytes.Buffer
	if err := svc.Run(Options{Target: target, Yes: true, Backup: true}, &out); err != nil {
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
	if err := os.Symlink(outside, filepath.Join(target, ".codex")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	plan, err := NewService().BuildPlan(Options{Target: target})
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
	if err := os.MkdirAll(filepath.Join(target, ".lufy", "managed-state"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(state.Path(target), []byte("{not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := NewService().Run(Options{Target: target, Yes: true}, &bytes.Buffer{})
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

func TestApplyInstallRejectsUnknownActionKind(t *testing.T) {
	target := t.TempDir()
	plan := Plan{
		TargetRoot: target,
		Actions:    []Action{{Kind: ActionKind("unknown-action"), Target: "x"}},
	}

	err := NewService().applyInstall(plan, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "acción install no soportada: unknown-action") {
		t.Fatalf("expected unknown action error, got %v", err)
	}
}

func TestInstallAndVerifyIntegration(t *testing.T) {
	source := minimalInstallerSource(t)
	chdirForTest(t, source)
	target := t.TempDir()
	if err := NewService().Run(Options{Target: target, Yes: true}, &bytes.Buffer{}); err != nil {
		t.Fatalf("install error = %v", err)
	}
	var out bytes.Buffer
	if err := verify.NewService().Run(verify.Options{Target: target}, &out); err != nil {
		t.Fatalf("verify error = %v, output=%s", err, out.String())
	}
	if !strings.Contains(out.String(), "ok: verify estructural completo") {
		t.Fatalf("verify output unexpected: %s", out.String())
	}
	registry, err := skillregistry.NewService().Inspect(skillregistry.Options{Target: target, Tool: domain.ToolCodex})
	if err != nil || registry.Status != "ready" {
		t.Fatalf("install should ensure skill registry: report=%#v err=%v", registry, err)
	}
}

func TestInstallSkillRegistryFailureIsBestEffort(t *testing.T) {
	svc := NewService()
	svc.skillRegistry = failingSkillRegistry{}
	var out bytes.Buffer
	svc.ensureSkillRegistry(Plan{TargetRoot: "repo", Harness: domain.DefaultHarnessConfig()}, &out)
	if !strings.Contains(out.String(), "[warn] no se pudo asegurar skill registry") || !strings.Contains(out.String(), "skills ensure") {
		t.Fatalf("missing best-effort warning and recovery: %s", out.String())
	}
}

type failingSkillRegistry struct{}

func (failingSkillRegistry) Ensure(skillregistry.Options, io.Writer) error {
	return errors.New("registry unavailable")
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

func hasAction(actions []Action, kind ActionKind, target string) bool {
	for _, action := range actions {
		if action.Kind == kind && action.Target == target {
			return true
		}
	}
	return false
}

func hasActionKind(actions []Action, kind ActionKind) bool {
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

func hasInstalledTargetPrefix(st *state.InstallState, prefix string) bool {
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

func minimalInstallerSource(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		filepath.Join(".agents", "skills", "lufy-mem-search", "SKILL.md"): "---\nname: lufy-mem-search\ndescription: Buscar memoria\n---\nBuscar memoria\n",
		"AGENTS.md":                         "agents root\n",
		"AGENTS.md.template":                "<!-- LUFY:BEGIN project-guide -->\nagents template\n<!-- LUFY:END project-guide -->\n",
		"lufy-ia.harness.md":                "harness template\n",
		"tui.json":                          "{}\n",
		filepath.Join(".lufy", "README.md"): "layout\n",
		filepath.Join(".lufy", "contracts", "README.md"):                        "contracts\n",
		filepath.Join(".lufy", "contracts", "delivery.md"):                      "delivery\n",
		filepath.Join(".lufy", "contracts", "result-contract.md"):               "result\n",
		filepath.Join(".lufy", "contracts", "pr-review", "review-framework.md"): "review framework\n",
		filepath.Join(".lufy", "contracts", "pr-review", "report.html"):         "<!DOCTYPE html>\n",
		filepath.Join(".agents", "skills", "lufy-mem-search", "SKILL.md"):       "close skill\n",
		filepath.Join(".lufy", "contracts", "work.md"):                          "sdd skill\n",
		filepath.Join(".codex", "README.md"):                                    "codex readme\n",
		filepath.Join(".codex", "config.toml"):                                  "project_doc_max_bytes = 32768\n\n[features]\nmulti_agent = true\n",
		filepath.Join(".codex", "lufy-agent-mapping.md"):                        "agent_execution_mode\n",
		filepath.Join(".codex", "hooks.json"):                                   "name = \"implementer\"\n",
		filepath.Join(".codex", "rules", "lufy.rules"):                          "# rules\n",
		filepath.Join(".opencode", ".gitignore"):                                "node_modules\n",
		filepath.Join(".opencode", "README.md"):                                 "readme\n",
		filepath.Join(".opencode", "package.json"):                              "{}\n",
		filepath.Join(".opencode", "package-lock.json"):                         "{}\n",
		filepath.Join(".opencode", "agents", "orchestrator.md"):                 "orchestrator\n",
		filepath.Join(".opencode", "commands", "opsx-apply.md"):                 "apply\n",
		filepath.Join(".opencode", "commands", "lufy.mem-search.md"):            "memory search\n",
		filepath.Join(".opencode", "hooks", "format-dispatch.sh"):               "hook\n",
		filepath.Join(".opencode", "hooks", "memory-validate.sh"):               "memory hook\n",
		filepath.Join(".opencode", "skills", "sdd-workflow", "x.md"):            "skill\n",
		filepath.Join(".opencode", "skills", "lufy.mem-search", "SKILL.md"):     "memory skill\n",
		filepath.Join(".opencode", "templates", "sdd-lite.md"):                  "lite\n",
		filepath.Join(".opencode", "templates", "result-contract.md"):           "result\n",
		filepath.Join(".opencode", "templates", "memory-note.md"):               "memory note\n",
		filepath.Join(".opencode", "policies", "delivery.md"):                   "delivery\n",
		filepath.Join(".opencode", "plugins", "agent-observatory.tsx"):          "plugin\n",
		filepath.Join(".opencode", "agent-observatory", "state.ts"):             "state\n",
		filepath.Join("openspec", "config.yaml"):                                "config\n",
		filepath.Join("openspec", "UPSTREAM.json"):                              "{}\n",
		filepath.Join("openspec", "README.md"):                                  "openspec\n",
		filepath.Join("openspec", "specs", ".gitkeep"):                          "",
		filepath.Join(".lufy", "sdd", "README.md"):                              "lufy-sdd\n",
		filepath.Join(".lufy", "sdd", "changes", ".gitkeep"):                    "",
		filepath.Join(".lufy", "sdd", "decisions", ".gitkeep"):                  "",
		filepath.Join(".lufy", "sdd", "specs", ".gitkeep"):                      "",
		filepath.Join(".lufy", "sdd", "verification", ".gitkeep"):               "",
		filepath.Join("tools", "lufy-cli-go", "go.mod"):                         "module github.com/adrianrojas/lufy-ai/tools/lufy-cli-go\n",
	}
	c, e := assets.BuildEmbeddedCatalog()
	if e != nil {
		t.Fatal(e)
	}
	for _, rel := range []string{".codex/config.toml", ".codex/hooks.json", ".codex/rules/lufy.rules"} {
		b, e := assets.ReadSourceFile(c.SourceRoot, rel)
		if e != nil {
			t.Fatal(e)
		}
		files[filepath.FromSlash(rel)] = string(b)
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
