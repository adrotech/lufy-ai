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
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/conflictplan"
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
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
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
		if opts.Interactive {
			plan, planErr := conflictplan.NewService().Build(conflictplan.Options{Target: report.TargetRoot, Scope: opts.Scope, Harness: opts.Harness})
			if planErr == nil {
				if err := runConflictReview(plan); err != nil {
					return err
				}
			}
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
		if opts.Interactive && hasPendingMutations(report.Features) {
			return s.applyInteractive(opts, &report)
		}
		return s.apply(opts, &report, stdout)
	}
	return nil
}

func (s Service) BuildReport(opts Options) (Report, error) {
	if err := validateSetupTargetInput(opts.Target); err != nil {
		return Report{}, err
	}
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
	report.Features = s.planFeatures(target, opts.Scope, opts.Harness)
	return report, nil
}

func validateSetupTargetInput(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	normalized := filepath.ToSlash(filepath.Clean(strings.TrimSpace(raw)))
	placeholders := map[string]bool{
		"<dir>":                 true,
		"<repo>":                true,
		"/path/to/project":      true,
		"/ruta/a/proyecto":      true,
		"/ruta/a/tu/proyecto":   true,
		"ruta/a/proyecto":       true,
		"ruta/a/tu/proyecto":    true,
		"/path/to/your/project": true,
	}
	if placeholders[normalized] {
		return fmt.Errorf("--target apunta a un placeholder (%s); usa una ruta real, por ejemplo --target . o --target /Users/tu-usuario/proyecto", raw)
	}
	return nil
}

func (s Service) planFeatures(target string, scope assets.Scope, harness domain.HarnessConfig) []FeatureAction {
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
		if plan, planErr := conflictplan.NewService().Build(conflictplan.Options{Target: target, Scope: scope, Harness: harness}); planErr == nil && len(plan.Items) > 0 {
			features = append(features, action("install", "conflict", fmt.Sprintf("manifest ausente y %d conflicto(s) de assets no gestionados", len(plan.Items)), "lufy-ai conflicts plan --target <dir>"))
		} else {
			features = append(features, action("install", "apply", "No existe manifest de instalacion", "lufy-ai install --target <dir> --yes"))
		}
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
	for i := range report.Features {
		err := s.applyFeature(opts, report, i, stdout)
		if err != nil {
			return err
		}
	}
	report.Applied = true
	return nil
}

func (s Service) applyFeature(opts Options, report *Report, index int, stdout io.Writer) error {
	if index < 0 || index >= len(report.Features) {
		return nil
	}
	feature := &report.Features[index]
	if feature.Status != "apply" {
		return nil
	}
	scope := opts.Scope
	if scope == "" {
		scope = assets.ScopeProject
	}
	harness := opts.Harness.WithDefaults()
	if harness.Tool == "" {
		harness = domain.HarnessConfig{}.WithDefaults()
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
			return nil
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
	width     int
	height    int
	done      bool
	cancelled bool
	keys      checklistKeys
	spinner   spinner.Model
	help      help.Model
	progress  progress.Model
}

type checklistKeys struct {
	Up      key.Binding
	Down    key.Binding
	Toggle  key.Binding
	Confirm key.Binding
	Cancel  key.Binding
}

func (k checklistKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Toggle, k.Confirm, k.Cancel}
}

func (k checklistKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down}, {k.Toggle, k.Confirm, k.Cancel}}
}

func newChecklistModel(report Report) checklistModel {
	indexes := []int{}
	selected := map[string]bool{}
	for i, feature := range report.Features {
		if feature.Status == "apply" || feature.Status == "conflict" || feature.Status == "error" {
			indexes = append(indexes, i)
		}
		if feature.Status == "apply" {
			selected[feature.ID] = true
		}
	}
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	helpModel := help.New()
	helpModel.Width = 100
	progressModel := progress.New(progress.WithWidth(24), progress.WithSolidFill("63"))
	progressModel.EmptyColor = "240"
	return checklistModel{report: report, indexes: indexes, selected: selected, width: 100, height: 30, spinner: spin, help: helpModel, progress: progressModel, keys: checklistKeys{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "subir")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "bajar")),
		Toggle:  key.NewBinding(key.WithKeys(" "), key.WithHelp("espacio", "activar/desactivar")),
		Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "aplicar seleccion")),
		Cancel:  key.NewBinding(key.WithKeys("esc", "ctrl+c", "q"), key.WithHelp("esc/q", "cancelar")),
	}}
}

func (m checklistModel) Init() tea.Cmd { return m.spinner.Tick }

func (m checklistModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if spin, ok := msg.(spinner.TickMsg); ok {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(spin)
		return m, cmd
	}
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = clampInt(size.Width, 56, 140)
		m.height = clampInt(size.Height, 18, 60)
		m.help.Width = contentWidth(m.width)
		m.progress.Width = clampInt(contentWidth(m.width)/4, 12, 28)
		return m, nil
	}
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
			feature := m.report.Features[m.indexes[m.cursor]]
			if feature.Status == "apply" {
				m.selected[feature.ID] = !m.selected[feature.ID]
			}
		}
	}
	return m, nil
}

