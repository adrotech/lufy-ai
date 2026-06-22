package setup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	contextapp "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/application"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/installer"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/layout"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/memory"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/verify"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/versioncheck"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Options struct {
	Target           string
	DryRun           bool
	Yes              bool
	JSON             bool
	SkipVersionCheck bool
	RequireLatest    bool
	CheckNewFeatures bool
	Interactive      bool
	VersionCheck     func() versioncheck.Result
	Scope            assets.Scope
	Harness          domain.HarnessConfig
}

type FeatureAction struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Reason   string `json:"reason"`
	Applied  bool   `json:"applied"`
	Error    string `json:"error,omitempty"`
	Recovery string `json:"recovery,omitempty"`
	Since    string `json:"since,omitempty"`
}

type Report struct {
	TargetRoot       string               `json:"targetRoot"`
	DryRun           bool                 `json:"dryRun"`
	Applied          bool                 `json:"applied"`
	CheckNewFeatures bool                 `json:"checkNewFeatures"`
	Version          *versioncheck.Result `json:"version,omitempty"`
	Features         []FeatureAction      `json:"features"`
}

type Service struct{}

type FeatureSpec struct {
	ID    string
	Name  string
	Since string
}

var featureRegistry = []FeatureSpec{
	{ID: "layout", Name: "Layout .lufy", Since: "v0.6.0"},
	{ID: "install", Name: "Assets gestionados", Since: "v0.1.0"},
	{ID: "project-config", Name: "Project config", Since: "v0.3.0"},
	{ID: "stack-profile", Name: "Stack y superficies", Since: "v0.4.0"},
	{ID: "sdd-methodology", Name: "Metodologia SDD", Since: "v0.6.0"},
	{ID: "memory", Name: "Memoria Obsidian", Since: "v0.6.0"},
	{ID: "context-graph", Name: "Context graph", Since: "v0.6.0"},
	{ID: "verify", Name: "Verify final", Since: "v0.1.0"},
}

func NewService() Service {
	return Service{}
}

func (s Service) Run(opts Options, stdout io.Writer) error {
	report, err := s.BuildReport(opts)
	if err != nil {
		return err
	}
	if report.Version != nil && opts.RequireLatest && (report.Version.UpdateAvailable || report.Version.Error != "") {
		if opts.JSON {
			_ = writeJSON(stdout, report)
		}
		if report.Version.Error != "" {
			return fmt.Errorf("setup requiere ultima version pero no pudo verificar releases: %s", report.Version.Error)
		}
		return fmt.Errorf("setup requiere ultima version; ejecuta lufy-ai upgrade --to %s", report.Version.LatestVersion)
	}
	if hasConflicts(report.Features) {
		if opts.JSON {
			_ = writeJSON(stdout, report)
		}
		if !opts.JSON {
			printReport(stdout, report)
		}
		return fmt.Errorf("setup bloqueado por conflictos; revisa las acciones recovery antes de aplicar")
	}
	if opts.JSON {
		if opts.DryRun {
			return writeJSON(stdout, report)
		}
		if !opts.Yes && hasPendingMutations(report.Features) {
			_ = writeJSON(stdout, report)
			return fmt.Errorf("setup requiere --yes para aplicar mutaciones reales; usa --dry-run para revisar el plan sin escribir")
		}
		if err := s.apply(opts, &report, io.Discard); err != nil {
			_ = writeJSON(stdout, report)
			return err
		}
		return writeJSON(stdout, report)
	}
	printReport(stdout, report)
	if opts.DryRun {
		fmt.Fprintln(stdout, "Modo dry-run: sin mutaciones en filesystem")
		return nil
	}
	if opts.Interactive && hasPendingMutations(report.Features) {
		selected, err := runChecklist(report)
		if err != nil {
			return err
		}
		if len(selected) == 0 {
			fmt.Fprintln(stdout, "Setup cancelado: no se seleccionaron acciones")
			return nil
		}
		report.Features = filterSelected(report.Features, selected)
		opts.Yes = true
	}
	if !opts.Yes && hasPendingMutations(report.Features) {
		return fmt.Errorf("setup requiere --yes para aplicar mutaciones reales; usa --dry-run para revisar el plan sin escribir")
	}
	if opts.Yes {
		return s.apply(opts, &report, stdout)
	}
	return nil
}

func (s Service) BuildReport(opts Options) (Report, error) {
	target, err := platform.ResolveTargetPath(opts.Target)
	if err != nil {
		return Report{}, err
	}
	report := Report{TargetRoot: target, DryRun: opts.DryRun, CheckNewFeatures: opts.CheckNewFeatures}
	if !opts.SkipVersionCheck {
		check := opts.VersionCheck
		if check == nil {
			check = func() versioncheck.Result { return versioncheck.NewService().Check(versioncheck.Options{}) }
		}
		result := check()
		report.Version = &result
	}
	report.Features = s.planFeatures(target)
	return report, nil
}

