package setup

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/conflictplan"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/versioncheck"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSetupDryRunDoesNotWrite(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	err := NewService().Run(Options{Target: target, DryRun: true, SkipVersionCheck: true}, &out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Version: check omitido") || !strings.Contains(out.String(), "[apply] install") {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy")); !os.IsNotExist(err) {
		t.Fatalf("dry-run wrote .lufy: %v", err)
	}
}

func TestSetupJSONIncludesVersionAndFeatures(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	err := NewService().Run(Options{
		Target: target,
		DryRun: true,
		JSON:   true,
		VersionCheck: func() versioncheck.Result {
			return versioncheck.Result{Checked: true, CurrentVersion: "v1.0.0", LatestVersion: "v1.0.1", UpdateAvailable: true, Recommendation: "upgrade"}
		},
	}, &out)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Report
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid JSON %q: %v", out.String(), err)
	}
	if decoded.Version == nil || !decoded.Version.UpdateAvailable {
		t.Fatalf("missing update in JSON: %#v", decoded.Version)
	}
	if len(decoded.Features) == 0 {
		t.Fatalf("missing features: %#v", decoded)
	}
}

func TestSetupRejectsPlaceholderTarget(t *testing.T) {
	var out bytes.Buffer
	err := NewService().Run(Options{Target: "/ruta/a/proyecto", SkipVersionCheck: true}, &out)
	if err == nil || !strings.Contains(err.Error(), "placeholder") || !strings.Contains(err.Error(), "ruta real") {
		t.Fatalf("expected placeholder target error, got %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("placeholder target should fail before rendering, got output:\n%s", out.String())
	}
}

func TestSetupJSONWithoutYesFailsWhenMutating(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	err := NewService().Run(Options{Target: target, JSON: true, SkipVersionCheck: true}, &out)
	if err == nil || !strings.Contains(err.Error(), "requiere --yes") {
		t.Fatalf("expected --yes error, got %v", err)
	}
	var decoded Report
	if jsonErr := json.Unmarshal(out.Bytes(), &decoded); jsonErr != nil {
		t.Fatalf("expected JSON report before error, got %q: %v", out.String(), jsonErr)
	}
	if len(decoded.Features) == 0 {
		t.Fatalf("missing feature plan: %#v", decoded)
	}
}

func TestSetupRequireLatestBlocks(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	err := NewService().Run(Options{
		Target:        target,
		RequireLatest: true,
		VersionCheck: func() versioncheck.Result {
			return versioncheck.Result{Checked: true, CurrentVersion: "v1.0.0", LatestVersion: "v1.0.1", UpdateAvailable: true}
		},
	}, &out)
	if err == nil || !strings.Contains(err.Error(), "requiere ultima version") {
		t.Fatalf("expected require latest error, got %v", err)
	}
}

func TestSetupRequireLatestBlocksWhenCheckFails(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	err := NewService().Run(Options{
		Target:        target,
		RequireLatest: true,
		VersionCheck: func() versioncheck.Result {
			return versioncheck.Result{Checked: true, CurrentVersion: "v1.0.0", Error: "network unavailable"}
		},
	}, &out)
	if err == nil || !strings.Contains(err.Error(), "no pudo verificar") {
		t.Fatalf("expected verification error, got %v", err)
	}
}

func TestSetupApplyEndToEnd(t *testing.T) {
	target := t.TempDir()
	err := NewService().Run(Options{Target: target, Yes: true, SkipVersionCheck: true}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{
		filepath.Join(".lufy", "README.md"),
		filepath.Join(".lufy", "config", "project.yaml"),
		filepath.Join(".lufy", "managed-state", "install-state.json"),
		filepath.Join(".lufy", "memory", "MEMORY.md"),
		filepath.Join(".lufy", "context", "graph.json"),
	} {
		if _, statErr := os.Stat(filepath.Join(target, rel)); statErr != nil {
			t.Fatalf("expected %s after setup: %v", rel, statErr)
		}
	}
	report, err := NewService().BuildReport(Options{Target: target, SkipVersionCheck: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, feature := range report.Features {
		if feature.ID == "layout" || feature.ID == "install" || feature.ID == "memory" || feature.ID == "context-graph" {
			if feature.Status != "skip" {
				t.Fatalf("expected %s to skip after setup, got %#v", feature.ID, feature)
			}
		}
	}
}

func TestSetupPlansInstallConflictRecovery(t *testing.T) {
	target := t.TempDir()
	writeSetupTestFile(t, filepath.Join(target, ".opencode", "agents", "orchestrator.md"), "local agent\n")
	report, err := NewService().BuildReport(Options{Target: target, SkipVersionCheck: true})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, feature := range report.Features {
		if feature.ID != "install" {
			continue
		}
		found = true
		if feature.Status != "conflict" || !strings.Contains(feature.Recovery, "conflicts plan") {
			t.Fatalf("unexpected install conflict feature: %#v", feature)
		}
	}
	if !found {
		t.Fatalf("missing install feature: %#v", report.Features)
	}
}

func TestChecklistModelRendersRichSetupPlan(t *testing.T) {
	report := Report{
		TargetRoot:       "/tmp/project",
		CheckNewFeatures: true,
		Version:          &versioncheck.Result{Checked: true, CurrentVersion: "v1.0.0", LatestVersion: "v1.0.0", UpToDate: true},
		Features: []FeatureAction{
			{ID: "layout", Name: "Layout .lufy", Status: "skip", Reason: "Layout .lufy listo"},
			{ID: "verify", Name: "Verify final", Status: "apply", Reason: "Validar instalacion", Recovery: "lufy-ai verify --target <dir>"},
		},
	}
	model := newChecklistModel(report)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 110, Height: 32})
	view := updated.(checklistModel).View()
	for _, want := range []string{"Lufy setup", "VERSION OK", "/tmp/project", "Setup plan", "Current step", "Console", "01 ✓ Layout .lufy", "02 ◆ Verify final", "Progress", "Qué hará", "comprueba que lo instalado", "Impacto", "lee el proyecto", "Comando", "lufy-ai verify", "Listo para aplicar 1 acciones", "espacio activar/desactivar"} {
		if !strings.Contains(view, want) {
			t.Fatalf("setup view missing %q:\n%s", want, view)
		}
	}
}

func TestChecklistModelResponsiveNavigationAndToggle(t *testing.T) {
	report := Report{TargetRoot: "/tmp/project", Features: []FeatureAction{
		{ID: "install", Name: "Assets gestionados", Status: "apply", Reason: "No existe manifest", Recovery: "lufy-ai install --target <dir> --yes"},
		{ID: "verify", Name: "Verify final", Status: "apply", Reason: "Validar instalacion", Recovery: "lufy-ai verify --target <dir>"},
	}}
	model := newChecklistModel(report)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 70, Height: 24})
	model = updated.(checklistModel)
	if view := model.View(); !strings.Contains(view, "Assets gestionados") || !strings.Contains(view, "Setup plan") || !strings.Contains(view, "Current step") || !strings.Contains(view, "Console") {
		t.Fatalf("compact layout should render dashboard panels:\n%s", view)
	}
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(checklistModel)
	if view := model.View(); !strings.Contains(view, "Verify final") || !strings.Contains(view, "lufy-ai verify") {
		t.Fatalf("down navigation should change active detail:\n%s", view)
	}
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(checklistModel)
	if model.selected["verify"] || !strings.Contains(model.View(), "02 ○ Verify final") {
		t.Fatalf("space should toggle active feature off:\n%s", model.View())
	}
}

