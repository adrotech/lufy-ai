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
	if surface.MaxHandoffFocus[0] != "authorization_state" {
		t.Fatalf("handoff focus = %v", surface.MaxHandoffFocus)
	}
}