func (m checklistModel) View() string {
	width := contentWidth(m.width)
	header := m.dashboardHeader("checklist")
	footer := m.checklistFooter(width)
	consoleHeight := clampInt(m.height/5, 5, 8)
	bodyHeight := clampInt(m.height-lipgloss.Height(header)-consoleHeight-lipgloss.Height(footer)-3, 7, 24)
	console := m.consolePreview(width, consoleHeight)

	sections := []string{header}
	if m.isWide() && width >= 84 {
		leftWidth := clampInt((width-1)*38/100, 30, 44)
		rightWidth := maxInt(28, width-leftWidth-1)
		body := lipgloss.JoinHorizontal(lipgloss.Top, m.planPanel(leftWidth, bodyHeight), " ", m.currentStepPanel(rightWidth, bodyHeight))
		sections = append(sections, body)
	} else {
		planHeight := clampInt(bodyHeight/2+2, 6, 12)
		detailHeight := clampInt(bodyHeight-planHeight+3, 6, 12)
		sections = append(sections, m.planPanel(width, planHeight), m.currentStepPanel(width, detailHeight))
	}
	sections = append(sections, console, footer)
	return strings.Join(sections, "\n")
}

var (
	setupTitleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	setupSectionStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	setupMutedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	setupFaintStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Faint(true)
	setupOmittedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Faint(true).Strikethrough(true)
	setupHelpStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	setupSelectedStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	setupApplyStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	setupSkipStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	setupConflictStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	setupErrorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	setupDetailStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(1, 2)
	setupLogStyle        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
	setupPanelStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1)
	setupHeaderBoxStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("238")).Padding(0, 1)
	setupFooterBoxStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	setupBadgeReadyStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("22")).Background(lipgloss.Color("120")).Padding(0, 1)
	setupBadgeDoStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("63")).Padding(0, 1)
	setupBadgeSkipStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250")).Background(lipgloss.Color("238")).Padding(0, 1)
	setupBadgeBlockStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("160")).Padding(0, 1)
	setupBadgeErrorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("196")).Padding(0, 1)
	setupBadgeRunStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("39")).Padding(0, 1)
)

func (m checklistModel) isWide() bool { return m.width >= 90 }

func contentWidth(width int) int { return clampInt(width-4, 52, 136) }

func (m checklistModel) header() string {
	return m.dashboardHeader("checklist")
}

func (m checklistModel) dashboardHeader(mode string) string {
	versionBadge := setupMutedStyle.Render("CHECK OMITIDO")
	if m.report.Version != nil {
		versionBadge = setupMutedStyle.Render(versionLabelFor(*m.report.Version))
	}
	width := contentWidth(m.width)
	title := setupTitleStyle.Render("Lufy setup")
	modeLabel := setupMutedStyle.Render(mode)
	status := modeLabel + " " + versionBadge
	header := truncateText(title+" · "+status, width-4)
	target := setupMutedStyle.Render(truncateText(m.report.TargetRoot, width-4))
	metrics := setupFaintStyle.Render(m.metricLine(mode))
	lines := []string{header, target, metrics}
	if m.report.CheckNewFeatures {
		lines = append(lines, setupApplyStyle.Render("Nuevas capacidades incluidas"))
	}
	return renderDashboardBox("", width, 0, strings.Join(lines, "\n"), lipgloss.Color("238"), setupHeaderBoxStyle)
}

func (m checklistModel) sidebar(width int) string {
	return m.planContent(width, 0)
}

func (m checklistModel) planPanel(width int, height int) string {
	return renderDashboardBox("Setup plan", width, height, m.planContent(panelInnerWidth(width), height), lipgloss.Color("240"), setupPanelStyle)
}

func (m checklistModel) planContent(width int, height int) string {
	apply, skip, conflicts := featureCounts(m.report.Features)
	selected := m.selectedCount()
	ready := len(m.report.Features) - apply - conflicts
	percent := progressPercent(ready, len(m.report.Features))
	bar := m.progress.ViewAs(percent)
	lines := []string{}
	rowBudget := len(m.report.Features)
	if height > 0 {
		rowBudget = clampInt(height-5, 1, len(m.report.Features))
	}
	start := visibleWindowStart(len(m.report.Features), rowBudget, m.activeFeatureIndex())
	end := minInt(len(m.report.Features), start+rowBudget)
	if start > 0 {
		lines = append(lines, setupFaintStyle.Render("… pasos anteriores"))
	}
	for i := start; i < end; i++ {
		feature := m.report.Features[i]
		lines = append(lines, m.featureRow(i, feature, width))
	}
	if end < len(m.report.Features) {
		lines = append(lines, setupFaintStyle.Render("… mas pasos"))
	}
	lines = append(lines, "")
	lines = append(lines, setupMutedStyle.Render("Progress ")+bar+setupMutedStyle.Render(fmt.Sprintf("  %d/%d", ready, len(m.report.Features))))
	lines = append(lines, setupFaintStyle.Render(fmt.Sprintf("%d pendientes · %d seleccionadas · %d listas · %d conflictos", apply, selected, skip, conflicts)))
	return strings.Join(lines, "\n")
}

func (m checklistModel) featureRow(index int, feature FeatureAction, width int) string {
	active := false
	for pos, idx := range m.indexes {
		if idx == index && pos == m.cursor {
			active = true
			break
		}
	}
	selected := m.selected[feature.ID]
	icon := featureMarker(feature, selected)
	prefix := " "
	if active {
		prefix = "›"
	}
	label := strings.ToLower(featureStatusLabel(feature, selected))
	if label == "hacer" {
		label = "apply"
	}
	if label == "listo" {
		label = "ready"
	}
	if label == "bloqueado" {
		label = "block"
	}
	nameWidth := maxInt(10, width-14)
	body := fmt.Sprintf("%s%02d %s %-*s %s", prefix, index+1, icon, nameWidth, truncateText(feature.Name, nameWidth), label)
	if feature.Status == "apply" && !selected {
		body = setupOmittedStyle.Render(body)
	} else if feature.Status == "conflict" || feature.Status == "error" {
		body = setupConflictStyle.Render(body)
	} else if active {
		body = setupSelectedStyle.Render(body)
	} else if feature.Status != "apply" && feature.Status != "conflict" {
		body = setupFaintStyle.Render(body)
	}
	return body
}

