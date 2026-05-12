package verify

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/config"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

type Options struct {
	Target                string
	NoEngram              bool
	JSON                  bool
	Quiet                 bool
	Verbose               bool
	Deep                  bool
	AllowCatalogNewAssets bool
	Scope                 assets.Scope
}

type Report struct {
	OK            bool    `json:"ok"`
	TargetRoot    string  `json:"targetRoot"`
	Scope         string  `json:"scope,omitempty"`
	GlobalRoot    string  `json:"globalRoot,omitempty"`
	SchemaVersion int     `json:"schemaVersion,omitempty"`
	Assets        int     `json:"assets,omitempty"`
	Failures      int     `json:"failures"`
	Warnings      int     `json:"warnings"`
	Infos         int     `json:"infos"`
	Checks        []Check `json:"checks"`
}

type Check struct {
	Level             string `json:"level"`
	Path              string `json:"path,omitempty"`
	Policy            string `json:"policy,omitempty"`
	RecommendedAction string `json:"recommendedAction,omitempty"`
	Message           string `json:"message"`
}

type Service struct{}

func NewService() Service {
	return Service{}
}

var fallbackRequiredDirs = []string{
	filepath.Join(".opencode", "agents"),
	filepath.Join(".opencode", "commands"),
	filepath.Join(".opencode", "skills"),
	filepath.Join(".opencode", "plugins"),
	filepath.Join(".opencode", "policies"),
}

var fallbackRequiredManagedFiles = []string{
	"AGENTS.md",
	filepath.Join(".opencode", "plugins", "agent-observatory.tsx"),
	"tui.json",
	filepath.Join("openspec", "config.yaml"),
}

var requiredStateFiles = []string{filepath.Join(".lufy-ai", "install-state.json")}