func (s Service) planFeatures(target string) []FeatureAction {
	features := []FeatureAction{}
	layoutReport, err := layout.BuildPlan(target)
	if err != nil {
		features = append(features, action("layout", "apply", err.Error(), "lufy-ai migrate-layout --target <dir> --dry-run"))
	} else if len(layoutReport.Conflicts) > 0 {
		features = append(features, action("layout", "conflict", fmt.Sprintf("%d conflicto(s) de layout legacy", len(layoutReport.Conflicts)), "lufy-ai migrate-layout --target <dir> --dry-run"))
	} else if hasLayoutMutations(layoutReport.Actions) {
		features = append(features, action("layout", "apply", layoutSummary(layoutReport.Actions), "lufy-ai migrate-layout --target <dir> --yes"))
	} else {
		features = append(features, action("layout", "skip", "Layout .lufy listo", ""))
	}
	if _, err := state.Load(target); err != nil || !existingInstallState(target) {
		features = append(features, action("install", "apply", "No existe manifest de instalacion", "lufy-ai install --target <dir> --yes"))
	} else {
		features = append(features, action("install", "skip", "Manifest de instalacion presente", ""))
	}
	configPath := projectconfig.Path(target)
	if !fileExists(configPath) {
		if existing, err := projectconfig.ExistingPath(target); err == nil && fileExists(existing) {
			features = append(features, action("project-config", "skip", "Config legacy/canonica existente", ""))
		} else {
			features = append(features, action("project-config", "apply", "Falta .lufy/config/project.yaml", "lufy-ai init --target <dir>"))
		}
	} else {
		features = append(features, action("project-config", "skip", "Project config presente", ""))
	}
	features = append(features, s.planStackProfile(target))
	features = append(features, s.planSDDMethodology(target))
	memStatus, memErr := memory.NewService().BuildStatus(memory.Options{Target: target})
	if memErr != nil || !memStatus.Status.Initialized {
		reason := "Memoria Obsidian no inicializada"
		if memErr != nil {
			reason = memErr.Error()
		}
		features = append(features, action("memory", "apply", reason, "lufy-ai memory init --target <dir>"))
	} else {
		features = append(features, action("memory", "skip", "Memoria inicializada", ""))
	}
	ctxStatus := contextapp.NewService().Status(target)
	if ctxStatus.Status != "ready" {
		reason := ctxStatus.Reason
		if reason == "" {
			reason = "Context graph no esta ready"
		}
		features = append(features, action("context-graph", "apply", reason, "lufy-ai context build --target <dir>"))
	} else {
		features = append(features, action("context-graph", "skip", "Context graph ready", ""))
	}
	features = append(features, action("verify", "apply", "Validar instalacion y drift despues de setup", "lufy-ai verify --target <dir>"))
	return features
}

func (s Service) apply(opts Options, report *Report, stdout io.Writer) error {
	scope := opts.Scope
	if scope == "" {
		scope = assets.ScopeProject
	}
	harness := opts.Harness.WithDefaults()
	if harness.Tool == "" {
		harness = domain.HarnessConfig{}.WithDefaults()
	}
	for i := range report.Features {
		feature := &report.Features[i]
		if feature.Status != "apply" {
			continue
		}
		var err error
		switch feature.ID {
		case "layout":
			layoutPlan, buildErr := layout.BuildPlan(report.TargetRoot)
			if buildErr != nil {
				err = buildErr
			} else if len(layoutPlan.Conflicts) > 0 {
				err = fmt.Errorf("layout bloqueado por %d conflicto(s); ejecuta lufy-ai migrate-layout --target %s --dry-run", len(layoutPlan.Conflicts), report.TargetRoot)
			} else {
				_, err = layout.Apply(report.TargetRoot, layoutPlan.Actions, stdout)
			}
		case "install":
			err = installer.NewService().Run(installer.Options{Target: report.TargetRoot, Yes: true, Scope: scope, Harness: harness}, stdout)
		case "project-config":
			if fileExists(projectconfig.Path(report.TargetRoot)) {
				feature.Status = "skip"
				feature.Reason = "Project config creado por install"
				continue
			}
			err = projectconfig.NewService().Run(projectconfig.Options{Target: report.TargetRoot}, stdout)
		case "memory":
			err = memory.NewService().Init(memory.Options{Target: report.TargetRoot}, stdout)
		case "stack-profile", "sdd-methodology":
			err = projectconfig.NewService().Run(projectconfig.Options{Target: report.TargetRoot, Rescan: true}, stdout)
		case "context-graph":
			_, err = contextapp.NewService().Build(report.TargetRoot)
			if err == nil {
				fmt.Fprintln(stdout, "Context graph generado")
			}
		case "verify":
			err = verify.NewService().Run(verify.Options{Target: report.TargetRoot, Scope: scope, ExpectedTool: harness.Tool}, stdout)
		}
		if err != nil {
			feature.Error = err.Error()
			return fmt.Errorf("setup %s fallo: %w", feature.ID, err)
		}
		feature.Applied = true
	}
	report.Applied = true
	return nil
}

