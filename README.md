# lufy-ai

AI-first Software Development Flow Manager

## What is lufy-ai?

lufy-ai is a configuration template that brings intelligent software development workflows to any project. It provides:

- **Orchestrated Agents**: A system of specialized AI agents that collaborate through clear responsibilities
- **SDD/OpenSpec Workflow**: Spec-Driven Development with proposal, apply, verify, and archive phases
- **Git Delivery**: Safe delivery with PR templates, commit conventions, and traceability
- **GitHub Project Sync**: Automatic project board synchronization
- **Memory**: Engram integration for persistent memory across sessions
- **TUI Observatory**: Real-time visibility into agent activity

## Quick Start

### Installation

```bash
# Clone and run installer
git clone https://github.com/adrianrojas/lufy-ai.git /tmp/lufy-ai
cd /tmp/lufy-ai
./scripts/install.sh
```

Or directly:

```bash
curl -fsSL https://raw.githubusercontent.com/adrianrojas/lufy-ai/main/scripts/install.sh | bash
```

### Requirements

- Git
- A project with:
  - `AGENTS.md` (recommended)
  - Or a recognized stack (`package.json`, `pom.xml`, `go.mod`, etc.)

## Features

### Agents

| Agent | Role |
|-------|------|
| `orchestrator` | Primary coordinator that routes work to subagents |
| `explorer` | Read-only impact analysis and file discovery |
| `implementer` | Bounded implementation of code, tests, docs |
| `validator` | Read-only validation evidence and diagnosis |
| `reviewer` | Quality review and merge risk assessment |
| `delivery` | Git/GH operations and traceability |

### Skills

- **sdd-workflow**: Spec-Driven Development lifecycle
- **git-delivery**: Safe branch, commit, PR, and delivery
- **project-sync**: GitHub Project board synchronization
- **memory**: Engram persistent memory integration
- **release**: Release workflow management

### Stack Templates

Templates for common stacks:

- `frontend-react`: React, Next.js, Vue
- `mobile-expo`: Expo, React Native
- `backend-spring`: Spring Boot, Java
- `backend-node`: Node.js, Express, Nest

## Usage

After installation, OpenCode will automatically use the agents defined in `.opencode/agents/`.

### Common Commands

- `/opsx-explore` - Explore the codebase without changes
- `/opsx-propose` - Create a new feature proposal
- `/opsx-apply` - Implement approved tasks
- `/opsx-verify` - Verify implementation against spec
- `/opsx-archive` - Archive a completed change

## Documentation

- [Getting Started](docs/getting-started.md)
- [Architecture](docs/architecture.md)
- [Agents Reference](docs/agents-reference.md)
- [Skills Reference](docs/skills-reference.md)
- [TUI Reference](docs/tui-reference.md)
- [Memory Reference](docs/memory-reference.md)

## Difference from gentle-ai

| gentle-ai | lufy-ai |
|-----------|---------|
| Configures AI agents (Claude, OpenCode, etc.) | Configures SDD/OpenSpec workflows |
| Installs to ~/.config/ | Installs to project .opencode/ |
| Deep integration with AI providers | Generic, adaptable to any project |
| Engram by default | Engram optional |

## License

MIT