var jsonValidationFiles = []string{
	config.OpenCodeFile,
	"tui.json",
	filepath.Join(".opencode", "package.json"),
	filepath.Join(".opencode", "package-lock.json"),
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
		globalRoot, err := platform.ResolveOpenCodeConfigRoot()
		if err != nil {
			return err
		}
		report.GlobalRoot = globalRoot
	}
	emit := func(level, path, format string, args ...any) {
		message := fmt.Sprintf(format, args...)
		report.Checks = append(report.Checks, Check{Level: level, Path: path, Message: message})
		switch level {
		case "fail":
			report.Failures++
		case "warn":
			report.Warnings++
		case "info":
			report.Infos++
		}
		if !opts.JSON && !opts.Quiet {
			if path == "" {
				fmt.Fprintf(stdout, "%s: %s\n", level, message)
				return
			}
			fmt.Fprintf(stdout, "%s: %s: %s\n", level, message, path)
		}
	}
	emitAsset := func(level, path, policy, recommendedAction, format string, args ...any) {
		message := fmt.Sprintf(format, args...)
		report.Checks = append(report.Checks, Check{Level: level, Path: path, Policy: policy, RecommendedAction: recommendedAction, Message: message})
		switch level {
		case "fail":
			report.Failures++
		case "warn":
			report.Warnings++
		case "info":
			report.Infos++
		}
		if !opts.JSON && !opts.Quiet {
			fmt.Fprintf(stdout, "%s: %s: %s\n", level, message, path)
		}
	}
	finish := func(err error) error {
		report.OK = report.Failures == 0
		if opts.JSON {
			body, jsonErr := json.MarshalIndent(report, "", "  ")
			if jsonErr != nil {
				return jsonErr
			}
			fmt.Fprintf(stdout, "%s\n", body)
		}
		return err
	}

	st, err := state.Load(target)
	if err != nil {
		return fmt.Errorf("fail: install-state.json inválido: %w", err)
	}
	if st == nil {
		emit("fail", state.Path(target), "falta estado de instalación")
		return finish(fmt.Errorf("verify falló con 1 problema(s) crítico(s)"))
	}
	report.SchemaVersion = st.SchemaVersion
	report.Assets = len(st.Assets)
	assetMap := st.AssetMap()
	requiredDirs, requiredManagedFiles := catalogRequirements(assetMap)
	if opts.Verbose {
		emit("info", "", "requirements derivados dirs=%d files=%d", len(requiredDirs), len(requiredManagedFiles))
	}
	for _, rel := range requiredStateFiles {
		path, err := platform.SafeJoin(target, rel)
		if err != nil {
			return err
		}
		if !regularFile(path) {
			emit("fail", rel, "falta archivo crítico")
			continue
		}
		emit("ok", rel, "archivo crítico")
	}
	if st.SourceChangeID == "" || st.SourceRootFingerprint == "" {
		emit("fail", "install-state.json", "install-state.json no contiene fingerprint de catálogo")
	}
	for _, rel := range jsonValidationFiles {
		status, err := validateJSONFile(target, rel)
		if err != nil {
			emit("fail", "", "JSON inválido en %s: %s", rel, err.Error())
			continue
		}
		if status != "" {
			emit("ok", rel, "JSON parseable")
		}
	}
	if opts.Deep {
		runDeepVerify(target, emit)
	}
	if status, err := config.NewService().ValidateManagedStructure(target); err != nil {
		emit("fail", "", "estructura gestionada inválida en %s: %s", config.OpenCodeFile, err.Error())
	} else if status.Exists {
		emit("ok", config.OpenCodeFile, "estructura merge-managed")
	}
	manifestTarget := st.TargetRoot
	if manifestTarget != "" {
		if resolved, err := platform.ResolveTargetPath(manifestTarget); err == nil {
			manifestTarget = resolved
		}
	}
	if manifestTarget != "" && manifestTarget != target {
		emit("fail", "install-state.json", "targetRoot del manifest no coincide: manifest=%s actual=%s", st.TargetRoot, target)
	}
	emit("ok", "install-state.json", "install-state.json schema=%d assets=%d", st.SchemaVersion, len(st.Assets))

	for _, required := range requiredDirs {
		clean, err := platform.EnsureRelativeSafe(required)
		if err != nil {
			return err
		}
		path, err := platform.SafeJoin(target, clean)
		if err != nil {
			emit("fail", clean, "directorio crítico inseguro: %s", err.Error())
			continue
		}
		if !directory(path) {
			emit("fail", clean, "falta directorio crítico")
			continue
		}
		emit("ok", clean, "directorio crítico")
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
			emit("fail", clean, "asset clave no está en manifest")
		}
		path, err := platform.SafeJoin(target, clean)
		if err != nil {
			emit("fail", clean, "asset clave inseguro: %s", err.Error())
			continue
		}
		if !regularFile(path) {
			emit("fail", clean, "falta archivo crítico")
			continue
		}
		emit("ok", clean, "archivo crítico")
	}

	for _, asset := range st.Assets {
		clean, err := platform.EnsureRelativeSafe(asset.TargetRel)
		if err != nil {
			return err
		}
		path, err := platform.SafeJoin(target, clean)
		if err != nil {
			emit("fail", clean, "asset inseguro: %s", err.Error())
			continue
		}
		if !regularFile(path) {
			emit("fail", clean, "falta asset gestionado")
			continue
		}
		actual, err := assets.FileSHA256(path)
		if err != nil {
			return err
		}
		if actual != asset.TargetSHA256 {
			if asset.Policy == string(assets.PolicyNoReplace) && regularFile(path+".lufy-new") {
				emitAsset("warn", clean, asset.Policy, "review-lufy-new", "drift no-replace con nueva versión en %s", clean+".lufy-new")
				continue
			}
			emit("fail", "", "drift en %s expected=%s actual=%s", clean, shortHash(asset.TargetSHA256), shortHash(actual))
			continue
		}
		emit("ok", clean, "sha256=%s", shortHash(actual))
	}
	reportExtraManagedDirFiles(target, requiredDirs, assetMap, emit)

	if opts.NoEngram {
		emit("warn", "", "chequeo de Engram omitido por --no-engram")
	} else if path, ok := platform.ResolveEngram(false, platform.OSResolver{}); ok {
		emit("ok", "", "engram detectado en PATH (%s)", path)
	} else {
		emit("warn", "", "engram no encontrado en PATH")
	}

	if report.Failures > 0 {
		return finish(fmt.Errorf("verify falló con %d problema(s) crítico(s)", report.Failures))
	}
	emit("ok", "", "verify estructural completo")
	return finish(nil)
}

func catalogRequirements(assetMap map[string]state.AssetState) ([]string, []string) {
	catalog, err := currentCatalog()
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

func runDeepVerify(target string, emit func(level, path, format string, args ...any)) {
	validatePluginConfig(target, "tui.json", emit)
	validatePluginConfig(target, config.OpenCodeFile, emit)
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
