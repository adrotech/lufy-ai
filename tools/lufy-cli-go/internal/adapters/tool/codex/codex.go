package codex

import (
	"context"
	"path/filepath"
	"runtime"

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

func (Adapter) SkillRoots(target ports.Target, env ports.Env) []ports.SkillRoot {
	roots := []ports.SkillRoot{{Path: filepath.Join(target.Root, ".agents", "skills"), Scope: "project", Priority: 0}}
	if env["HOME"] != "" {
		roots = append(roots, ports.SkillRoot{Path: filepath.Join(env["HOME"], ".agents", "skills"), Scope: "global", Priority: 10})
	}
	if runtime.GOOS != "windows" {
		roots = append(roots, ports.SkillRoot{Path: filepath.Join(string(filepath.Separator), "etc", "codex", "skills"), Scope: "global", Priority: 20})
	}
	return roots
}

func (Adapter) RenderSurface(ports.HarnessModel) ([]ports.AssetSpec, error) {
	paths := []string{".codex/README.md", ".codex/config.toml", ".codex/hooks.json", ".codex/rules/lufy.rules", ".codex/lufy-agent-mapping.md"}
	for _, skill := range []string{"lufy-check", "lufy-mem-search", "lufy-mem-capture", "lufy-mem-connect", "lufy-mem-document"} {
		paths = append(paths, ".agents/skills/"+skill)
	}
	out := make([]ports.AssetSpec, 0, len(paths))
	for _, path := range paths {
		out = append(out, ports.AssetSpec{ID: path, TargetRel: path, Policy: "managed", Scope: "project"})
	}
	return out, nil
}

func (Adapter) Verify(ports.Target) ([]ports.Check, error) {
	return []ports.Check{
		{Level: "info", Message: "codex adapter verification is structural in the installer layer"},
		{Level: "info", Message: "Principal inline; delegación según execution.delegation y capacidades efectivas del host"},
	}, nil
}
