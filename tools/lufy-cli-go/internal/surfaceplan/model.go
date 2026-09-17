package surfaceplan

import "fmt"

const SchemaVersion = "surface-execution-plan/v1"

type ExecutionMode string

const (
	ModeSingle   ExecutionMode = "single"
	ModeComposed ExecutionMode = "composed"
)

type Options struct {
	Target           string
	RequestedSurface string
	Base             string
	Files            []string
	Capabilities     []string
}

type ProjectDefinition struct {
	Surfaces []SurfaceDefinition
	Stacks   map[string]StackDefinition
}

type SurfaceDefinition struct {
	ID                     string   `json:"id"`
	Type                   string   `json:"type"`
	Roots                  []string `json:"roots"`
	Stacks                 []string `json:"stacks,omitempty"`
	Frameworks             []string `json:"frameworks,omitempty"`
	Connects               []string `json:"connects,omitempty"`
	Capabilities           []string `json:"capabilities,omitempty"`
	Architecture           string   `json:"architecture,omitempty"`
	PrimaryConcerns        []string `json:"primary_concerns,omitempty"`
	StructuralExpectations []string `json:"structural_expectations,omitempty"`
	ValidationExpectations []string `json:"validation_expectations,omitempty"`
}

type StackDefinition struct {
	ID              string
	TestCommand     string
	CoverageCommand string
	LintCommand     string
	StaticCommand   string
}

type Decision struct {
	Code     string   `json:"code"`
	Reason   string   `json:"reason"`
	Evidence []string `json:"evidence,omitempty"`
}

type ContractEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
}

type ValidationRule struct {
	ID        string   `json:"id"`
	Category  string   `json:"category"`
	Trigger   string   `json:"trigger"`
	Required  bool     `json:"required"`
	Surfaces  []string `json:"surfaces,omitempty"`
	Commands  []string `json:"commands,omitempty"`
	Evidence  []string `json:"evidence,omitempty"`
	Rationale string   `json:"rationale"`
}

type ExecutionPlan struct {
	SchemaVersion    string              `json:"schema_version"`
	Target           string              `json:"target"`
	Base             string              `json:"base"`
	Source           string              `json:"source"`
	RequestedSurface string              `json:"requested_surface"`
	Mode             ExecutionMode       `json:"mode"`
	PrimarySurface   string              `json:"primary_surface"`
	ActiveSurfaces   []SurfaceDefinition `json:"active_surfaces"`
	ChangedFiles     []string            `json:"changed_files"`
	Capabilities     []string            `json:"capabilities,omitempty"`
	Signals          []string            `json:"signals,omitempty"`
	Contracts        []ContractEdge      `json:"contracts,omitempty"`
	Decisions        []Decision          `json:"decisions"`
	ValidationRules  []ValidationRule    `json:"validation_rules"`
}

type AmbiguousSurfaceError struct {
	Choices []string
}

func (e AmbiguousSurfaceError) Error() string {
	return fmt.Sprintf("superficie ambigua; usa --surface con una de estas opciones: %v", e.Choices)
}
