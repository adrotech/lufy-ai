package opencode

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
	return domain.ToolInitialDefault
}

func (Adapter) Capabilities() domain.ToolCapabilities {
	return domain.ToolCapabilities{
		Subagents:     true,
		SlashCommands: true,
		Skills:        true,
		Hooks:         true,
		MCP:           true,
		TUI:           true,
		GlobalConfig:  true,
		ProjectConfig: true,
		SystemPrompt:  true,
	}
}

func (Adapter) Detect(context.Context, ports.Env) ports.DetectionResult {
	return ports.DetectionResult{Detected: true, Reason: "default adapter"}
}

func (Adapter) RenderSurface(ports.HarnessModel) ([]ports.AssetSpec, error) {
	return []ports.AssetSpec{
		{ID: "opencode.agents", TargetRel: ".opencode/agents", Policy: "managed", Scope: "project"},
		{ID: "opencode.commands", TargetRel: ".opencode/commands", Policy: "managed", Scope: "project"},
		{ID: "opencode.hooks", TargetRel: ".opencode/hooks", Policy: "managed", Scope: "project"},
		{ID: "opencode.skills", TargetRel: ".opencode/skills", Policy: "managed", Scope: "project"},
		{ID: "opencode.templates", TargetRel: ".opencode/templates", Policy: "managed", Scope: "project"},
		{ID: "opencode.policies", TargetRel: ".opencode/policies", Policy: "managed", Scope: "project"},
		{ID: "opencode.plugins", TargetRel: ".opencode/plugins", Policy: "managed", Scope: "project"},
		{ID: "opencode.agent-observatory", TargetRel: ".opencode/agent-observatory", Policy: "managed", Scope: "project"},
		{ID: "opencode.readme", TargetRel: ".opencode/README.md", Policy: "managed", Scope: "project"},
		{ID: "opencode.package", TargetRel: ".opencode/package.json", Policy: "managed", Scope: "project"},
		{ID: "opencode.package-lock", TargetRel: ".opencode/package-lock.json", Policy: "managed", Scope: "project"},
		{ID: "opencode.gitignore", TargetRel: ".opencode/.gitignore", Policy: "managed", Scope: "project"},
		{ID: "opencode.config", TargetRel: "opencode.json", Policy: "merge-json", Scope: "project"},
		{ID: "opencode.tui", TargetRel: "tui.json", Policy: "no-replace", Scope: "project"},
	}, nil
}

func (Adapter) Verify(ports.Target) ([]ports.Check, error) {
	return []ports.Check{
		{Level: "info", Message: "opencode adapter verification is structural in the installer layer"},
	}, nil
}