func (m checklistModel) detailPanel(width int) string {
	return m.currentStepPanel(width, 0)
}

func (m checklistModel) currentStepPanel(width int, height int) string {
	idx := m.activeFeatureIndex()
	if idx < 0 || idx >= len(m.report.Features) {
		body := strings.Join([]string{
			setupApplyStyle.Render("Todo listo"),
			setupMutedStyle.Render("No hay pasos pendientes para aplicar."),
		}, "\n")
		return renderDashboardBox("Current step", width, height, body, lipgloss.Color("240"), setupPanelStyle)
	}
	feature := m.report.Features[idx]
	state := "no se aplicara"
	if m.selected[feature.ID] {
		state = "se aplicara al presionar enter"
	}
	inner := panelInnerWidth(width)
	status := featureStatusLabel(feature, m.selected[feature.ID])
	if height > 0 && height <= 8 {
		lines := []string{
			fmt.Sprintf("%s  %s  %s", featureMarker(feature, m.selected[feature.ID]), setupSectionStyle.Render(feature.Name), setupMutedStyle.Render(strings.ToLower(status))),
			wrapLine(feature.Reason, inner),
		}
		if feature.Recovery != "" {
			lines = append(lines, setupMutedStyle.Render("Comando"), wrapLine(feature.Recovery, inner))
		} else {
			lines = append(lines, setupMutedStyle.Render("Decision"), state)
		}
		border := lipgloss.Color("240")
		if feature.Status == "conflict" || feature.Status == "error" {
			border = lipgloss.Color("196")
		} else if m.selected[feature.ID] {
			border = lipgloss.Color("39")
		}
		return renderDashboardBox("Current step", width, height, limitLines(strings.Join(lines, "\n"), maxInt(1, height-2)), border, setupPanelStyle)
	}
	lines := []string{
		fmt.Sprintf("%s  %s  %s", featureMarker(feature, m.selected[feature.ID]), setupSectionStyle.Render(feature.Name), setupMutedStyle.Render(strings.ToLower(status))),
		"",
		setupMutedStyle.Render("Qué hará"),
		wrapLine(featureExplanation(feature.ID), inner),
		"",
		setupMutedStyle.Render("Impacto"),
		wrapLine(featureEffect(feature.ID), inner),
	}
	if feature.Recovery != "" {
		lines = append(lines, "", setupMutedStyle.Render("Comando"), wrapLine(feature.Recovery, inner))
	} else {
		lines = append(lines, "", setupMutedStyle.Render("Decision"), state)
	}
	lines = append(lines, "", setupMutedStyle.Render("Motivo"), wrapLine(feature.Reason, inner))
	border := lipgloss.Color("240")
	if feature.Status == "conflict" || feature.Status == "error" {
		border = lipgloss.Color("196")
	} else if m.selected[feature.ID] {
		border = lipgloss.Color("39")
	}
	return renderDashboardBox("Current step", width, height, limitLines(strings.Join(lines, "\n"), maxInt(0, height-2)), border, setupPanelStyle)
}

func (m checklistModel) consolePreview(width int, height int) string {
	selected := m.selectedCount()
	_, _, conflicts := featureCounts(m.report.Features)
	lines := []string{}
	switch {
	case conflicts > 0:
		lines = append(lines, setupConflictStyle.Render(fmt.Sprintf("Hay %d conflicto(s). Revisa recovery antes de aplicar.", conflicts)))
	case selected > 0:
		lines = append(lines, setupApplyStyle.Render(fmt.Sprintf("Listo para aplicar %d acciones.", selected)))
	default:
		lines = append(lines, setupMutedStyle.Render("No hay acciones seleccionadas para aplicar."))
	}
	if idx := m.activeFeatureIndex(); idx >= 0 && idx < len(m.report.Features) {
		feature := m.report.Features[idx]
		lines = append(lines, setupFaintStyle.Render("Actual: "+feature.Name+" · "+feature.Reason))
		if feature.Recovery != "" {
			lines = append(lines, setupFaintStyle.Render("Recovery: "+feature.Recovery))
		}
	}
	lines = append(lines, setupMutedStyle.Render("Enter aplica selección · espacio alterna paso · q cancela."))
	return renderDashboardBox("Console", width, height, limitLines(strings.Join(lines, "\n"), maxInt(1, height-2)), lipgloss.Color("240"), setupPanelStyle)
}

func (m checklistModel) checklistFooter(width int) string {
	status := "checklist"
	if selected := m.selectedCount(); selected > 0 {
		status = fmt.Sprintf("%d seleccionadas", selected)
	}
	line := "↑/↓ navegar · espacio activar/desactivar · enter aplicar · q salir"
	return setupFooterBoxStyle.Width(width).Render(truncateText(line+" · "+status, width))
}

func (m checklistModel) activeFeatureIndex() int {
	if len(m.indexes) == 0 {
		return -1
	}
	return m.indexes[clampInt(m.cursor, 0, len(m.indexes)-1)]
}

func (m checklistModel) metricLine(mode string) string {
	apply, skip, conflicts := featureCounts(m.report.Features)
	selected := m.selectedCount()
	if mode == "apply" {
		completed, failed, pending := applyCounts(m.report.Features)
		return fmt.Sprintf("%d ok · %d pendientes · %d errores", completed, pending, failed)
	}
	return fmt.Sprintf("%d pendientes · %d seleccionadas · %d listas · %d conflictos", apply, selected, skip, conflicts)
}

