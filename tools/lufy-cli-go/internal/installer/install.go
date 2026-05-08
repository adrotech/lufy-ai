package installer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/backup"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/config"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/verify"
)

type Options struct {
	Target   string
	DryRun   bool
	Yes      bool
	NoEngram bool
	Backup   bool
}

type Action struct {
	Kind                  string
	Source                string
	Target                string
	Reason                string
	Risk                  string
	SourceHash            string
	CurrentHash           string
	PreviousInstalledHash string
}

type Conflict struct {
	Path        string
	Reason      string
	Risk        string
	CurrentHash string
	SourceHash  string
}

type Plan struct {
	SourceRoot string
	TargetRoot string
	NoEngram   bool
	Catalog    assets.Catalog
	Previous   *state.InstallState
	Actions    []Action
	Conflicts  []Conflict
}

type Service struct{}

func NewService() Service {
	return Service{}
}

func (s Service) Run(opts Options, stdout io.Writer) error {
	plan, err := s.BuildPlan(opts)
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Plan de instalación para %s\n", plan.TargetRoot)
	fmt.Fprintf(stdout, "Source root: %s\n", plan.SourceRoot)
	for _, a := range plan.Actions {
		fmt.Fprintf(stdout, "- [%s] %s (%s)", a.Kind, a.Target, a.Reason)
		if a.SourceHash != "" {
			fmt.Fprintf(stdout, " source=%s", shortHash(a.SourceHash))
		}
		if a.CurrentHash != "" {
			fmt.Fprintf(stdout, " current=%s", shortHash(a.CurrentHash))
		}
		fmt.Fprintln(stdout)
	}
	for _, c := range plan.Conflicts {
		fmt.Fprintf(stdout, "- [warn-conflict] %s (%s) current=%s source=%s\n", c.Path, c.Reason, shortHash(c.CurrentHash), shortHash(c.SourceHash))
	}

	if opts.NoEngram {
		fmt.Fprintln(stdout, "Engram: omitido por --no-engram")
	} else if path, ok := platform.ResolveEngram(false, platform.OSResolver{}); ok {
		fmt.Fprintf(stdout, "Engram: detectado en PATH (%s)\n", path)
	} else {
		fmt.Fprintln(stdout, "Engram: no encontrado en PATH (instalación base continúa)")
	}

	if opts.DryRun {
		fmt.Fprintln(stdout, "Modo dry-run: sin mutaciones en filesystem")
		return nil
	}
	if len(plan.Conflicts) > 0 {
		return fmt.Errorf("install bloqueado por %d conflicto(s); resuelve manualmente y reintenta", len(plan.Conflicts))
	}
	if !opts.Yes && requiresConfirmation(plan.Actions) {
		return fmt.Errorf("install requiere --yes para aplicar mutaciones reales; usa --dry-run para revisar el plan sin escribir")
	}

	if err := s.applyInstall(plan, stdout); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Install real completado")
	return nil
}

