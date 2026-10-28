package cli

import (
	"flag"
	"fmt"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adapters/registry"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
)

type harnessFlagValues struct {
	Tool               *string
	ToolExplicit       *bool
	Delegation         *string
	DelegationExplicit *bool
}
type trackedStringFlag struct {
	value    *string
	explicit *bool
}

func (f trackedStringFlag) String() string {
	if f.value == nil {
		return ""
	}
	return *f.value
}
func (f trackedStringFlag) Set(value string) error { *f.value = value; *f.explicit = true; return nil }
func addToolFlag(fs *flag.FlagSet) harnessFlagValues {
	tool, explicit := string(domain.ToolCodex), false
	fs.Var(trackedStringFlag{&tool, &explicit}, "tool", "Tool adapter: "+writableToolList())
	return harnessFlagValues{Tool: &tool, ToolExplicit: &explicit}
}
func addHarnessFlags(fs *flag.FlagSet) harnessFlagValues {
	values := addToolFlag(fs)
	delegation, explicit := string(domain.DelegationAuto), false
	fs.Var(trackedStringFlag{&delegation, &explicit}, "delegation", "Delegación: auto, on_request u off; no amplía permisos")
	values.Delegation, values.DelegationExplicit = &delegation, &explicit
	return values
}
func parseHarnessFlags(values harnessFlagValues) (domain.HarnessConfig, error) {
	cfg := domain.DefaultHarnessConfig()
	if values.Tool != nil && (values.ToolExplicit == nil || *values.ToolExplicit) {
		cfg.Tool = domain.ToolID(strings.TrimSpace(*values.Tool))
		cfg.Provenance.Tool = domain.HarnessSourceExplicit
		if !writableToolSupported(cfg.Tool) {
			return domain.HarnessConfig{}, fmt.Errorf("tool adapter no soportado para escritura: %q; disponible: codex", cfg.Tool)
		}
	}
	if values.Delegation != nil && (values.DelegationExplicit == nil || *values.DelegationExplicit) {
		cfg.Execution.Delegation = domain.DelegationPolicy(strings.TrimSpace(*values.Delegation))
		cfg.Provenance.Execution = domain.HarnessSourceExplicit
		if cfg.Execution.Delegation == "" {
			return domain.HarnessConfig{}, fmt.Errorf("--delegation no puede estar vacío")
		}
	}
	if err := cfg.ValidateSupported(); err != nil {
		return domain.HarnessConfig{}, err
	}
	return cfg, nil
}
func writableToolIDs() []domain.ToolID { return registry.Default().WritableToolIDs() }
func writableToolSupported(tool domain.ToolID) bool {
	for _, id := range writableToolIDs() {
		if id == tool {
			return true
		}
	}
	return false
}
func writableToolList() string  { return strings.Join(toolIDStrings(writableToolIDs()), ", ") }
func writableToolUsage() string { return strings.Join(toolIDStrings(writableToolIDs()), "|") }
func toolIDStrings(ids []domain.ToolID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}
