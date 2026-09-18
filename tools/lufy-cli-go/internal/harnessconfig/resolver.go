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
	Tier           domain.Tier
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
	if err != nil {
		return Resolution{}, err
	}
	result.ProjectConfig = cfg
	result.InstallState = st
	return result, nil
}

func ResolveLoaded(requested domain.HarnessConfig, cfg *projectconfig.ProjectConfig, st *state.InstallState, blockOnMismatch bool) (Resolution, error) {
	projectHarness := harnessFromProject(cfg)
	installedHarness := harnessFromState(st)
	drifts := Compare(projectHarness, installedHarness)

	if blockOnMismatch {
		unresolved := unresolvedDrifts(drifts, requested)
		if len(unresolved) > 0 {
			return Resolution{}, mismatchError(unresolved)
		}
	}

	defaults := domain.DefaultHarnessConfig()
	effective := defaults
	if requested.ToolIsExplicit() {
		effective.Tool = requested.Tool
		effective.Provenance.Tool = domain.HarnessSourceExplicit
	} else if projectHarness != nil {
		effective.Tool = projectHarness.Tool
		effective.Provenance.Tool = domain.HarnessSourceProjectConfig
	} else if installedHarness != nil {
		effective.Tool = installedHarness.Tool
		effective.Provenance.Tool = domain.HarnessSourceInstallState
	}
	for _, tier := range []domain.Tier{domain.TierT1, domain.TierT2, domain.TierT3} {
		switch {
		case requested.TierIsExplicit(tier):
			effective.MethodologyByTier[tier] = requested.MethodologyByTier[tier]
			effective.Provenance.MethodologyByTier[tier] = domain.HarnessSourceExplicit
		case projectHarness != nil:
			effective.MethodologyByTier[tier] = projectHarness.MethodologyByTier[tier]
			effective.Provenance.MethodologyByTier[tier] = domain.HarnessSourceProjectConfig
		case installedHarness != nil:
			effective.MethodologyByTier[tier] = installedHarness.MethodologyByTier[tier]
			effective.Provenance.MethodologyByTier[tier] = domain.HarnessSourceInstallState
		}
	}
	if err := effective.ValidateSupported(); err != nil {
		return Resolution{}, err
	}
	if err := effective.MethodologyByTier.ValidateRoutingPolicy(domain.RoutingPolicyOptions{}); err != nil {
		return Resolution{}, err
	}
	return Resolution{Config: effective, Drifts: drifts}, nil
}

func Compare(projectHarness, installedHarness *domain.HarnessConfig) []Drift {
	if projectHarness == nil || installedHarness == nil {
		return nil
	}
	project := projectHarness.WithDefaults()
	installed := installedHarness.WithDefaults()
	drifts := []Drift{}
	if project.Tool != installed.Tool {
		drifts = append(drifts, Drift{Field: "tool", ProjectValue: string(project.Tool), InstalledValue: string(installed.Tool)})
	}
	for _, tier := range []domain.Tier{domain.TierT1, domain.TierT2, domain.TierT3} {
		projectSelection := project.MethodologyByTier[tier]
		installedSelection := installed.MethodologyByTier[tier]
		if projectSelection != installedSelection {
			drifts = append(drifts, Drift{
				Field:          "methodology",
				Tier:           tier,
				ProjectValue:   formatSelection(projectSelection),
				InstalledValue: formatSelection(installedSelection),
			})
		}
	}
	return drifts
}

func RecoveryMessage(drift Drift) string {
	if drift.Field == "tool" {
		return fmt.Sprintf("tool project=%s install-state=%s; Recovery: ejecuta install con --tool explícito para reconciliar sin sobrescribir otros campos", drift.ProjectValue, drift.InstalledValue)
	}
	return fmt.Sprintf("methodology %s project=%s install-state=%s; Recovery: ejecuta install con --methodology-tier %s:<methodology>/<mode> explícito", drift.Tier, drift.ProjectValue, drift.InstalledValue, drift.Tier)
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
		return nil, fmt.Errorf("leer project config para resolver harness: %w", err)
	}
	return &cfg, nil
}

func harnessFromProject(cfg *projectconfig.ProjectConfig) *domain.HarnessConfig {
	if cfg == nil {
		return nil
	}
	harness := domain.HarnessConfig{Tool: cfg.Tool, MethodologyByTier: cfg.MethodologyByTier}.WithDefaults()
	return &harness
}

func harnessFromState(st *state.InstallState) *domain.HarnessConfig {
	if st == nil {
		return nil
	}
	harness := domain.HarnessConfig{Tool: st.Tool, MethodologyByTier: st.MethodologyByTier}.WithDefaults()
	return &harness
}

func unresolvedDrifts(drifts []Drift, requested domain.HarnessConfig) []Drift {
	unresolved := []Drift{}
	for _, drift := range drifts {
		if drift.Field == "tool" && requested.ToolIsExplicit() {
			continue
		}
		if drift.Field == "methodology" && requested.TierIsExplicit(drift.Tier) {
			continue
		}
		unresolved = append(unresolved, drift)
	}
	return unresolved
}

func mismatchError(drifts []Drift) error {
	messages := make([]string, 0, len(drifts))
	for _, drift := range drifts {
		messages = append(messages, RecoveryMessage(drift))
	}
	return fmt.Errorf("harness inconsistente; no se modificaron archivos: %s", strings.Join(messages, "; "))
}

func formatSelection(selection domain.MethodologySelection) string {
	return fmt.Sprintf("%s/%s(required=%t)", selection.ID, selection.Mode, selection.Required)
}
