package openspec

import (
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

func TestRenderWorkflowIncludesOpenSpecBaseAssets(t *testing.T) {
	adapter := New()
	if adapter.ID() != domain.MethodologySpecWorkflow {
		t.Fatalf("ID = %s", adapter.ID())
	}
	if modes := adapter.SupportedModes(); len(modes) != 2 || modes[0] != domain.MethodologyModeFull || modes[1] != domain.MethodologyModeLite {
		t.Fatalf("SupportedModes = %#v", modes)
	}
	assets, err := adapter.RenderWorkflow(ports.WorkflowModel{
		Tier: domain.TierT1,
		Selection: domain.MethodologySelection{
			ID:       domain.MethodologySpecWorkflow,
			Mode:     domain.MethodologyModeFull,
			Required: true,
		},
	})
	if err != nil {
		t.Fatalf("render workflow: %v", err)
	}
	byTarget := map[string]bool{}
	for _, asset := range assets {
		byTarget[asset.TargetRel] = true
	}
	for _, target := range []string{"openspec/config.yaml", "openspec/README.md", "openspec/UPSTREAM.json", "openspec/specs"} {
		if !byTarget[target] {
			t.Fatalf("missing target %s", target)
		}
	}
	checks, err := adapter.VerifyWorkflow(ports.Target{}, domain.TierT1)
	if err != nil {
		t.Fatalf("verify workflow: %v", err)
	}
	if len(checks) != 1 || checks[0].Level != "info" {
		t.Fatalf("checks = %#v", checks)
	}
}

func TestRenderWorkflowIgnoresOtherMethodology(t *testing.T) {
	assets, err := New().RenderWorkflow(ports.WorkflowModel{
		Tier: domain.TierT3,
		Selection: domain.MethodologySelection{
			ID:   domain.MethodologyNone,
			Mode: domain.MethodologyModeNone,
		},
	})
	if err != nil {
		t.Fatalf("render workflow: %v", err)
	}
	if len(assets) != 0 {
		t.Fatalf("assets = %#v, want none", assets)
	}
}
