package syncer

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
}

type Action struct {
	Kind               string
	Source             string
	Target             string
	Reason             string
	SourceHash         string
	CurrentHash        string
	RecordedSourceHash string
	RecordedTargetHash string
}

type Conflict struct {
	Path               string
	Reason             string
	SourceHash         string
	CurrentHash        string
	RecordedSourceHash string
	RecordedTargetHash string
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

	fmt.Fprintf(stdout, "Plan de sync para %s\n", plan.TargetRoot)
	fmt.Fprintf(stdout, "Source root: %s\n", plan.SourceRoot)
	for _, action := range plan.Actions {
		fmt.Fprintf(stdout, "- [%s] %s (%s)", action.Kind, action.Target, action.Reason)
		if action.SourceHash != "" {
			fmt.Fprintf(stdout, " source=%s", shortHash(action.SourceHash))
		}
		if action.CurrentHash != "" {
			fmt.Fprintf(stdout, " current=%s", shortHash(action.CurrentHash))
		}
		if action.RecordedSourceHash != "" {
			fmt.Fprintf(stdout, " recordedSource=%s", shortHash(action.RecordedSourceHash))
		}
		if action.RecordedTargetHash != "" {
			fmt.Fprintf(stdout, " recordedTarget=%s", shortHash(action.RecordedTargetHash))
		}
		fmt.Fprintln(stdout)
	}
	for _, conflict := range plan.Conflicts {
		fmt.Fprintf(stdout, "- [conflict] %s (%s) source=%s current=%s recordedSource=%s recordedTarget=%s\n", conflict.Path, conflict.Reason, shortHash(conflict.SourceHash), shortHash(conflict.CurrentHash), shortHash(conflict.RecordedSourceHash), shortHash(conflict.RecordedTargetHash))
	}
	if opts.NoEngram {
		fmt.Fprintln(stdout, "Engram: omitido por --no-engram")
	}
	if opts.DryRun {
		fmt.Fprintln(stdout, "Modo dry-run: sin mutaciones en filesystem")
		return nil
	}
	if len(plan.Conflicts) > 0 {
		return fmt.Errorf("sync bloqueado por %d conflicto(s); resuelve drift/estado antes de reintentar", len(plan.Conflicts))
	}
	if requiresConfirmation(plan.Actions) && !opts.Yes {
		return fmt.Errorf("sync requiere --yes para aplicar mutaciones reales; usa --dry-run para revisar el plan sin escribir")
	}
	if err := s.apply(plan, stdout); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Sync real completado")
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
	if previous == nil {
		return Plan{}, fmt.Errorf("sync requiere %s; ejecuta install/verify antes de sincronizar", state.Path(target))
	}
	previousAssets := previous.AssetMap()
	plan := Plan{SourceRoot: sourceRoot, TargetRoot: target, NoEngram: opts.NoEngram, Catalog: catalog, Previous: previous}
	catalogTargets := map[string]bool{}

	for _, asset := range catalog.Assets {
		if asset.Kind != assets.KindFile {
			continue
		}
		catalogTargets[asset.TargetRel] = true
		targetPath, err := platform.SafeJoin(target, asset.TargetRel)
		if err != nil {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Reason: err.Error(), SourceHash: asset.SourceSHA256})
			continue
		}
		prev, managed := previousAssets[asset.TargetRel]
		info, err := os.Lstat(targetPath)
		if os.IsNotExist(err) {
			if managed {
				plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Reason: "asset gestionado ausente; no se recrea sin estrategia explícita", SourceHash: asset.SourceSHA256, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
			} else {
				plan.Actions = append(plan.Actions, Action{Kind: "skip", Source: asset.SourceRel, Target: asset.TargetRel, Reason: "asset del catálogo ausente y no registrado; sync no instala nuevos assets", SourceHash: asset.SourceSHA256})
			}
			continue
		}
		if err != nil {
			return Plan{}, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Reason: "destino symlink no permitido", SourceHash: asset.SourceSHA256})
			continue
		}
		if !info.Mode().IsRegular() {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Reason: "destino no es archivo regular seguro", SourceHash: asset.SourceSHA256})
			continue
		}
		currentHash, err := assets.FileSHA256(targetPath)
		if err != nil {
			return Plan{}, err
		}
		if !managed {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Reason: "archivo existente no gestionado", SourceHash: asset.SourceSHA256, CurrentHash: currentHash})
			continue
		}
		if currentHash != prev.TargetSHA256 {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Reason: "archivo gestionado con drift local", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
			continue
		}
		if asset.SourceSHA256 == prev.SourceSHA256 {
			plan.Actions = append(plan.Actions, Action{Kind: "skip", Source: asset.SourceRel, Target: asset.TargetRel, Reason: "source y target coinciden con el estado registrado", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
			continue
		}
		plan.Actions = append(plan.Actions,
			Action{Kind: "backup", Source: asset.SourceRel, Target: asset.TargetRel, Reason: "source cambió y target no tiene drift; backup requerido", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256},
			Action{Kind: "update-managed", Source: asset.SourceRel, Target: asset.TargetRel, Reason: "source cambió y target coincide con hash registrado", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256},
		)
	}
	for _, prev := range previous.Assets {
		if catalogTargets[prev.TargetRel] {
			continue
		}
		targetPath, err := platform.SafeJoin(target, prev.TargetRel)
		if err != nil {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: prev.TargetRel, Reason: err.Error(), RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
			continue
		}
		info, err := os.Lstat(targetPath)
		if os.IsNotExist(err) {
			plan.Actions = append(plan.Actions, Action{Kind: "retired", Target: prev.TargetRel, Reason: "asset previamente gestionado ya no está en el catálogo y tampoco existe en target", RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
			continue
		}
		if err != nil {
			return Plan{}, err
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: prev.TargetRel, Reason: "asset retirado no es archivo regular seguro", RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
			continue
		}
		currentHash, err := assets.FileSHA256(targetPath)
		if err != nil {
			return Plan{}, err
		}
		if currentHash != prev.TargetSHA256 {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: prev.TargetRel, Reason: "asset retirado con drift local; cleanup manual requerido", CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
			continue
		}
		plan.Actions = append(plan.Actions, Action{Kind: "retired", Target: prev.TargetRel, Reason: "asset previamente gestionado ya no está en el catálogo; se preserva y queda rastreado", CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
	}
	configPlan, err := config.NewService().Plan(config.Options{TargetRoot: target, NoEngram: opts.NoEngram})
	if err != nil {
		return Plan{}, err
	}
	if configPlan.Action == "merge-json" {
		if fileExists(filepath.Join(target, config.OpenCodeFile)) {
			plan.Actions = append(plan.Actions, Action{Kind: "backup", Target: config.OpenCodeFile, Reason: "opencode.json existente será mergeado conservadoramente"})
		}
		plan.Actions = append(plan.Actions, Action{Kind: "merge-json", Target: config.OpenCodeFile, Reason: "configuración OpenCode merge-managed parcial"})
	} else {
		plan.Actions = append(plan.Actions, Action{Kind: "skip", Target: config.OpenCodeFile, Reason: "configuración OpenCode merge-managed sin cambios"})
	}
	sort.SliceStable(plan.Actions, func(i, j int) bool {
		order := map[string]int{"backup": 0, "update-managed": 1, "merge-json": 2, "retired": 3, "skip": 4}
		if order[plan.Actions[i].Kind] == order[plan.Actions[j].Kind] {
			return plan.Actions[i].Target < plan.Actions[j].Target
		}
		return order[plan.Actions[i].Kind] < order[plan.Actions[j].Kind]
	})
	return plan, nil
}

func (s Service) apply(plan Plan, stdout io.Writer) error {
	updates := targetsForKind(plan.Actions, "update-managed")
	merges := targetsForKind(plan.Actions, "merge-json")
	if len(updates) == 0 && len(merges) == 0 {
		fmt.Fprintln(stdout, "- [skip] sin cambios gestionados")
		return nil
	}
	backupTargets := uniqueTargets(append(append([]string{}, updates...), targetsForKind(plan.Actions, "backup")...))
	manifestPath := ""
	if len(backupTargets) > 0 {
		backupDir, err := backup.BackupFiles(plan.TargetRoot, backupTargets, "sync", stdout)
		if err != nil {
			return err
		}
		manifestPath = filepath.Join(backupDir, "manifest.json")
		fmt.Fprintf(stdout, "- [backup] %s\n", manifestPath)
	}
	applied := 0
	for _, action := range plan.Actions {
		switch action.Kind {
		case "update-managed":
			if err := copyFile(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
				return syncRecoveryError(err, manifestPath, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [update-managed] %s\n", action.Target)
		case "merge-json":
			if _, err := config.NewService().Ensure(config.Options{TargetRoot: plan.TargetRoot, NoEngram: plan.NoEngram}); err != nil {
				return syncRecoveryError(err, manifestPath, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [merge-json] %s\n", action.Target)
		default:
			continue
		}
	}
	assetStates, err := buildAssetStates(plan, updates)
	if err != nil {
		return syncRecoveryError(err, manifestPath, applied)
	}
	fingerprint, err := plan.Catalog.Fingerprint()
	if err != nil {
		return syncRecoveryError(err, manifestPath, applied)
	}
	st := state.New(plan.TargetRoot, plan.Previous, assetStates, fingerprint)
	if err := state.WriteAtomic(plan.TargetRoot, st); err != nil {
		return syncRecoveryError(err, manifestPath, applied)
	}
	fmt.Fprintf(stdout, "- [write] %s\n", filepath.Join(".lufy-ai", "install-state.json"))
	if err := verify.NewService().Run(verify.Options{Target: plan.TargetRoot, NoEngram: plan.NoEngram}, stdout); err != nil {
		return syncRecoveryError(err, manifestPath, applied)
	}
	fmt.Fprintf(stdout, "- [verify] %s\n", plan.TargetRoot)
	return nil
}

func buildAssetStates(plan Plan, updates []string) ([]state.AssetState, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	updated := map[string]bool{}
	for _, rel := range updates {
		updated[rel] = true
	}
	previous := plan.Previous.AssetMap()
	var out []state.AssetState
	for _, asset := range plan.Catalog.Assets {
		if asset.Kind != assets.KindFile {
			continue
		}
		prev, managed := previous[asset.TargetRel]
		if !managed && !updated[asset.TargetRel] {
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
		installedAt := prev.InstalledAt
		if installedAt == "" {
			installedAt = now
		}
		lastAction := prev.LastAction
		if updated[asset.TargetRel] {
			lastAction = "sync-update-managed"
		}
		out = append(out, state.AssetState{ID: asset.ID, SourceRel: asset.SourceRel, TargetRel: asset.TargetRel, SourceSHA256: asset.SourceSHA256, TargetSHA256: targetHash, InstalledAt: installedAt, LastAction: lastAction})
	}
	catalogTargets := map[string]bool{}
	for _, asset := range plan.Catalog.Assets {
		if asset.Kind == assets.KindFile {
			catalogTargets[asset.TargetRel] = true
		}
	}
	for _, prev := range plan.Previous.Assets {
		if catalogTargets[prev.TargetRel] {
			continue
		}
		targetPath, err := platform.SafeJoin(plan.TargetRoot, prev.TargetRel)
		if err != nil {
			return nil, err
		}
		currentHash, err := assets.FileSHA256(targetPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		out = append(out, state.AssetState{ID: prev.ID, SourceRel: prev.SourceRel, TargetRel: prev.TargetRel, SourceSHA256: prev.SourceSHA256, TargetSHA256: currentHash, InstalledAt: prev.InstalledAt, LastAction: "retired"})
	}
	return out, nil
}

func copyFile(sourceRoot, sourceRel, targetRoot, targetRel string) error {
	var body []byte
	if sourceRoot == assets.EmbeddedSourceRoot {
		var err error
		body, err = assets.ReadSourceFile(sourceRoot, sourceRel)
		if err != nil {
			return err
		}
	} else {
		src := filepath.Join(sourceRoot, sourceRel)
		info, err := os.Lstat(src)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("source no es archivo regular seguro: %s", src)
		}
		body, err = os.ReadFile(src)
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
	return platform.WriteFileAtomic(dst, body, 0o644)
}

func requiresConfirmation(actions []Action) bool {
	for _, action := range actions {
		if action.Kind == "update-managed" || action.Kind == "backup" || action.Kind == "merge-json" {
			return true
		}
	}
	return false
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

func uniqueTargets(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, target := range in {
		if seen[target] {
			continue
		}
		seen[target] = true
		out = append(out, target)
	}
	return out
}

func fileExists(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0
}

func syncRecoveryError(err error, manifestPath string, applied int) error {
	if manifestPath == "" {
		return err
	}
	return fmt.Errorf("sync falló después de crear backup en %s; acciones aplicadas=%d; restaura con: lufy-ai restore --target <target> --backup %s --yes: %w", manifestPath, applied, manifestPath, err)
}

func shortHash(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}
