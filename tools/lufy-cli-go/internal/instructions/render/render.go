package render

import (
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/instructions/registry"
)

type RoleSurface struct {
	RoleID          domain.RoleID
	Kind            string
	Purpose         string
	DirectSkills    []SkillRef
	OutputSchema    string
	AllowedStatus   []string
	CompactPayload  []string
	MaxHandoffFocus []string
}

type SkillRef struct {
	Slot            domain.SkillSlot
	Name            string
	Path            string
	Category        string
	MissingBehavior string
}

func BuildRoleSurface(role registry.RoleDefinition, binding registry.SkillBinding) (RoleSurface, error) {
	resolved, err := registry.ResolveDirectSkills(role, binding)
	if err != nil {
		return RoleSurface{}, err
	}

	return RoleSurface{
		RoleID:          role.ID,
		Kind:            role.Kind,
		Purpose:         role.Purpose,
		DirectSkills:    toSkillRefs(resolved),
		OutputSchema:    role.Output.Schema,
		AllowedStatus:   append([]string(nil), role.Output.AllowedStatus...),
		CompactPayload:  append([]string(nil), role.Output.CompactPayload...),
		MaxHandoffFocus: append([]string(nil), role.Output.MaxHandoffFocus...),
	}, nil
}

func toSkillRefs(in []registry.ResolvedSkill) []SkillRef {
	out := make([]SkillRef, 0, len(in))
	for _, skill := range in {
		out = append(out, SkillRef{
			Slot:            skill.Slot,
			Name:            skill.Name,
			Path:            skill.Path,
			Category:        skill.Category,
			MissingBehavior: skill.MissingBehavior,
		})
	}
	return out
}
