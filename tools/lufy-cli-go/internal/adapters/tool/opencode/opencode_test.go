package opencode

import (
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

func TestAdapterCapabilities(t *testing.T) {
	caps := New().Capabilities()
	if !caps.Subagents || !caps.SlashCommands || !caps.Skills || !caps.TUI {
		t.Fatalf("expected opencode capabilities to include subagents, commands, skills and TUI: %+v", caps)
	}
}

func TestRenderSurfaceIncludesCurrentOpenCodeAssets(t *testing.T) {
	assets, err := New().RenderSurface(ports.HarnessModel{Tool: domain.ToolInitialDefault})
	if err != nil {
		t.Fatalf("render surface: %v", err)
	}
	byTarget := map[string]ports.AssetSpec{}
	for _, asset := range assets {
		byTarget[asset.TargetRel] = asset
	}
	for _, target := range []string{".opencode/agents", ".opencode/commands", ".opencode/skills", "opencode.json", "tui.json"} {
		if _, ok := byTarget[target]; !ok {
			t.Fatalf("missing target %s", target)
		}
	}
	if byTarget["opencode.json"].Policy != "merge-json" {
		t.Fatalf("opencode.json policy = %s", byTarget["opencode.json"].Policy)
	}
}
