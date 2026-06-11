package layout

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufypaths"
)

func TestRunMigratesLegacyArtifactsToUnifiedLayout(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.LegacyProjectConfig), "schema_version: 1\n")
	writeFile(t, filepath.Join(target, lufypaths.LegacyInstallState), "{}\n")
	writeFile(t, filepath.Join(target, lufypaths.LegacyBackups, "old", "manifest.json"), "{}\n")
	writeFile(t, filepath.Join(target, lufypaths.LegacyAncestors, "agent.md"), "ancestor\n")
	writeFile(t, filepath.Join(target, lufypaths.LegacyOpenSpecCache, "UPSTREAM.json"), "{}\n")
	writeFile(t, filepath.Join(target, lufypaths.LegacyLufySDD, "README.md"), "sdd\n")

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Yes: true}, &out); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{
		lufypaths.ProjectConfig,
		lufypaths.InstallState,
		filepath.Join(lufypaths.Backups, "old", "manifest.json"),
		filepath.Join(lufypaths.Ancestors, "agent.md"),
		filepath.Join(lufypaths.OpenSpecCache, "UPSTREAM.json"),
		filepath.Join(lufypaths.LufySDD, "README.md"),
		lufypaths.Readme,
	} {
		if _, err := os.Stat(filepath.Join(target, rel)); err != nil {
			t.Fatalf("missing migrated %s: %v", rel, err)
		}
	}
	if _, err := os.Stat(filepath.Join(target, lufypaths.LegacyInstallState)); err != nil {
		t.Fatalf("legacy install state should be preserved: %v", err)
	}
	if !strings.Contains(out.String(), "legacy layout migrated; .lufy-ai preserved for rollback") {
		t.Fatalf("migration summary missing: %s", out.String())
	}
}

func TestRunDryRunDoesNotWrite(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.LegacyInstallState), "{}\n")

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, DryRun: true}, &out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(target, lufypaths.InstallState)); !os.IsNotExist(err) {
		t.Fatalf("dry-run wrote canonical install state: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, lufypaths.Readme)); !os.IsNotExist(err) {
		t.Fatalf("dry-run wrote README: %v", err)
	}
}

func TestRunDryRunJSONIsMachineReadable(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.LegacyInstallState), "{}\n")

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, DryRun: true, JSON: true}, &out); err != nil {
		t.Fatal(err)
	}
	var report Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("invalid json output %q: %v", out.String(), err)
	}
	if report.Applied || len(report.Legacy) != 1 {
		t.Fatalf("report = %#v", report)
	}
	if strings.Contains(out.String(), "[dry-run]") {
		t.Fatalf("json output contains human log: %s", out.String())
	}
}

func TestRunJSONAfterApplyDoesNotMixHumanLogs(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.LegacyInstallState), "{}\n")

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Yes: true, JSON: true}, &out); err != nil {
		t.Fatal(err)
	}
	var report Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("invalid json output %q: %v", out.String(), err)
	}
	if !report.Applied {
		t.Fatalf("expected applied report: %#v", report)
	}
	if strings.Contains(out.String(), "[migrate-layout]") {
		t.Fatalf("json output contains human log: %s", out.String())
	}
}

func TestRunJSONReportsConflicts(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.InstallState), "{\"new\":true}\n")
	writeFile(t, filepath.Join(target, lufypaths.LegacyInstallState), "{\"legacy\":true}\n")

	var out bytes.Buffer
	err := NewService().Run(Options{Target: target, JSON: true}, &out)
	if err == nil || !strings.Contains(err.Error(), "conflicto") {
		t.Fatalf("expected conflict error, got %v", err)
	}
	var report Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("invalid json output %q: %v", out.String(), err)
	}
	if len(report.Conflicts) != 1 || report.Applied {
		t.Fatalf("report = %#v", report)
	}
}

func TestRunJSONReportsMissingYes(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.LegacyInstallState), "{}\n")

	var out bytes.Buffer
	err := NewService().Run(Options{Target: target, JSON: true}, &out)
	if err == nil || !strings.Contains(err.Error(), "requiere --yes") {
		t.Fatalf("expected --yes error, got %v", err)
	}
	var report Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("invalid json output %q: %v", out.String(), err)
	}
	if report.Applied || len(report.Actions) == 0 {
		t.Fatalf("report = %#v", report)
	}
}