func (m checklistModel) selectedCount() int {
	count := 0
	for _, idx := range m.indexes {
		if m.selected[m.report.Features[idx].ID] {
			count++
		}
	}
	return count
}

func featureCounts(features []FeatureAction) (apply int, skip int, conflicts int) {
	for _, feature := range features {
		switch feature.Status {
		case "apply":
			apply++
		case "conflict":
			conflicts++
		default:
			skip++
		}
	}
	return apply, skip, conflicts
}

func featureMarker(feature FeatureAction, selected bool) string {
	switch feature.Status {
	case "apply":
		if selected {
			return "◆"
		}
		return "○"
	case "conflict":
		return "!"
	case "error":
		return "✕"
	default:
		return "✓"
	}
}

func featureStatusBadge(feature FeatureAction, selected bool) string {
	label := featureStatusLabel(feature, selected)
	switch label {
	case "LISTO":
		return setupBadgeReadyStyle.Render(label)
	case "HACER":
		return setupBadgeDoStyle.Render(label)
	case "OMITIDO":
		return setupBadgeSkipStyle.Render(label)
	case "BLOQUEADO":
		return setupBadgeBlockStyle.Render(label)
	case "ERROR":
		return setupBadgeErrorStyle.Render(label)
	default:
		return setupBadgeSkipStyle.Render(label)
	}
}

func featureStatusLabel(feature FeatureAction, selected bool) string {
	switch feature.Status {
	case "apply":
		if selected {
			return "HACER"
		}
		return "OMITIDO"
	case "conflict":
		return "BLOQUEADO"
	case "error":
		return "ERROR"
	default:
		return "LISTO"
	}
}

func featureExplanation(id string) string {
	switch id {
	case "layout":
		return "ordena la carpeta .lufy y migra rutas antiguas si existen."
	case "install":
		return "instala agentes, skills, comandos y politicas gestionadas por Lufy."
	case "project-config":
		return "crea .lufy/config/project.yaml, el perfil local del repositorio."
	case "stack-profile":
		return "detecta stacks y superficies para orientar validacion y reviews."
	case "sdd-methodology":
		return "configura el routing T1/T2/T3 y los limites de workflow."
	case "memory":
		return "prepara memoria Obsidian portable para aprendizajes durables."
	case "context-graph":
		return "genera un indice local para buscar contexto del proyecto."
	case "verify":
		return "comprueba que lo instalado quede coherente y sin drift obvio."
	default:
		return "capacidad local de Lufy."
	}
}

func featureEffect(id string) string {
	switch id {
	case "layout":
		return "puede escribir o mover archivos dentro de .lufy."
	case "install":
		return "puede escribir archivos bajo .opencode, .agents, openspec y .lufy gestionados."
	case "project-config", "stack-profile", "sdd-methodology":
		return "actualiza solo configuracion del proyecto bajo .lufy/config."
	case "memory":
		return "crea estructura de memoria bajo .lufy/memory."
	case "context-graph":
		return "crea o refresca datos locales bajo .lufy/context."
	case "verify":
		return "lee el proyecto y reporta problemas; no deberia cambiar archivos."
	default:
		return "aplica cambios locales relacionados con esta capacidad."
	}
}

func progressPercent(done, total int) float64 {
	if total <= 0 {
		return 0
	}
	percent := float64(done) / float64(total)
	if percent < 0 {
		return 0
	}
	if percent > 1 {
		return 1
	}
	return percent
}

func renderVersionLine(version versioncheck.Result) string {
	if version.Error != "" {
		return setupConflictStyle.Render("Version no verificada") + setupMutedStyle.Render(" · "+version.Error)
	}
	if version.UpdateAvailable {
		return setupConflictStyle.Render(fmt.Sprintf("Version local=%s latest=%s", version.CurrentVersion, version.LatestVersion)) + setupMutedStyle.Render(" · recomendado: lufy-ai upgrade --to "+version.LatestVersion)
	}
	if version.UpToDate {
		return setupApplyStyle.Render(fmt.Sprintf("Lufy AI esta al dia · %s", version.CurrentVersion))
	}
	return setupMutedStyle.Render(fmt.Sprintf("Version local=%s latest=%s", version.CurrentVersion, version.LatestVersion))
}

func versionBadgeFor(version versioncheck.Result) string {
	if version.Error != "" {
		return setupBadgeBlockStyle.Render("VERSION SIN CHECK")
	}
	if version.UpdateAvailable {
		return setupBadgeBlockStyle.Render("UPDATE DISPONIBLE")
	}
	if version.UpToDate {
		return setupBadgeReadyStyle.Render("VERSION OK")
	}
	if version.CurrentVersion != "" {
		return setupBadgeSkipStyle.Render(version.CurrentVersion)
	}
	return setupBadgeSkipStyle.Render("VERSION")
}

func versionLabelFor(version versioncheck.Result) string {
	if version.Error != "" {
		return "VERSION SIN CHECK"
	}
	if version.UpdateAvailable {
		return "UPDATE DISPONIBLE"
	}
	if version.UpToDate {
		return "VERSION OK"
	}
	if version.CurrentVersion != "" {
		return version.CurrentVersion
	}
	return "VERSION"
}

func applyStatusBadge(label string) string {
	switch label {
	case "LISTO":
		return setupBadgeReadyStyle.Render(label)
	case "EJECUTANDO":
		return setupBadgeRunStyle.Render(label)
	case "ERROR":
		return setupBadgeErrorStyle.Render(label)
	case "OMITIDO":
		return setupBadgeSkipStyle.Render(label)
	default:
		return setupBadgeSkipStyle.Render(label)
	}
}

