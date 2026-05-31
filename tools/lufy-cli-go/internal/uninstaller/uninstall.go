package uninstaller

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/agentsref"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/backup"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/toolruntime"
)

type Options struct {
	Target    string
	DryRun    bool
	Yes       bool
	KeepState bool
}

type Action struct {
	Kind         string
	Target       string
	Reason       string
	CurrentHash  string
	RecordedHash string
}

type Conflict struct {
	Path   string
	Reason string
}

type Plan struct {
	TargetRoot string
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
	printPlan(plan, opts, stdout)
	if len(plan.Conflicts) > 0 {
		return fmt.Errorf("uninstall bloqueado por %d conflicto(s); resuelve drift local o restaura backup antes de desinstalar", len(plan.Conflicts))
	}
	if opts.DryRun {
		fmt.Fprintln(stdout, "Modo dry-run: sin mutaciones en filesystem")
		return nil
	}
	if requiresConfirmation(plan.Actions) && !opts.Yes {
		return fmt.Errorf("uninstall requiere --yes para aplicar mutaciones reales; usa --dry-run para revisar el plan sin escribir")
	}
	if !hasMutation(plan.Actions) {
		fmt.Fprintln(stdout, "- [skip] sin assets gestionados para remover")
		return nil
	}

	lock, err := platform.AcquireLock(plan.TargetRoot)
	if err != nil {
		return err
	}
	defer lock.Release()

	backupTargets := backupTargetsFor(plan.Actions)
	backupDir := ""
	if len(backupTargets) > 0 {
		backupDir, err = backup.BackupFiles(plan.TargetRoot, backupTargets, "uninstall", stdout)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "- [backup] %s\n", backupDir)
	}
	for _, action := range plan.Actions {
		switch action.Kind {
		case "remove-file", "remove-ancestor", "remove-state":
			if err := removeFile(plan.TargetRoot, action.Target); err != nil {
				return uninstallRecoveryError(err, backupDir)
			}
			fmt.Fprintf(stdout, "- [%s] %s\n", action.Kind, action.Target)
		case "remove-agents-reference":
			changed, err := agentsref.RemoveReference(plan.TargetRoot)
			if err != nil {
				return uninstallRecoveryError(err, backupDir)
			}
			if changed {
				fmt.Fprintf(stdout, "- [remove-agents-reference] %s\n", action.Target)
			}
		}
	}
	for _, dir := range emptyDirsFor(plan.Actions) {
		if err := removeEmptyDir(plan.TargetRoot, dir); err != nil {
			return uninstallRecoveryError(err, backupDir)
		}
	}
	fmt.Fprintln(stdout, "Uninstall real completado")
	return nil
}

func (s Service) BuildPlan(opts Options) (Plan, error) {
	target, err := platform.ResolveTargetPath(opts.Target)
	if err != nil {
		return Plan{}, err
	}
	previous, err := state.Load(target)
	if err != nil {
		return Plan{}, err
	}
	if previous == nil {
		return Plan{}, fmt.Errorf("uninstall requiere %s; no hay instalación gestionada de Lufy", state.Path(target))
	}
	plan := Plan{TargetRoot: target, Previous: previous}
	for _, assetState := range previous.Assets {
		if filepath.ToSlash(assetState.TargetRel) == agentsref.AgentsFile {
			continue
		}
		action, conflict, ok := planManagedRemoval(target, assetState, "remove-file")
		if conflict != nil {
			plan.Conflicts = append(plan.Conflicts, *conflict)
			continue
		}
		if ok {
			plan.Actions = append(plan.Actions, action)
		}
		if assetState.AncestorRel != "" {
			ancestorState := state.AssetState{TargetRel: assetState.AncestorRel, TargetSHA256: assetState.AncestorHash}
			action, conflict, ok := planManagedRemoval(target, ancestorState, "remove-ancestor")
			if conflict != nil {
				plan.Conflicts = append(plan.Conflicts, *conflict)
				continue
			}
			if ok {
				plan.Actions = append(plan.Actions, action)
			}
		}
	}
	if exists, hasReference, err := agentsref.Status(target); err != nil {
		plan.Conflicts = append(plan.Conflicts, Conflict{Path: agentsref.AgentsFile, Reason: err.Error()})
	} else if exists && hasReference {
		plan.Actions = append(plan.Actions, Action{Kind: "remove-agents-reference", Target: agentsref.AgentsFile, Reason: "remover referencia user-owned al harness Lufy"})
	}
	if projectConfig, err := toolruntime.ProjectConfigFile(previous.Tool); err == nil {
		plan.Actions = append(plan.Actions, Action{Kind: "preserve", Target: projectConfig, Reason: "configuración merge-managed/user-owned preservada"})
	}
	if !opts.KeepState {
		plan.Actions = append(plan.Actions, Action{Kind: "remove-state", Target: stateRelPath(), Reason: "remover manifest de instalación gestionada"})
	}
	sort.SliceStable(plan.Actions, func(i, j int) bool {
		order := map[string]int{"remove-file": 0, "remove-ancestor": 1, "remove-agents-reference": 2, "remove-state": 3, "preserve": 4, "skip": 5}
		if order[plan.Actions[i].Kind] == order[plan.Actions[j].Kind] {
			return plan.Actions[i].Target < plan.Actions[j].Target
		}
		return order[plan.Actions[i].Kind] < order[plan.Actions[j].Kind]
	})
	return plan, nil
}

