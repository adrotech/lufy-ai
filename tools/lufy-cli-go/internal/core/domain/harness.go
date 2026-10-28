package domain

import "fmt"

type ToolID string

const ToolCodex ToolID = "codex"

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

type HarnessSource string

const (
	HarnessSourceExplicit      HarnessSource = "explicit"
	HarnessSourceProjectConfig HarnessSource = "project-config"
	HarnessSourceInstallState  HarnessSource = "install-state"
	HarnessSourceDefault       HarnessSource = "default"
)

type HarnessProvenance struct {
	Tool      HarnessSource `json:"tool" yaml:"tool"`
	Execution HarnessSource `json:"execution" yaml:"execution"`
}

type HarnessConfig struct {
	Tool       ToolID            `json:"tool"`
	Execution  ExecutionConfig   `json:"execution"`
	Provenance HarnessProvenance `json:"provenance"`
}

func DefaultHarnessConfig() HarnessConfig {
	return HarnessConfig{Tool: ToolCodex, Execution: DefaultExecutionConfig(), Provenance: HarnessProvenance{Tool: HarnessSourceDefault, Execution: HarnessSourceDefault}}
}

func (c HarnessConfig) WithDefaults() HarnessConfig {
	if c.Tool == "" {
		c.Tool = ToolCodex
	}
	c.Execution = c.Execution.WithDefaults()
	if c.Provenance.Tool == "" {
		c.Provenance.Tool = HarnessSourceDefault
	}
	if c.Provenance.Execution == "" {
		c.Provenance.Execution = HarnessSourceDefault
	}
	return c
}

func (c HarnessConfig) ToolIsExplicit() bool {
	return c.Provenance.Tool == HarnessSourceExplicit || (c.Provenance.Tool == "" && c.Tool != "")
}

func (c HarnessConfig) ExecutionIsExplicit() bool {
	return c.Provenance.Execution == HarnessSourceExplicit || (c.Provenance.Execution == "" && c.Execution != (ExecutionConfig{}))
}

func (c HarnessConfig) ValidateSupported() error {
	c = c.WithDefaults()
	if c.Tool != ToolCodex {
		return fmt.Errorf("tool no soportada por el flujo inline: %s; usa codex", c.Tool)
	}
	return c.Execution.Validate()
}

func (t ToolID) Valid() bool { return t == ToolCodex }