func (s Service) BuildPlan(opts Options) (Plan, error) {
	target, err := platform.ResolveTargetPath(opts.Target)
	if err != nil {
		return Plan{}, err
	}
	sourceRoot, err := platform.ResolveSourceRoot("")
	if err != nil {
		sourceRoot = assets.EmbeddedSourceRoot
	}
	var catalog assets.Catalog
	if sourceRoot == assets.EmbeddedSourceRoot {
		catalog, err = assets.BuildEmbeddedCatalog()
	} else {
		catalog, err = assets.BuildCatalog(sourceRoot)
	}
	if err != nil {
		return Plan{}, err
	}
	previous, err := state.Load(target)
	if err != nil {
		return Plan{}, err
	}
	previousAssets := map[string]state.AssetState{}
	if previous != nil {
		previousAssets = previous.AssetMap()
	}

	plan := Plan{SourceRoot: sourceRoot, TargetRoot: target, NoEngram: opts.NoEngram, Catalog: catalog, Previous: previous}
	seenDirs := map[string]bool{}
	for _, asset := range catalog.Assets {
		if asset.Kind == assets.KindDir {
			if !dirExists(filepath.Join(target, asset.TargetRel)) {
				plan.Actions = append(plan.Actions, Action{Kind: "mkdir", Source: asset.SourceRel, Target: asset.TargetRel, Reason: "directorio gestionado ausente", Risk: "low"})
			}
			seenDirs[asset.TargetRel] = true
			continue
		}
		for _, dir := range parentDirs(asset.TargetRel) {
			if seenDirs[dir] {
				continue
			}
			seenDirs[dir] = true
			if !dirExists(filepath.Join(target, dir)) {
				plan.Actions = append(plan.Actions, Action{Kind: "mkdir", Target: dir, Reason: "directorio padre requerido", Risk: "low"})
			}
		}

		targetPath, err := platform.SafeJoin(target, asset.TargetRel)
		if err != nil {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Reason: err.Error(), Risk: "high", SourceHash: asset.SourceSHA256})
			continue
		}
		info, err := os.Lstat(targetPath)
		if os.IsNotExist(err) {
			plan.Actions = append(plan.Actions, Action{Kind: "copy", Source: asset.SourceRel, Target: asset.TargetRel, Reason: "archivo gestionado ausente", Risk: "low", SourceHash: asset.SourceSHA256})
			continue
		}
		if err != nil {
			return Plan{}, err
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Reason: "destino no es archivo regular seguro", Risk: "high", SourceHash: asset.SourceSHA256})
			continue
		}
		currentHash, err := assets.FileSHA256(targetPath)
		if err != nil {
			return Plan{}, err
		}
		if currentHash == asset.SourceSHA256 {
			plan.Actions = append(plan.Actions, Action{Kind: "skip", Source: asset.SourceRel, Target: asset.TargetRel, Reason: "hash destino coincide con source", Risk: "none", SourceHash: asset.SourceSHA256, CurrentHash: currentHash})
			continue
		}
		prev, managed := previousAssets[asset.TargetRel]
		if !managed {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Reason: "archivo existente no gestionado con contenido distinto", Risk: "high", CurrentHash: currentHash, SourceHash: asset.SourceSHA256})
			continue
		}
		if currentHash != prev.TargetSHA256 {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Reason: "archivo gestionado con drift local", Risk: "high", CurrentHash: currentHash, SourceHash: asset.SourceSHA256})
			continue
		}
		plan.Actions = append(plan.Actions,
			Action{Kind: "backup", Source: asset.SourceRel, Target: asset.TargetRel, Reason: "source gestionado cambió; backup requerido antes de actualizar", Risk: "medium", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, PreviousInstalledHash: prev.TargetSHA256},
			Action{Kind: "update-managed", Source: asset.SourceRel, Target: asset.TargetRel, Reason: "source gestionado cambió sin drift local", Risk: "medium", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, PreviousInstalledHash: prev.TargetSHA256},
		)
	}
	if opts.Backup && previous != nil {
		alreadyBackup := map[string]bool{}
		for _, action := range plan.Actions {
			if action.Kind == "backup" {
				alreadyBackup[action.Target] = true
			}
		}
		for _, asset := range previous.Assets {
			if alreadyBackup[asset.TargetRel] {
				continue
			}
			assetPath, err := platform.SafeJoin(target, asset.TargetRel)
			if err != nil {
				return Plan{}, err
			}
			if fileExists(assetPath) {
				plan.Actions = append(plan.Actions, Action{Kind: "backup", Target: asset.TargetRel, Reason: "backup solicitado explícitamente", Risk: "low", CurrentHash: asset.TargetSHA256})
			}
		}
	}
	configPlan, err := config.NewService().Plan(config.Options{TargetRoot: target, NoEngram: opts.NoEngram})
	if err != nil {
		return Plan{}, err
	}
	if configPlan.Action == "merge-json" && fileExists(filepath.Join(target, config.OpenCodeFile)) {
		plan.Actions = append(plan.Actions, Action{Kind: "backup", Target: config.OpenCodeFile, Reason: "opencode.json existente será mergeado", Risk: "medium"})
	}
	plan.Actions = append(plan.Actions, Action{Kind: configPlan.Action, Target: config.OpenCodeFile, Reason: "configuración OpenCode gestionada con merge conservador", Risk: "low"})
	plan.Actions = append(plan.Actions, Action{Kind: "verify", Target: target, Reason: "verificación estructural posterior a install", Risk: "none"})
	sort.SliceStable(plan.Actions, func(i, j int) bool {
		order := map[string]int{"mkdir": 0, "backup": 1, "copy": 2, "update-managed": 3, "merge-json": 4, "verify": 5, "skip": 6}
		if order[plan.Actions[i].Kind] == order[plan.Actions[j].Kind] {
			return plan.Actions[i].Target < plan.Actions[j].Target
		}
		return order[plan.Actions[i].Kind] < order[plan.Actions[j].Kind]
	})
	return plan, nil
}

