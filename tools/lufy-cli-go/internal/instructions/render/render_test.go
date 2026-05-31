package render

import (
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/instructions/registry"
)

func TestBuildRoleSurfaceIncludesDirectSkillRefsAndOutputContract(t *testing.T) {
	role := registry.RoleDefinition{
		ID:      domain.RoleDelivery,
		Kind:    "primary",
		Purpose: "deliver authorized work",
		SkillSlots: registry.RoleSkillSlots{
			Direct: []domain.SkillSlot{domain.SkillSlotDeliveryPRContent},
		},
		Output: registry.RoleOutput{
			Schema:          "result-contract/v1",
			AllowedStatus:   []string{"blocked", "delivered"},
			CompactPayload:  []string{"evidence", "next_recommended"},
			MaxHandoffFocus: []string{"authorization_state"},
		},
	}
	binding := registry.SkillBinding{
		Skills: map[domain.SkillSlot]registry.SkillSpec{
			domain.SkillSlotDeliveryPRContent: {
				Name:     "pr-content",
				Path:     "skills/pr-content/SKILL.md",
				Category: "core",
			},
		},
	}

	surface, err := BuildRoleSurface(role, binding)
	if err != nil {
		t.Fatalf("build role surface: %v", err)
	}
	if surface.RoleID != domain.RoleDelivery {
		t.Fatalf("role = %s", surface.RoleID)
	}
	if len(surface.DirectSkills) != 1 {
		t.Fatalf("direct skills = %d", len(surface.DirectSkills))
	}
	if surface.DirectSkills[0].Path != "skills/pr-content/SKILL.md" {
		t.Fatalf("skill path = %s", surface.DirectSkills[0].Path)
	}
	if surface.OutputSchema != "result-contract/v1" {
		t.Fatalf("schema = %s", surface.OutputSchema)
	}
	if surface.ResultContractContext.Tool != domain.ToolInitialDefault {
		t.Fatalf("tool context = %s", surface.ResultContractContext.Tool)
	}
	if surface.ResultContractContext.Methodology != domain.MethodologySpecWorkflow || surface.ResultContractContext.MethodologyMode != domain.MethodologyModeFull {
		t.Fatalf("methodology context = %+v", surface.ResultContractContext)
	}
	if surface.MaxHandoffFocus[0] != "authorization_state" {
		t.Fatalf("handoff focus = %v", surface.MaxHandoffFocus)
	}
}

func TestBuildRoleSurfaceForTierUsesConfiguredMethodologySelection(t *testing.T) {
	role := registry.RoleDefinition{
		ID:      domain.RoleRouter,
		Kind:    "primary",
		Purpose: "route work",
		Output: registry.RoleOutput{
			Schema:          "result-contract/v1",
			AllowedStatus:   []string{"ready"},
			CompactPayload:  []string{"workflow_decision"},
			MaxHandoffFocus: []string{"tier"},
		},
	}
	binding := registry.SkillBinding{
		Tool:        domain.ToolInitialDefault,
		Methodology: domain.MethodologySpecWorkflow,
		Skills:      map[domain.SkillSlot]registry.SkillSpec{},
	}

	surface, err := BuildRoleSurfaceForTier(role, binding, domain.HarnessConfig{}, domain.TierT3)
	if err != nil {
		t.Fatalf("build role surface for tier: %v", err)
	}
	if surface.ResultContractContext.Tier != domain.TierT3 {
		t.Fatalf("tier context = %s", surface.ResultContractContext.Tier)
	}
	if surface.ResultContractContext.Methodology != domain.MethodologyNone || surface.ResultContractContext.MethodologyMode != domain.MethodologyModeNone || surface.ResultContractContext.MethodologyRequired {
		t.Fatalf("unexpected T3 adapter context: %+v", surface.ResultContractContext)
	}
}
