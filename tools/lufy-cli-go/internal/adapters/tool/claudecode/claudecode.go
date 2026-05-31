package claudecode

import (
	"context"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

type Adapter struct{}

func New() Adapter {
	return Adapter{}
}

func (Adapter) ID() domain.ToolID {
	return domain.ToolClaudeCode
}

func (Adapter) Capabilities() domain.ToolCapabilities {
	return domain.ToolCapabilities{
		Subagents:     false,
		SlashCommands: false,
		Skills:        false,
		Hooks:         false,
		MCP:           false,
		TUI:           false,
		GlobalConfig:  false,
		ProjectConfig: true,
		SystemPrompt:  true,
		DryRunOnly:    true,
	}
}

func (Adapter) Detect(context.Context, ports.Env) ports.DetectionResult {
	return ports.DetectionResult{Detected: false, Reason: "claude-code adapter is dry-run only"}
}

func (Adapter) RenderSurface(ports.HarnessModel) ([]ports.AssetSpec, error) {
	return []ports.AssetSpec{
		{ID: "claude-code.instructions-preview", TargetRel: "CLAUDE.md", Policy: "dry-run", Scope: "preview"},
		{ID: "claude-code.inline-roles-preview", TargetRel: "CLAUDE.md#lufy-inline-roles", Policy: "dry-run", Scope: "preview"},
		{ID: "claude-code.gaps-preview", TargetRel: "CLAUDE.md#lufy-capability-gaps", Policy: "dry-run", Scope: "preview"},
	}, nil
}

func (Adapter) Verify(ports.Target) ([]ports.Check, error) {
	return []ports.Check{
		{Level: "info", Message: "claude-code adapter is dry-run only; no repository assets are installed"},
		{Level: "info", Message: "roles must run as inline phases because native subagent isolation is not declared"},
	}, nil
}
