package syncer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/agentsref"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/backup"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/harnesscatalog"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/managedcontent"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/mergeblock"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/toolruntime"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/verify"
)

type Options struct {
	Target   string
	DryRun   bool
	Yes      bool
	NoEngram bool
	Scope    assets.Scope
	Harness  domain.HarnessConfig
}

type Action struct {
	Kind               string
	Source             string
	Target             string
	Policy             assets.Policy
	Scope              assets.Scope
	Reason             string
	SourceHash         string
	CurrentHash        string
	RecordedSourceHash string
	RecordedTargetHash string
}

type Conflict struct {
	Path               string
	Policy             assets.Policy
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
	Scope      assets.Scope
	GlobalRoot string
	Harness    domain.HarnessConfig
	Actions    []Action
	Conflicts  []Conflict
}

type Service struct{}

func NewService() Service {
	return Service{}
}

func (s Service) Run(opts Options, stdout io.Writer) error {
	if !opts.DryRun {
		target, err := platform.ResolveTargetPath(opts.Target)
		if err != nil {
			return err
		}
		lock, err := platform.AcquireLock(target)
		if err != nil {
			return err
		}
		defer lock.Release()
		opts.Target = target
		if created, err := projectconfig.NewService().Ensure(target); err != nil {
			return err
		} else if created {
			fmt.Fprintf(stdout, "- [project-config] %s\n", projectconfig.ProjectConfigPath)
		}
	}
	plan, err := s.BuildPlan(opts)
	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Plan de sync para %s\n", plan.TargetRoot)
	fmt.Fprintf(stdout, "Source root: %s\n", plan.SourceRoot)
	fmt.Fprintf(stdout, "Scope: %s projectRoot=%s", plan.Scope, plan.TargetRoot)
	if plan.GlobalRoot != "" {
		fmt.Fprintf(stdout, " globalRoot=%s", plan.GlobalRoot)
	}
	fmt.Fprintln(stdout)
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
	scope, err := assets.ParseScope(string(opts.Scope))
	if err != nil {
		return Plan{}, err
	}
	globalRoot := ""
	if scope == assets.ScopeGlobal || scope == assets.ScopeBoth {
		harness := opts.Harness.WithDefaults()
		globalRoot, err = toolruntime.GlobalRoot(harness.Tool)
		if err != nil {
			return Plan{}, err
		}
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
	harness := opts.Harness.WithDefaults()
	if err := harness.ValidateSupported(); err != nil {
		return Plan{}, err
	}
	if previous.Tool != harness.Tool {
		return Plan{}, fmt.Errorf("sync bloqueado por tool mismatch: manifest=%s solicitado=%s", previous.Tool, harness.Tool)
	}
	installedHarness := domain.HarnessConfig{Tool: previous.Tool, MethodologyByTier: previous.MethodologyByTier}.WithDefaults()
	catalog, err = harnesscatalog.Effective(catalog, installedHarness)
	if err != nil {
		return Plan{}, err
	}
	catalog, err = managedcontent.CatalogWithRenderedHashes(catalog, target)
	if err != nil {
		return Plan{}, err
	}
	previousAssets := previous.AssetMap()
	plan := Plan{SourceRoot: sourceRoot, TargetRoot: target, NoEngram: opts.NoEngram, Catalog: catalog, Previous: previous, Scope: scope, GlobalRoot: globalRoot, Harness: installedHarness}
	catalogTargets := map[string]bool{}

	for _, asset := range catalog.Assets {
		if asset.Kind != assets.KindFile {
			continue
		}
		catalogTargets[asset.TargetRel] = true
		targetPath, err := platform.SafeJoin(target, asset.TargetRel)
		if err != nil {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Policy: asset.Policy, Reason: err.Error(), SourceHash: asset.SourceSHA256})
			continue
		}
		prev, managed := previousAssets[asset.TargetRel]
		info, err := os.Lstat(targetPath)
		if os.IsNotExist(err) {
			if managed {
				plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Reason: "asset gestionado ausente; no se recrea sin estrategia explícita", SourceHash: asset.SourceSHA256, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
			} else {
				plan.Actions = append(plan.Actions, Action{Kind: "create-managed", Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: "asset nuevo del catálogo ausente; se instala sin tocar drift local", SourceHash: asset.SourceSHA256})
			}
			continue
		}
		if err != nil {
			return Plan{}, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Policy: asset.Policy, Reason: "destino symlink no permitido", SourceHash: asset.SourceSHA256})
			continue
		}
		if !info.Mode().IsRegular() {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Policy: asset.Policy, Reason: "destino no es archivo regular seguro", SourceHash: asset.SourceSHA256})
			continue
		}
		currentHash, err := assets.FileSHA256(targetPath)
		if err != nil {
			return Plan{}, err
		}
		if !managed {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Policy: asset.Policy, Reason: "archivo existente no gestionado", SourceHash: asset.SourceSHA256, CurrentHash: currentHash})
			continue
		}
		if currentHash != prev.TargetSHA256 {
			if asset.Policy == assets.PolicyMergeBlock {
				if _, err := renderMergeBlock(plan.SourceRoot, asset.SourceRel, plan.TargetRoot, asset.TargetRel); err != nil {
					plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Policy: asset.Policy, Reason: err.Error(), SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
					continue
				}
				plan.Actions = append(plan.Actions,
					Action{Kind: "backup", Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: "merge-block preserva texto local fuera de bloques; backup requerido", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256},
					Action{Kind: "merge-block", Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: "asset merge-block con drift local; se actualizan solo bloques LUFY", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256},
				)
				continue
			}
			if asset.Policy == assets.PolicyNoReplace && asset.SourceSHA256 != prev.SourceSHA256 {
				plan.Actions = append(plan.Actions, Action{Kind: "write-lufy-new", Source: asset.SourceRel, Target: asset.TargetRel + ".lufy-new", Policy: asset.Policy, Scope: asset.Scope, Reason: "asset no-replace con drift local; se escribe nueva versión sin tocar original", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
				continue
			}
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Policy: asset.Policy, Reason: "archivo gestionado con drift local", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
			continue
		}
		if asset.SourceSHA256 == prev.SourceSHA256 {
			plan.Actions = append(plan.Actions, Action{Kind: "skip", Source: asset.SourceRel, Target: asset.TargetRel, Reason: "source y target coinciden con el estado registrado", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
			continue
		}
		kind := "update-managed"
		reason := "source cambió y target coincide con hash registrado"
		if asset.Policy == assets.PolicyMergeBlock {
			if _, err := renderMergeBlock(plan.SourceRoot, asset.SourceRel, plan.TargetRoot, asset.TargetRel); err != nil {
				plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Policy: asset.Policy, Reason: err.Error(), SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
				continue
			}
			kind = "merge-block"
			reason = "source merge-block cambió; se actualizan solo bloques LUFY"
		}
		plan.Actions = append(plan.Actions,
			Action{Kind: "backup", Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: "source cambió y target no tiene drift; backup requerido", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256},
			Action{Kind: kind, Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: reason, SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256},
		)
	}
	for _, prev := range previous.Assets {
		if prev.TargetRel == agentsref.AgentsFile {
			exists, hasReference, err := agentsref.Status(target)
			if err != nil {
				plan.Conflicts = append(plan.Conflicts, Conflict{Path: prev.TargetRel, Reason: err.Error(), RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
				continue
			}
			if exists && !hasReference {
				currentHash, hashErr := assets.FileSHA256(filepath.Join(target, prev.TargetRel))
				if hashErr != nil {
					return Plan{}, hashErr
				}
				if currentHash != prev.TargetSHA256 {
					plan.Conflicts = append(plan.Conflicts, Conflict{Path: prev.TargetRel, Reason: "AGENTS.md legacy gestionado tiene drift local y falta @lufy-ia.harness.md; preservado, agrega la referencia manualmente o resuelve el drift antes de reintentar sync", CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
				} else {
					plan.Actions = append(plan.Actions, Action{Kind: "warn-agents-reference", Target: prev.TargetRel, Reason: "AGENTS.md legacy gestionado sale del manifest; agrega @lufy-ia.harness.md con install --yes o edición manual", CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
				}
			}
			continue
		}
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
	if previousAssets[agentsref.AgentsFile].TargetRel == "" {
		exists, hasReference, err := agentsref.Status(target)
		if err != nil {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: agentsref.AgentsFile, Reason: err.Error()})
		} else if !exists || !hasReference {
			plan.Actions = append(plan.Actions, Action{Kind: "warn-agents-reference", Target: agentsref.AgentsFile, Reason: "sync preserva AGENTS.md; agrega @lufy-ia.harness.md con install --yes o edición manual"})
		}
	}
	configPlan, err := toolruntime.PlanProjectConfig(harness.Tool, target, opts.NoEngram)
	if err != nil {
		return Plan{}, err
	}
	if configPlan.Action == "merge-json" {
		if fileExists(filepath.Join(target, configPlan.File)) {
			plan.Actions = append(plan.Actions, Action{Kind: "backup", Target: configPlan.File, Reason: "opencode.json existente será mergeado conservadoramente"})
		}
		plan.Actions = append(plan.Actions, Action{Kind: "merge-json", Target: configPlan.File, Reason: "configuración OpenCode merge-managed parcial"})
	} else {
		plan.Actions = append(plan.Actions, Action{Kind: "skip", Target: configPlan.File, Reason: "configuración OpenCode merge-managed sin cambios"})
	}
	sort.SliceStable(plan.Actions, func(i, j int) bool {
		order := map[string]int{"backup": 0, "create-managed": 1, "update-managed": 2, "merge-block": 3, "write-lufy-new": 4, "merge-json": 5, "retired": 6, "warn-agents-reference": 7, "skip": 8}
		if order[plan.Actions[i].Kind] == order[plan.Actions[j].Kind] {
			return plan.Actions[i].Target < plan.Actions[j].Target
		}
		return order[plan.Actions[i].Kind] < order[plan.Actions[j].Kind]
	})
	return plan, nil
}

func (s Service) apply(plan Plan, stdout io.Writer) error {
	updates := uniqueTargets(append(append(targetsForKind(plan.Actions, "create-managed"), targetsForKind(plan.Actions, "update-managed")...), targetsForKind(plan.Actions, "merge-block")...))
	lufyNew := targetsForKind(plan.Actions, "write-lufy-new")
	merges := targetsForKind(plan.Actions, "merge-json")
	stateOnly := hasActionKind(plan.Actions, "retired") || hasActionKind(plan.Actions, "warn-agents-reference")
	if len(updates) == 0 && len(lufyNew) == 0 && len(merges) == 0 && !stateOnly {
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
		case "create-managed":
			if err := copyFile(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
				return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
			}
			if action.Policy.SupportsAncestor() {
				if err := writeAncestor(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
					return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
				}
			}
			applied++
			fmt.Fprintf(stdout, "- [create-managed] %s\n", action.Target)
		case "update-managed":
			if err := copyFile(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
				return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
			}
			if action.Policy.SupportsAncestor() {
				if err := writeAncestor(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
					return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
				}
			}
			applied++
			fmt.Fprintf(stdout, "- [update-managed] %s\n", action.Target)
		case "merge-block":
			merged, err := renderMergeBlock(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target)
			if err != nil {
				return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
			}
			if err := writeTargetFile(plan.TargetRoot, action.Target, merged); err != nil {
				return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
			}
			if err := writeAncestor(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
				return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [merge-block] %s\n", action.Target)
		case "write-lufy-new":
			if err := copyFile(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
				return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [write-lufy-new] %s\n", action.Target)
		case "merge-json":
			if _, err := toolruntime.EnsureProjectConfig(plan.Harness.Tool, plan.TargetRoot, plan.NoEngram); err != nil {
				return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [merge-json] %s\n", action.Target)
		default:
			continue
		}
	}
	assetStates, err := buildAssetStates(plan, updates)
	if err != nil {
		return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
	}
	fingerprint, err := plan.Catalog.Fingerprint()
	if err != nil {
		return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
	}
	st := state.New(plan.TargetRoot, plan.Previous, assetStates, fingerprint)
	if err := state.WriteAtomic(plan.TargetRoot, st); err != nil {
		return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
	}
	fmt.Fprintf(stdout, "- [write] %s\n", filepath.Join(".lufy-ai", "install-state.json"))
	if err := verify.NewService().Run(verify.Options{Target: plan.TargetRoot, NoEngram: plan.NoEngram, AllowCatalogNewAssets: true, AllowMissingAgentsRef: true}, stdout); err != nil {
		return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
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
	lufyNew := map[string]bool{}
	for _, action := range plan.Actions {
		if action.Kind == "write-lufy-new" {
			lufyNew[trimLufyNewSuffix(action.Target)] = true
		}
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
		if lufyNew[asset.TargetRel] {
			prev.LastAction = "write-lufy-new"
			out = append(out, prev)
			continue
		}
		lastAction := prev.LastAction
		if updated[asset.TargetRel] {
			lastAction = "sync-update-managed"
		}
		assetState := state.AssetState{ID: asset.ID, SourceRel: asset.SourceRel, TargetRel: asset.TargetRel, SourceSHA256: asset.SourceSHA256, TargetSHA256: targetHash, Policy: string(asset.Policy), Scope: string(asset.Scope), Tool: string(asset.Tool), Methodology: string(asset.Methodology), Component: asset.Component, AncestorRel: prev.AncestorRel, AncestorHash: prev.AncestorHash, InstalledAt: installedAt, LastAction: lastAction}
		if asset.Policy.SupportsAncestor() && updated[asset.TargetRel] {
			ancestorRel, err := state.AncestorRel(asset.TargetRel)
			if err != nil {
				return nil, err
			}
			assetState.AncestorRel = ancestorRel
			assetState.AncestorHash = asset.SourceSHA256
		}
		out = append(out, assetState)
	}
	catalogTargets := map[string]bool{}
	for _, asset := range plan.Catalog.Assets {
		if asset.Kind == assets.KindFile {
			catalogTargets[asset.TargetRel] = true
		}
	}
	for _, prev := range plan.Previous.Assets {
		if prev.TargetRel == agentsref.AgentsFile {
			continue
		}
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
		out = append(out, state.AssetState{ID: prev.ID, SourceRel: prev.SourceRel, TargetRel: prev.TargetRel, SourceSHA256: prev.SourceSHA256, TargetSHA256: currentHash, Policy: prev.Policy, Scope: prev.Scope, Tool: prev.Tool, Methodology: prev.Methodology, Component: prev.Component, AncestorRel: prev.AncestorRel, AncestorHash: prev.AncestorHash, InstalledAt: prev.InstalledAt, LastAction: "retired"})
	}
	return out, nil
}

func trimLufyNewSuffix(path string) string {
	return strings.TrimSuffix(path, ".lufy-new")
}

func copyFile(sourceRoot, sourceRel, targetRoot, targetRel string) error {
	body, err := managedcontent.Render(sourceRoot, sourceRel, targetRoot, targetRel)
	if err != nil {
		return err
	}
	return writeTargetFile(targetRoot, targetRel, body)
}

func writeAncestor(sourceRoot, sourceRel, targetRoot, targetRel string) error {
	body, err := managedcontent.Render(sourceRoot, sourceRel, targetRoot, targetRel)
	if err != nil {
		return err
	}
	ancestorPath, err := state.AncestorPath(targetRoot, targetRel)
	if err != nil {
		return err
	}
	return platform.WriteFileAtomic(ancestorPath, body, 0o644)
}

func renderMergeBlock(sourceRoot, sourceRel, targetRoot, targetRel string) ([]byte, error) {
	sourceContent, err := readSourceContent(sourceRoot, sourceRel)
	if err != nil {
		return nil, err
	}
	targetPath, err := platform.SafeJoin(targetRoot, targetRel)
	if err != nil {
		return nil, err
	}
	targetContent, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, err
	}
	return mergeblock.Render(targetContent, sourceContent)
}

func readSourceContent(sourceRoot, sourceRel string) ([]byte, error) {
	var body []byte
	if sourceRoot == assets.EmbeddedSourceRoot {
		var err error
		body, err = assets.ReadSourceFile(sourceRoot, sourceRel)
		if err != nil {
			return nil, err
		}
	} else {
		src := filepath.Join(sourceRoot, sourceRel)
		info, err := os.Lstat(src)
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("source no es archivo regular seguro: %s", src)
		}
		body, err = os.ReadFile(src)
		if err != nil {
			return nil, err
		}
	}
	return body, nil
}

func writeTargetFile(targetRoot, targetRel string, body []byte) error {
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
		if action.Kind == "create-managed" || action.Kind == "update-managed" || action.Kind == "merge-block" || action.Kind == "write-lufy-new" || action.Kind == "backup" || action.Kind == "merge-json" {
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

func hasActionKind(actions []Action, kind string) bool {
	for _, action := range actions {
		if action.Kind == kind {
			return true
		}
	}
	return false
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

func syncRecoveryError(err error, targetRoot string, manifestPath string, applied int) error {
	if manifestPath == "" {
		return err
	}
	restored, rollbackErr := backup.RestoreCapturedFiles(targetRoot, manifestPath, io.Discard)
	if rollbackErr != nil {
		return fmt.Errorf("sync falló después de crear backup en %s; acciones aplicadas=%d; rollback automático falló: %v; restaura con: lufy-ai restore --target <target> --backup %s --yes: %w", manifestPath, applied, rollbackErr, manifestPath, err)
	}
	return fmt.Errorf("sync falló después de crear backup en %s; acciones aplicadas=%d; rollback automático restauró %d archivo(s): %w", manifestPath, applied, restored, err)
}

func shortHash(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}
