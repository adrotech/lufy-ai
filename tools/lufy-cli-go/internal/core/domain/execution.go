package domain

import "fmt"

type ExecutionMode string
type DelegationPolicy string

const (
	ExecutionInline     ExecutionMode    = "inline"
	DelegationAuto      DelegationPolicy = "auto"
	DelegationOnRequest DelegationPolicy = "on_request"
	DelegationOff       DelegationPolicy = "off"
)

type ExecutionConfig struct {
	Mode       ExecutionMode    `yaml:"mode" json:"mode"`
	Delegation DelegationPolicy `yaml:"delegation" json:"delegation"`
}

func DefaultExecutionConfig() ExecutionConfig {
	return ExecutionConfig{Mode: ExecutionInline, Delegation: DelegationAuto}
}

func (c ExecutionConfig) WithDefaults() ExecutionConfig {
	if c.Mode == "" {
		c.Mode = ExecutionInline
	}
	if c.Delegation == "" {
		c.Delegation = DelegationAuto
	}
	return c
}

func (c ExecutionConfig) Validate() error {
	if c.Mode != ExecutionInline {
		return fmt.Errorf("execution.mode no soportado: %q; usa inline", c.Mode)
	}
	switch c.Delegation {
	case DelegationAuto, DelegationOnRequest, DelegationOff:
		return nil
	default:
		return fmt.Errorf("execution.delegation no soportada: %q; usa auto, on_request u off", c.Delegation)
	}
}
