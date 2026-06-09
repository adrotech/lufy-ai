package verify

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/agentsref"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/harnesscatalog"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/memory"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/toolruntime"
)

type Options struct {
	Target                string
	NoEngram              bool
	JSON                  bool
	Quiet                 bool
	Verbose               bool
	Deep                  bool
	AllowCatalogNewAssets bool
	AllowMissingAgentsRef bool
	Scope                 assets.Scope
	ExpectedTool          domain.ToolID
}

type Report struct {
	OK                bool    `json:"ok"`
	TargetRoot        string  `json:"targetRoot"`
	Scope             string  `json:"scope,omitempty"`
	GlobalRoot        string  `json:"globalRoot,omitempty"`
	SchemaVersion     int     `json:"schemaVersion,omitempty"`
	Tool              string  `json:"tool,omitempty"`
	MethodologyByTier any     `json:"methodologyByTier,omitempty"`
	Assets            int     `json:"assets,omitempty"`
	Failures          int     `json:"failures"`
	Warnings          int     `json:"warnings"`
	Infos             int     `json:"infos"`
	Checks            []Check `json:"checks"`
}

type Check struct {
	Level             string `json:"level"`
	Path              string `json:"path,omitempty"`
	Policy            string `json:"policy,omitempty"`
	RecommendedAction string `json:"recommendedAction,omitempty"`
	Message           string `json:"message"`
}

type reportRecorder struct {
	report *Report
}

type checkBuilder interface {
	Build(Options, *Report) error
}

type reportPresenter interface {
	Present(Report, Options, io.Writer, error) error
}

type projectConfigEnsurer interface {
	Ensure(string) (bool, error)
}

type CheckBuilder struct{}

type ReportPresenter struct{}

type Service struct {
	checkBuilder  checkBuilder
	presenter     reportPresenter
	projectConfig projectConfigEnsurer
}

func NewService() Service {
	return Service{
		checkBuilder:  CheckBuilder{},
		presenter:     ReportPresenter{},
		projectConfig: projectconfig.NewService(),
	}
}

var fallbackRequiredDirs = []string{
	filepath.Join(".opencode", "agents"),
	filepath.Join(".opencode", "commands"),
	filepath.Join(".opencode", "skills"),
	filepath.Join(".opencode", "templates"),
	filepath.Join(".opencode", "plugins"),
	filepath.Join(".opencode", "policies"),
}

var fallbackRequiredManagedFiles = []string{
	agentsref.HarnessFile,
	filepath.Join(".opencode", "plugins", "agent-observatory.tsx"),
	"tui.json",
}

var requiredStateFiles = []string{filepath.Join(".lufy-ai", "install-state.json")}

var jsonValidationFiles = []string{
	toolruntime.OpenCodeProjectConfigFile,
	"tui.json",
	filepath.Join(".opencode", "package.json"),
	filepath.Join(".opencode", "package-lock.json"),
	filepath.Join("openspec", "UPSTREAM.json"),
}

func (s Service) Run(opts Options, stdout io.Writer) error {
	target, err := platform.ResolveTargetPath(opts.Target)
	if err != nil {
		return err
	}
	scope, err := assets.ParseScope(string(opts.Scope))
	if err != nil {
		return err
	}
	report := Report{TargetRoot: target, Scope: string(scope)}
	if scope == assets.ScopeGlobal || scope == assets.ScopeBoth {
		tool := opts.ExpectedTool
		if tool == "" {
			tool = domain.ToolInitialDefault
		}
		globalRoot, err := toolruntime.GlobalRoot(tool)
		if err != nil {
			return err
		}
		report.GlobalRoot = globalRoot
	}
	recorder := reportRecorder{report: &report}

	if created, err := s.projectConfig.Ensure(target); err != nil {
		recorder.emit("fail", projectconfig.ProjectConfigPath, "no se pudo crear configuración project-local: %s", err.Error())
		return s.presenter.Present(report, opts, stdout, fmt.Errorf("verify falló con 1 problema(s) crítico(s)"))
	} else if created {
		recorder.emit("ok", projectconfig.ProjectConfigPath, "configuración project-local creada")
	} else {
		recorder.emit("ok", projectconfig.ProjectConfigPath, "configuración project-local")
	}

	opts.Target = target
	opts.Scope = scope
	if err := s.checkBuilder.Build(opts, &report); err != nil {
		return err
	}
	recorder = reportRecorder{report: &report}
	if report.Failures > 0 {
		return s.presenter.Present(report, opts, stdout, fmt.Errorf("verify falló con %d problema(s) crítico(s)", report.Failures))
	}
	recorder.emit("ok", "", "verify estructural completo")
	return s.presenter.Present(report, opts, stdout, nil)
}

