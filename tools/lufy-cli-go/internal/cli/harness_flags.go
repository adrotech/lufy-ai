package cli

import (
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adapters/registry"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
)

type methodologyTierFlags []string

func (m *methodologyTierFlags) String() string {
	return strings.Join(*m, ",")
}

func (m *methodologyTierFlags) Set(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("methodology-tier no puede estar vacío")
	}
	*m = append(*m, value)
	return nil
}

type harnessFlagValues struct {
	Tool            *string
	MethodologyTier *methodologyTierFlags
}

func addHarnessFlags(fs *flag.FlagSet) harnessFlagValues {
	tool := fs.String("tool", string(domain.ToolInitialDefault), "Tool adapter efectivo: opencode o codex")
	methodologyTier := methodologyTierFlags{}
	fs.Var(&methodologyTier, "methodology-tier", "Override por tier: T1:openspec/full, T2:openspec/lite o T3:none; repetible")
	return harnessFlagValues{Tool: tool, MethodologyTier: &methodologyTier}
}

func addToolFlag(fs *flag.FlagSet) harnessFlagValues {
	tool := fs.String("tool", string(domain.ToolInitialDefault), "Tool adapter efectivo: opencode o codex")
	return harnessFlagValues{Tool: tool}
}

func parseHarnessFlags(values harnessFlagValues) (domain.HarnessConfig, error) {
	cfg := domain.DefaultHarnessConfig()
	if values.Tool != nil {
		tool := domain.ToolID(strings.TrimSpace(*values.Tool))
		if tool == "" {
			return domain.HarnessConfig{}, fmt.Errorf("--tool no puede estar vacío")
		}
		adapter, err := registry.Default().Tool(tool)
		if err != nil || adapter.Capabilities().DryRunOnly {
			return domain.HarnessConfig{}, fmt.Errorf("tool adapter no soportado para escritura: %s; disponibles: %s", tool, strings.Join(writableToolIDs(), ", "))
		}
		cfg.Tool = tool
	}
	if values.MethodologyTier != nil {
		for _, raw := range *values.MethodologyTier {
			tier, selection, err := parseMethodologyTier(raw)
			if err != nil {
				return domain.HarnessConfig{}, err
			}
			cfg.MethodologyByTier[tier] = selection
		}
	}
	if err := cfg.ValidateSupported(); err != nil {
		return domain.HarnessConfig{}, err
	}
	if err := cfg.MethodologyByTier.ValidateRoutingPolicy(domain.RoutingPolicyOptions{}); err != nil {
		return domain.HarnessConfig{}, err
	}
	return cfg, nil
}

func writableToolIDs() []string {
	reg := registry.Default()
	out := []string{}
	for _, id := range reg.ToolIDs() {
		adapter, err := reg.Tool(id)
		if err == nil && !adapter.Capabilities().DryRunOnly {
			out = append(out, string(id))
		}
	}
	sort.Strings(out)
	return out
}

func parseMethodologyTier(raw string) (domain.Tier, domain.MethodologySelection, error) {
	parts := strings.Split(strings.TrimSpace(raw), ":")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", domain.MethodologySelection{}, fmt.Errorf("--methodology-tier debe usar formato TIER:METHODOLOGY[/MODE]")
	}
	tier := domain.Tier(strings.TrimSpace(parts[0]))
	if !tier.Valid() {
		return "", domain.MethodologySelection{}, fmt.Errorf("tier no soportado en --methodology-tier: %s", tier)
	}
	methodParts := strings.Split(strings.TrimSpace(parts[1]), "/")
	if len(methodParts) > 2 || strings.TrimSpace(methodParts[0]) == "" {
		return "", domain.MethodologySelection{}, fmt.Errorf("--methodology-tier debe usar formato TIER:METHODOLOGY[/MODE]")
	}
	methodology := domain.MethodologyID(strings.TrimSpace(methodParts[0]))
	mode := domain.MethodologyMode("")
	if len(methodParts) == 2 {
		mode = domain.MethodologyMode(strings.TrimSpace(methodParts[1]))
		if mode == "" {
			return "", domain.MethodologySelection{}, fmt.Errorf("modo vacío en --methodology-tier: %s", raw)
		}
	}
	selection, err := inferMethodologySelection(tier, methodology, mode)
	if err != nil {
		return "", domain.MethodologySelection{}, err
	}
	return tier, selection, nil
}

func inferMethodologySelection(tier domain.Tier, methodology domain.MethodologyID, mode domain.MethodologyMode) (domain.MethodologySelection, error) {
	switch methodology {
	case domain.MethodologySpecWorkflow:
		if mode == "" {
			if tier == domain.TierT1 {
				mode = domain.MethodologyModeFull
			} else {
				mode = domain.MethodologyModeLite
			}
		}
		if mode != domain.MethodologyModeFull && mode != domain.MethodologyModeLite {
			return domain.MethodologySelection{}, fmt.Errorf("openspec requiere mode full o lite para %s", tier)
		}
		return domain.MethodologySelection{ID: methodology, Mode: mode, Required: true}, nil
	case domain.MethodologyNone:
		if mode == "" {
			mode = domain.MethodologyModeNone
		}
		if mode != domain.MethodologyModeNone {
			return domain.MethodologySelection{}, fmt.Errorf("none requiere mode none para %s", tier)
		}
		return domain.MethodologySelection{ID: methodology, Mode: mode, Required: false}, nil
	case domain.MethodologyLufyWorkflow:
		if mode == "" {
			if tier == domain.TierT1 {
				mode = domain.MethodologyModeFull
			} else {
				mode = domain.MethodologyModeLite
			}
		}
		if mode != domain.MethodologyModeFull && mode != domain.MethodologyModeLite {
			return domain.MethodologySelection{}, fmt.Errorf("lufy-sdd requiere mode full o lite para %s", tier)
		}
		return domain.MethodologySelection{ID: methodology, Mode: mode, Required: true}, nil
	default:
		return domain.MethodologySelection{}, fmt.Errorf("metodologia no soportada en CLI: %s; disponibles operativas: openspec, none", methodology)
	}
}