func planManagedRemoval(target string, assetState state.AssetState, kind string) (Action, *Conflict, bool) {
	rel, err := platform.EnsureRelativeSafe(assetState.TargetRel)
	if err != nil {
		return Action{}, &Conflict{Path: assetState.TargetRel, Reason: err.Error()}, false
	}
	path, err := platform.SafeJoin(target, rel)
	if err != nil {
		return Action{}, &Conflict{Path: rel, Reason: err.Error()}, false
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return Action{Kind: "skip", Target: rel, Reason: "archivo gestionado ya ausente"}, nil, false
	}
	if err != nil {
		return Action{}, &Conflict{Path: rel, Reason: err.Error()}, false
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return Action{}, &Conflict{Path: rel, Reason: "no es archivo regular seguro"}, false
	}
	current, err := assets.FileSHA256(path)
	if err != nil {
		return Action{}, &Conflict{Path: rel, Reason: err.Error()}, false
	}
	recorded := assetState.TargetSHA256
	if recorded == "" {
		recorded = assetState.SourceSHA256
	}
	if recorded != "" && current != recorded {
		return Action{}, &Conflict{Path: rel, Reason: fmt.Sprintf("drift local detectado: esperado=%s actual=%s", shortHash(recorded), shortHash(current))}, false
	}
	reason := "remover archivo gestionado sin drift local"
	if kind == "remove-ancestor" {
		reason = "remover ancestor gestionado sin drift local"
	}
	return Action{Kind: kind, Target: rel, Reason: reason, CurrentHash: current, RecordedHash: recorded}, nil, true
}

func printPlan(plan Plan, opts Options, stdout io.Writer) {
	fmt.Fprintf(stdout, "Plan de uninstall para %s\n", plan.TargetRoot)
	for _, conflict := range plan.Conflicts {
		fmt.Fprintf(stdout, "- [conflict] %s (%s)\n", conflict.Path, conflict.Reason)
	}
	for _, action := range plan.Actions {
		if action.CurrentHash != "" {
			fmt.Fprintf(stdout, "- [%s] %s (%s) current=%s recorded=%s\n", action.Kind, action.Target, action.Reason, shortHash(action.CurrentHash), shortHash(action.RecordedHash))
			continue
		}
		fmt.Fprintf(stdout, "- [%s] %s (%s)\n", action.Kind, action.Target, action.Reason)
	}
	if opts.KeepState {
		fmt.Fprintln(stdout, "Estado: se preservará .lufy-ai/install-state.json por --keep-state")
	}
}

func backupTargetsFor(actions []Action) []string {
	var out []string
	for _, action := range actions {
		switch action.Kind {
		case "remove-file", "remove-ancestor", "remove-agents-reference", "remove-state":
			out = append(out, action.Target)
		}
	}
	return out
}

func requiresConfirmation(actions []Action) bool {
	return hasMutation(actions)
}

func hasMutation(actions []Action) bool {
	for _, action := range actions {
		switch action.Kind {
		case "remove-file", "remove-ancestor", "remove-agents-reference", "remove-state":
			return true
		}
	}
	return false
}

func removeFile(target, rel string) error {
	path, err := platform.SafeJoin(target, rel)
	if err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s no es archivo regular seguro", rel)
	}
	return os.Remove(path)
}

func emptyDirsFor(actions []Action) []string {
	seen := map[string]bool{}
	for _, action := range actions {
		switch action.Kind {
		case "remove-file", "remove-ancestor":
			dir := filepath.ToSlash(filepath.Dir(action.Target))
			for dir != "." && dir != "/" && dir != "" {
				if strings.HasPrefix(dir, ".lufy-ai/backups") {
					break
				}
				seen[dir] = true
				parent := filepath.ToSlash(filepath.Dir(dir))
				if parent == dir {
					break
				}
				dir = parent
			}
		}
	}
	out := make([]string, 0, len(seen))
	for dir := range seen {
		out = append(out, dir)
	}
	sort.Slice(out, func(i, j int) bool {
		if strings.Count(out[i], "/") == strings.Count(out[j], "/") {
			return out[i] > out[j]
		}
		return strings.Count(out[i], "/") > strings.Count(out[j], "/")
	})
	return out
}

func removeEmptyDir(target, rel string) error {
	path, err := platform.SafeJoin(target, rel)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		if strings.Contains(err.Error(), "directory not empty") || strings.Contains(err.Error(), "not empty") {
			return nil
		}
		return err
	}
	return nil
}

func stateRelPath() string {
	return filepath.ToSlash(filepath.Join(".lufy-ai", "install-state.json"))
}

func uninstallRecoveryError(err error, backupDir string) error {
	if backupDir == "" {
		return err
	}
	return fmt.Errorf("uninstall falló después de crear backup en %s; puedes restaurar con: lufy-ai restore --target <target> --backup %s --yes: %w", backupDir, backupDir, err)
}

func shortHash(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}
