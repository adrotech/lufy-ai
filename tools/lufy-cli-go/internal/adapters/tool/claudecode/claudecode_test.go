package claudecode

import (
	"context"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

func TestAdapterCapabilitiesAreDryRunOnly(t *testing.T) {
	if New().ID() != domain.ToolClaudeCode {
		t.Fatalf("unexpected adapter id: %s", New().ID())
	}
	caps := New().Capabilities()
	if !caps.DryRunOnly || !caps.ProjectConfig || !caps.SystemPrompt {
		t.Fatalf("expected claude-code dry-run project/system capabilities: %+v", caps)
	}
	if caps.Subagents || caps.SlashCommands || caps.Skills || caps.Hooks || caps.TUI || caps.GlobalConfig {
		t.Fatalf("claude-code capabilities overpromise native tool support: %+v", caps)
	}
}

func TestRenderSurfaceIsPreviewOnly(t *testing.T) {
	assets, err := New().RenderSurface(ports.HarnessModel{Tool: domain.ToolClaudeCode})
	if err != nil {
		t.Fatalf("render surface: %v", err)
	}
	if len(assets) == 0 {
		t.Fatal("expected preview assets")
	}
	for _, asset := range assets {
		if asset.Policy != "dry-run" || asset.Scope != "preview" {
			t.Fatalf("asset is not preview-only: %+v", asset)
		}
		if !strings.HasPrefix(asset.TargetRel, "CLAUDE.md") {
			t.Fatalf("claude-code preview target should be CLAUDE.md-based: %+v", asset)
		}
	}
}

func TestRenderSurfaceDoesNotLeakForbiddenToolPaths(t *testing.T) {
	assets, err := New().RenderSurface(ports.HarnessModel{Tool: domain.ToolClaudeCode})
	if err != nil {
		t.Fatalf("render surface: %v", err)
	}
	forbiddenPaths := []string{"." + "opencode", "opencode" + ".json"}
	for _, asset := range assets {
		text := strings.Join([]string{asset.ID, asset.TargetRel, asset.Policy, asset.Scope}, " ")
		for _, forbidden := range forbiddenPaths {
			if strings.Contains(text, forbidden) {
				t.Fatalf("claude-code render leaked %q in %+v", forbidden, asset)
			}
		}
	}
}

func TestDetectAndVerifyExplainDryRunOnly(t *testing.T) {
	detection := New().Detect(context.Background(), ports.Env{})
	if detection.Detected || !strings.Contains(detection.Reason, "dry-run") {
		t.Fatalf("detection = %+v", detection)
	}
	checks, err := New().Verify(ports.Target{Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if len(checks) == 0 || !strings.Contains(checks[0].Message, "dry-run") {
		t.Fatalf("checks = %+v", checks)
	}
}