func (s Service) planStackProfile(target string) FeatureAction {
	path, err := projectconfig.ExistingPath(target)
	if err != nil || !fileExists(path) {
		return action("stack-profile", "skip", "Se genera junto con project-config", "lufy-ai scan --target <dir>")
	}
	cfg, err := projectconfig.Load(path)
	if err != nil {
		return action("stack-profile", "apply", "Project config no se pudo leer para validar stacks/superficies", "lufy-ai scan --target <dir>")
	}
	if len(cfg.Stacks) == 0 || len(cfg.ProjectProfile.Surfaces) == 0 {
		return action("stack-profile", "apply", "Faltan stacks o superficies detectadas", "lufy-ai scan --target <dir>")
	}
	return action("stack-profile", "skip", "Stacks y superficies presentes", "")
}

func (s Service) planSDDMethodology(target string) FeatureAction {
	path, err := projectconfig.ExistingPath(target)
	if err != nil || !fileExists(path) {
		return action("sdd-methodology", "skip", "Se configura junto con project-config/install", "lufy-ai install --target <dir> --yes")
	}
	cfg, err := projectconfig.Load(path)
	if err != nil {
		return action("sdd-methodology", "apply", "Project config no se pudo leer para validar metodologia", "lufy-ai init --target <dir> --rescan")
	}
	if len(cfg.MethodologyByTier) == 0 || cfg.Tool == "" {
		return action("sdd-methodology", "apply", "Falta tool o methodology_by_tier", "lufy-ai init --target <dir> --rescan")
	}
	return action("sdd-methodology", "skip", "Tool y methodology_by_tier presentes", "")
}

func printReport(out io.Writer, report Report) {
	fmt.Fprintf(out, "Setup para %s\n", report.TargetRoot)
	if report.Version == nil {
		fmt.Fprintln(out, "Version: check omitido")
	} else if report.Version.Error != "" {
		fmt.Fprintf(out, "Version: no verificada (%s)\n", report.Version.Error)
		fmt.Fprintf(out, "Recomendacion: %s\n", report.Version.Recommendation)
	} else if report.Version.UpdateAvailable {
		fmt.Fprintf(out, "Version: local=%s latest=%s update_disponible=true\n", report.Version.CurrentVersion, report.Version.LatestVersion)
		fmt.Fprintf(out, "Recomendacion: %s\n", report.Version.Recommendation)
	} else if report.Version.UpToDate {
		fmt.Fprintf(out, "Version: local=%s latest=%s\n", report.Version.CurrentVersion, report.Version.LatestVersion)
		fmt.Fprintln(out, "Lufy AI esta al dia")
	} else {
		fmt.Fprintf(out, "Version: local=%s latest=%s\n", report.Version.CurrentVersion, report.Version.LatestVersion)
		fmt.Fprintf(out, "Recomendacion: %s\n", report.Version.Recommendation)
	}
	if report.CheckNewFeatures {
		fmt.Fprintln(out, "Modo check-new-features: revisando capacidades configurables pendientes")
	}
	for _, feature := range report.Features {
		fmt.Fprintf(out, "- [%s] %s: %s", feature.Status, feature.ID, feature.Reason)
		if feature.Recovery != "" && feature.Status == "apply" {
			fmt.Fprintf(out, " (%s)", feature.Recovery)
		}
		fmt.Fprintln(out)
	}
}

func hasPendingMutations(features []FeatureAction) bool {
	for _, feature := range features {
		if feature.Status == "apply" {
			return true
		}
	}
	return false
}