func (b CheckBuilder) Build(opts Options, report *Report) error {
	target := opts.Target
	recorder := reportRecorder{report: report}
	st, err := state.Load(target)
	if err != nil {
		return fmt.Errorf("fail: install-state.json inválido: %w", err)
	}
	if st == nil {
		recorder.emit("fail", state.Path(target), "falta estado de instalación")
		return nil
	}
	report.SchemaVersion = st.SchemaVersion
	report.Tool = string(st.Tool)
	report.MethodologyByTier = st.MethodologyByTier
	report.Assets = len(st.Assets)
	if opts.ExpectedTool != "" && st.Tool != opts.ExpectedTool {
		recorder.emit("fail", "install-state.json", "tool del manifest no coincide: esperado=%s actual=%s", opts.ExpectedTool, st.Tool)
	}
	assetMap := st.AssetMap()
	requiredDirs, requiredManagedFiles := catalogRequirements(st)
	if opts.Verbose {
		recorder.emit("info", "", "requirements derivados dirs=%d files=%d", len(requiredDirs), len(requiredManagedFiles))
	}
	for _, rel := range requiredStateFiles {
		path, err := platform.SafeJoin(target, rel)
		if err != nil {
			return err
		}
		if !regularFile(path) {
			recorder.emit("fail", rel, "falta archivo crítico")
			continue
		}
		recorder.emit("ok", rel, "archivo crítico")
	}
	if st.SourceChangeID == "" || st.SourceRootFingerprint == "" {
		recorder.emit("fail", "install-state.json", "install-state.json no contiene fingerprint de catálogo")
	}
	for _, rel := range jsonValidationFiles {
		status, err := validateJSONFile(target, rel)
		if err != nil {
			recorder.emit("fail", "", "JSON inválido en %s: %s", rel, err.Error())
			continue
		}
		if status != "" {
			recorder.emit("ok", rel, "JSON parseable")
		}
	}
	if opts.Deep {
		runDeepVerify(st.Tool, target, recorder.emit)
		runDeepMemoryVerify(target, recorder.emit)
	}
	projectConfigFile, err := toolruntime.ProjectConfigFile(st.Tool)
	if err != nil {
		recorder.emit("fail", "", "runtime de configuración inválido: %s", err.Error())
	} else if exists, err := toolruntime.ValidateProjectConfig(st.Tool, target); err != nil {
		recorder.emit("fail", "", "estructura gestionada inválida en %s: %s", projectConfigFile, err.Error())
	} else if exists {
		recorder.emit("ok", projectConfigFile, "estructura merge-managed")
	}
	manifestTarget := st.TargetRoot
	if manifestTarget != "" {
		if resolved, err := platform.ResolveTargetPath(manifestTarget); err == nil {
			manifestTarget = resolved
		}
	}
	if manifestTarget != "" && manifestTarget != target {
		recorder.emit("fail", "install-state.json", "targetRoot del manifest no coincide: manifest=%s actual=%s", st.TargetRoot, target)
	}
	recorder.emit("ok", "install-state.json", "install-state.json schema=%d assets=%d", st.SchemaVersion, len(st.Assets))

	for _, required := range requiredDirs {
		clean, err := platform.EnsureRelativeSafe(required)
		if err != nil {
			return err
		}
		path, err := platform.SafeJoin(target, clean)
		if err != nil {
			recorder.emit("fail", clean, "directorio crítico inseguro: %s", err.Error())
			continue
		}
		if !directory(path) {
			recorder.emit("fail", clean, "falta directorio crítico")
			continue
		}
		recorder.emit("ok", clean, "directorio crítico")
	}

	for _, required := range requiredManagedFiles {
		clean, err := platform.EnsureRelativeSafe(required)
		if err != nil {
			return err
		}
		if _, ok := assetMap[clean]; !ok {
			if opts.AllowCatalogNewAssets {
				continue
			}
			recorder.emit("fail", clean, "asset clave no está en manifest")
		}
		path, err := platform.SafeJoin(target, clean)
		if err != nil {
			recorder.emit("fail", clean, "asset clave inseguro: %s", err.Error())
			continue
		}
		if !regularFile(path) {
			recorder.emit("fail", clean, "falta archivo crítico")
			continue
		}
		recorder.emit("ok", clean, "archivo crítico")
	}
	verifyAgentsReference(target, opts.AllowMissingAgentsRef, recorder.emitAsset)

	for _, asset := range st.Assets {
		clean, err := platform.EnsureRelativeSafe(asset.TargetRel)
		if err != nil {
			return err
		}
		path, err := platform.SafeJoin(target, clean)
		if err != nil {
			recorder.emit("fail", clean, "asset inseguro: %s", err.Error())
			continue
		}
		if !regularFile(path) {
			recorder.emit("fail", clean, "falta asset gestionado")
			continue
		}
		actual, err := assets.FileSHA256(path)
		if err != nil {
			return err
		}
		if actual != asset.TargetSHA256 {
			if asset.Policy == string(assets.PolicyNoReplace) && regularFile(path+".lufy-new") {
				recorder.emitAsset("warn", clean, asset.Policy, "review-lufy-new", "drift no-replace con nueva versión en %s", clean+".lufy-new")
				continue
			}
			recorder.emit("fail", "", "drift en %s expected=%s actual=%s", clean, shortHash(asset.TargetSHA256), shortHash(actual))
			continue
		}
		recorder.emit("ok", clean, "sha256=%s", shortHash(actual))
	}
	reportExtraManagedDirFiles(target, requiredDirs, assetMap, recorder.emit)

	if opts.NoEngram {
		recorder.emit("warn", "", "chequeo de Engram omitido por --no-engram")
	} else if path, ok := platform.ResolveEngram(false, platform.OSResolver{}); ok {
		recorder.emit("ok", "", "engram detectado en PATH (%s)", path)
	} else {
		recorder.emit("warn", "", "engram no encontrado en PATH")
	}
	return nil
}

