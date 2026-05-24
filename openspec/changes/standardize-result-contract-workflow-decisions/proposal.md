## Why

Result contracts and workflow-limit decisions exist today, but they are partially applied and shaped differently across agents. This makes handoffs harder to resume, weakens enforcement of `workflow_limits`, and leaves recurrent delivery/rework risks to human memory instead of explicit workflow outputs.

## What Changes

- Define a canonical Result Contract envelope v1 for routed agent outputs and final handoffs.
- Require agents to emit the envelope when returning substantive workflow results, while allowing a documented fallback for legacy/third-party outputs.
- Require router/orchestrator workflow decisions to report the exact `workflow_limits` inputs used, derived decisions, stop/preflight status, and review/delivery slicing guidance.
- Connect configured stop rules to explicit pause/escalation behavior before large or risky implementation/delivery steps continue.
- Clarify that delivery batching remains advisory until explicit delivery authorization.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `sdd-harness-routing`: standardize the result contract envelope and require workflow-limit-driven routing outputs.
- `systemic-workflow`: make `workflow_limits` stop/preflight decisions observable gates with structured evidence.

## Impact

- Affects `.opencode/agents/orchestrator.md`, `.opencode/agents/sdd-router.md`, `.opencode/agents/implementer.md`, `.opencode/agents/validator.md`, `.opencode/agents/reviewer.md`, `.opencode/agents/delivery.md`, and `.opencode/policies/delivery.md`.
- Affects human guidance in `AGENTS.md` and possibly `AGENTS.md.template` so future installed harnesses preserve the same envelope and decision model.
- Affects embedded managed assets under `tools/lufy-cli-go/internal/assets/embedded/**` if implementation changes installed OpenCode assets.
- No public CLI contract, database schema, network API, or release process changes are intended.
