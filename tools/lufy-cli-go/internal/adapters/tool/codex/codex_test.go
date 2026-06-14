package codex

import (
	"context"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

func TestAdapterCapabilitiesAreWritableProjectSurface(t *testing.T) {
	if New().ID() != domain.ToolCodex {
		t.Fatalf("unexpected adapter id: %s", New().ID())
	}
	caps := New().Capabilities()
	if caps.DryRunOnly || !caps.ProjectConfig || !caps.SystemPrompt || !caps.Subagents || !caps.Skills || !caps.Hooks || !caps.MCP {
		t.Fatalf("expected codex writable project/system capabilities: %+v", caps)
	}
	if caps.SlashCommands || caps.TUI || caps.GlobalConfig {
		t.Fatalf("codex capabilities overpromise native tool support: %+v", caps)
	}
}

func TestRenderSurfaceIsManagedProjectSurface(t *testing.T) {
	assets, err := New().RenderSurface(ports.HarnessModel{Tool: domain.ToolCodex})
	if err != nil {
		t.Fatalf("render surface: %v", err)
	}
	if len(assets) == 0 {
		t.Fatal("expected managed assets")
	}
	targets := map[string]bool{}
	for _, asset := range assets {
		if asset.Policy != "managed" || asset.Scope != "project" {
			t.Fatalf("asset is not managed project surface: %+v", asset)
		}
		targets[asset.TargetRel] = true
	}
	for _, want := range []string{".agents/skills", ".codex"} {
		if !targets[want] {
			t.Fatalf("codex surface missing %s: %+v", want, assets)
		}
	}
}

func TestRenderSurfaceDoesNotLeakForbiddenToolPaths(t *testing.T) {
	assets, err := New().RenderSurface(ports.HarnessModel{Tool: domain.ToolCodex})
	if err != nil {
		t.Fatalf("render surface: %v", err)
	}
	forbiddenPaths := []string{"." + "opencode", "opencode" + ".json"}
	for _, asset := range assets {
		text := strings.Join([]string{asset.ID, asset.TargetRel, asset.Policy, asset.Scope}, " ")
		for _, forbidden := range forbiddenPaths {
			if strings.Contains(text, forbidden) {
				t.Fatalf("codex render leaked %q in %+v", forbidden, asset)
			}
		}
	}
}

func TestDetectAndVerifyExplainWritableAdapter(t *testing.T) {
	detection := New().Detect(context.Background(), ports.Env{})
	if !detection.Detected || !strings.Contains(detection.Reason, "codex") {
		t.Fatalf("detection = %+v", detection)
	}
	checks, err := New().Verify(ports.Target{Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if len(checks) < 2 || !strings.Contains(checks[0].Message, "structural") || !strings.Contains(checks[1].Message, "waited delegation") || !strings.Contains(checks[1].Message, "lufy-agent-mapping.md") {
		t.Fatalf("checks = %+v", checks)
	}
}