func hasConflicts(features []FeatureAction) bool {
	for _, feature := range features {
		if feature.Status == "conflict" {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func writeJSON(out io.Writer, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = io.Copy(out, bytes.NewReader(append(body, '\n')))
	return err
}

func StatePathForTest(target string) string {
	return filepath.ToSlash(state.Path(target))
}

func action(id, status, reason, recovery string) FeatureAction {
	spec := featureSpec(id)
	return FeatureAction{ID: id, Name: spec.Name, Status: status, Reason: reason, Recovery: recovery, Since: spec.Since}
}

func featureSpec(id string) FeatureSpec {
	for _, spec := range featureRegistry {
		if spec.ID == id {
			return spec
		}
	}
	return FeatureSpec{ID: id, Name: id}
}

func hasLayoutMutations(actions []layout.Action) bool {
	for _, item := range actions {
		if item.Kind == "migrate-copy" || item.Kind == "write-readme" {
			return true
		}
	}
	return false
}

func layoutSummary(actions []layout.Action) string {
	parts := []string{}
	for _, item := range actions {
		if item.Kind == "migrate-copy" || item.Kind == "write-readme" {
			parts = append(parts, item.Kind+":"+item.Target)
		}
	}
	if len(parts) == 0 {
		return "Layout .lufy listo"
	}
	return strings.Join(parts, ", ")
}

func existingInstallState(target string) bool {
	path, err := state.ExistingPath(target)
	return err == nil && fileExists(path)
}

func filterSelected(features []FeatureAction, selected map[string]bool) []FeatureAction {
	out := make([]FeatureAction, 0, len(features))
	for _, feature := range features {
		if feature.Status == "apply" && !selected[feature.ID] {
			feature.Status = "skip"
			feature.Reason = "Omitido por seleccion interactiva"
		}
		out = append(out, feature)
	}
	return out
}

type checklistModel struct {
	report    Report
	indexes   []int
	selected  map[string]bool
	cursor    int
	done      bool
	cancelled bool
	keys      checklistKeys
}

type checklistKeys struct {
	Up      key.Binding
	Down    key.Binding
	Toggle  key.Binding
	Confirm key.Binding
	Cancel  key.Binding
}

func newChecklistModel(report Report) checklistModel {
	indexes := []int{}
	selected := map[string]bool{}
	for i, feature := range report.Features {
		if feature.Status == "apply" {
			indexes = append(indexes, i)
			selected[feature.ID] = true
		}
	}
	return checklistModel{report: report, indexes: indexes, selected: selected, keys: checklistKeys{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "subir")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "bajar")),
		Toggle:  key.NewBinding(key.WithKeys(" "), key.WithHelp("espacio", "activar/desactivar")),
		Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "aplicar seleccion")),
		Cancel:  key.NewBinding(key.WithKeys("esc", "ctrl+c", "q"), key.WithHelp("esc/q", "cancelar")),
	}}
}

func (m checklistModel) Init() tea.Cmd { return nil }

func (m checklistModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(keyMsg, m.keys.Cancel):
		m.cancelled = true
		return m, tea.Quit
	case key.Matches(keyMsg, m.keys.Confirm):
		m.done = true
		return m, tea.Quit
	case key.Matches(keyMsg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(keyMsg, m.keys.Down):
		if m.cursor < len(m.indexes)-1 {
			m.cursor++
		}
	case key.Matches(keyMsg, m.keys.Toggle):
		if len(m.indexes) > 0 {
			id := m.report.Features[m.indexes[m.cursor]].ID
			m.selected[id] = !m.selected[id]
		}
	}
	return m, nil
}

func (m checklistModel) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).Render("Lufy setup")
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", title)
	if m.report.Version != nil {
		if m.report.Version.UpdateAvailable {
			fmt.Fprintf(&b, "Version local=%s latest=%s. Recomendado: lufy-ai upgrade --to %s\n\n", m.report.Version.CurrentVersion, m.report.Version.LatestVersion, m.report.Version.LatestVersion)
		} else if m.report.Version.UpToDate {
			fmt.Fprintln(&b, "Lufy AI esta al dia")
			fmt.Fprintln(&b)
		}
	}
	if len(m.indexes) == 0 {
		fmt.Fprintln(&b, "No hay acciones pendientes. enter para salir.")
		return b.String()
	}
	fmt.Fprintln(&b, "Selecciona acciones a aplicar:")
	for pos, idx := range m.indexes {
		feature := m.report.Features[idx]
		cursor := " "
		if pos == m.cursor {
			cursor = ">"
		}
		checked := " "
		if m.selected[feature.ID] {
			checked = "x"
		}
		line := fmt.Sprintf("%s [%s] %s (%s) - %s", cursor, checked, feature.Name, feature.ID, feature.Reason)
		if pos == m.cursor {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render(line)
		}
		fmt.Fprintln(&b, line)
	}
	fmt.Fprintf(&b, "\n%s · %s · %s\n", m.keys.Toggle.Help().Key, m.keys.Confirm.Help().Key, m.keys.Cancel.Help().Key)
	return b.String()
}

func runChecklist(report Report) (map[string]bool, error) {
	model := newChecklistModel(report)
	finalModel, err := tea.NewProgram(model).Run()
	if err != nil {
		return nil, err
	}
	result, ok := finalModel.(checklistModel)
	if !ok || result.cancelled {
		return nil, fmt.Errorf("setup cancelado por el usuario")
	}
	selected := map[string]bool{}
	for id, enabled := range result.selected {
		if enabled {
			selected[id] = true
		}
	}
	return selected, nil
}