func (e reportRecorder) emit(level, path, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	e.report.Checks = append(e.report.Checks, Check{Level: level, Path: path, Message: message})
	e.count(level)
}

func (e reportRecorder) emitAsset(level, path, policy, recommendedAction, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	e.report.Checks = append(e.report.Checks, Check{Level: level, Path: path, Policy: policy, RecommendedAction: recommendedAction, Message: message})
	e.count(level)
}

func (e reportRecorder) count(level string) {
	switch level {
	case "fail":
		e.report.Failures++
	case "warn":
		e.report.Warnings++
	case "info":
		e.report.Infos++
	}
}

func (p ReportPresenter) Present(report Report, opts Options, stdout io.Writer, err error) error {
	report.OK = report.Failures == 0
	if opts.JSON {
		body, jsonErr := json.MarshalIndent(report, "", "  ")
		if jsonErr != nil {
			return jsonErr
		}
		fmt.Fprintf(stdout, "%s\n", body)
		return err
	}
	if opts.Quiet {
		return err
	}
	for _, check := range report.Checks {
		if check.Path == "" {
			fmt.Fprintf(stdout, "%s: %s\n", check.Level, check.Message)
			continue
		}
		fmt.Fprintf(stdout, "%s: %s: %s\n", check.Level, check.Message, check.Path)
	}
	return err
}

func verifyAgentsReference(target string, allowMissing bool, emitAsset func(level, path, policy, recommendedAction, format string, args ...any)) {
	exists, hasReference, err := agentsref.Status(target)
	if err != nil {
		emitAsset("fail", agentsref.AgentsFile, "user-owned-reference", agentsref.RecommendedInstallAction(), "integración AGENTS insegura: %s", err.Error())
		return
	}
	if !exists {
		level := "fail"
		if allowMissing {
			level = "warn"
		}
		emitAsset(level, agentsref.AgentsFile, "user-owned-reference", agentsref.RecommendedInstallAction(), "falta AGENTS.md con referencia %s", agentsref.Reference)
		return
	}
	if !hasReference {
		level := "fail"
		if allowMissing {
			level = "warn"
		}
		emitAsset(level, agentsref.AgentsFile, "user-owned-reference", agentsref.RecommendedInstallAction(), "AGENTS.md no referencia %s", agentsref.Reference)
		return
	}
	emitAsset("ok", agentsref.AgentsFile, "user-owned-reference", "", "referencia %s presente", agentsref.Reference)
}

func catalogRequirements(st *state.InstallState) ([]string, []string) {
	assetMap := st.AssetMap()
	catalog, err := currentCatalogForHarness(domain.HarnessConfig{Tool: st.Tool, MethodologyByTier: st.MethodologyByTier})
	if err != nil {
		return fallbackRequiredDirs, fallbackRequiredManagedFiles
	}
	dirs := map[string]bool{}
	files := map[string]bool{}
	for _, dir := range fallbackRequiredDirs {
		dirs[dir] = true
	}
	for _, file := range fallbackRequiredManagedFiles {
		files[file] = true
	}
	for _, asset := range catalog.Assets {
		if asset.Policy != assets.PolicyManaged {
			continue
		}
		if asset.Kind == assets.KindFile && assetMap[asset.TargetRel].TargetRel != "" {
			files[asset.TargetRel] = true
			for _, dir := range parentDirs(asset.TargetRel) {
				dirs[dir] = true
			}
		}
	}
	return sortedKeys(dirs), sortedKeys(files)
}

