package syncer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/agentsref"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/backup"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/harnesscatalog"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/managedcontent"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/managedio"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/toolruntime"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/verify"
)

type Options struct {
	Target  string
	DryRun  bool
	Yes     bool
	Scope   assets.Scope
	Harness domain.HarnessConfig
}

type Action struct {
	Kind               ActionKind
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
	Catalog    assets.Catalog
	Previous   *state.InstallState
	Scope      assets.Scope
	GlobalRoot string
	Harness    domain.HarnessConfig
	Actions    []Action
	Conflicts  []Conflict
}

type planBuilder interface {
	Build(Options) (Plan, error)
}

type actionExecutor interface {
	Apply(Plan, io.Writer) error
}

type projectConfigEnsurer interface {
	Ensure(string) (bool, error)
}

type PlanBuilder struct{}

type ActionExecutor struct{}

type Service struct {
	planBuilder    planBuilder
	actionExecutor actionExecutor
	projectConfig  projectConfigEnsurer
}

func NewService() Service {
	return Service{
		planBuilder:    PlanBuilder{},
		actionExecutor: ActionExecutor{},
		projectConfig:  projectconfig.NewService(),
	}
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
	}
	plan, err := s.BuildPlan(opts)
	if err != nil {
		return err
	}

	printPlan(plan, stdout)
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
	if created, err := s.projectConfig.Ensure(plan.TargetRoot); err != nil {
		return err
	} else if created {
		fmt.Fprintf(stdout, "- [project-config] %s\n", projectconfig.ProjectConfigPath)
		plan, err = s.BuildPlan(opts)
		if err != nil {
			return err
		}
		if len(plan.Conflicts) > 0 {
			return fmt.Errorf("sync bloqueado por %d conflicto(s); resuelve drift/estado antes de reintentar", len(plan.Conflicts))
		}
	}
	if err := s.apply(plan, stdout); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Sync real completado")
	return nil
}

func (s Service) BuildPlan(opts Options) (Plan, error) {
	return s.planBuilder.Build(opts)
}

