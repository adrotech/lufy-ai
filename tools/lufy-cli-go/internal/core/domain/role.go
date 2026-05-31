package domain

type SkillSlot string

const (
	SkillSlotMethodologyExplore SkillSlot = "methodology.explore"
	SkillSlotMethodologyPropose SkillSlot = "methodology.propose"
	SkillSlotMethodologyApply   SkillSlot = "methodology.apply"
	SkillSlotMethodologyVerify  SkillSlot = "methodology.verify"
	SkillSlotMethodologySync    SkillSlot = "methodology.sync"
	SkillSlotMethodologyArchive SkillSlot = "methodology.archive"
	SkillSlotDeliveryPRContent  SkillSlot = "delivery.pr_content"
	SkillSlotDeliveryGit        SkillSlot = "delivery.git"
	SkillSlotRoleTestWriter     SkillSlot = "role.test_writer"
	SkillSlotSkillRegistry      SkillSlot = "skill_registry.lookup"
	SkillSlotStackConfig        SkillSlot = "stack_config.lookup"
	SkillSlotValidationGrouped  SkillSlot = "validation.grouped"
)

type RoleContract struct {
	ID              RoleID
	Kind            string
	DirectSlots     []SkillSlot
	DelegatedSlots  []SkillSlot
	ReferencedSlots []SkillSlot
	Output          OutputContract
}

type OutputContract struct {
	Schema          string
	AllowedStatus   []string
	CompactPayload  []string
	MaxHandoffFocus []string
}

func DefaultRoleContracts() []RoleContract {
	return []RoleContract{
		{
			ID:   RoleOrchestrator,
			Kind: "primary",
			DirectSlots: []SkillSlot{
				SkillSlotMethodologyExplore,
				SkillSlotMethodologyPropose,
				SkillSlotMethodologyApply,
				SkillSlotMethodologyVerify,
				SkillSlotMethodologySync,
				SkillSlotMethodologyArchive,
			},
			ReferencedSlots: []SkillSlot{SkillSlotSkillRegistry, SkillSlotDeliveryPRContent},
			Output:          compactOutput([]string{"ready", "blocked", "escalated", "delivery_pending", "sync_pending", "closed"}),
		},
		{
			ID:              RoleRouter,
			Kind:            "primary",
			ReferencedSlots: []SkillSlot{SkillSlotSkillRegistry},
			Output:          compactOutput([]string{"ready", "blocked", "escalated", "delivery_pending"}),
		},
		{
			ID:              RoleExplorer,
			Kind:            "subagent",
			ReferencedSlots: []SkillSlot{SkillSlotMethodologyExplore, SkillSlotSkillRegistry},
			Output:          compactOutput([]string{"ready", "blocked", "escalated"}),
		},
		{
			ID:              RoleImplementer,
			Kind:            "subagent",
			DelegatedSlots:  []SkillSlot{SkillSlotRoleTestWriter},
			ReferencedSlots: []SkillSlot{SkillSlotMethodologyApply, SkillSlotValidationGrouped},
			Output:          compactOutput([]string{"implemented", "validated", "blocked", "escalated", "delivery_pending"}),
		},
		{
			ID:              RoleTestWriter,
			Kind:            "subagent",
			ReferencedSlots: []SkillSlot{SkillSlotStackConfig},
			Output:          compactOutput([]string{"implemented", "validated", "blocked", "escalated"}),
		},
		{
			ID:              RoleValidator,
			Kind:            "subagent",
			ReferencedSlots: []SkillSlot{SkillSlotMethodologyVerify, SkillSlotValidationGrouped},
			Output:          compactOutput([]string{"validated", "blocked", "escalated", "delivery_pending"}),
		},
		{
			ID:              RoleReviewer,
			Kind:            "subagent",
			ReferencedSlots: []SkillSlot{SkillSlotStackConfig},
			Output:          compactOutput([]string{"ready", "blocked", "escalated"}),
		},
		{
			ID:              RoleDelivery,
			Kind:            "primary",
			DirectSlots:     []SkillSlot{SkillSlotDeliveryPRContent, SkillSlotDeliveryGit},
			ReferencedSlots: []SkillSlot{SkillSlotMethodologySync, SkillSlotValidationGrouped},
			Output:          compactOutput([]string{"delivery_pending", "sync_pending", "blocked", "delivered", "closed"}),
		},
	}
}

func compactOutput(status []string) OutputContract {
	return OutputContract{
		Schema:          "result-contract/v1",
		AllowedStatus:   status,
		CompactPayload:  []string{"artifacts", "evidence", "risks", "next_recommended"},
		MaxHandoffFocus: []string{"intent", "decision", "evidence_gap", "next_owner"},
	}
}