func (s Service) applyInstall(plan Plan, stdout io.Writer) error {
	if err := os.MkdirAll(plan.TargetRoot, 0o755); err != nil {
		return err
	}
	backupTargets := targetsForKind(plan.Actions, "backup")
	recoveryBackup := ""
	if len(backupTargets) > 0 {
		backupDir, err := backup.BackupFiles(plan.TargetRoot, backupTargets, "install-update-managed", stdout)
		if err != nil {
			return err
		}
		recoveryBackup = backupDir
		fmt.Fprintf(stdout, "- [backup] %s\n", backupDir)
	}
	applied := 0
	for _, action := range plan.Actions {
		switch action.Kind {
		case "mkdir":
			targetPath, err := platform.SafeJoin(plan.TargetRoot, action.Target)
			if err != nil {
				return installRecoveryError(err, recoveryBackup, applied)
			}
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return installRecoveryError(err, recoveryBackup, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [mkdir] %s\n", action.Target)
		case "copy", "update-managed":
			if err := copyFile(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
				return installRecoveryError(err, recoveryBackup, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [%s] %s\n", action.Kind, action.Target)
		case "merge-json":
			if _, err := config.NewService().Ensure(config.Options{TargetRoot: plan.TargetRoot, NoEngram: plan.NoEngram}); err != nil {
				return installRecoveryError(err, recoveryBackup, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [merge-json] %s\n", action.Target)
		case "verify":
		case "skip", "backup":
		}
	}
	if plan.Previous != nil && !hasContentMutation(plan.Actions) {
		if err := runPostInstallVerify(plan, stdout); err != nil {
			return installRecoveryError(err, recoveryBackup, applied)
		}
		fmt.Fprintf(stdout, "- [skip] %s sin cambios\n", filepath.Join(".lufy-ai", "install-state.json"))
		return nil
	}
	assetStates, err := buildAssetStates(plan)
	if err != nil {
		return installRecoveryError(err, recoveryBackup, applied)
	}
	fingerprint, err := plan.Catalog.Fingerprint()
	if err != nil {
		return installRecoveryError(err, recoveryBackup, applied)
	}
	st := state.New(plan.TargetRoot, plan.Previous, assetStates, fingerprint)
	if err := state.WriteAtomic(plan.TargetRoot, st); err != nil {
		return installRecoveryError(err, recoveryBackup, applied)
	}
	fmt.Fprintf(stdout, "- [write] %s\n", filepath.Join(".lufy-ai", "install-state.json"))
	if err := runPostInstallVerify(plan, stdout); err != nil {
		if rollbackErr := restoreStateAfterVerifyFailure(plan); rollbackErr != nil {
			err = fmt.Errorf("%w; además falló restaurar install-state previo: %v", err, rollbackErr)
		}
		return installRecoveryError(err, recoveryBackup, applied)
	}
	return nil
}

func restoreStateAfterVerifyFailure(plan Plan) error {
	if plan.Previous == nil {
		err := os.Remove(state.Path(plan.TargetRoot))
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return state.WriteAtomic(plan.TargetRoot, *plan.Previous)
}

func runPostInstallVerify(plan Plan, stdout io.Writer) error {
	if err := verify.NewService().Run(verify.Options{Target: plan.TargetRoot, NoEngram: plan.NoEngram}, stdout); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "- [verify] %s\n", plan.TargetRoot)
	return nil
}

func installRecoveryError(err error, recoveryBackup string, applied int) error {
	if recoveryBackup == "" {
		return err
	}
	return fmt.Errorf("install falló después de crear backup de recovery en %s; acciones aplicadas=%d: %w", recoveryBackup, applied, err)
}

func targetsForKind(actions []Action, kind string) []string {
	var out []string
	for _, action := range actions {
		if action.Kind == kind {
			out = append(out, action.Target)
		}
	}
	return out
}

func hasContentMutation(actions []Action) bool {
	for _, action := range actions {
		if action.Kind == "copy" || action.Kind == "update-managed" || action.Kind == "merge-json" {
			return true
		}
	}
	return false
}

func requiresConfirmation(actions []Action) bool {
	for _, action := range actions {
		switch action.Kind {
		case "mkdir", "copy", "update-managed", "merge-json", "backup":
			return true
		}
	}
	return false
}

func buildAssetStates(plan Plan) ([]state.AssetState, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	lastAction := map[string]string{}
	for _, action := range plan.Actions {
		if action.Kind == "copy" || action.Kind == "update-managed" || action.Kind == "skip" {
			lastAction[action.Target] = action.Kind
		}
	}
	var out []state.AssetState
	for _, asset := range plan.Catalog.Assets {
		if asset.Kind != assets.KindFile {
			continue
		}
		targetPath, err := platform.SafeJoin(plan.TargetRoot, asset.TargetRel)
		if err != nil {
			return nil, err
		}
		targetHash, err := assets.FileSHA256(targetPath)
		if err != nil {
			return nil, err
		}
		action := lastAction[asset.TargetRel]
		if action == "" {
			action = "record"
		}
		out = append(out, state.AssetState{ID: asset.ID, SourceRel: asset.SourceRel, TargetRel: asset.TargetRel, SourceSHA256: asset.SourceSHA256, TargetSHA256: targetHash, InstalledAt: now, LastAction: action})
	}
	return out, nil
}

func copyFile(sourceRoot, sourceRel, targetRoot, targetRel string) error {
	var content []byte
	if sourceRoot == assets.EmbeddedSourceRoot {
		var err error
		content, err = assets.ReadSourceFile(sourceRoot, sourceRel)
		if err != nil {
			return err
		}
	} else {
		src := filepath.Join(sourceRoot, sourceRel)
		if info, err := os.Lstat(src); err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			if err != nil {
				return err
			}
			return fmt.Errorf("source no es archivo regular seguro: %s", src)
		}
		var err error
		content, err = os.ReadFile(src)
		if err != nil {
			return err
		}
	}
	dst, err := platform.SafeJoin(targetRoot, targetRel)
	if err != nil {
		return err
	}
	if info, err := os.Lstat(dst); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("destino symlink no permitido: %s", dst)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return platform.WriteFileAtomic(dst, content, 0o644)
}

func parentDirs(path string) []string {
	var dirs []string
	for dir := filepath.Dir(path); dir != "." && dir != string(filepath.Separator); dir = filepath.Dir(dir) {
		dirs = append(dirs, dir)
	}
	return dirs
}

func dirExists(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0
}

func fileExists(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0
}

func shortHash(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}
