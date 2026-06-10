package installer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
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
	Target   string
	DryRun   bool
	Yes      bool
	NoEngram bool
	Backup   bool
	Scope    assets.Scope
	Harness  domain.HarnessConfig
}

type Action struct {
	Kind                  ActionKind
	Source                string
	Target                string
	Policy                assets.Policy
	Scope                 assets.Scope
	Reason                string
	Risk                  string
	SourceHash            string
	CurrentHash           string
	PreviousInstalledHash string
}

type Conflict struct {
	Path        string
	Policy      assets.Policy
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

	printPlan(plan, opts.NoEngram, stdout)

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
	if created, err := s.projectConfig.Ensure(plan.TargetRoot); err != nil {
		return err
	} else if created {
		fmt.Fprintf(stdout, "- [project-config] %s\n", projectconfig.ProjectConfigPath)
		plan, err = s.BuildPlan(opts)
		if err != nil {
			return err
		}
		if len(plan.Conflicts) > 0 {
			return fmt.Errorf("install bloqueado por %d conflicto(s); resuelve manualmente y reintenta", len(plan.Conflicts))
		}
	}

	if err := s.applyInstall(plan, stdout); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Install real completado")
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
	harness := opts.Harness.WithDefaults()
	if err := harness.ValidateSupported(); err != nil {
		return Plan{}, err
	}
	if err := harness.MethodologyByTier.ValidateRoutingPolicy(domain.RoutingPolicyOptions{}); err != nil {
		return Plan{}, err
	}
	globalRoot := ""
	if scope == assets.ScopeGlobal || scope == assets.ScopeBoth {
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
	catalog, err = harnesscatalog.Effective(catalog, harness)
	if err != nil {
		return Plan{}, err
	}
	catalog, err = managedcontent.CatalogWithRenderedHashes(catalog, target)
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

	plan := Plan{SourceRoot: sourceRoot, TargetRoot: target, NoEngram: opts.NoEngram, Catalog: catalog, Previous: previous, Scope: scope, GlobalRoot: globalRoot, Harness: harness}
	seenDirs := map[string]bool{}
	for _, asset := range catalog.Assets {
		if asset.Kind == assets.KindDir {
			if !dirExists(filepath.Join(target, asset.TargetRel)) {
				plan.Actions = append(plan.Actions, Action{Kind: ActionMkdir, Source: asset.SourceRel, Target: asset.TargetRel, Reason: "directorio gestionado ausente", Risk: "low"})
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
				plan.Actions = append(plan.Actions, Action{Kind: ActionMkdir, Target: dir, Reason: "directorio padre requerido", Risk: "low"})
			}
		}

		targetPath, err := platform.SafeJoin(target, asset.TargetRel)
		if err != nil {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Reason: err.Error(), Risk: "high", SourceHash: asset.SourceSHA256})
			continue
		}
		info, err := os.Lstat(targetPath)
		if os.IsNotExist(err) {
			plan.Actions = append(plan.Actions, Action{Kind: ActionCopy, Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: "archivo gestionado ausente", Risk: "low", SourceHash: asset.SourceSHA256})
			continue
		}
		if err != nil {
			return Plan{}, err
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Policy: asset.Policy, Reason: "destino no es archivo regular seguro", Risk: "high", SourceHash: asset.SourceSHA256})
			continue
		}
		currentHash, err := assets.FileSHA256(targetPath)
		if err != nil {
			return Plan{}, err
		}
		if currentHash == asset.SourceSHA256 {
			plan.Actions = append(plan.Actions, Action{Kind: ActionSkip, Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: "hash destino coincide con source", Risk: "none", SourceHash: asset.SourceSHA256, CurrentHash: currentHash})
			continue
		}
		prev, managed := previousAssets[asset.TargetRel]
		if !managed {
			if asset.Policy == assets.PolicyMergeBlock {
				plan.Actions = append(plan.Actions, Action{Kind: ActionAdoptMergeBlock, Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: "archivo merge-block existente no gestionado; se adopta preservando contenido local", Risk: "low", SourceHash: asset.SourceSHA256, CurrentHash: currentHash})
				continue
			}
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Policy: asset.Policy, Reason: "archivo existente no gestionado con contenido distinto", Risk: "high", CurrentHash: currentHash, SourceHash: asset.SourceSHA256})
			continue
		}
		if currentHash != prev.TargetSHA256 {
			if asset.Policy == assets.PolicyMergeBlock {
				if _, err := managedio.RenderMergeBlock(plan.SourceRoot, asset.SourceRel, plan.TargetRoot, asset.TargetRel); err != nil {
					plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Policy: asset.Policy, Reason: err.Error(), Risk: "high", CurrentHash: currentHash, SourceHash: asset.SourceSHA256})
					continue
				}
				plan.Actions = append(plan.Actions,
					Action{Kind: ActionBackup, Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: "merge-block preserva texto local fuera de bloques; backup requerido", Risk: "medium", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, PreviousInstalledHash: prev.TargetSHA256},
					Action{Kind: ActionMergeBlock, Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: "archivo merge-block con drift local; se actualizan solo bloques LUFY", Risk: "medium", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, PreviousInstalledHash: prev.TargetSHA256},
				)
				continue
			}
			if asset.Policy == assets.PolicyNoReplace {
				plan.Actions = append(plan.Actions, Action{Kind: ActionWriteLufyNew, Source: asset.SourceRel, Target: asset.TargetRel + ".lufy-new", Policy: asset.Policy, Scope: asset.Scope, Reason: "archivo no-replace con drift local; se escribe nueva versión sin tocar original", Risk: "low", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, PreviousInstalledHash: prev.TargetSHA256})
				continue
			}
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Policy: asset.Policy, Reason: "archivo gestionado con drift local", Risk: "high", CurrentHash: currentHash, SourceHash: asset.SourceSHA256})
			continue
		}
		kind := ActionUpdateManaged
		reason := "source gestionado cambió sin drift local"
		if asset.Policy == assets.PolicyMergeBlock {
			if _, err := managedio.RenderMergeBlock(plan.SourceRoot, asset.SourceRel, plan.TargetRoot, asset.TargetRel); err != nil {
				plan.Conflicts = append(plan.Conflicts, Conflict{Path: asset.TargetRel, Policy: asset.Policy, Reason: err.Error(), Risk: "high", CurrentHash: currentHash, SourceHash: asset.SourceSHA256})
				continue
			}
			kind = ActionMergeBlock
			reason = "source merge-block cambió; se actualizan solo bloques LUFY"
		}
		plan.Actions = append(plan.Actions,
			Action{Kind: ActionBackup, Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: "source gestionado cambió; backup requerido antes de actualizar", Risk: "medium", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, PreviousInstalledHash: prev.TargetSHA256},
			Action{Kind: kind, Source: asset.SourceRel, Target: asset.TargetRel, Policy: asset.Policy, Scope: asset.Scope, Reason: reason, Risk: "medium", SourceHash: asset.SourceSHA256, CurrentHash: currentHash, PreviousInstalledHash: prev.TargetSHA256},
		)
	}
	if legacy, ok := previousAssets[agentsref.AgentsFile]; ok {
		exists, hasReference, err := agentsref.Status(target)
		if err != nil {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: agentsref.AgentsFile, Reason: err.Error(), Risk: "high", CurrentHash: legacy.TargetSHA256, SourceHash: legacy.SourceSHA256})
		} else if exists && !hasReference {
			currentHash, hashErr := assets.FileSHA256(filepath.Join(target, agentsref.AgentsFile))
			if hashErr != nil {
				return Plan{}, hashErr
			}
			if currentHash != legacy.TargetSHA256 {
				plan.Conflicts = append(plan.Conflicts, Conflict{Path: agentsref.AgentsFile, Reason: "AGENTS.md legacy gestionado tiene drift local y falta referencia; preservado, agrega @lufy-ia.harness.md manualmente", Risk: "high", CurrentHash: currentHash, SourceHash: legacy.SourceSHA256})
			}
		}
	}
	if !hasConflictForPath(plan.Conflicts, agentsref.AgentsFile) {
		exists, hasReference, err := agentsref.Status(target)
		if err != nil {
			plan.Conflicts = append(plan.Conflicts, Conflict{Path: agentsref.AgentsFile, Reason: err.Error(), Risk: "high"})
		} else if !exists {
			plan.Actions = append(plan.Actions, Action{Kind: ActionAgentsReferenceCreate, Target: agentsref.AgentsFile, Reason: "AGENTS.md user-owned ausente; se crea referencia mínima al harness", Risk: "low"})
		} else if hasReference {
			plan.Actions = append(plan.Actions, Action{Kind: ActionAgentsReferenceSkip, Target: agentsref.AgentsFile, Reason: "referencia al harness ya presente; AGENTS.md no se reescribe", Risk: "none"})
		} else {
			plan.Actions = append(plan.Actions,
				Action{Kind: ActionBackup, Target: agentsref.AgentsFile, Reason: "AGENTS.md user-owned recibirá referencia al harness", Risk: "medium"},
				Action{Kind: ActionAgentsReferenceInsert, Target: agentsref.AgentsFile, Reason: "se agrega solo @lufy-ia.harness.md sin copiar contenido completo de Lufy", Risk: "medium"},
			)
		}
	}
	if opts.Backup && previous != nil {
		alreadyBackup := map[string]bool{}
		for _, action := range plan.Actions {
			if action.Kind == ActionBackup {
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
			if managedio.FileExists(assetPath) {
				plan.Actions = append(plan.Actions, Action{Kind: ActionBackup, Target: asset.TargetRel, Reason: "backup solicitado explícitamente", Risk: "low", CurrentHash: asset.TargetSHA256})
			}
		}
	}
	configPlan, err := toolruntime.PlanProjectConfig(harness.Tool, target, opts.NoEngram)
	if err != nil {
		return Plan{}, err
	}
	if configPlan.Action == string(ActionMergeJSON) && managedio.FileExists(filepath.Join(target, configPlan.File)) {
		plan.Actions = append(plan.Actions, Action{Kind: ActionBackup, Target: configPlan.File, Reason: "opencode.json existente será mergeado", Risk: "medium"})
	}
	plan.Actions = append(plan.Actions, Action{Kind: ActionKind(configPlan.Action), Target: configPlan.File, Reason: "configuración OpenCode gestionada con merge conservador", Risk: "low"})
	plan.Actions = append(plan.Actions, Action{Kind: ActionVerify, Target: target, Reason: "verificación estructural posterior a install", Risk: "none"})
	sort.SliceStable(plan.Actions, func(i, j int) bool {
		if actionOrder[plan.Actions[i].Kind] == actionOrder[plan.Actions[j].Kind] {
			return plan.Actions[i].Target < plan.Actions[j].Target
		}
		return actionOrder[plan.Actions[i].Kind] < actionOrder[plan.Actions[j].Kind]
	})
	return plan, nil
}

