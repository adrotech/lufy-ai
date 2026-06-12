package harnesscatalog

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

func TestEffectiveDefaultPreservesOpenCodeOpenSpecPreset(t *testing.T) {
	effective, err := Effective(testCatalog(), domain.DefaultHarnessConfig())
	if err != nil {
		t.Fatalf("effective catalog: %v", err)
	}

	for _, target := range []string{
		"lufy-ia.harness.md",
		filepath.Join(".opencode", "agents", "orchestrator.md"),
		filepath.Join(".opencode", "commands", "opsx-apply.md"),
		filepath.Join(".opencode", "skills", "sdd-workflow", "openspec-sync", "SKILL.md"),
		filepath.Join("openspec", "config.yaml"),
		"tui.json",
	} {
		if !hasTarget(effective, target) {
			t.Fatalf("default effective catalog missing %s", target)
		}
	}
	if hasTarget(effective, filepath.Join(".lufy", "workflows", "sdd", "README.md")) {
		t.Fatalf("default effective catalog includes lufy-sdd")
	}
}

func TestEffectiveLufySDDLiteOmitsOpenSpecAndSpecs(t *testing.T) {
	harness := domain.HarnessConfig{
		Tool: domain.ToolInitialDefault,
		MethodologyByTier: domain.MethodologyByTier{
			domain.TierT1: {ID: domain.MethodologyLufyWorkflow, Mode: domain.MethodologyModeLite, Required: true},
			domain.TierT2: {ID: domain.MethodologyLufyWorkflow, Mode: domain.MethodologyModeLite, Required: true},
			domain.TierT3: {ID: domain.MethodologyNone, Mode: domain.MethodologyModeNone, Required: false},
		},
	}
	effective, err := Effective(testCatalog(), harness)
	if err != nil {
		t.Fatalf("effective catalog: %v", err)
	}

	if hasTarget(effective, filepath.Join("openspec", "config.yaml")) {
		t.Fatalf("lufy-sdd catalog includes openspec config")
	}
	if hasTarget(effective, filepath.Join(".opencode", "commands", "opsx-apply.md")) {
		t.Fatalf("lufy-sdd catalog includes openspec command")
	}
	if !hasTarget(effective, filepath.Join(".lufy", "workflows", "sdd", "README.md")) {
		t.Fatalf("lufy-sdd catalog missing readme")
	}
	if !hasTarget(effective, filepath.Join(".lufy", "workflows", "sdd", "changes", ".gitkeep")) {
		t.Fatalf("lufy-sdd lite catalog missing changes")
	}
	if hasTarget(effective, filepath.Join(".lufy", "workflows", "sdd", "specs", ".gitkeep")) {
		t.Fatalf("lufy-sdd lite catalog includes specs")
	}
}

func TestEffectiveLufySDDFullIncludesSpecs(t *testing.T) {
	harness := domain.HarnessConfig{
		Tool: domain.ToolInitialDefault,
		MethodologyByTier: domain.MethodologyByTier{
			domain.TierT1: {ID: domain.MethodologyLufyWorkflow, Mode: domain.MethodologyModeFull, Required: true},
			domain.TierT2: {ID: domain.MethodologyNone, Mode: domain.MethodologyModeNone, Required: false},
			domain.TierT3: {ID: domain.MethodologyNone, Mode: domain.MethodologyModeNone, Required: false},
		},
	}
	effective, err := Effective(testCatalog(), harness)
	if err != nil {
		t.Fatalf("effective catalog: %v", err)
	}
	if !hasTarget(effective, filepath.Join(".lufy", "workflows", "sdd", "specs", ".gitkeep")) {
		t.Fatalf("lufy-sdd full catalog missing specs")
	}
}

func TestEffectiveCodexIncludesCodexSurfaceAndOmitsOpenCode(t *testing.T) {
	harness := domain.HarnessConfig{
		Tool: domain.ToolCodex,
		MethodologyByTier: domain.MethodologyByTier{
			domain.TierT1: {ID: domain.MethodologySpecWorkflow, Mode: domain.MethodologyModeFull, Required: true},
			domain.TierT2: {ID: domain.MethodologySpecWorkflow, Mode: domain.MethodologyModeLite, Required: true},
			domain.TierT3: {ID: domain.MethodologyNone, Mode: domain.MethodologyModeNone, Required: false},
		},
	}
	effective, err := Effective(testCatalog(), harness)
	if err != nil {
		t.Fatalf("effective catalog: %v", err)
	}
	for _, target := range []string{
		filepath.Join(".agents", "skills", "lufy-close", "SKILL.md"),
		filepath.Join(".agents", "skills", "sdd-workflow", "SKILL.md"),
		filepath.Join(".codex", "config.toml"),
		filepath.Join(".codex", "agents", "implementer.toml"),
		filepath.Join("openspec", "config.yaml"),
	} {
		if !hasTarget(effective, target) {
			t.Fatalf("codex effective catalog missing %s", target)
		}
	}
	if hasTarget(effective, filepath.Join(".opencode", "agents", "orchestrator.md")) {
		t.Fatalf("codex effective catalog includes opencode agent")
	}
	if hasTarget(effective, "tui.json") {
		t.Fatalf("codex effective catalog includes opencode tui")
	}
}