func TestRunRequiresYesForMutations(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.LegacyInstallState), "{}\n")

	err := NewService().Run(Options{Target: target}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "requiere --yes") {
		t.Fatalf("expected --yes error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, lufypaths.InstallState)); !os.IsNotExist(err) {
		t.Fatalf("run without --yes wrote canonical install state: %v", err)
	}
}

func TestRunWithoutMutationsSucceedsWithoutYes(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.Readme), ReadmeContent())

	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Layout .lufy ya está actualizado") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestEnsureAppliesReadmeWhenMissing(t *testing.T) {
	target := t.TempDir()
	if err := Ensure(target, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(target, lufypaths.Readme))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "managed-state") {
		t.Fatalf("unexpected README: %s", string(body))
	}
}

func TestBuildPlanBlocksConflictingCanonicalAndLegacyContent(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.InstallState), "{\"new\":true}\n")
	writeFile(t, filepath.Join(target, lufypaths.LegacyInstallState), "{\"legacy\":true}\n")

	report, err := BuildPlan(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Conflicts) != 1 {
		t.Fatalf("conflicts = %#v", report.Conflicts)
	}
}

func TestBuildPlanBlocksConflictingDirectories(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.Backups, "one", "manifest.json"), "new\n")
	writeFile(t, filepath.Join(target, lufypaths.LegacyBackups, "one", "manifest.json"), "legacy\n")

	report, err := BuildPlan(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Conflicts) != 1 {
		t.Fatalf("conflicts = %#v", report.Conflicts)
	}
}

func TestBuildPlanMarksSameFileAsAlreadyMigrated(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.InstallState), "{}\n")
	writeFile(t, filepath.Join(target, lufypaths.LegacyInstallState), "{}\n")

	report, err := BuildPlan(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Conflicts) != 0 {
		t.Fatalf("unexpected conflicts = %#v", report.Conflicts)
	}
	if !hasAction(report.Actions, "legacy-stale", lufypaths.LegacyInstallState, lufypaths.InstallState) {
		t.Fatalf("missing stale install-state action: %#v", report.Actions)
	}
}

func TestBuildPlanMarksSameDirectoryAsAlreadyMigrated(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.Backups, "old", "manifest.json"), "{}\n")
	writeFile(t, filepath.Join(target, lufypaths.LegacyBackups, "old", "manifest.json"), "{}\n")

	report, err := BuildPlan(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Conflicts) != 0 {
		t.Fatalf("unexpected conflicts = %#v", report.Conflicts)
	}
	if !hasAction(report.Actions, "legacy-stale", lufypaths.LegacyBackups, lufypaths.Backups) {
		t.Fatalf("missing stale backups action: %#v", report.Actions)
	}
}

func TestBuildPlanTreatsCanonicalDirectorySupersetAsAlreadyMigrated(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.Backups, "old", "manifest.json"), "{}\n")
	writeFile(t, filepath.Join(target, lufypaths.Backups, "layout-migration", "manifest.json"), "{\"cause\":\"layout-migration\"}\n")
	writeFile(t, filepath.Join(target, lufypaths.LegacyBackups, "old", "manifest.json"), "{}\n")

	report, err := BuildPlan(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Conflicts) != 0 {
		t.Fatalf("unexpected conflicts = %#v", report.Conflicts)
	}
	if !hasAction(report.Actions, "legacy-stale", lufypaths.LegacyBackups, lufypaths.Backups) {
		t.Fatalf("missing stale backups action: %#v", report.Actions)
	}
}

func TestBuildPlanMergesNonOverlappingLegacyDirectoryEntries(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.Backups, "new", "manifest.json"), "new\n")
	writeFile(t, filepath.Join(target, lufypaths.LegacyBackups, "old", "manifest.json"), "old\n")

	report, err := BuildPlan(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Conflicts) != 0 {
		t.Fatalf("unexpected conflicts = %#v", report.Conflicts)
	}
	if !hasAction(report.Actions, "migrate-copy", lufypaths.LegacyBackups, lufypaths.Backups) {
		t.Fatalf("missing merge migrate-copy action: %#v", report.Actions)
	}
}

