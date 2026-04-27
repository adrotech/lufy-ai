# OpenCode Local Tooling

This directory contains repository-local lufy-ai configuration.

## Agents

- `agents/orchestrator.md`: default primary coordinator
- `agents/explorer.md`: read-only exploration subagent
- `agents/implementer.md`: implementation subagent
- `agents/validator.md`: read-only validation subagent
- `agents/reviewer.md`: read-only review subagent
- `agents/delivery.md`: Git/GH and PR delivery subagent

Shared project rules live in `../AGENTS.md`. Shared delivery rules live in `policies/delivery.md`.

## Commands

Slash commands live in `commands/`.

- `opsx-explore`: explore codebase without implementation
- `opsx-propose`: create OpenSpec proposal artifacts
- `opsx-apply`: implement OpenSpec tasks
- `opsx-verify`: verify implementation against spec
- `opsx-archive`: archive completed change
- `opsx-delivery`: create PR through delivery

## Skills

- `skills/sdd-workflow`: OpenSpec/SDD lifecycle
- `skills/git-delivery`: delivery templates and traceability helpers
- `skills/project-sync`: GitHub Project sync helpers
- `skills/memory`: Engram memory integration
- `skills/release`: Release workflow

## Agent Observatory TUI Plugin

The TUI sidebar plugin is loaded by root `tui.json`:

```json
{
  "$schema": "https://opencode.ai/tui.json",
  "plugin": ["./.opencode/plugins/agent-observatory.tsx"],
  "plugin_enabled": {
    "lufy-ai.observatory": true
  }
}
```

Runtime toggles exposed as slash commands:

- `/observatory`: show/hide the panel
- `/observatory-agents`, `/obs-agent-list`: collapse/expand agent list
- `/observatory-subagents`, `/obs-agents`: collapse/expand subagent section
- `/observatory-tools`, `/obs-tools`: collapse/expand tool summaries
- `/observatory-cost`: show/hide cost
- `/observatory-emoji`, `/obs-emoji`: show/hide emojis

V1 is local/TUI-only. Do not add external telemetry without a separate proposal.