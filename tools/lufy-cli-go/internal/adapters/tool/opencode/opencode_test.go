package opencode

import (
	"context"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

func TestAdapterCapabilities(t *testing.T) {
	if New().ID() != domain.ToolInitialDefault {
		t.Fatalf("unexpected adapter id: %s", New().ID())
	}
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

func TestDetectAndVerifyReportStructuralAdapter(t *testing.T) {
	detection := New().Detect(context.Background(), ports.Env{})
	if !detection.Detected || !strings.Contains(detection.Reason, "default") {
		t.Fatalf("detection = %+v", detection)
	}
	checks, err := New().Verify(ports.Target{Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if len(checks) != 1 || checks[0].Level != "info" || !strings.Contains(checks[0].Message, "structural") {
		t.Fatalf("checks = %+v", checks)
	}
}