func (b PlanBuilder) Build(opts Options) (Plan, error) {
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
	plan := Plan{SourceRoot: sourceRoot, TargetRoot: target, Catalog: catalog, Previous: previous, Scope: scope, GlobalRoot: globalRoot, Harness: installedHarness}
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
		if managed && prev.Pinned {
			plan.Actions = append(plan.Actions, Action{Kind: ActionPinnedSkip, Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: "asset pinned; sync no lo modifica", SourceHash: asset.SourceSHA256, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
			continue
		}
		info, err := os.Lstat(targetPath)
		if os.IsNotExist(err) {
			if managed {
				plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Reason: "asset gestionado ausente; no se recrea sin estrategia explícita", SourceHash: asset.SourceSHA256, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
			} else {
				plan.Actions = append(plan.Actions, Action{Kind: ActionCreateManaged, Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: "asset nuevo del catálogo ausente; se instala sin tocar drift local", SourceHash: asset.SourceSHA256})
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
				if _, err := managedio.RenderMergeBlock(plan.SourceRoot, asset.SourceRel, plan.TargetRoot, asset.TargetRel); err != nil {
					plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Policy: asset.Policy, Reason: err.Error(), SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
					continue
				}
				plan.Actions = append(plan.Actions,
					Action{Kind: ActionBackup, Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: "merge-block preserva texto local fuera de bloques; backup requerido", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256},
					Action{Kind: ActionMergeBlock, Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: "asset merge-block con drift local; se actualizan solo bloques LUFY", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256},
				)
				continue
			}
			if asset.Policy == assets.PolicyNoReplace && asset.SourceSHA256 != prev.SourceSHA256 {
				plan.Actions = append(plan.Actions, Action{Kind: ActionWriteLufyNew, Source: asset.SourceRel, Target: asset.TargetRel + ".lufy-new", Policy: asset.Policy, Scope: asset.Scope, Reason: "asset no-replace con drift local; se escribe nueva versión sin tocar original", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
				continue
			}
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Policy: asset.Policy, Reason: "archivo gestionado con drift local", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
			continue
		}
		if asset.SourceSHA256 == prev.SourceSHA256 {
			plan.Actions = append(plan.Actions, Action{Kind: ActionSkip, Source: asset.SourceRel, Target: asset.TargetRel, Reason: "source y target coinciden con el estado registrado", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
			continue
		}
		kind := ActionUpdateManaged
		reason := "source cambió y target coincide con hash registrado"
		if asset.Policy == assets.PolicyMergeBlock {
			if _, err := managedio.RenderMergeBlock(plan.SourceRoot, asset.SourceRel, plan.TargetRoot, asset.TargetRel); err != nil {
				plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Policy: asset.Policy, Reason: err.Error(), SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
				continue
			}
			kind = ActionMergeBlock
			reason = "source merge-block cambió; se actualizan solo bloques LUFY"
		}
		plan.Actions = append(plan.Actions,
			Action{Kind: ActionBackup, Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: "source cambió y target no tiene drift; backup requerido", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256},
			Action{Kind: kind, Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: reason, SourceHash: asset.SourceSHA256, CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256},
		)
	}
	for _, prev := range previous.Assets {
		if prev.TargetRel == agentsref.AgentsFile {
			if prev.Pinned {
				plan.Actions = append(plan.Actions, Action{Kind: ActionPinnedSkip, Target: prev.TargetRel, Reason: "asset legacy pinned; sync no lo modifica", RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
				continue
			}
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
					plan.Conflicts = append(plan.Conflicts, Conflict{Path: prev.TargetRel, Reason: "AGENTS.md legacy gestionado tiene drift local y falta integración LUFY; preservado, agrega el bloque gestionado manualmente o resuelve el drift antes de reintentar sync", CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
				} else {
					plan.Actions = append(plan.Actions, Action{Kind: ActionWarnAgentsReference, Target: prev.TargetRel, Reason: "AGENTS.md legacy gestionado sale del manifest; agrega el bloque gestionado LUFY con install --yes o edición manual", CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
				}
			}
			continue
		}
		if catalogTargets[prev.TargetRel] {
			continue
		}
		if prev.Pinned {
			plan.Actions = append(plan.Actions, Action{Kind: ActionPinnedSkip, Target: prev.TargetRel, Reason: "asset retirado pinned; sync lo preserva en manifest", RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
			continue
		}
		targetPath, err := platform.SafeJoin(target, prev.TargetRel)
		if err != nil {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: prev.TargetRel, Reason: err.Error(), RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
			continue
		}
		info, err := os.Lstat(targetPath)
		if os.IsNotExist(err) {
			plan.Actions = append(plan.Actions, Action{Kind: ActionRetired, Target: prev.TargetRel, Reason: "asset previamente gestionado ya no está en el catálogo y tampoco existe en target", RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
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
		plan.Actions = append(plan.Actions, Action{Kind: ActionRetired, Target: prev.TargetRel, Reason: "asset previamente gestionado ya no está en el catálogo; se preserva y queda rastreado", CurrentHash: currentHash, RecordedSourceHash: prev.SourceSHA256, RecordedTargetHash: prev.TargetSHA256})
	}
	if previousAssets[agentsref.AgentsFile].TargetRel == "" {
		exists, hasReference, err := agentsref.Status(target)
		if err != nil {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: agentsref.AgentsFile, Reason: err.Error()})
		} else if !exists || !hasReference {
			plan.Actions = append(plan.Actions, Action{Kind: ActionWarnAgentsReference, Target: agentsref.AgentsFile, Reason: "sync preserva AGENTS.md; agrega el bloque gestionado LUFY con install --yes o edición manual"})
		}
	}
	configPlan, err := toolruntime.PlanProjectConfig(harness.Tool, target)
	if err != nil {
		return Plan{}, err
	}
	if configPlan.Action == string(ActionMergeJSON) {
		if managedio.FileExists(filepath.Join(target, configPlan.File)) {
			plan.Actions = append(plan.Actions, Action{Kind: ActionBackup, Target: configPlan.File, Reason: "opencode.json existente será mergeado conservadoramente"})
		}
		plan.Actions = append(plan.Actions, Action{Kind: ActionMergeJSON, Target: configPlan.File, Reason: "configuración OpenCode merge-managed parcial"})
	} else if configPlan.Action != "" {
		plan.Actions = append(plan.Actions, Action{Kind: ActionSkip, Target: configPlan.File, Reason: "configuración OpenCode merge-managed sin cambios"})
	}
	sort.SliceStable(plan.Actions, func(i, j int) bool {
		if actionOrder[plan.Actions[i].Kind] == actionOrder[plan.Actions[j].Kind] {
			return plan.Actions[i].Target < plan.Actions[j].Target
		}
		return actionOrder[plan.Actions[i].Kind] < actionOrder[plan.Actions[j].Kind]
	})
	return plan, nil
}

func (s Service) apply(plan Plan, stdout io.Writer) error {
	return s.actionExecutor.Apply(plan, stdout)
}

func (e ActionExecutor) Apply(plan Plan, stdout io.Writer) error {
	if err := validateActionKinds(plan.Actions); err != nil {
		return err
	}
	updates := managedio.UniqueTargets(append(append(targetsForKind(plan.Actions, ActionCreateManaged), targetsForKind(plan.Actions, ActionUpdateManaged)...), targetsForKind(plan.Actions, ActionMergeBlock)...))
	lufyNew := targetsForKind(plan.Actions, ActionWriteLufyNew)
	merges := targetsForKind(plan.Actions, ActionMergeJSON)
	stateOnly := hasActionKind(plan.Actions, ActionRetired) || hasActionKind(plan.Actions, ActionWarnAgentsReference)
	if len(updates) == 0 && len(lufyNew) == 0 && len(merges) == 0 && !stateOnly {
		fmt.Fprintln(stdout, "- [skip] sin cambios gestionados")
		return nil
	}
	backupTargets := managedio.UniqueTargets(append(append([]string{}, updates...), targetsForKind(plan.Actions, ActionBackup)...))
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
		case ActionCreateManaged:
			if err := managedio.CopyRenderedFile(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
				return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
			}
			if action.Policy.SupportsAncestor() {
				if err := managedio.WriteAncestor(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
					return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
				}
			}
			applied++
			fmt.Fprintf(stdout, "- [create-managed] %s\n", action.Target)
		case ActionUpdateManaged:
			if err := managedio.CopyRenderedFile(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
				return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
			}
			if action.Policy.SupportsAncestor() {
				if err := managedio.WriteAncestor(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
					return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
				}
			}
			applied++
			fmt.Fprintf(stdout, "- [update-managed] %s\n", action.Target)
		case ActionMergeBlock:
			merged, err := managedio.RenderMergeBlock(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target)
			if err != nil {
				return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
			}
			if err := managedio.WriteTargetFile(plan.TargetRoot, action.Target, merged); err != nil {
				return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
			}
			if err := managedio.WriteAncestor(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
				return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [merge-block] %s\n", action.Target)
		case ActionWriteLufyNew:
			if err := managedio.CopyRenderedFile(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
				return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [write-lufy-new] %s\n", action.Target)
		case ActionMergeJSON:
			if _, err := toolruntime.EnsureProjectConfig(plan.Harness.Tool, plan.TargetRoot); err != nil {
				return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [merge-json] %s\n", action.Target)
		case ActionBackup, ActionRetired, ActionWarnAgentsReference, ActionPinnedSkip, ActionSkip:
			continue
		default:
			err := fmt.Errorf("acción sync no soportada: %s", action.Kind)
			return syncRecoveryError(err, plan.TargetRoot, manifestPath, applied)
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
	fmt.Fprintf(stdout, "- [write] %s\n", filepath.ToSlash(filepath.Join(".lufy", "managed-state", "install-state.json")))
	if err := verify.NewService().Run(verify.Options{Target: plan.TargetRoot, AllowCatalogNewAssets: true, AllowMissingAgentsRef: true}, stdout); err != nil {
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
		if action.Kind == ActionWriteLufyNew {
			lufyNew[managedio.TrimLufyNewSuffix(action.Target)] = true
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
		if managed && prev.Pinned {
			out = append(out, prev)
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
			prev.LastAction = string(ActionWriteLufyNew)
			out = append(out, prev)
			continue
		}
		lastAction := prev.LastAction
		if updated[asset.TargetRel] {
			lastAction = "sync-update-managed"
		}
		assetState := state.AssetState{ID: asset.ID, SourceRel: asset.SourceRel, TargetRel: asset.TargetRel, SourceSHA256: asset.SourceSHA256, TargetSHA256: targetHash, Policy: string(asset.Policy), Scope: string(asset.Scope), Tool: string(asset.Tool), Methodology: string(asset.Methodology), Component: asset.Component, AncestorRel: prev.AncestorRel, AncestorHash: prev.AncestorHash, Pinned: prev.Pinned, PinnedAt: prev.PinnedAt, PinnedReason: prev.PinnedReason, InstalledAt: installedAt, LastAction: lastAction}
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
			if prev.Pinned {
				out = append(out, prev)
			}
			continue
		}
		if catalogTargets[prev.TargetRel] {
			continue
		}
		if prev.Pinned {
			out = append(out, prev)
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
		out = append(out, state.AssetState{ID: prev.ID, SourceRel: prev.SourceRel, TargetRel: prev.TargetRel, SourceSHA256: prev.SourceSHA256, TargetSHA256: currentHash, Policy: prev.Policy, Scope: prev.Scope, Tool: prev.Tool, Methodology: prev.Methodology, Component: prev.Component, AncestorRel: prev.AncestorRel, AncestorHash: prev.AncestorHash, Pinned: prev.Pinned, PinnedAt: prev.PinnedAt, PinnedReason: prev.PinnedReason, InstalledAt: prev.InstalledAt, LastAction: "retired"})
	}
	return out, nil
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
