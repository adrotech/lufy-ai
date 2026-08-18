package lufysdd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufypaths"
)

func TestNewValidateAndStatus(t *testing.T) {
	target := workflowFixture(t)
	service := NewService()
	report, err := service.New(target, "add-audit-log", "audit-log")
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if report.Schema != ReportSchema || report.Status != "proposed" {
		t.Fatalf("report = %#v", report)
	}
	overviewPath := filepath.Join(target, lufypaths.LufySDD, "changes", "add-audit-log", overviewFile)
	overview := mustRead(t, overviewPath)
	if !strings.Contains(overview, "Lufy SDD Change Overview") || !strings.Contains(overview, "Spec: audit-log") || strings.Contains(overview, "<script") || strings.Contains(overview, "<link") || strings.Contains(overview, "http://") || strings.Contains(overview, "https://") {
		t.Fatalf("overview full inválido: %s", overview)
	}
	if _, err := service.New(target, "add-audit-log", "audit-log"); err == nil {
		t.Fatal("expected existing change error")
	}
	validated, err := service.Validate(target, "add-audit-log", true)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if !validated.Valid() || validated.Status != "valid" {
		t.Fatalf("validated = %#v", validated)
	}
	firstRendered := mustRead(t, overviewPath)
	validatedAgain, err := service.Validate(target, "add-audit-log", true)
	if err != nil || validated.DeltaDigest == "" || validated.DeltaDigest != validatedAgain.DeltaDigest {
		t.Fatalf("digest no determinístico: first=%q second=%q err=%v", validated.DeltaDigest, validatedAgain.DeltaDigest, err)
	}
	if secondRendered := mustRead(t, overviewPath); firstRendered != secondRendered {
		t.Fatal("overview no determinístico")
	}
	beforeStatus := mustRead(t, overviewPath)
	status, err := service.Status(target, "add-audit-log")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status.Status != "in_progress" || status.Progress.Total != 3 {
		t.Fatalf("status = %#v", status)
	}
	if afterStatus := mustRead(t, overviewPath); beforeStatus != afterStatus {
		t.Fatal("status modificó el overview")
	}
}