func runChecklist(report Report) (map[string]bool, error) {
	model := newChecklistModel(report)
	finalModel, err := tea.NewProgram(model, tea.WithAltScreen()).Run()
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

func (s Service) applyInteractive(opts Options, report *Report) error {
	model := newApplyModel(s, opts, report)
	finalModel, err := tea.NewProgram(model, tea.WithAltScreen()).Run()
	if err != nil {
		return err
	}
	result, ok := finalModel.(applyModel)
	if !ok {
		return nil
	}
	if result.err != nil {
		return result.err
	}
	return nil
}

type applyModel struct {
	service Service
	opts    Options
	report  *Report
	events  chan tea.Msg

	current  int
	width    int
	height   int
	done     bool
	err      error
	started  bool
	logText  string
	view     viewport.Model
	keys     applyKeys
	spinner  spinner.Model
	help     help.Model
	progress progress.Model
}

type applyKeys struct {
	Up   key.Binding
	Down key.Binding
	Quit key.Binding
}

func (k applyKeys) ShortHelp() []key.Binding { return []key.Binding{k.Up, k.Down, k.Quit} }

func (k applyKeys) FullHelp() [][]key.Binding { return [][]key.Binding{{k.Up, k.Down}, {k.Quit}} }

type applyLogMsg string
type applyStepStartMsg int
type applyStepDoneMsg struct {
	index int
	err   error
}
type applyFinishedMsg struct{ err error }
type applyNoopMsg struct{}

func newApplyModel(service Service, opts Options, report *Report) applyModel {
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	logView := viewport.New(96, 16)
	logView.Style = lipgloss.NewStyle()
	helpModel := help.New()
	helpModel.Width = 100
	progressModel := progress.New(progress.WithWidth(24), progress.WithSolidFill("63"))
	progressModel.EmptyColor = "240"
	return applyModel{
		service:  service,
		opts:     opts,
		report:   report,
		events:   make(chan tea.Msg, 64),
		current:  -1,
		width:    100,
		height:   30,
		view:     logView,
		spinner:  spin,
		help:     helpModel,
		progress: progressModel,
		keys: applyKeys{
			Up:   key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "subir logs")),
			Down: key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "bajar logs")),
			Quit: key.NewBinding(key.WithKeys("esc", "ctrl+c", "q", "enter"), key.WithHelp("enter/q", "salir al terminar")),
		},
	}
}

func (m applyModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.startApply(), waitApplyEvent(m.events))
}

func (m applyModel) startApply() tea.Cmd {
	return func() tea.Msg {
		go m.runApply()
		return applyNoopMsg{}
	}
}

func waitApplyEvent(events <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-events
		if !ok {
			return applyFinishedMsg{}
		}
		return msg
	}
}

func (m applyModel) runApply() {
	var err error
	defer func() {
		m.events <- applyFinishedMsg{err: err}
		close(m.events)
	}()
	for i := range m.report.Features {
		if m.report.Features[i].Status != "apply" {
			continue
		}
		m.events <- applyStepStartMsg(i)
		stepErr := m.service.applyFeature(m.opts, m.report, i, channelWriter{events: m.events})
		m.events <- applyStepDoneMsg{index: i, err: stepErr}
		if stepErr != nil {
			err = stepErr
			return
		}
	}
	m.report.Applied = true
}

type channelWriter struct{ events chan<- tea.Msg }

func (w channelWriter) Write(p []byte) (int, error) {
	if len(p) > 0 {
		w.events <- applyLogMsg(string(p))
	}
	return len(p), nil
}

func (m applyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if spin, ok := msg.(spinner.TickMsg); ok {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(spin)
		if !m.done {
			return m, cmd
		}
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = clampInt(msg.Width, 56, 140)
		m.height = clampInt(msg.Height, 18, 60)
		m.view.Width = contentWidth(m.width)
		m.view.Height = clampInt(m.height-15, 8, 24)
		m.help.Width = contentWidth(m.width)
		m.progress.Width = clampInt(contentWidth(m.width)/4, 12, 28)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			if m.done || m.err != nil {
				return m, tea.Quit
			}
		case key.Matches(msg, m.keys.Up):
			m.view.LineUp(1)
		case key.Matches(msg, m.keys.Down):
			m.view.LineDown(1)
		}
	case applyNoopMsg:
		return m, waitApplyEvent(m.events)
	case applyStepStartMsg:
		m.started = true
		m.current = int(msg)
		m.appendLog(m.stepLogHeader(m.current))
		return m, waitApplyEvent(m.events)
	case applyLogMsg:
		m.appendLog(string(msg))
		return m, waitApplyEvent(m.events)
	case applyStepDoneMsg:
		m.current = msg.index
		if msg.err != nil {
			m.err = msg.err
			m.appendLog(fmt.Sprintf("✗ Falló: %s\n", msg.err.Error()))
		} else {
			m.appendLog(fmt.Sprintf("✓ Completado: %s\n", m.report.Features[msg.index].Name))
		}
		return m, waitApplyEvent(m.events)
	case applyFinishedMsg:
		m.done = true
		if msg.err != nil {
			m.err = msg.err
			m.appendLog("\nSetup detenido. Revisa el error y los logs antes de salir.\n")
		} else {
			m.appendLog("\nSetup completado. enter/q para salir.\n")
		}
	}
	return m, nil
}

func (m *applyModel) appendLog(text string) {
	m.logText += text
	m.view.SetContent(m.logText)
	m.view.GotoBottom()
}

func (m applyModel) stepLogHeader(index int) string {
	if index < 0 || index >= len(m.report.Features) {
		return "\n▶ Ejecutando paso\n"
	}
	separator := ""
	if strings.TrimSpace(m.logText) != "" {
		separator = "\n"
	}
	feature := m.report.Features[index]
	return fmt.Sprintf("%s▶ Ejecutando: %s\n  estado: ejecutando · %s\n", separator, feature.Name, feature.Reason)
}