func TestEffectiveReturnsAdapterErrors(t *testing.T) {
	_, err := EffectiveWithRegistry(testCatalog(), domain.DefaultHarnessConfig(), failingRegistry{})
	if err == nil {
		t.Fatalf("expected adapter error")
	}
}

type failingRegistry struct{}

func (failingRegistry) Tool(domain.ToolID) (ports.ToolAdapter, error) {
	return fakeToolAdapter{}, nil
}

func (failingRegistry) Methodology(id domain.MethodologyID) (ports.MethodologyAdapter, error) {
	return nil, fmt.Errorf("missing methodology %s", id)
}

type fakeToolAdapter struct{}

func (fakeToolAdapter) ID() domain.ToolID { return domain.ToolInitialDefault }

func (fakeToolAdapter) Capabilities() domain.ToolCapabilities { return domain.ToolCapabilities{} }

func (fakeToolAdapter) Detect(context.Context, ports.Env) ports.DetectionResult {
	return ports.DetectionResult{Detected: true}
}

func (fakeToolAdapter) RenderSurface(ports.HarnessModel) ([]ports.AssetSpec, error) {
	return nil, nil
}

func (fakeToolAdapter) Verify(ports.Target) ([]ports.Check, error) {
	return nil, nil
}

func testCatalog() assets.Catalog {
	return assets.Catalog{SourceRoot: "test", Assets: []assets.Asset{
		{TargetRel: "lufy-ia.harness.md", Kind: assets.KindFile, Policy: assets.PolicyManaged, Scope: assets.ScopeProject, Methodology: domain.MethodologyNone, Component: "harness-reference"},
		{TargetRel: filepath.Join(".opencode", "agents", "orchestrator.md"), Kind: assets.KindFile, Policy: assets.PolicyManaged, Scope: assets.ScopeProject, Methodology: domain.MethodologyNone, Component: "instruction-surface"},
		{TargetRel: filepath.Join(".opencode", "commands", "opsx-apply.md"), Kind: assets.KindFile, Policy: assets.PolicyManaged, Scope: assets.ScopeProject, Methodology: domain.MethodologySpecWorkflow, Component: "methodology-command"},
		{TargetRel: filepath.Join(".opencode", "skills", "sdd-workflow", "openspec-sync", "SKILL.md"), Kind: assets.KindFile, Policy: assets.PolicyManaged, Scope: assets.ScopeProject, Methodology: domain.MethodologySpecWorkflow, Component: "methodology-skill"},
		{TargetRel: filepath.Join(".agents", "skills", "lufy-close", "SKILL.md"), Kind: assets.KindFile, Policy: assets.PolicyManaged, Scope: assets.ScopeProject, Tool: domain.ToolCodex, Methodology: domain.MethodologyNone, Component: "instruction-surface"},
		{TargetRel: filepath.Join(".agents", "skills", "sdd-workflow", "SKILL.md"), Kind: assets.KindFile, Policy: assets.PolicyManaged, Scope: assets.ScopeProject, Tool: domain.ToolCodex, Methodology: domain.MethodologySpecWorkflow, Component: "methodology-skill"},
		{TargetRel: filepath.Join(".codex", "config.toml"), Kind: assets.KindFile, Policy: assets.PolicyManaged, Scope: assets.ScopeProject, Tool: domain.ToolCodex, Methodology: domain.MethodologyNone, Component: "instruction-surface"},
		{TargetRel: filepath.Join(".codex", "agents", "implementer.toml"), Kind: assets.KindFile, Policy: assets.PolicyManaged, Scope: assets.ScopeProject, Tool: domain.ToolCodex, Methodology: domain.MethodologyNone, Component: "instruction-surface"},
		{TargetRel: filepath.Join("openspec", "config.yaml"), Kind: assets.KindFile, Policy: assets.PolicyManaged, Scope: assets.ScopeProject, Methodology: domain.MethodologySpecWorkflow, Component: "methodology-surface"},
		{TargetRel: "tui.json", Kind: assets.KindFile, Policy: assets.PolicyNoReplace, Scope: assets.ScopeProject, Methodology: domain.MethodologyNone, Component: "tool-ui"},
		{TargetRel: filepath.Join(".lufy", "workflows", "sdd", "README.md"), Kind: assets.KindFile, Policy: assets.PolicyManaged, Scope: assets.ScopeProject, Methodology: domain.MethodologyLufyWorkflow, Component: "methodology-surface"},
		{TargetRel: filepath.Join(".lufy", "workflows", "sdd", "changes", ".gitkeep"), Kind: assets.KindFile, Policy: assets.PolicyManaged, Scope: assets.ScopeProject, Methodology: domain.MethodologyLufyWorkflow, Component: "methodology-surface"},
		{TargetRel: filepath.Join(".lufy", "workflows", "sdd", "specs", ".gitkeep"), Kind: assets.KindFile, Policy: assets.PolicyManaged, Scope: assets.ScopeProject, Methodology: domain.MethodologyLufyWorkflow, Component: "methodology-surface"},
	}}
}

func hasTarget(catalog assets.Catalog, target string) bool {
	target = filepath.ToSlash(target)
	for _, asset := range catalog.Assets {
		if filepath.ToSlash(asset.TargetRel) == target {
			return true
		}
	}
	return false
}
