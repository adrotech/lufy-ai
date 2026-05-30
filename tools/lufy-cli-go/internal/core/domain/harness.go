package domain

import "fmt"

type Tier string

const (
	TierT1 Tier = "T1"
	TierT2 Tier = "T2"
	TierT3 Tier = "T3"
)

type ToolID string

const (
	ToolInitialDefault ToolID = "opencode"
)

type MethodologyID string

const (
	MethodologySpecWorkflow MethodologyID = "openspec"
	MethodologyLufyWorkflow MethodologyID = "lufy-sdd"
	MethodologyNone         MethodologyID = "none"
)

type MethodologyMode string

const (
	MethodologyModeFull MethodologyMode = "full"
	MethodologyModeLite MethodologyMode = "lite"
	MethodologyModeNone MethodologyMode = "none"
)

type RoleID string

const (
	RoleOrchestrator RoleID = "orchestrator"
	RoleRouter       RoleID = "router"
	RoleExplorer     RoleID = "explorer"
	RoleImplementer  RoleID = "implementer"
	RoleTestWriter   RoleID = "test-writer"
	RoleValidator    RoleID = "validator"
	RoleReviewer     RoleID = "reviewer"
	RoleDelivery     RoleID = "delivery"
)

type ToolCapabilities struct {
	Subagents     bool
	SlashCommands bool
	Skills        bool
	Hooks         bool
	MCP           bool
	TUI           bool
	GlobalConfig  bool
	ProjectConfig bool
	SystemPrompt  bool
}

type MethodologySelection struct {
	ID       MethodologyID
	Mode     MethodologyMode
	Required bool
}

type MethodologyByTier map[Tier]MethodologySelection

func DefaultMethodologyByTier() MethodologyByTier {
	return MethodologyByTier{
		TierT1: {ID: MethodologySpecWorkflow, Mode: MethodologyModeFull, Required: true},
		TierT2: {ID: MethodologySpecWorkflow, Mode: MethodologyModeLite, Required: true},
		TierT3: {ID: MethodologyNone, Mode: MethodologyModeNone, Required: false},
	}
}

func (m MethodologyByTier) ValidateSupported() error {
	for tier, selection := range m {
		if !tier.Valid() {
			return fmt.Errorf("tier no soportado: %s", tier)
		}
		if !selection.ID.Valid() {
			return fmt.Errorf("metodologia no soportada para %s: %s", tier, selection.ID)
		}
		if !selection.Mode.Valid() {
			return fmt.Errorf("modo de metodologia no soportado para %s: %s", tier, selection.Mode)
		}
		if selection.ID == MethodologyNone && selection.Mode != MethodologyModeNone {
			return fmt.Errorf("metodologia none requiere mode none para %s", tier)
		}
	}
	return nil
}

func (t Tier) Valid() bool {
	switch t {
	case TierT1, TierT2, TierT3:
		return true
	default:
		return false
	}
}

func (m MethodologyID) Valid() bool {
	switch m {
	case MethodologySpecWorkflow, MethodologyLufyWorkflow, MethodologyNone:
		return true
	default:
		return false
	}
}

func (m MethodologyMode) Valid() bool {
	switch m {
	case MethodologyModeFull, MethodologyModeLite, MethodologyModeNone:
		return true
	default:
		return false
	}
}
