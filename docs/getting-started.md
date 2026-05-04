# Getting Started with lufy-ai

## What is lufy-ai?

lufy-ai is a template that brings intelligent software development workflows to any project. It provides:

- **Orchestrated Agents**: Specialized AI agents with clear responsibilities
- **SDD/OpenSpec Workflow**: Spec-Driven Development lifecycle
- **Git Delivery Policy**: Safe delivery rules and a delivery subagent
- **Memory-ready config**: Optional Engram MCP configuration in `opencode.json`
- **TUI Observatory**: Real-time visibility into agent activity

## Installation

### Quick Install

```bash
# Run installer directly
curl -fsSL https://raw.githubusercontent.com/adrotech/lufy-ai/main/scripts/install.sh | bash
```

### Manual Install

```bash
# Clone repository
git clone https://github.com/adrotech/lufy-ai.git /tmp/lufy-ai

# Go to your project and run installer
cd my-project
/tmp/lufy-ai/scripts/install.sh
```

### What the installer does

1. Detects if `.opencode/` already exists
2. Detects project stack (React, Next.js, Astro, Expo, Spring Boot, Go, Python)
3. Copies configuration files to `.opencode/`
4. Creates `AGENTS.md` from template
5. Sets up `tui.json`, `opencode.json`, and `openspec/`

## Available Commands

After installation, use these slash commands in OpenCode:

| Command | Description |
|---------|-------------|
| `/opsx-explore` | Explore codebase without changes |
| `/opsx-propose` | Create new feature proposal |
| `/opsx-apply` | Implement approved tasks |
| `/opsx-verify` | Verify implementation |
| `/opsx-archive` | Archive completed change |

## Agent Roles

| Agent | Role |
|-------|------|
| `orchestrator` | Routes work to subagents |
| `explorer` | Read-only file discovery |
| `implementer` | Bounded code changes |
| `validator` | Validation evidence |
| `reviewer` | Quality review |
| `delivery` | Git/GH delivery |

## Project Setup

After installation:

1. **Review AGENTS.md**: Update with your project conventions
2. **Restart OpenCode**: Load new agents
3. **Start exploring**: Use `/opsx-explore` to understand your codebase

## Stack Detection

The installer detects common project markers and records the detected stack in its output. Stack-specific templates are a planned evolution; the current installer copies the same base agent, command, policy, OpenSpec, and observability assets for every stack.

## Documentation

- [OpenSpec Overview](../openspec/README.md) - OpenSpec structure and lifecycle
- [Local OpenCode Tooling](../.opencode/README.md) - Installed agents, commands, skills, and observability notes

## Troubleshooting

### Agents not loading

1. Restart OpenCode
2. Check `.opencode/agents/` exists
3. Verify `AGENTS.md` has project conventions

### TUI plugin not showing

1. Check `tui.json` exists in project root
2. Verify plugin path in `tui.json`
3. Use `/observatory` command to toggle

## Further Support

- GitHub: https://github.com/adrotech/lufy-ai
- Issues: https://github.com/adrotech/lufy-ai/issues