func TestChecklistModelRendersConflictDashboard(t *testing.T) {
	report := Report{TargetRoot: "/tmp/project", Features: []FeatureAction{
		{ID: "install", Name: "Assets gestionados", Status: "conflict", Reason: "manifest ausente y 2 conflicto(s)", Recovery: "lufy-ai conflicts plan --target <dir>"},
	}}
	view := newChecklistModel(report).View()
	for _, want := range []string{"Setup plan", "Current step", "Console", "01 ! Assets gestionados", "bloqueado", "Hay 1 conflicto", "manifest ausente", "lufy-ai conflicts plan"} {
		if !strings.Contains(view, want) {
			t.Fatalf("conflict view missing %q:\n%s", want, view)
		}
	}
}

func TestConflictReviewModelRendersAndNavigates(t *testing.T) {
	plan := conflictplan.Report{
		TargetRoot: "/tmp/project",
		Summary:    conflictplan.Summary{Conflicts: 2, Groups: 2, LegacyDeprecated: 1},
		Groups: []conflictplan.Group{
			{Category: ".opencode/agents", Risk: "medium", Count: 1, ParallelGroup: "opencode-agents"},
			{Category: "openspec/specs", Risk: "medium", Count: 1, ParallelGroup: "openspec-specs"},
		},
		Items: []conflictplan.Item{
			{Path: ".opencode/agents/orchestrator.md", Category: ".opencode/agents", Risk: "medium", Recommendation: "merge", Reason: "local", AvailableActions: []string{"keep-local", "merge"}},
			{Path: "openspec/specs/foo/spec.md", Category: "openspec/specs", Risk: "medium", Recommendation: "merge", Reason: "spec", AvailableActions: []string{"keep-local", "merge"}},
		},
		LegacyDeprecated: []conflictplan.LegacyItem{{Path: ".lufy-ai/backups", Recommendation: "migrate-layout"}},
	}
	model := newConflictReviewModel(plan)
	if view := model.View(); !strings.Contains(view, "Lufy setup: conflictos detectados") || !strings.Contains(view, ".opencode/agents") || !strings.Contains(view, "Legacy/deprecated") {
		t.Fatalf("unexpected view:\n%s", view)
	}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	next := updated.(conflictReviewModel)
	if next.groupIndex != 1 || !strings.Contains(next.View(), "openspec/specs") {
		t.Fatalf("right navigation failed: %#v\n%s", next, next.View())
	}
	updated, cmd := next.Update(tea.KeyMsg{Type: tea.KeyEnter})
	quit := updated.(conflictReviewModel)
	if !quit.quit || cmd == nil {
		t.Fatalf("enter should quit: %#v cmd=%v", quit, cmd)
	}
}

