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
		Subagents:     true,
		SlashCommands: false,
		Skills:        true,
		Hooks:         true,
		MCP:           true,
		TUI:           false,
		GlobalConfig:  false,
		ProjectConfig: true,
		SystemPrompt:  true,
		DryRunOnly:    false,
	}
}

func (Adapter) Detect(context.Context, ports.Env) ports.DetectionResult {
	return ports.DetectionResult{Detected: true, Reason: "codex project adapter"}
}

func (Adapter) RenderSurface(ports.HarnessModel) ([]ports.AssetSpec, error) {
	return []ports.AssetSpec{
		{ID: "codex.skills", TargetRel: ".agents/skills", Policy: "managed", Scope: "project"},
		{ID: "codex.project-config", TargetRel: ".codex", Policy: "managed", Scope: "project"},
	}, nil
}

func (Adapter) Verify(ports.Target) ([]ports.Check, error) {
	return []ports.Check{
		{Level: "info", Message: "codex adapter verification is structural in the installer layer"},
		{Level: "info", Message: "codex Lufy roles should use native agents when discovered, with emulated/inline fallback documented in .codex/lufy-agent-mapping.md"},
	}, nil
}
