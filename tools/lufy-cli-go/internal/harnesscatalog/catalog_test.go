package harnesscatalog

import (
	"fmt"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
	"testing"
)

func TestEffectiveExcludesUnselectedSources(t *testing.T) {
	b := assets.Catalog{Assets: []assets.Asset{
		{TargetRel: "lufy-ia.harness.md"}, {TargetRel: ".codex/config.toml"}, {TargetRel: ".codex/agents/implementer.toml"}, {TargetRel: ".agents/skills/lufy-mem-search/SKILL.md"}, {TargetRel: ".agents/skills/sdd-workflow/SKILL.md"}, {TargetRel: "openspec/config.yaml"},
	}}
	c, err := Effective(b, domain.DefaultHarnessConfig())
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Assets) != 3 {
		t.Fatalf("unexpected selected assets: %#v", c.Assets)
	}
	if _, err := Effective(b, domain.HarnessConfig{Tool: domain.ToolID("opencode")}); err == nil {
		t.Fatal("accepted inactive adapter")
	}
}
func TestEffectiveReturnsRegistryError(t *testing.T) {
	if _, err := EffectiveWithRegistry(assets.Catalog{}, domain.DefaultHarnessConfig(), failingRegistry{}); err == nil {
		t.Fatal("missing registry error")
	}
}

type failingRegistry struct{}

func (failingRegistry) Tool(domain.ToolID) (ports.ToolAdapter, error) {
	return nil, fmt.Errorf("unavailable")
}