func TestApplyModelRendersLogsAndStepResults(t *testing.T) {
	report := Report{
		TargetRoot: "/tmp/project",
		Features: []FeatureAction{
			{ID: "install", Name: "Assets gestionados", Status: "apply", Reason: "No existe manifest"},
			{ID: "verify", Name: "Verify final", Status: "apply", Reason: "Validar instalacion"},
		},
	}
	model := newApplyModel(NewService(), Options{}, &report)
	updated, _ := model.Update(applyStepStartMsg(0))
	model = updated.(applyModel)
	updated, _ = model.Update(applyLogMsg("instalando assets\n"))
	model = updated.(applyModel)
	updated, _ = model.Update(applyStepDoneMsg{index: 0})
	model = updated.(applyModel)
	updated, _ = model.Update(applyFinishedMsg{})
	model = updated.(applyModel)

	view := model.View()
	for _, want := range []string{"Lufy setup", "Setup plan", "Current step", "Console", "▶ Ejecutando: Assets gestionados", "instalando assets", "✓ Completado", "Resumen final: setup completado"} {
		if !strings.Contains(view, want) {
			t.Fatalf("apply view missing %q:\n%s", want, view)
		}
	}
}

func TestApplyModelLogAppendSurvivesValueCopies(t *testing.T) {
	report := Report{
		TargetRoot: "/tmp/project",
		Features: []FeatureAction{
			{ID: "memory", Name: "Memoria Obsidian", Status: "apply", Reason: "Inicializar memoria"},
		},
	}
	model := newApplyModel(NewService(), Options{}, &report)

	updated, _ := model.Update(applyStepStartMsg(0))
	model = updated.(applyModel)
	copiedAfterFirstLog := model
	updated, _ = copiedAfterFirstLog.Update(applyLogMsg("creando vault\n"))
	model = updated.(applyModel)
	updated, _ = model.Update(applyStepDoneMsg{index: 0})
	model = updated.(applyModel)
	updated, _ = model.Update(applyFinishedMsg{})
	model = updated.(applyModel)

	view := model.View()
	for _, want := range []string{"▶ Ejecutando: Memoria Obsidian", "creando vault", "✓ Completado: Memoria Obsidian", "enter/q para salir"} {
		if !strings.Contains(view, want) {
			t.Fatalf("apply copied model view missing %q:\n%s", want, view)
		}
	}
}

