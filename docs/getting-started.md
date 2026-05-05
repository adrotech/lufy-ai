# Getting Started with lufy-ai

## What is lufy-ai?

lufy-ai is a template that brings intelligent software development workflows to any project. It provides:

- **Orchestrated Agents**: Specialized AI agents with clear responsibilities
- **SDD/OpenSpec Workflow**: Spec-Driven Development lifecycle
- **Git Delivery Policy**: Safe delivery rules and a delivery subagent
- **Memory-ready config**: Optional Engram MCP configuration in `opencode.json`
- **TUI Observatory**: Real-time visibility into agent activity

## Installation

### Quick Install con CLI Go

El instalador vigente es la CLI Go `lufy-ai`. El script `scripts/install.sh` es solo un wrapper de compatibilidad que ejecuta `lufy-ai install`; no contiene fallback legacy de Bash.

Si tienes `lufy-ai` en `PATH`:

```bash
lufy-ai install --target my-project --dry-run --yes --no-engram
```

### Manual Install

```bash
# Clone repository
git clone https://github.com/adrotech/lufy-ai.git /tmp/lufy-ai

# Build local CLI for the wrapper
cd /tmp/lufy-ai/tools/lufy-cli-go
mkdir -p bin
go build -o bin/lufy-ai ./cmd/lufy-ai

# Run wrapper from repository root or any project path
/tmp/lufy-ai/scripts/install.sh --target /ruta/a/tu/proyecto --dry-run --yes --no-engram
```

También se conserva el argumento posicional histórico como target:

```bash
/tmp/lufy-ai/scripts/install.sh /ruta/a/tu/proyecto --dry-run --yes --no-engram
```

### What the installer does

1. Delegates to `lufy-ai install`
2. Resolves `--target` safely
3. Prints an installation plan
4. Honors `--dry-run` without filesystem mutations
5. Handles `--yes`, `--no-engram`, and `--backup`
6. Resolves Engram from `PATH` when it is not disabled

If neither `lufy-ai` in `PATH` nor `tools/lufy-cli-go/bin/lufy-ai` exists, the wrapper fails with this build instruction:

```bash
cd tools/lufy-cli-go && mkdir -p bin && go build -o bin/lufy-ai ./cmd/lufy-ai
```

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

Stack-specific templates are a planned evolution. The current Bash wrapper does not detect stacks; installation behavior belongs to the Go CLI.

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