func (m applyModel) View() string {
	width := contentWidth(m.width)
	header := m.applyHeader()
	footer := m.applyFooter(width)
	consoleHeight := clampInt(m.height/3, 8, 18)
	bodyHeight := clampInt(m.height-lipgloss.Height(header)-consoleHeight-lipgloss.Height(footer)-3, 7, 20)
	sections := []string{header}
	if m.width >= 90 && width >= 84 {
		leftWidth := clampInt((width-1)*38/100, 30, 44)
		rightWidth := maxInt(28, width-leftWidth-1)
		body := lipgloss.JoinHorizontal(lipgloss.Top, m.applyPlanPanel(leftWidth, bodyHeight), " ", m.applyCurrentStepPanel(rightWidth, bodyHeight))
		sections = append(sections, body)
	} else {
		planHeight := clampInt(bodyHeight/2+2, 6, 12)
		detailHeight := clampInt(bodyHeight-planHeight+3, 6, 12)
		sections = append(sections, m.applyPlanPanel(width, planHeight), m.applyCurrentStepPanel(width, detailHeight))
	}
	sections = append(sections, m.applyConsolePanel(width, consoleHeight), footer)
	return strings.Join(sections, "\n")
}

func (m applyModel) applyHeader() string {
	status := "Aplicando pasos seleccionados"
	badge := setupMutedStyle.Render("APLICANDO")
	if m.err != nil {
		status = "Setup detenido por error"
		badge = setupErrorStyle.Render("ERROR")
	} else if m.done {
		status = "Setup completado"
		badge = setupApplyStyle.Render("LISTO")
	}
	width := contentWidth(m.width)
	title := setupTitleStyle.Render("Lufy setup")
	mode := setupMutedStyle.Render("apply")
	right := mode + " " + badge
	header := truncateText(title+" · "+right, width-4)
	body := strings.Join([]string{header, setupMutedStyle.Render(truncateText(m.report.TargetRoot, width-4)), setupFaintStyle.Render(status + " · " + m.applyMetricLine())}, "\n")
	return renderDashboardBox("", width, 0, body, lipgloss.Color("238"), setupHeaderBoxStyle)
}

func (m applyModel) applySummary() string {
	completed := 0
	failed := 0
	pending := 0
	for _, feature := range m.report.Features {
		if feature.Status != "apply" {
			continue
		}
		switch {
		case feature.Error != "":
			failed++
		case feature.Applied:
			completed++
		default:
			pending++
		}
	}
	total := completed + failed + pending
	return fmt.Sprintf("%s\n%s  %s  %s  %s",
		setupMutedStyle.Render("Progreso de aplicacion"),
		m.progress.ViewAs(progressPercent(completed+failed, total)),
		setupApplyStyle.Render(fmt.Sprintf("%d ok", completed)),
		setupConflictStyle.Render(fmt.Sprintf("%d error", failed)),
		setupMutedStyle.Render(fmt.Sprintf("%d pendientes", pending)),
	)
}

func (m applyModel) applyPlanPanel(width int, height int) string {
	contentWidth := panelInnerWidth(width)
	rows := []string{}
	rowBudget := len(m.report.Features)
	if height > 0 {
		rowBudget = clampInt(height-5, 1, len(m.report.Features))
	}
	start := visibleWindowStart(len(m.report.Features), rowBudget, m.current)
	end := minInt(len(m.report.Features), start+rowBudget)
	if start > 0 {
		rows = append(rows, setupFaintStyle.Render("… pasos anteriores"))
	}
	for i := start; i < end; i++ {
		rows = append(rows, m.applyCompactRow(i, m.report.Features[i], contentWidth))
	}
	if end < len(m.report.Features) {
		rows = append(rows, setupFaintStyle.Render("… mas pasos"))
	}
	rows = append(rows, "", m.applySummary())
	return renderDashboardBox("Setup plan", width, height, limitLines(strings.Join(rows, "\n"), maxInt(1, height-2)), lipgloss.Color("240"), setupPanelStyle)
}

func (m applyModel) applyCurrentStepPanel(width int, height int) string {
	idx := m.current
	if idx < 0 || idx >= len(m.report.Features) {
		idx = firstApplyFeature(m.report.Features)
	}
	if idx < 0 || idx >= len(m.report.Features) {
		body := strings.Join([]string{setupApplyStyle.Render("Setup completado"), setupMutedStyle.Render("No quedan pasos aplicables.")}, "\n")
		return renderDashboardBox("Current step", width, height, body, lipgloss.Color("240"), setupPanelStyle)
	}
	feature := m.report.Features[idx]
	inner := panelInnerWidth(width)
	label := "pendiente"
	border := lipgloss.Color("240")
	if idx == m.current && !m.done && feature.Status == "apply" && feature.Error == "" && !feature.Applied {
		label = "ejecutando"
		border = lipgloss.Color("39")
	}
	if feature.Applied {
		label = "listo"
		border = lipgloss.Color("42")
	}
	if feature.Error != "" || (m.err != nil && idx == m.current) {
		label = "error"
		border = lipgloss.Color("196")
	}
	lines := []string{
		fmt.Sprintf("%s  %s  %s", applyMarkerFor(feature, idx == m.current && !m.done, m.spinner.View()), setupSectionStyle.Render(feature.Name), setupMutedStyle.Render(label)),
		"",
		setupMutedStyle.Render("Detalle"),
		wrapLine(featureExplanation(feature.ID), inner),
		"",
		setupMutedStyle.Render("Motivo"),
		wrapLine(feature.Reason, inner),
	}
	if feature.Error != "" {
		lines = append(lines, "", setupMutedStyle.Render("Error"), wrapLine(feature.Error, inner))
	}
	return renderDashboardBox("Current step", width, height, limitLines(strings.Join(lines, "\n"), maxInt(1, height-2)), border, setupPanelStyle)
}