func parentDirs(path string) []string {
	var dirs []string
	for dir := filepath.Dir(path); dir != "." && dir != string(filepath.Separator); dir = filepath.Dir(dir) {
		dirs = append(dirs, dir)
	}
	return dirs
}

func currentCatalog() (assets.Catalog, error) {
	if sourceRoot, err := platform.ResolveSourceRoot(""); err == nil {
		return assets.BuildCatalog(sourceRoot)
	}
	return assets.BuildEmbeddedCatalog()
}

func currentCatalogForHarness(harness domain.HarnessConfig) (assets.Catalog, error) {
	catalog, err := currentCatalog()
	if err != nil {
		return assets.Catalog{}, err
	}
	return harnesscatalog.Effective(catalog, harness)
}

func sortedKeys(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func regularFile(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0
}

func directory(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0
}

func validateJSONFile(target, rel string) (string, error) {
	path, err := platform.SafeJoin(target, rel)
	if err != nil {
		return "", err
	}
	body, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return "", err
	}
	return "ok", nil
}

func runDeepVerify(tool domain.ToolID, target string, emit func(level, path, format string, args ...any)) {
	files, err := toolruntime.PluginConfigFiles(tool)
	if err != nil {
		emit("fail", "", "runtime de configuración inválido: %s", err.Error())
		return
	}
	for _, rel := range files {
		validatePluginConfig(target, rel, emit)
	}
}

func runDeepMemoryVerify(target string, emit func(level, path, format string, args ...any)) {
	report, err := memory.NewService().BuildValidate(memory.Options{Target: target})
	if err != nil {
		emit("fail", projectconfig.ProjectConfigPath, "memoria no evaluable: %s", err.Error())
		return
	}
	if !report.Status.Initialized {
		emit("warn", report.Root, "memoria Obsidian no inicializada")
		return
	}
	for _, check := range report.Checks {
		if check.Level == "fail" || check.Level == "warn" {
			emit(check.Level, check.Path, "memoria: %s", check.Message)
		}
	}
	if report.OK {
		emit("ok", report.Root, "memoria Obsidian schema=%d notas=%d", report.Status.SchemaVersion, report.Status.Notes)
	}
}

func validatePluginConfig(target, rel string, emit func(level, path, format string, args ...any)) {
	path, err := platform.SafeJoin(target, rel)
	if err != nil {
		emit("fail", rel, "config plugin insegura: %s", err.Error())
		return
	}
	body, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		emit("fail", rel, "no se pudo leer config plugin: %s", err.Error())
		return
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return
	}
	plugins, exists := decoded["plugin"]
	if !exists {
		return
	}
	items, ok := plugins.([]any)
	if !ok {
		emit("fail", rel, "plugin debe ser array para deep verify")
		return
	}
	for _, item := range items {
		pluginRel, ok := item.(string)
		if !ok {
			emit("fail", rel, "plugin contiene entrada no string")
			continue
		}
		pluginRel = filepath.Clean(pluginRel)
		clean, err := platform.EnsureRelativeSafe(pluginRel)
		if err != nil {
			emit("fail", rel, "plugin path inseguro %q: %s", pluginRel, err.Error())
			continue
		}
		pluginPath, err := platform.SafeJoin(target, clean)
		if err != nil {
			emit("fail", clean, "plugin inseguro: %s", err.Error())
			continue
		}
		if !regularFile(pluginPath) {
			emit("fail", clean, "plugin referenciado no existe o no es archivo regular seguro")
			continue
		}
		emit("ok", clean, "plugin referenciado por %s", rel)
	}
}

func reportExtraManagedDirFiles(target string, dirs []string, assetMap map[string]state.AssetState, emit func(level, path, format string, args ...any)) {
	seen := map[string]bool{}
	for _, dir := range dirs {
		root, err := platform.SafeJoin(target, dir)
		if err != nil || !directory(root) {
			continue
		}
		_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(target, path)
			if err != nil {
				return nil
			}
			clean, err := platform.EnsureRelativeSafe(rel)
			if err != nil || seen[clean] {
				return nil
			}
			seen[clean] = true
			if _, managed := assetMap[clean]; !managed {
				emit("info", clean, "archivo extra en directorio gestionado")
			}
			return nil
		})
	}
}

func shortHash(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}