func TestSetupRenderHelpersCoverStatusAndVersions(t *testing.T) {
	features := []FeatureAction{
		{ID: "layout", Name: "Layout .lufy", Status: "skip", Reason: "listo"},
		{ID: "install", Name: "Assets gestionados", Status: "apply", Reason: "pendiente"},
		{ID: "memory", Name: "Memoria Obsidian", Status: "apply", Reason: "omitida"},
		{ID: "verify", Name: "Verify final", Status: "conflict", Reason: "bloqueado"},
		{ID: "context-graph", Name: "Context graph", Status: "error", Reason: "fallo"},
	}

	cases := []struct {
		feature  FeatureAction
		selected bool
		marker   string
		label    string
	}{
		{features[0], false, "✓", "LISTO"},
		{features[1], true, "◆", "HACER"},
		{features[2], false, "○", "OMITIDO"},
		{features[3], false, "!", "BLOQUEADO"},
		{features[4], false, "✕", "ERROR"},
	}
	for _, tc := range cases {
		if got := featureMarker(tc.feature, tc.selected); got != tc.marker {
			t.Fatalf("featureMarker(%s) = %q, want %q", tc.feature.ID, got, tc.marker)
		}
		if got := featureStatusLabel(tc.feature, tc.selected); got != tc.label {
			t.Fatalf("featureStatusLabel(%s) = %q, want %q", tc.feature.ID, got, tc.label)
		}
		if badge := featureStatusBadge(tc.feature, tc.selected); !strings.Contains(badge, tc.label) {
			t.Fatalf("featureStatusBadge(%s) missing %q: %q", tc.feature.ID, tc.label, badge)
		}
	}

	for _, id := range []string{"layout", "install", "project-config", "stack-profile", "sdd-methodology", "memory", "context-graph", "verify", "unknown"} {
		if featureExplanation(id) == "" || featureEffect(id) == "" {
			t.Fatalf("missing explanation/effect for %s", id)
		}
	}

	versions := []versioncheck.Result{
		{Error: "offline"},
		{UpdateAvailable: true, CurrentVersion: "v1.0.0", LatestVersion: "v1.0.1"},
		{UpToDate: true, CurrentVersion: "v1.0.1", LatestVersion: "v1.0.1"},
		{CurrentVersion: "dev"},
		{},
	}
	for _, version := range versions {
		if versionLabelFor(version) == "" || versionBadgeFor(version) == "" || renderVersionLine(version) == "" {
			t.Fatalf("version helpers returned empty output for %#v", version)
		}
	}

	if progressPercent(-1, 10) != 0 || progressPercent(11, 10) != 1 || progressPercent(0, 0) != 0 {
		t.Fatalf("progressPercent clamp failed")
	}
	if start := visibleWindowStart(10, 3, 8); start != 7 {
		t.Fatalf("visibleWindowStart = %d, want 7", start)
	}
	if firstApplyFeature(features) != 1 {
		t.Fatalf("firstApplyFeature mismatch")
	}
}

func TestApplyModelErrorRowsAndMarkers(t *testing.T) {
	report := Report{TargetRoot: "/tmp/project", Features: []FeatureAction{
		{ID: "layout", Name: "Layout .lufy", Status: "skip", Reason: "Layout listo"},
		{ID: "install", Name: "Assets gestionados", Status: "apply", Reason: "Pendiente", Applied: true},
		{ID: "verify", Name: "Verify final", Status: "apply", Reason: "Validar", Error: "boom"},
	}}
	model := newApplyModel(NewService(), Options{}, &report)
	model.current = 2
	model.err = os.ErrInvalid
	model.done = true

	if got := applyStatusBadge("EJECUTANDO"); !strings.Contains(got, "EJECUTANDO") {
		t.Fatalf("applyStatusBadge missing label: %q", got)
	}
	if got := applyStatusBadge("OMITIDO"); !strings.Contains(got, "OMITIDO") {
		t.Fatalf("applyStatusBadge missing omitted label: %q", got)
	}
	if marker := applyMarkerFor(report.Features[2], false, "spin"); marker != "✕" {
		t.Fatalf("applyMarkerFor error = %q", marker)
	}
	if marker := applyMarkerFor(report.Features[1], false, "spin"); marker != "✓" {
		t.Fatalf("applyMarkerFor applied = %q", marker)
	}
	if marker := applyMarkerFor(FeatureAction{Status: "apply"}, true, "spin"); marker != "spin" {
		t.Fatalf("applyMarkerFor running = %q", marker)
	}
	if !reasonLooksReady("Layout listo") || !reasonLooksReady("context graph ready") || reasonLooksReady("pendiente") {
		t.Fatalf("reasonLooksReady mismatch")
	}

	view := strings.Join([]string{
		model.applyCurrentStepPanel(72, 12),
		model.applyConsolePanel(72, 10),
		model.applyRow(0, report.Features[0]),
		model.applyRow(1, report.Features[1]),
		model.applyRow(2, report.Features[2]),
	}, "\n")
	for _, want := range []string{"Current step", "error", "boom", "Console", "Fallo", "Layout .lufy", "LISTO", "Verify final", "ERROR"} {
		if !strings.Contains(view, want) {
			t.Fatalf("apply helper view missing %q:\n%s", want, view)
		}
	}
}

func writeSetupTestFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
