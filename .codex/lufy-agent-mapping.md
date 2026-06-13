# Lufy Agent Mapping for Codex

This file is managed by `lufy-ai`.

Codex project-local files under `.codex/agents/*.toml` define custom agents for Lufy role behavior. Use Lufy roles in `native` mode when tool discovery exposes the exact role as invokable. If a specific Codex runtime exposes only generic multi-agent roles such as `default`, `explorer`, and `worker`, degrade explicitly to `emulated` or `inline`.

## Required Runtime Statement

Before delegating Lufy workflow work, the assistant must state one of these modes in the handoff or result:

- `native`: the exact Lufy role is available as an invokable Codex role.
- `emulated`: the exact Lufy role is not available and the assistant is mapping it to a generic Codex role.
- `inline`: no suitable subagent role is available and the assistant is executing the phase inline.

## Default Mapping

First try the exact native Lufy role. Use this mapping only when native Lufy roles are unavailable:

| Lufy role | Codex role | Notes |
| --- | --- | --- |
| `orchestrator` | `default` | Coordinate locally; do not claim isolated orchestration. |
| `sdd-router` | `explorer` | Read-only routing, sizing, risks, and next owner. |
| `explorer` | `explorer` | Read-only discovery. |
| `implementer` | `worker` | Scoped edits only; no delivery. |
| `test-writer` | `worker` | Test-focused edits only; report RED/GREEN evidence. |
| `validator` | `explorer` | Read-only validation and diagnosis. |
| `reviewer` | `explorer` | Read-only review; lead with findings. |
| `delivery` | `worker` | Only after explicit delivery authorization. |

## Guardrails

- Never tell the user that Lufy roles were used natively when they were only emulated.
- Pass the Lufy role instructions and minimum context into the generic Codex role prompt.
- Preserve role permissions: read-only Lufy roles stay read-only even when mapped to a more capable generic role.
- Include the selected `agent_execution_mode` and `role_mapping` in substantive Result Contract evidence.