func (m applyModel) applyConsolePanel(width int, height int) string {
	view := m.view
	view.Width = panelInnerWidth(width)
	view.Height = clampInt(height-4, 3, 14)
	body := view.View()
	if strings.TrimSpace(m.logText) == "" {
		body = setupMutedStyle.Render("Consola Lufy lista. Los logs reales apareceran aqui.")
	}
	if m.err != nil {
		body += "\n" + setupConflictStyle.Render("Fallo: "+m.err.Error())
	} else if m.done {
		body += "\n" + setupApplyStyle.Render("Resumen final: setup completado. enter/q para salir.")
	}
	return renderDashboardBox("Console", width, height, limitLines(body, maxInt(1, height-2)), lipgloss.Color("240"), setupPanelStyle)
}

func (m applyModel) applyFooter(width int) string {
	line := "↑/↓ scroll logs · enter/q salir al terminar · " + m.applyMetricLine()
	return setupFooterBoxStyle.Width(width).Render(truncateText(line, width))
}

func (m applyModel) applyMetricLine() string {
	completed, failed, pending := applyCounts(m.report.Features)
	return fmt.Sprintf("%d ok · %d pendientes · %d errores", completed, pending, failed)
}

func (m applyModel) applyCompactRow(index int, feature FeatureAction, width int) string {
	running := index == m.current && !m.done && feature.Status == "apply" && feature.Error == "" && !feature.Applied
	marker := applyMarkerFor(feature, running, m.spinner.View())
	label := "omitido"
	style := setupFaintStyle
	if feature.Status == "apply" {
		label = "pending"
		style = setupMutedStyle
	}
	if running {
		label = "running"
		style = setupSelectedStyle
	}
	if feature.Applied {
		label = "ready"
		style = setupApplyStyle
	}
	if feature.Error != "" {
		label = "error"
		style = setupConflictStyle
	}
	nameWidth := maxInt(10, width-15)
	body := fmt.Sprintf(" %02d %s %-*s %s", index+1, marker, nameWidth, truncateText(feature.Name, nameWidth), label)
	return style.Render(body)
}

func (m applyModel) applyRow(index int, feature FeatureAction) string {
	marker := "○"
	label := "OMITIDO"
	style := setupFaintStyle
	if feature.Status == "apply" {
		marker = "○"
		label = "PENDIENTE"
		style = setupMutedStyle
		if index == m.current && !feature.Applied && feature.Error == "" && !m.done {
			marker = m.spinner.View()
			label = "EJECUTANDO"
			style = setupSelectedStyle
		}
		if feature.Applied {
			marker = "✓"
			label = "LISTO"
			style = setupApplyStyle
		}
		if feature.Error != "" {
			marker = "✕"
			label = "ERROR"
			style = setupConflictStyle
		}
	}
	if feature.Status != "apply" && reasonLooksReady(feature.Reason) {
		label = "LISTO"
		marker = "✓"
	}
	body := fmt.Sprintf("%s  %s  %s", marker, style.Render(feature.Name), setupMutedStyle.Render(label))
	if feature.Status == "apply" || feature.Error != "" {
		body += "\n" + setupMutedStyle.Render("   "+feature.Reason)
	}
	rowStyle := lipgloss.NewStyle().PaddingLeft(1)
	if feature.Error != "" {
		return rowStyle.Foreground(lipgloss.Color("196")).Render(body)
	}
	if index == m.current && !m.done {
		return rowStyle.Foreground(lipgloss.Color("39")).Render(body)
	}
	return rowStyle.Render(body)
}

func reasonLooksReady(reason string) bool {
	normalized := strings.ToLower(reason)
	return strings.Contains(normalized, "ready") || strings.Contains(normalized, "listo") || strings.Contains(normalized, "lista")
}

func renderDashboardBox(title string, width int, height int, body string, border lipgloss.Color, base lipgloss.Style) string {
	style := base.BorderForeground(border)
	frameW, frameH := style.GetFrameSize()
	innerWidth := maxInt(1, width-frameW)
	if height > 0 {
		style = style.Height(maxInt(1, height-frameH))
	}
	style = style.Width(innerWidth)
	content := body
	if title != "" {
		content = setupSectionStyle.Render(title) + "\n" + body
	}
	return style.Render(content)
}

func panelInnerWidth(width int) int {
	frameW, _ := setupPanelStyle.GetFrameSize()
	return maxInt(12, width-frameW)
}

func visibleWindowStart(total, budget, active int) int {
	if total <= 0 || budget <= 0 || budget >= total {
		return 0
	}
	if active < 0 {
		active = 0
	}
	if active >= total {
		active = total - 1
	}
	half := budget / 2
	start := active - half
	if start < 0 {
		return 0
	}
	if start+budget > total {
		return total - budget
	}
	return start
}

func limitLines(text string, maxLines int) string {
	if maxLines <= 0 {
		return text
	}
	lines := strings.Split(text, "\n")
	if len(lines) <= maxLines {
		return text
	}
	if maxLines == 1 {
		return truncateText(lines[0], 80)
	}
	trimmed := append([]string{}, lines[:maxLines-1]...)
	trimmed = append(trimmed, setupFaintStyle.Render("…"))
	return strings.Join(trimmed, "\n")
}