func TestValidateRefreshesOverviewAfterMarkdownEdit(t *testing.T) {
	target := workflowFixture(t)
	service := NewService()
	if _, err := service.New(target, "refresh-overview", "overview"); err != nil {
		t.Fatal(err)
	}
	changeRoot := filepath.Join(target, lufypaths.LufySDD, "changes", "refresh-overview")
	proposalPath := filepath.Join(changeRoot, "proposal.md")
	updated := "# Proposal: refresh-overview\n\n## Why\n\nContenido renovado y visible.\n"
	if err := os.WriteFile(proposalPath, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Validate(target, "refresh-overview", false); err != nil {
		t.Fatal(err)
	}
	first := mustRead(t, filepath.Join(changeRoot, overviewFile))
	if !strings.Contains(first, "Contenido renovado y visible") {
		t.Fatalf("overview stale: %s", first)
	}
	if _, err := service.Validate(target, "refresh-overview", false); err != nil {
		t.Fatal(err)
	}
	if second := mustRead(t, filepath.Join(changeRoot, overviewFile)); first != second {
		t.Fatal("dos refresh sin cambios produjeron bytes distintos")
	}
}

func TestLiteLifecycleUsesBoundedArtifacts(t *testing.T) {
	target := workflowFixture(t)
	service := NewService()
	report, err := service.NewWithMode(target, "quick-fix", "", ModeLite)
	if err != nil || report.Mode != ModeLite {
		t.Fatalf("new lite report=%#v err=%v", report, err)
	}
	changeRoot := filepath.Join(target, lufypaths.LufySDD, "changes", "quick-fix")
	for _, rel := range []string{"design.md", "specs"} {
		if _, statErr := os.Lstat(filepath.Join(changeRoot, rel)); !os.IsNotExist(statErr) {
			t.Fatalf("lite creó artifact full %s: %v", rel, statErr)
		}
	}
	overview := mustRead(t, filepath.Join(changeRoot, overviewFile))
	if !strings.Contains(overview, "mode: lite") || strings.Contains(overview, "Spec:") || strings.Contains(overview, ">Design<") {
		t.Fatalf("overview lite inválido: %s", overview)
	}
	validated, err := service.Validate(target, "quick-fix", true)
	if err != nil || !validated.Valid() || validated.Mode != ModeLite {
		t.Fatalf("validate lite report=%#v err=%v", validated, err)
	}
	synced, err := service.Sync(target, "quick-fix")
	if err != nil || synced.Status != "not_applicable" || len(synced.Actions) != 0 {
		t.Fatalf("sync lite report=%#v err=%v", synced, err)
	}
	if entries, readErr := os.ReadDir(filepath.Join(target, lufypaths.LufySDD, "specs")); readErr != nil || len(entries) != 0 {
		t.Fatalf("sync lite mutó specs: entries=%v err=%v", entries, readErr)
	}
}

func TestLiteArchiveDoesNotRequireSyncDigest(t *testing.T) {
	target := workflowFixture(t)
	service := NewService()
	service.now = func() time.Time { return time.Date(2026, 8, 18, 10, 0, 0, 0, time.UTC) }
	if _, err := service.NewWithMode(target, "archive-lite", "", ModeLite); err != nil {
		t.Fatal(err)
	}
	changeRoot := filepath.Join(target, lufypaths.LufySDD, "changes", "archive-lite")
	if err := os.WriteFile(filepath.Join(changeRoot, "tasks.md"), []byte("# Tasks\n\n- [x] Implementar.\n- [x] Validar.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	evidenceRoot := filepath.Join(target, lufypaths.LufySDD, "verification", "archive-lite")
	if err := os.MkdirAll(evidenceRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(evidenceRoot, "report.md"), []byte("validated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status, err := service.Status(target, "archive-lite")
	if err != nil || status.Status != "completed" {
		t.Fatalf("status lite report=%#v err=%v", status, err)
	}
	report, err := service.Archive(target, "archive-lite")
	if err != nil || !report.Valid() || report.Status != "archived" {
		t.Fatalf("archive lite report=%#v err=%v", report, err)
	}
	destination := filepath.Join(target, lufypaths.LufySDD, "archive", "2026-08-18-archive-lite")
	if !strings.Contains(mustRead(t, filepath.Join(destination, overviewFile)), "mode: lite") {
		t.Fatal("archive lite no preservó overview")
	}
}

func TestValidateRejectsSymlinkOverview(t *testing.T) {
	target := workflowFixture(t)
	service := NewService()
	if _, err := service.New(target, "linked-overview", "audit-log"); err != nil {
		t.Fatal(err)
	}
	changeRoot := filepath.Join(target, lufypaths.LufySDD, "changes", "linked-overview")
	overviewPath := filepath.Join(changeRoot, overviewFile)
	if err := os.Remove(overviewPath); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(target, "outside-overview.html")
	if err := os.WriteFile(outside, []byte("outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, overviewPath); err != nil {
		t.Skipf("symlink no soportado: %v", err)
	}
	if _, err := service.Validate(target, "linked-overview", true); err == nil || !strings.Contains(err.Error(), "overview no es un archivo regular") {
		t.Fatalf("expected overview symlink rejection, got %v", err)
	}
	if got := mustRead(t, outside); got != "outside\n" {
		t.Fatalf("overview externo mutado: %q", got)
	}
}

func TestSyncAddsModifiesAndRemovesRequirement(t *testing.T) {
	target := workflowFixture(t)
	service := NewService()
	if _, err := service.New(target, "change-audit-log", "audit-log"); err != nil {
		t.Fatal(err)
	}
	report, err := service.Sync(target, "change-audit-log")
	if err != nil || !report.Valid() || report.Status != "synced" {
		t.Fatalf("sync add report=%#v err=%v", report, err)
	}
	mainSpec := filepath.Join(target, lufypaths.LufySDD, "specs", "audit-log", "spec.md")
	body := mustRead(t, mainSpec)
	if !strings.Contains(body, "Describe the required behavior") {
		t.Fatalf("main spec sin requirement agregado: %s", body)
	}
	if repeated, repeatErr := service.Sync(target, "change-audit-log"); repeatErr != nil || repeated.Status != "synced" || len(repeated.Actions) != 0 {
		t.Fatalf("sync idempotente report=%#v err=%v", repeated, repeatErr)
	}

	deltaPath := filepath.Join(target, lufypaths.LufySDD, "changes", "change-audit-log", "specs", "audit-log", "spec.md")
	modified := `# audit-log Specification Delta

## MODIFIED Requirements

### Requirement: Describe the required behavior

El sistema SHALL guardar auditoría durable.

#### Scenario: Audit is durable

- **WHEN** una acción termina
- **THEN** la auditoría queda persistida
`
	if err := os.WriteFile(deltaPath, []byte(modified), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err = service.Sync(target, "change-audit-log")
	if err != nil || !report.Valid() || !strings.Contains(mustRead(t, mainSpec), "auditoría durable") {
		t.Fatalf("sync modify report=%#v err=%v spec=%s", report, err, mustRead(t, mainSpec))
	}

	removed := `# audit-log Specification Delta

## REMOVED Requirements

### Requirement: Describe the required behavior

Se elimina porque fue reemplazado por otro contrato.
`
	if err := os.WriteFile(deltaPath, []byte(removed), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err = service.Sync(target, "change-audit-log")
	if err != nil || !report.Valid() || strings.Contains(mustRead(t, mainSpec), "Describe the required behavior") {
		t.Fatalf("sync remove report=%#v err=%v spec=%s", report, err, mustRead(t, mainSpec))
	}
}

func TestSyncAmbiguityDoesNotWrite(t *testing.T) {
	target := workflowFixture(t)
	service := NewService()
	if _, err := service.New(target, "modify-missing", "audit-log"); err != nil {
		t.Fatal(err)
	}
	deltaPath := filepath.Join(target, lufypaths.LufySDD, "changes", "modify-missing", "specs", "audit-log", "spec.md")
	body := `## MODIFIED Requirements
### Requirement: Missing
#### Scenario: Missing remains safe
- **WHEN** sync runs
- **THEN** no file is written
`
	if err := os.WriteFile(deltaPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := service.Sync(target, "modify-missing")
	if err != nil {
		t.Fatal(err)
	}
	if report.Valid() || report.Status != "blocked" || !hasCode(report.Diagnostics, "ambiguous_requirement") {
		t.Fatalf("report = %#v", report)
	}
	mainSpec := filepath.Join(target, lufypaths.LufySDD, "specs", "audit-log", "spec.md")
	if _, err := os.Lstat(mainSpec); !os.IsNotExist(err) {
		t.Fatalf("sync ambiguo escribió %s: %v", mainSpec, err)
	}
}

func TestSyncRollsBackWhenMetadataWriteFails(t *testing.T) {
	target := workflowFixture(t)
	service := NewService()
	if _, err := service.New(target, "rollback-sync", "audit-log"); err != nil {
		t.Fatal(err)
	}
	writes := 0
	service.write = func(path string, body []byte, mode os.FileMode) error {
		writes++
		if writes == 2 {
			return errors.New("injected metadata failure")
		}
		return atomicWrite(path, body, mode)
	}
	if _, err := service.Sync(target, "rollback-sync"); err == nil {
		t.Fatal("expected injected sync error")
	}
	mainSpec := filepath.Join(target, lufypaths.LufySDD, "specs", "audit-log", "spec.md")
	if _, err := os.Lstat(mainSpec); !os.IsNotExist(err) {
		t.Fatalf("rollback dejó spec nueva: %v", err)
	}
	metadata := mustRead(t, filepath.Join(target, lufypaths.LufySDD, "changes", "rollback-sync", "change.yaml"))
	if !strings.Contains(metadata, `syncedDigest: ""`) {
		t.Fatalf("rollback alteró metadata: %s", metadata)
	}
}

func TestSyncRejectsSymlinkTarget(t *testing.T) {
	target := workflowFixture(t)
	service := NewService()
	if _, err := service.New(target, "linked-target", "audit-log"); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(target, "outside.md")
	if err := os.WriteFile(outside, []byte("outside\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	capabilityRoot := filepath.Join(target, lufypaths.LufySDD, "specs", "audit-log")
	if err := os.MkdirAll(capabilityRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(capabilityRoot, "spec.md")); err != nil {
		t.Skipf("symlink no soportado: %v", err)
	}
	if _, err := service.Sync(target, "linked-target"); err == nil || !strings.Contains(err.Error(), "symlink no permitido") {
		t.Fatalf("expected symlink rejection, got %v", err)
	}
	if got := mustRead(t, outside); got != "outside\n" {
		t.Fatalf("symlink target mutado: %q", got)
	}
}

func TestArchiveRequiresGatesAndMovesReadyChange(t *testing.T) {
	target := workflowFixture(t)
	service := NewService()
	service.now = func() time.Time { return time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC) }
	if _, err := service.New(target, "archive-ready", "archive-ready"); err != nil {
		t.Fatal(err)
	}
	blocked, err := service.Archive(target, "archive-ready")
	if err != nil || blocked.Valid() || blocked.Status != "blocked" {
		t.Fatalf("archive should block: report=%#v err=%v", blocked, err)
	}
	changeRoot := filepath.Join(target, lufypaths.LufySDD, "changes", "archive-ready")
	tasks := "# Tasks\n\n- [x] Implementar.\n- [x] Validar.\n"
	if err := os.WriteFile(filepath.Join(changeRoot, "tasks.md"), []byte(tasks), 0o644); err != nil {
		t.Fatal(err)
	}
	if report, err := service.Sync(target, "archive-ready"); err != nil || !report.Valid() {
		t.Fatalf("sync: report=%#v err=%v", report, err)
	}
	evidenceRoot := filepath.Join(target, lufypaths.LufySDD, "verification", "archive-ready")
	if err := os.MkdirAll(evidenceRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(evidenceRoot, "report.md"), []byte("validated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(target, lufypaths.LufySDD, "archive", "2026-08-17-archive-ready")
	if err := os.Mkdir(destination, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Archive(target, "archive-ready"); err == nil || !strings.Contains(err.Error(), "ya existe") {
		t.Fatalf("expected archive collision, got %v", err)
	}
	if err := os.Remove(destination); err != nil {
		t.Fatal(err)
	}
	report, err := service.Archive(target, "archive-ready")
	if err != nil || !report.Valid() || report.Status != "archived" {
		t.Fatalf("archive report=%#v err=%v", report, err)
	}
	if info, err := os.Stat(destination); err != nil || !info.IsDir() {
		t.Fatalf("archive destination: %v", err)
	}
	if _, err := os.Lstat(changeRoot); !os.IsNotExist(err) {
		t.Fatalf("active change still exists: %v", err)
	}
}

func TestRejectsUnsafeIDsAndSymlinkSpecs(t *testing.T) {
	target := workflowFixture(t)
	service := NewService()
	if _, err := service.New(target, "../escape", "audit-log"); err == nil {
		t.Fatal("expected unsafe change error")
	}
	if _, err := service.New(target, "safe-change", "../escape"); err == nil {
		t.Fatal("expected unsafe capability error")
	}
	if _, err := service.New(target, "linked-spec", "audit-log"); err != nil {
		t.Fatal(err)
	}
	spec := filepath.Join(target, lufypaths.LufySDD, "changes", "linked-spec", "specs", "audit-log", "spec.md")
	if err := os.Remove(spec); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(target, "outside.md"), spec); err != nil {
		t.Skipf("symlink no soportado: %v", err)
	}
	report, err := service.Validate(target, "linked-spec", true)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if report.Valid() || !hasCode(report.Diagnostics, "missing_spec") {
		t.Fatalf("report = %#v", report)
	}
}

func workflowFixture(t *testing.T) string {
	t.Helper()
	target := t.TempDir()
	for _, rel := range []string{"", "changes", "specs", "decisions", "verification", "archive"} {
		path := filepath.Join(target, lufypaths.LufySDD, rel)
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return target
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}
