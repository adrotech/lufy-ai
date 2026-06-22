package governance

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	contextapp "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/application"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/harnesscatalog"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/memory"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/status"
)

type Options struct {
	Target string
	JSON   bool
	Scope  assets.Scope
}

type PinOptions struct {
	Target string
	Path   string
	Reason string
}

type InfoReport struct {
	TargetRoot            string   `json:"targetRoot"`
	Installed             bool     `json:"installed"`
	Tool                  string   `json:"tool"`
	MethodologyByTier     any      `json:"methodologyByTier,omitempty"`
	CatalogAssets         int      `json:"catalogAssets"`
	ManifestAssets        int      `json:"manifestAssets,omitempty"`
	Pinned                int      `json:"pinned,omitempty"`
	ConflictsPending      int      `json:"conflictsPending,omitempty"`
	SourceRootFingerprint string   `json:"sourceRootFingerprint,omitempty"`
	ProjectConfig         string   `json:"projectConfig"`
	Stacks                []string `json:"stacks,omitempty"`
	Surfaces              []string `json:"surfaces,omitempty"`
}

type DoctorReport struct {
	OK         bool          `json:"ok"`
	TargetRoot string        `json:"targetRoot"`
	Checks     []DoctorCheck `json:"checks"`
}

type DoctorCheck struct {
	Level   string `json:"level"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

type Service struct{}

func NewService() Service {
	return Service{}
}

func (s Service) Info(opts Options, stdout io.Writer) error {
	report, err := s.BuildInfo(opts)
	if err != nil {
		return err
	}
	if opts.JSON {
		return writeJSON(stdout, report)
	}
	fmt.Fprintf(stdout, "Info para %s\n", report.TargetRoot)
	fmt.Fprintf(stdout, "Instalado: %s\n", yesNo(report.Installed))
	fmt.Fprintf(stdout, "Tool: %s\n", report.Tool)
	fmt.Fprintf(stdout, "Assets catalogo efectivo: %d\n", report.CatalogAssets)
	if report.Installed {
		fmt.Fprintf(stdout, "Assets manifest: %d\n", report.ManifestAssets)
		fmt.Fprintf(stdout, "Pinned/frozen: %d\n", report.Pinned)
		fmt.Fprintf(stdout, "Conflictos pendientes: %d\n", report.ConflictsPending)
		fmt.Fprintf(stdout, "Fingerprint catalogo: %s\n", short(report.SourceRootFingerprint))
	}
	fmt.Fprintf(stdout, "Project config: %s\n", report.ProjectConfig)
	if len(report.Stacks) > 0 {
		fmt.Fprintf(stdout, "Stacks: %s\n", strings.Join(report.Stacks, ", "))
	}
	if len(report.Surfaces) > 0 {
		fmt.Fprintf(stdout, "Surfaces: %s\n", strings.Join(report.Surfaces, ", "))
	}
	return nil
}

func (s Service) Doctor(opts Options, stdout io.Writer) error {
	report, err := s.BuildDoctor(opts)
	if err != nil {
		return err
	}
	if opts.JSON {
		if err := writeJSON(stdout, report); err != nil {
			return err
		}
		if !report.OK {
			return fmt.Errorf("doctor detectó problemas")
		}
		return nil
	}
	fmt.Fprintf(stdout, "Doctor para %s\n", report.TargetRoot)
	for _, check := range report.Checks {
		path := ""
		if check.Path != "" {
			path = " " + check.Path
		}
		fmt.Fprintf(stdout, "- [%s]%s %s\n", check.Level, path, check.Message)
	}
	if !report.OK {
		return fmt.Errorf("doctor detectó problemas")
	}
	fmt.Fprintln(stdout, "Doctor OK")
	return nil
}

func (s Service) Pin(opts PinOptions, stdout io.Writer) error {
	target, targetRel, st, release, err := loadMutableAsset(opts.Target, opts.Path)
	if err != nil {
		return err
	}
	defer release()
	now := time.Now().UTC().Format(time.RFC3339)
	for i := range st.Assets {
		if st.Assets[i].TargetRel != targetRel {
			continue
		}
		st.Assets[i].Pinned = true
		st.Assets[i].PinnedAt = now
		st.Assets[i].PinnedReason = opts.Reason
		st.Assets[i].LastAction = "pin"
		break
	}
	st.UpdatedAt = now
	if err := state.WriteAtomic(target, *st); err != nil {
		return err
	}
	if opts.Reason == "" {
		fmt.Fprintf(stdout, "Asset pinned/frozen: %s\n", targetRel)
	} else {
		fmt.Fprintf(stdout, "Asset pinned/frozen: %s (%s)\n", targetRel, opts.Reason)
	}
	return nil
}

func (s Service) Unpin(opts PinOptions, stdout io.Writer) error {
	target, targetRel, st, release, err := loadMutableAsset(opts.Target, opts.Path)
	if err != nil {
		return err
	}
	defer release()
	now := time.Now().UTC().Format(time.RFC3339)
	for i := range st.Assets {
		if st.Assets[i].TargetRel != targetRel {
			continue
		}
		st.Assets[i].Pinned = false
		st.Assets[i].PinnedAt = ""
		st.Assets[i].PinnedReason = ""
		st.Assets[i].LastAction = "unpin"
		break
	}
	st.UpdatedAt = now
	if err := state.WriteAtomic(target, *st); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Asset unpinned/unfrozen: %s\n", targetRel)
	return nil
}

func (s Service) BuildInfo(opts Options) (InfoReport, error) {
	target, cfg, cfgStatus, st, err := loadContext(opts.Target)
	if err != nil {
		return InfoReport{}, err
	}
	harness := harnessFromContext(cfg, st)
	catalogAssets, err := effectiveCatalogAssets(harness)
	if err != nil {
		return InfoReport{}, err
	}
	report := InfoReport{
		TargetRoot:        target,
		Tool:              string(harness.Tool),
		MethodologyByTier: harness.MethodologyByTier,
		CatalogAssets:     catalogAssets,
		ProjectConfig:     cfgStatus,
	}
	if st != nil {
		report.Installed = true
		report.ManifestAssets = len(st.Assets)
		report.SourceRootFingerprint = st.SourceRootFingerprint
		statusReport, err := status.NewService().Build(target, false, opts.Scope)
		if err != nil {
			return InfoReport{}, err
		}
		report.Pinned = statusReport.Pinned
		report.ConflictsPending = statusReport.ConflictsPending
	}
	if cfg != nil {
		for _, stack := range cfg.Stacks {
			report.Stacks = append(report.Stacks, stack.ID)
		}
		for _, surface := range cfg.ProjectProfile.Surfaces {
			report.Surfaces = append(report.Surfaces, surface.ID+":"+surface.Type)
		}
	}
	return report, nil
}

func (s Service) BuildDoctor(opts Options) (DoctorReport, error) {
	target, cfg, cfgStatus, st, err := loadContext(opts.Target)
	if err != nil {
		return DoctorReport{}, err
	}
	report := DoctorReport{OK: true, TargetRoot: target}
	emit := func(level, path, message string) {
		if level == "fail" {
			report.OK = false
		}
		report.Checks = append(report.Checks, DoctorCheck{Level: level, Path: path, Message: message})
	}
	emit("ok", "", "target resuelto")
	if cfg == nil {
		emit("fail", projectconfig.ProjectConfigPath, cfgStatus)
	} else {
		emit("ok", projectconfig.ProjectConfigPath, fmt.Sprintf("project config parseable; stacks=%d surfaces=%d", len(cfg.Stacks), len(cfg.ProjectProfile.Surfaces)))
	}
	reportMemoryDoctor(target, emit)
	reportContextDoctor(target, emit)
	reportOpenCodeMemoryHookDoctor(target, emit)
	if st == nil {
		emit("fail", state.Path(target), "falta manifest de instalación")
		return report, nil
	}
	emit("ok", ".lufy/managed-state/install-state.json", fmt.Sprintf("manifest schema=%d assets=%d", st.SchemaVersion, len(st.Assets)))
	statusReport, err := status.NewService().Build(target, false, opts.Scope)
	if err != nil {
		emit("fail", ".lufy/managed-state/install-state.json", err.Error())
		return report, nil
	}
	if !statusReport.Installed {
		emit("fail", ".lufy/managed-state/install-state.json", "target no instalado")
		return report, nil
	}
	if statusReport.Missing > 0 || statusReport.Drifted > 0 || statusReport.Errors > 0 {
		emit("fail", "", fmt.Sprintf("estado con drift=%d faltantes=%d errores=%d", statusReport.Drifted, statusReport.Missing, statusReport.Errors))
	} else {
		emit("ok", "", "sin drift, faltantes ni errores de assets")
	}
	if statusReport.Pinned > 0 {
		emit("info", "", fmt.Sprintf("assets pinned/frozen=%d; sync los preserva sin modificar", statusReport.Pinned))
	}
	if statusReport.ConflictsPending > 0 {
		emit("fail", "", fmt.Sprintf("conflictos pendientes .lufy-new=%d; ejecuta lufy-ai merge", statusReport.ConflictsPending))
	}
	return report, nil
}

func reportMemoryDoctor(target string, emit func(level, path, message string)) {
	report, err := memory.NewService().BuildStatus(memory.Options{Target: target})
	if err != nil {
		emit("warn", projectconfig.ProjectConfigPath, fmt.Sprintf("memoria no evaluable: %s", err.Error()))
		return
	}
	if !report.Status.Initialized {
		emit("info", report.Root, "memoria Obsidian no inicializada; ejecuta lufy-ai memory init")
		return
	}
	if report.Status.BrokenBacklinks > 0 {
		emit("warn", report.Root, fmt.Sprintf("memoria con backlinks rotos=%d", report.Status.BrokenBacklinks))
	}
	if report.Status.Drafts > 0 {
		emit("info", report.Root, fmt.Sprintf("drafts pendientes=%d", report.Status.Drafts))
	}
	for _, check := range report.Checks {
		if check.Level == "fail" || check.Level == "warn" {
			level := "warn"
			emit(level, check.Path, "memoria: "+check.Message)
		}
	}
	if report.Status.BrokenBacklinks == 0 {
		emit("ok", report.Root, fmt.Sprintf("memoria Obsidian ok notas=%d drafts=%d", report.Status.Notes, report.Status.Drafts))
	}
}

func reportContextDoctor(target string, emit func(level, path, message string)) {
	status := contextapp.NewService().Status(target)
	path := status.GraphPath
	if path == "" {
		path = filepath.Join(".lufy", "context", "graph.json")
	} else if rel, err := filepath.Rel(target, path); err == nil {
		path = filepath.ToSlash(rel)
	}
	switch status.Status {
	case "ready":
		emit("ok", path, fmt.Sprintf("context graph ready sources=%d nodes=%d edges=%d", status.Sources, status.Nodes, status.Edges))
	case "stale":
		emit("warn", path, fmt.Sprintf("context graph stale: %s; recovery: %s", status.Reason, status.Recovery))
	default:
		emit("warn", path, fmt.Sprintf("context graph not_available: %s; recovery: %s", status.Reason, status.Recovery))
	}
}

func reportOpenCodeMemoryHookDoctor(target string, emit func(level, path, message string)) {
	hooks := []string{
		filepath.Join(".opencode", "hooks", "memory-orient.sh"),
		filepath.Join(".opencode", "hooks", "memory-validate.sh"),
	}
	for _, rel := range hooks {
		if !regularFileForGovernance(filepath.Join(target, rel)) {
			emit("warn", filepath.ToSlash(rel), "hook de memoria no instalado; ejecuta lufy-ai sync --tool opencode --scope project")
			return
		}
	}
	plugin := filepath.Join(".opencode", "plugins", "lufy-memory-context.ts")
	if !regularFileForGovernance(filepath.Join(target, plugin)) {
		emit("warn", filepath.ToSlash(plugin), "plugin lifecycle de memoria/contexto no instalado; ejecuta lufy-ai sync --tool opencode --scope project")
		return
	}
	emit("ok", filepath.ToSlash(plugin), "OpenCode cargará plugin local para orientación y validación best-effort de memoria")
}

func regularFileForGovernance(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0
}

func loadMutableAsset(targetValue, pathValue string) (string, string, *state.InstallState, func(), error) {
	target, err := platform.ResolveTargetPath(targetValue)
	if err != nil {
		return "", "", nil, nil, err
	}
	targetRel, err := platform.EnsureRelativeSafe(pathValue)
	if err != nil {
		return "", "", nil, nil, err
	}
	lock, err := platform.AcquireLock(target)
	if err != nil {
		return "", "", nil, nil, err
	}
	release := func() { lock.Release() }
	st, err := state.Load(target)
	if err != nil {
		release()
		return "", "", nil, nil, err
	}
	if st == nil {
		release()
		return "", "", nil, nil, fmt.Errorf("pin/unpin requiere %s", state.Path(target))
	}
	if _, ok := st.AssetMap()[targetRel]; !ok {
		release()
		return "", "", nil, nil, fmt.Errorf("asset no gestionado en install-state: %s", targetRel)
	}
	return target, targetRel, st, release, nil
}

func loadContext(target string) (string, *projectconfig.ProjectConfig, string, *state.InstallState, error) {
	resolved, err := platform.ResolveTargetPath(target)
	if err != nil {
		return "", nil, "", nil, err
	}
	cfgPath, err := projectconfig.ExistingPath(resolved)
	if err != nil {
		return "", nil, "", nil, err
	}
	var cfg *projectconfig.ProjectConfig
	cfgStatus := "missing"
	loaded, err := projectconfig.Load(cfgPath)
	if err == nil {
		cfg = &loaded
		cfgStatus = "ok"
	} else if os.IsNotExist(err) {
		cfgStatus = "missing"
	} else {
		cfgStatus = "invalid: " + err.Error()
	}
	st, err := state.Load(resolved)
	if err != nil {
		return "", nil, "", nil, err
	}
	return resolved, cfg, cfgStatus, st, nil
}

func harnessFromContext(cfg *projectconfig.ProjectConfig, st *state.InstallState) domain.HarnessConfig {
	if st != nil {
		return domain.HarnessConfig{Tool: st.Tool, MethodologyByTier: st.MethodologyByTier}.WithDefaults()
	}
	if cfg != nil {
		return domain.HarnessConfig{Tool: cfg.Tool, MethodologyByTier: cfg.MethodologyByTier}.WithDefaults()
	}
	return domain.DefaultHarnessConfig()
}

func effectiveCatalogAssets(harness domain.HarnessConfig) (int, error) {
	catalog, err := assets.BuildEmbeddedCatalog()
	if err != nil {
		return 0, err
	}
	effective, err := harnesscatalog.Effective(catalog, harness)
	if err != nil {
		return 0, err
	}
	return len(effective.Assets), nil
}

func writeJSON(stdout io.Writer, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "%s\n", body)
	return err
}

func yesNo(value bool) string {
	if value {
		return "sí"
	}
	return "no"
}

func short(value string) string {
	if len(value) <= 12 {
		return value
	}
	return value[:12]
}
