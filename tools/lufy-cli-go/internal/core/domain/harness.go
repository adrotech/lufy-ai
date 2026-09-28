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
	ToolCodex          ToolID = "codex"
	ToolClaudeCode     ToolID = "claude-code"
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
	DryRunOnly    bool
}

type MethodologySelection struct {
	ID       MethodologyID   `yaml:"id" json:"id"`
	Mode     MethodologyMode `yaml:"mode" json:"mode"`
	Required bool            `yaml:"required" json:"required"`
}

type MethodologyByTier map[Tier]MethodologySelection

type HarnessSource string

const (
	HarnessSourceExplicit      HarnessSource = "explicit"
	HarnessSourceProjectConfig HarnessSource = "project-config"
	HarnessSourceInstallState  HarnessSource = "install-state"
	HarnessSourceDefault       HarnessSource = "default"
)

type HarnessProvenance struct {
	Tool              HarnessSource          `json:"tool" yaml:"tool"`
	MethodologyByTier map[Tier]HarnessSource `json:"methodologyByTier" yaml:"methodology_by_tier"`
}

type HarnessConfig struct {
	Tool              ToolID
	MethodologyByTier MethodologyByTier
	Provenance        HarnessProvenance
}

func DefaultHarnessConfig() HarnessConfig {
	return HarnessConfig{
		Tool:              ToolInitialDefault,
		MethodologyByTier: DefaultMethodologyByTier(),
		Provenance: HarnessProvenance{
			Tool: HarnessSourceDefault,
			MethodologyByTier: map[Tier]HarnessSource{
				TierT1: HarnessSourceDefault,
				TierT2: HarnessSourceDefault,
				TierT3: HarnessSourceDefault,
			},
		},
	}
}

func DefaultMethodologyByTier() MethodologyByTier {
	return MethodologyByTier{
		TierT1: {ID: MethodologySpecWorkflow, Mode: MethodologyModeFull, Required: true},
		TierT2: {ID: MethodologySpecWorkflow, Mode: MethodologyModeLite, Required: true},
		TierT3: {ID: MethodologyNone, Mode: MethodologyModeNone, Required: false},
	}
}

func (m MethodologyByTier) WithDefaults() MethodologyByTier {
	defaults := DefaultMethodologyByTier()
	for tier, selection := range m {
		defaults[tier] = selection
	}
	return defaults
}

func (m MethodologyByTier) SelectionFor(tier Tier) (MethodologySelection, error) {
	if !tier.Valid() {
		return MethodologySelection{}, fmt.Errorf("tier no soportado: %s", tier)
	}
	selection, ok := m.WithDefaults()[tier]
	if !ok {
		return MethodologySelection{}, fmt.Errorf("tier sin metodologia configurada: %s", tier)
	}
	if err := (MethodologyByTier{tier: selection}).ValidateSupported(); err != nil {
		return MethodologySelection{}, err
	}
	return selection, nil
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

func (c HarnessConfig) WithDefaults() HarnessConfig {
	defaults := DefaultHarnessConfig()
	if c.Tool != "" {
		defaults.Tool = c.Tool
	}
	if len(c.MethodologyByTier) > 0 {
		defaults.MethodologyByTier = c.MethodologyByTier.WithDefaults()
	}
	if c.Provenance.Tool != "" {
		defaults.Provenance.Tool = c.Provenance.Tool
	}
	for tier, source := range c.Provenance.MethodologyByTier {
		if source != "" {
			defaults.Provenance.MethodologyByTier[tier] = source
		}
	}
	return defaults
}

func (c HarnessConfig) ToolIsExplicit() bool {
	if c.Provenance.Tool == HarnessSourceExplicit {
		return true
	}
	if c.Provenance.Tool == "" {
		return c.Tool != ""
	}
	return c.Provenance.Tool == HarnessSourceDefault && c.Tool != "" && c.Tool != DefaultHarnessConfig().Tool
}

func (c HarnessConfig) TierIsExplicit(tier Tier) bool {
	if c.Provenance.MethodologyByTier[tier] == HarnessSourceExplicit {
		return true
	}
	selection, present := c.MethodologyByTier[tier]
	if !present {
		return false
	}
	source := c.Provenance.MethodologyByTier[tier]
	if source == "" {
		return true
	}
	return source == HarnessSourceDefault && selection != DefaultMethodologyByTier()[tier]
}

func (c HarnessConfig) ValidateSupported() error {
	normalized := c.WithDefaults()
	if !normalized.Tool.Valid() {
		return fmt.Errorf("tool no soportada: %s", normalized.Tool)
	}
	return normalized.MethodologyByTier.ValidateSupported()
}

func (t Tier) Valid() bool {
	switch t {
	case TierT1, TierT2, TierT3:
		return true
	default:
		return false
	}
}

func (t ToolID) Valid() bool {
	switch t {
	case ToolInitialDefault, ToolCodex, ToolClaudeCode:
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
