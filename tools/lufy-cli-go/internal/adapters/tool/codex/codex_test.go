package codex

import (
	"context"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

func TestAdapterCapabilitiesAreDryRunOnly(t *testing.T) {
	caps := New().Capabilities()
	if !caps.DryRunOnly || !caps.ProjectConfig || !caps.SystemPrompt {
		t.Fatalf("expected codex dry-run project/system capabilities: %+v", caps)
	}
	if caps.Subagents || caps.SlashCommands || caps.Skills || caps.Hooks || caps.TUI || caps.GlobalConfig {
		t.Fatalf("codex capabilities overpromise native tool support: %+v", caps)
	}
}

func TestRenderSurfaceIsPreviewOnly(t *testing.T) {
	assets, err := New().RenderSurface(ports.HarnessModel{Tool: domain.ToolCodex})
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
		if !strings.HasPrefix(asset.TargetRel, "AGENTS.md") {
			t.Fatalf("codex preview target should be AGENTS.md-based: %+v", asset)
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
