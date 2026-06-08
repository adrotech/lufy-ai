package none

import (
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

func TestRenderWorkflowReturnsNoArtifacts(t *testing.T) {
	adapter := New()
	if adapter.ID() != domain.MethodologyNone {
		t.Fatalf("ID = %s", adapter.ID())
	}
	if modes := adapter.SupportedModes(); len(modes) != 1 || modes[0] != domain.MethodologyModeNone {
		t.Fatalf("SupportedModes = %#v", modes)
	}
	assets, err := adapter.RenderWorkflow(ports.WorkflowModel{
		Tier:      domain.TierT3,
		Selection: domain.MethodologySelection{ID: domain.MethodologyNone, Mode: domain.MethodologyModeNone},
	})
	if err != nil {
		t.Fatalf("render workflow: %v", err)
	}
	if len(assets) != 0 {
		t.Fatalf("none methodology assets = %d, want 0", len(assets))
	}
	checks, err := adapter.VerifyWorkflow(ports.Target{}, domain.TierT3)
	if err != nil {
		t.Fatalf("verify workflow: %v", err)
	}
	if len(checks) != 1 || checks[0].Level != "info" {
		t.Fatalf("checks = %#v", checks)
	}
}
