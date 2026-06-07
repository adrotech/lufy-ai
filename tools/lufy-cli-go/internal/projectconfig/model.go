package projectconfig

import (
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
)

const (
	ProjectConfigPath = ".lufy/project.yaml"
	SchemaVersion     = 1
)

type ProjectConfig struct {
	SchemaVersion     int                      `yaml:"schema_version"`
	DetectedAt        time.Time                `yaml:"detected_at"`
	Tool              domain.ToolID            `yaml:"tool"`
	MethodologyByTier domain.MethodologyByTier `yaml:"methodology_by_tier"`
	ProjectProfile    ProjectProfile           `yaml:"project_profile"`
	Stacks            []Stack                  `yaml:"stacks"`
	CI                CIConfig                 `yaml:"ci"`
	TDD               TDDConfig                `yaml:"tdd"`
	Validation        ValidationConfig         `yaml:"validation"`
	WorkflowLimits    WorkflowLimits           `yaml:"workflow_limits"`
	Extra             map[string]any           `yaml:",inline,omitempty"`
}

type ProjectProfile struct {
	Surfaces []ProjectSurface `yaml:"surfaces"`
	Extra    map[string]any   `yaml:",inline,omitempty"`
}

type ProjectSurface struct {
	ID         string         `yaml:"id"`
	Type       string         `yaml:"type"`
	Roots      []string       `yaml:"roots"`
	Stacks     []string       `yaml:"stacks"`
	Frameworks []string       `yaml:"frameworks"`
	Connects   []string       `yaml:"connects,omitempty"`
	AgentLens  AgentLens      `yaml:"agent_lens"`
	Extra      map[string]any `yaml:",inline,omitempty"`
}

type AgentLens struct {
	PrimaryConcerns        []string `yaml:"primary_concerns"`
	ValidationExpectations []string `yaml:"validation_expectations"`
}

type Stack struct {
	ID                string         `yaml:"id"`
	Supported         bool           `yaml:"supported"`
	Deprecated        bool           `yaml:"deprecated,omitempty"`
	Version           string         `yaml:"version,omitempty"`
	PackageManager    string         `yaml:"package_manager,omitempty"`
	Frameworks        []string       `yaml:"frameworks"`
	TestRunner        CommandConfig  `yaml:"test_runner"`
	Linter            CommandConfig  `yaml:"linter"`
	Formatter         Formatter      `yaml:"formatter"`
	StaticAnalysis    CommandConfig  `yaml:"static_analysis"`
	AntiPatterns      []string       `yaml:"anti_patterns"`
	ObservabilityLibs []string       `yaml:"observability_libs"`
	Notes             string         `yaml:"notes,omitempty"`
	Extra             map[string]any `yaml:",inline,omitempty"`
}

type CommandConfig struct {
	Command           string `yaml:"command,omitempty"`
	CoverageCommand   string `yaml:"coverage_command,omitempty"`
	CoverageThreshold int    `yaml:"coverage_threshold,omitempty"`
	AutoFix           string `yaml:"auto_fix,omitempty"`
}

type Formatter struct {
	Command        string   `yaml:"command,omitempty"`
	FileExtensions []string `yaml:"file_extensions"`
}

type CIConfig struct {
	Detected  bool     `yaml:"detected"`
	Provider  string   `yaml:"provider,omitempty"`
	Workflows []string `yaml:"workflows"`
}

type TDDConfig struct {
	Strict              bool     `yaml:"strict"`
	TriangulateRequired bool     `yaml:"triangulate_required"`
	EdgeCaseCategories  []string `yaml:"edge_case_categories"`
}

type ValidationConfig struct {
	AllowedCommands ValidationAllowedCommands `yaml:"allowed_commands"`
}

type ValidationAllowedCommands struct {
	Implementer []string `yaml:"implementer"`
}

type WorkflowLimits struct {
	Sizing                  WorkflowSizing  `yaml:"sizing"`
	Routing                 WorkflowRouting `yaml:"routing"`
	ProposalSlicingStrategy string          `yaml:"proposal_slicing_strategy"`
	DeliveryBatchStrategy   string          `yaml:"delivery_batch_strategy"`
	StopRules               []string        `yaml:"stop_rules"`
	Preflight               []string        `yaml:"preflight"`
	Extra                   map[string]any  `yaml:",inline,omitempty"`
}

type WorkflowSizing struct {
	LOCBudget int            `yaml:"loc_budget"`
	Extra     map[string]any `yaml:",inline,omitempty"`
}

type WorkflowRouting struct {
	Strategy string         `yaml:"strategy"`
	Extra    map[string]any `yaml:",inline,omitempty"`
}
