package render

import (
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/instructions/registry"
)

type RoleSurface struct {
	RoleID                domain.RoleID
	Kind                  string
	Purpose               string
	DirectSkills          []SkillRef
	ResultContractContext AdapterContext
	OutputSchema          string
	AllowedStatus         []string
	CompactPayload        []string
	MaxHandoffFocus       []string
}

type SkillRef struct {
	Slot            domain.SkillSlot
	Name            string
	Path            string
	Category        string
	MissingBehavior string
}

type AdapterContext struct {
	Tool                domain.ToolID
	Tier                domain.Tier
	Methodology         domain.MethodologyID
	MethodologyMode     domain.MethodologyMode
	MethodologyRequired bool
}

func BuildRoleSurface(role registry.RoleDefinition, binding registry.SkillBinding) (RoleSurface, error) {
	return BuildRoleSurfaceForTier(role, binding, domain.HarnessConfig{}, domain.TierT1)
}

func BuildRoleSurfaceForTier(role registry.RoleDefinition, binding registry.SkillBinding, cfg domain.HarnessConfig, tier domain.Tier) (RoleSurface, error) {
	resolved, err := registry.ResolveDirectSkills(role, binding)
	if err != nil {
		return RoleSurface{}, err
	}
	adapterContext, err := BuildAdapterContext(binding, cfg, tier)
	if err != nil {
		return RoleSurface{}, err
	}

	return RoleSurface{
		RoleID:                role.ID,
		Kind:                  role.Kind,
		Purpose:               role.Purpose,
		DirectSkills:          toSkillRefs(resolved),
		ResultContractContext: adapterContext,
		OutputSchema:          role.Output.Schema,
		AllowedStatus:         append([]string(nil), role.Output.AllowedStatus...),
		CompactPayload:        append([]string(nil), role.Output.CompactPayload...),
		MaxHandoffFocus:       append([]string(nil), role.Output.MaxHandoffFocus...),
	}, nil
}

func BuildAdapterContext(binding registry.SkillBinding, cfg domain.HarnessConfig, tier domain.Tier) (AdapterContext, error) {
	normalized := cfg.WithDefaults()
	if cfg.Tool != "" && binding.Tool != "" && cfg.Tool != binding.Tool {
		return AdapterContext{}, domain.HarnessConfig{Tool: cfg.Tool, MethodologyByTier: normalized.MethodologyByTier}.ValidateSupported()
	}
	if binding.Tool != "" {
		normalized.Tool = binding.Tool
	}
	if err := normalized.ValidateSupported(); err != nil {
		return AdapterContext{}, err
	}
	selection, err := normalized.MethodologyByTier.SelectionFor(tier)
	if err != nil {
		return AdapterContext{}, err
	}
	return AdapterContext{
		Tool:                normalized.Tool,
		Tier:                tier,
		Methodology:         selection.ID,
		MethodologyMode:     selection.Mode,
		MethodologyRequired: selection.Required,
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
