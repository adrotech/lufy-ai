package openspec

import (
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

func TestRenderWorkflowIncludesOpenSpecBaseAssets(t *testing.T) {
	assets, err := New().RenderWorkflow(ports.WorkflowModel{
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
}