func applyCounts(features []FeatureAction) (completed int, failed int, pending int) {
	for _, feature := range features {
		if feature.Status != "apply" {
			continue
		}
		switch {
		case feature.Error != "":
			failed++
		case feature.Applied:
			completed++
		default:
			pending++
		}
	}
	return completed, failed, pending
}

func firstApplyFeature(features []FeatureAction) int {
	for i, feature := range features {
		if feature.Status == "apply" {
			return i
		}
	}
	return -1
}

func applyMarkerFor(feature FeatureAction, running bool, spinner string) string {
	switch {
	case feature.Error != "":
		return "✕"
	case feature.Applied:
		return "✓"
	case running:
		return spinner
	case feature.Status == "apply":
		return "◆"
	default:
		return "✓"
	}
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncateText(text string, width int) string {
	if width <= 0 || lipgloss.Width(text) <= width {
		return text
	}
	if width <= 1 {
		return "…"
	}
	runes := []rune(text)
	for len(runes) > 0 && lipgloss.Width(string(runes))+1 > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

func wrapLine(text string, width int) string {
	width = maxInt(12, width)
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	lines := []string{}
	current := words[0]
	for _, word := range words[1:] {
		candidate := current + " " + word
		if lipgloss.Width(candidate) > width {
			lines = append(lines, current)
			current = word
			continue
		}
		current = candidate
	}
	lines = append(lines, current)
	return strings.Join(lines, "\n")
}

type conflictReviewModel struct {
	plan       conflictplan.Report
	groupIndex int
	itemIndex  int
	quit       bool
	keys       conflictReviewKeys
}

type conflictReviewKeys struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	Quit  key.Binding
}

func newConflictReviewModel(plan conflictplan.Report) conflictReviewModel {
	return conflictReviewModel{plan: plan, keys: conflictReviewKeys{
		Up:    key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "item anterior")),
		Down:  key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "item siguiente")),
		Left:  key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "grupo anterior")),
		Right: key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "grupo siguiente")),
		Quit:  key.NewBinding(key.WithKeys("enter", "esc", "ctrl+c", "q"), key.WithHelp("enter/q", "salir")),
	}}
}

func (m conflictReviewModel) Init() tea.Cmd { return nil }

func (m conflictReviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(keyMsg, m.keys.Quit):
		m.quit = true
		return m, tea.Quit
	case key.Matches(keyMsg, m.keys.Left):
		if m.groupIndex > 0 {
			m.groupIndex--
			m.itemIndex = 0
		}
	case key.Matches(keyMsg, m.keys.Right):
		if m.groupIndex < len(m.plan.Groups)-1 {
			m.groupIndex++
			m.itemIndex = 0
		}
	case key.Matches(keyMsg, m.keys.Up):
		if m.itemIndex > 0 {
			m.itemIndex--
		}
	case key.Matches(keyMsg, m.keys.Down):
		if m.itemIndex < len(m.currentItems())-1 {
			m.itemIndex++
		}
	}
	return m, nil
}

func (m conflictReviewModel) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("Lufy setup: conflictos detectados")
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", title)
	fmt.Fprintf(&b, "Target: %s\n", m.plan.TargetRoot)
	fmt.Fprintf(&b, "Conflictos: %d · grupos: %d · legacy/deprecated: %d\n\n", m.plan.Summary.Conflicts, m.plan.Summary.Groups, m.plan.Summary.LegacyDeprecated)
	if len(m.plan.Groups) == 0 {
		fmt.Fprintln(&b, "No hay grupos de conflicto. enter/q para salir.")
		return b.String()
	}
	group := m.plan.Groups[m.groupIndex]
	fmt.Fprintf(&b, "Grupo %d/%d: %s riesgo=%s parallel_group=%s\n", m.groupIndex+1, len(m.plan.Groups), group.Category, group.Risk, group.ParallelGroup)
	items := m.currentItems()
	for i, item := range items {
		cursor := " "
		if i == m.itemIndex {
			cursor = ">"
		}
		line := fmt.Sprintf("%s %s [%s] recomendacion=%s", cursor, item.Path, item.Risk, item.Recommendation)
		if i == m.itemIndex {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render(line)
		}
		fmt.Fprintln(&b, line)
	}
	if len(items) > 0 {
		selected := items[m.itemIndex]
		fmt.Fprintf(&b, "\nDetalle: %s\n", selected.Reason)
		fmt.Fprintf(&b, "Acciones disponibles: %s\n", strings.Join(selected.AvailableActions, ", "))
	}
	if len(m.plan.LegacyDeprecated) > 0 {
		fmt.Fprintln(&b, "\nLegacy/deprecated: ejecutar migrate-layout antes de borrar rutas legacy.")
	}
	fmt.Fprintln(&b, "\nRecovery: lufy-ai conflicts plan --target <dir> --json")
	fmt.Fprintf(&b, "%s · %s · %s · %s · %s\n", m.keys.Left.Help().Key, m.keys.Right.Help().Key, m.keys.Up.Help().Key, m.keys.Down.Help().Key, m.keys.Quit.Help().Key)
	return b.String()
}

func (m conflictReviewModel) currentItems() []conflictplan.Item {
	if len(m.plan.Groups) == 0 || m.groupIndex >= len(m.plan.Groups) {
		return nil
	}
	category := m.plan.Groups[m.groupIndex].Category
	items := []conflictplan.Item{}
	for _, item := range m.plan.Items {
		if item.Category == category {
			items = append(items, item)
		}
	}
	return items
}

func runConflictReview(plan conflictplan.Report) error {
	model := newConflictReviewModel(plan)
	_, err := tea.NewProgram(model).Run()
	return err
}