func (s Service) applyInstall(plan Plan, stdout io.Writer) error {
	return s.actionExecutor.Apply(plan, stdout)
}

func (e ActionExecutor) Apply(plan Plan, stdout io.Writer) error {
	if err := os.MkdirAll(plan.TargetRoot, 0o755); err != nil {
		return err
	}
	backupTargets := targetsForKind(plan.Actions, ActionBackup)
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
		case ActionMkdir:
			targetPath, err := platform.SafeJoin(plan.TargetRoot, action.Target)
			if err != nil {
				return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
			}
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [mkdir] %s\n", action.Target)
		case ActionCopy, ActionUpdateManaged:
			if err := managedio.CopyRenderedFile(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
				return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
			}
			if action.Policy.SupportsAncestor() {
				if err := managedio.WriteAncestor(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
					return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
				}
			}
			applied++
			fmt.Fprintf(stdout, "- [%s] %s\n", action.Kind, action.Target)
		case ActionMergeBlock:
			merged, err := managedio.RenderMergeBlock(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target)
			if err != nil {
				return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
			}
			if err := managedio.WriteTargetFile(plan.TargetRoot, action.Target, merged); err != nil {
				return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
			}
			if err := managedio.WriteAncestor(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
				return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [merge-block] %s\n", action.Target)
		case ActionAdoptMergeBlock:
			if err := managedio.WriteAncestor(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
				return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [adopt-merge-block] %s\n", action.Target)
		case ActionWriteLufyNew:
			if err := managedio.CopyRenderedFile(plan.SourceRoot, action.Source, plan.TargetRoot, action.Target); err != nil {
				return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [write-lufy-new] %s\n", action.Target)
		case ActionMergeJSON:
			if _, err := toolruntime.EnsureProjectConfig(plan.Harness.Tool, plan.TargetRoot, plan.NoEngram); err != nil {
				return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [merge-json] %s\n", action.Target)
		case ActionAgentsReferenceCreate, ActionAgentsReferenceInsert:
			if err := agentsref.InsertReference(plan.TargetRoot); err != nil {
				return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
			}
			applied++
			fmt.Fprintf(stdout, "- [%s] %s\n", action.Kind, action.Target)
		case ActionVerify:
		case ActionSkip, ActionBackup, ActionAgentsReferenceSkip:
		default:
			err := fmt.Errorf("acción install no soportada: %s", action.Kind)
			return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
		}
	}
	if plan.Previous != nil && !hasContentMutation(plan.Actions) && !harnessConfigChanged(plan.Previous, plan.Harness) {
		if err := runPostInstallVerify(plan, stdout); err != nil {
			return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
		}
		fmt.Fprintf(stdout, "- [skip] %s sin cambios\n", filepath.ToSlash(filepath.Join(".lufy", "managed-state", "install-state.json")))
		return nil
	}
	assetStates, err := buildAssetStates(plan)
	if err != nil {
		return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
	}
	fingerprint, err := plan.Catalog.Fingerprint()
	if err != nil {
		return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
	}
	st := state.NewWithHarness(plan.TargetRoot, plan.Previous, assetStates, fingerprint, plan.Harness)
	if err := state.WriteAtomic(plan.TargetRoot, st); err != nil {
		return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
	}
	fmt.Fprintf(stdout, "- [write] %s\n", filepath.ToSlash(filepath.Join(".lufy", "managed-state", "install-state.json")))
	if err := runPostInstallVerify(plan, stdout); err != nil {
		if rollbackErr := restoreStateAfterVerifyFailure(plan); rollbackErr != nil {
			err = fmt.Errorf("%w; además falló restaurar install-state previo: %v", err, rollbackErr)
		}
		return installRecoveryError(err, plan.TargetRoot, recoveryBackup, applied)
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

func installRecoveryError(err error, targetRoot string, recoveryBackup string, applied int) error {
	if recoveryBackup == "" {
		return err
	}
	restored, rollbackErr := backup.RestoreCapturedFiles(targetRoot, recoveryBackup, io.Discard)
	if rollbackErr != nil {
		return fmt.Errorf("install falló después de crear backup de recovery en %s; acciones aplicadas=%d; rollback automático falló: %v: %w", recoveryBackup, applied, rollbackErr, err)
	}
	return fmt.Errorf("install falló después de crear backup de recovery en %s; acciones aplicadas=%d; rollback automático restauró %d archivo(s): %w", recoveryBackup, applied, restored, err)
}

func harnessConfigChanged(previous *state.InstallState, current domain.HarnessConfig) bool {
	if previous == nil {
		return true
	}
	prev := domain.HarnessConfig{Tool: previous.Tool, MethodologyByTier: previous.MethodologyByTier}.WithDefaults()
	next := current.WithDefaults()
	return !reflect.DeepEqual(prev, next)
}

func hasConflictForPath(conflicts []Conflict, path string) bool {
	for _, conflict := range conflicts {
		if conflict.Path == path {
			return true
		}
	}
	return false
}

func buildAssetStates(plan Plan) ([]state.AssetState, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	lastAction := map[string]string{}
	lufyNew := map[string]bool{}
	for _, action := range plan.Actions {
		if action.Kind == ActionCopy || action.Kind == ActionUpdateManaged || action.Kind == ActionMergeBlock || action.Kind == ActionAdoptMergeBlock || action.Kind == ActionSkip {
			lastAction[action.Target] = string(action.Kind)
		}
		if action.Kind == ActionWriteLufyNew {
			lufyNew[managedio.TrimLufyNewSuffix(action.Target)] = true
		}
	}
	previous := map[string]state.AssetState{}
	if plan.Previous != nil {
		previous = plan.Previous.AssetMap()
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
		if lufyNew[asset.TargetRel] {
			prev := previous[asset.TargetRel]
			prev.LastAction = string(ActionWriteLufyNew)
			out = append(out, prev)
			continue
		}
		action := lastAction[asset.TargetRel]
		if action == "" {
			action = "record"
		}
		assetState := state.AssetState{ID: asset.ID, SourceRel: asset.SourceRel, TargetRel: asset.TargetRel, SourceSHA256: asset.SourceSHA256, TargetSHA256: targetHash, Policy: string(asset.Policy), Scope: string(asset.Scope), Tool: string(asset.Tool), Methodology: string(asset.Methodology), Component: asset.Component, InstalledAt: now, LastAction: action}
		if prev, ok := previous[asset.TargetRel]; ok {
			assetState.InstalledAt = prev.InstalledAt
			assetState.AncestorRel = prev.AncestorRel
			assetState.AncestorHash = prev.AncestorHash
		}
		if asset.Policy.SupportsAncestor() && (action == string(ActionCopy) || action == string(ActionUpdateManaged) || action == string(ActionMergeBlock) || action == string(ActionAdoptMergeBlock)) {
			ancestorRel, err := state.AncestorRel(asset.TargetRel)
			if err != nil {
				return nil, err
			}
			assetState.AncestorRel = ancestorRel
			assetState.AncestorHash = asset.SourceSHA256
		}
		out = append(out, assetState)
	}
	return out, nil
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

func shortHash(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}
