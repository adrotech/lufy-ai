## Why

`.opencode/project.yaml` currently exposes workflow sizing and delivery controls through legacy top-level `loc_budget` and `delivery_strategy`, creating multiple places for agents and policy to interpret workflow limits. This change makes a clean break by introducing `workflow_limits` as the single canonical configuration source for routing, slicing, delivery batching, stop rules and preflight behavior.

## What Changes

- **BREAKING**: remove legacy top-level `loc_budget` and `delivery_strategy` from generated and accepted `.opencode/project.yaml` as valid workflow-limit sources.
- Add top-level `workflow_limits` to generated project configuration as the only canonical limits block.
- Centralize `sizing`, `routing`, `proposal_slicing_strategy`, `delivery_batch_strategy`, `stop_rules` and `preflight` under `workflow_limits`.
- Preserve user overrides inside `workflow_limits` during `lufy-ai init --rescan` while continuing to refresh generated stack/tooling/CI evidence safely.
- Require agents, workflow documentation, delivery policy and result contracts to consume and report workflow limits only from `workflow_limits`.
- Explicitly distinguish proposal/review slicing decisions from delivery batching decisions.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `project-stack-config`: generated and rescanned `.opencode/project.yaml` changes from legacy top-level `loc_budget`/`delivery_strategy` to canonical top-level `workflow_limits`.
- `systemic-workflow`: agents, policy and result contracts consume/report routing, slicing, delivery batching, stop rules and preflight from `workflow_limits` only.

## Impact

- Generated `.opencode/project.yaml` schema and examples.
- `lufy-ai init` and `lufy-ai init --rescan` config generation/merge behavior.
- Agent-facing workflow contracts for `sdd-router`, `orchestrator`, implementation/validation/reporting roles and delivery policy.
- Documentation and policy references that previously mentioned `loc_budget` or `delivery_strategy` as top-level project config.
