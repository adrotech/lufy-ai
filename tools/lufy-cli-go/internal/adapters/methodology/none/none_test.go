package none

import (
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

func TestRenderWorkflowReturnsNoArtifacts(t *testing.T) {
	assets, err := New().RenderWorkflow(ports.WorkflowModel{
		Tier:      domain.TierT3,
		Selection: domain.MethodologySelection{ID: domain.MethodologyNone, Mode: domain.MethodologyModeNone},
	})
	if err != nil {
		t.Fatalf("render workflow: %v", err)
	}
	if len(assets) != 0 {
		t.Fatalf("none methodology assets = %d, want 0", len(assets))
	}
}
