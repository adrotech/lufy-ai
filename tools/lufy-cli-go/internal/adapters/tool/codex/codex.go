package codex

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
	return domain.ToolCodex
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
	return ports.DetectionResult{Detected: false, Reason: "codex adapter is dry-run only"}
}

func (Adapter) RenderSurface(ports.HarnessModel) ([]ports.AssetSpec, error) {
	return []ports.AssetSpec{
		{ID: "codex.agents-preview", TargetRel: "AGENTS.md", Policy: "dry-run", Scope: "preview"},
		{ID: "codex.inline-roles-preview", TargetRel: "AGENTS.md#lufy-inline-roles", Policy: "dry-run", Scope: "preview"},
		{ID: "codex.gaps-preview", TargetRel: "AGENTS.md#lufy-capability-gaps", Policy: "dry-run", Scope: "preview"},
	}, nil
}

func (Adapter) Verify(ports.Target) ([]ports.Check, error) {
	return []ports.Check{
		{Level: "info", Message: "codex adapter is dry-run only; no repository assets are installed"},
		{Level: "info", Message: "roles must run as inline phases because native subagent isolation is not declared"},
	}, nil
}