func TestRunMigrationIsIdempotentAfterBackingUpLegacyBackups(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.LegacyBackups, "old", "manifest.json"), "{}\n")

	if err := NewService().Run(Options{Target: target, Yes: true}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	report, err := BuildPlan(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Conflicts) != 0 {
		t.Fatalf("migration should be idempotent, conflicts = %#v", report.Conflicts)
	}
	if !hasAction(report.Actions, "legacy-stale", lufypaths.LegacyBackups, lufypaths.Backups) {
		t.Fatalf("missing stale backups action after migration: %#v", report.Actions)
	}
}

func TestApplyRejectsSymlinkSource(t *testing.T) {
	target := t.TempDir()
	dst := filepath.Join(target, "source.txt")
	writeFile(t, dst, "source\n")
	if err := os.Symlink(dst, filepath.Join(target, "legacy-link")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	_, err := Apply(target, []Action{{Kind: "migrate-copy", Source: "legacy-link", Target: "new-link"}}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("expected symlink error, got %v", err)
	}
}

func TestApplyRejectsNestedSymlinkInDirectory(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, "legacy-dir", "file.txt"), "source\n")
	if err := os.Symlink(filepath.Join(target, "legacy-dir", "file.txt"), filepath.Join(target, "legacy-dir", "link.txt")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	_, err := Apply(target, []Action{{Kind: "migrate-copy", Source: "legacy-dir", Target: "new-dir"}}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("expected symlink error, got %v", err)
	}
}

func TestSameContentHandlesMismatchedKindsAndSymlink(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, "file.txt"), "file\n")
	if err := os.MkdirAll(filepath.Join(target, "dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	same, err := sameContent(filepath.Join(target, "file.txt"), filepath.Join(target, "dir"))
	if err != nil {
		t.Fatal(err)
	}
	if same {
		t.Fatal("file and dir should not be same content")
	}
	if err := os.Symlink(filepath.Join(target, "file.txt"), filepath.Join(target, "link.txt")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	same, err = sameContent(filepath.Join(target, "link.txt"), filepath.Join(target, "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if same {
		t.Fatal("symlink should not compare as same regular content")
	}
}

func TestCopyPathCopiesRegularFileAndMissingSourceFails(t *testing.T) {
	target := t.TempDir()
	src := filepath.Join(target, "src.txt")
	dst := filepath.Join(target, "nested", "dst.txt")
	writeFile(t, src, "content\n")

	if err := copyPath(src, dst); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "content\n" {
		t.Fatalf("copied body = %q", string(body))
	}
	if err := copyPath(filepath.Join(target, "missing.txt"), filepath.Join(target, "x")); err == nil {
		t.Fatal("expected missing source error")
	}
}

func TestCopyDirCopiesNestedRegularFiles(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, "src", "nested", "file.txt"), "nested\n")

	if err := copyDir(filepath.Join(target, "src"), filepath.Join(target, "dst")); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(target, "dst", "nested", "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "nested\n" {
		t.Fatalf("copied body = %q", string(body))
	}
}

func TestApplyNoActions(t *testing.T) {
	applied, err := Apply(t.TempDir(), nil, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if applied {
		t.Fatal("expected no applied actions")
	}
}

func TestEnsureBlocksConflicts(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.InstallState), "{\"new\":true}\n")
	writeFile(t, filepath.Join(target, lufypaths.LegacyInstallState), "{\"legacy\":true}\n")

	err := Ensure(target, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "bloqueado") {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestBuildPlanPrefersCanonicalProjectConfigWhenBothExist(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, lufypaths.ProjectConfig), "new\n")
	writeFile(t, filepath.Join(target, lufypaths.LegacyProjectConfig), "legacy\n")

	report, err := BuildPlan(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Conflicts) != 0 {
		t.Fatalf("unexpected conflicts = %#v", report.Conflicts)
	}
	if !hasAction(report.Actions, "legacy-stale", lufypaths.LegacyProjectConfig, lufypaths.ProjectConfig) {
		t.Fatalf("missing stale project config action: %#v", report.Actions)
	}
}

func TestRunCreatesOnlyNewLayoutWhenNoLegacyExists(t *testing.T) {
	target := t.TempDir()

	if err := NewService().Run(Options{Target: target, Yes: true}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(target, lufypaths.Readme)); err != nil {
		t.Fatalf("README not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy-ai")); !os.IsNotExist(err) {
		t.Fatalf("created legacy root: %v", err)
	}
}

func hasAction(actions []Action, kind, source, target string) bool {
	for _, action := range actions {
		if action.Kind == kind && action.Source == source && action.Target == target {
			return true
		}
	}
	return false
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
