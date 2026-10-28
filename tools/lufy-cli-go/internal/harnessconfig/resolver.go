package harnessconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

type Options struct {
	Target          string
	Requested       domain.HarnessConfig
	BlockOnMismatch bool
}
type Drift struct {
	Field          string
	ProjectValue   string
	InstalledValue string
}
type Resolution struct {
	Config        domain.HarnessConfig
	ProjectConfig *projectconfig.ProjectConfig
	InstallState  *state.InstallState
	Drifts        []Drift
}

func Resolve(opts Options) (Resolution, error) {
	target, err := platform.ResolveTargetPath(opts.Target)
	if err != nil {
		return Resolution{}, err
	}
	cfg, err := loadProjectConfig(target)
	if err != nil {
		return Resolution{}, err
	}
	st, err := state.Load(target)
	if err != nil {
		return Resolution{}, err
	}
	result, err := ResolveLoaded(opts.Requested, cfg, st, opts.BlockOnMismatch)
	result.ProjectConfig, result.InstallState = cfg, st
	return result, err
}
func ResolveLoaded(requested domain.HarnessConfig, cfg *projectconfig.ProjectConfig, st *state.InstallState, block bool) (Resolution, error) {
	project, installed := harnessFromProject(cfg), harnessFromState(st)
	for _, value := range []*domain.HarnessConfig{project, installed} {
		if value != nil {
			if err := value.ValidateSupported(); err != nil {
				return Resolution{}, err
			}
		}
	}
	drifts := Compare(project, installed)
	if block {
		var unresolved []string
		for _, drift := range drifts {
			if drift.Field == "tool" && requested.ToolIsExplicit() {
				continue
			}
			if drift.Field == "execution" && requested.ExecutionIsExplicit() {
				continue
			}
			unresolved = append(unresolved, RecoveryMessage(drift))
		}
		if len(unresolved) > 0 {
			return Resolution{}, fmt.Errorf("harness inconsistente; no se modificaron archivos: %s", strings.Join(unresolved, "; "))
		}
	}
	effective := domain.DefaultHarnessConfig()
	if installed != nil {
		effective = *installed
		effective.Provenance = domain.HarnessProvenance{Tool: domain.HarnessSourceInstallState, Execution: domain.HarnessSourceInstallState}
	}
	if project != nil {
		effective = *project
		effective.Provenance = domain.HarnessProvenance{Tool: domain.HarnessSourceProjectConfig, Execution: domain.HarnessSourceProjectConfig}
	}
	if requested.ToolIsExplicit() {
		effective.Tool = requested.Tool
		effective.Provenance.Tool = domain.HarnessSourceExplicit
	}
	if requested.ExecutionIsExplicit() {
		effective.Execution = requested.Execution
		effective.Provenance.Execution = domain.HarnessSourceExplicit
	}
	effective = effective.WithDefaults()
	if err := effective.ValidateSupported(); err != nil {
		return Resolution{}, err
	}
	return Resolution{Config: effective, Drifts: drifts}, nil
}
func Compare(project, installed *domain.HarnessConfig) []Drift {
	if project == nil || installed == nil {
		return nil
	}
	p, i := project.WithDefaults(), installed.WithDefaults()
	var result []Drift
	if p.Tool != i.Tool {
		result = append(result, Drift{Field: "tool", ProjectValue: string(p.Tool), InstalledValue: string(i.Tool)})
	}
	if p.Execution != i.Execution {
		result = append(result, Drift{Field: "execution", ProjectValue: string(p.Execution.Delegation), InstalledValue: string(i.Execution.Delegation)})
	}
	return result
}
func RecoveryMessage(drift Drift) string {
	if drift.Field == "tool" {
		return fmt.Sprintf("tool project=%s install-state=%s; usa install --tool codex para reconciliar", drift.ProjectValue, drift.InstalledValue)
	}
	return fmt.Sprintf("execution.delegation project=%s install-state=%s; usa install --delegation %s para reconciliar", drift.ProjectValue, drift.InstalledValue, drift.ProjectValue)
}
func loadProjectConfig(target string) (*projectconfig.ProjectConfig, error) {
	path, err := projectconfig.ExistingPath(target)
	if err != nil {
		return nil, err
	}
	cfg, err := projectconfig.Load(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("leer project config: %w", err)
	}
	return &cfg, nil
}
func harnessFromProject(cfg *projectconfig.ProjectConfig) *domain.HarnessConfig {
	if cfg == nil {
		return nil
	}
	value := domain.HarnessConfig{Tool: cfg.Tool, Execution: cfg.Execution}.WithDefaults()
	return &value
}
func harnessFromState(st *state.InstallState) *domain.HarnessConfig {
	if st == nil {
		return nil
	}
	value := domain.HarnessConfig{Tool: st.Tool, Execution: st.Execution}.WithDefaults()
	return &value
}
