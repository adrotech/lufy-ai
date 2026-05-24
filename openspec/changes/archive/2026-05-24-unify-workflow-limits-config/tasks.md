## 1. Config model and generation

- [x] 1.1 Locate `.opencode/project.yaml` generation and config model code for `lufy-ai init`.
- [x] 1.2 Replace generated top-level `loc_budget` and `delivery_strategy` with top-level `workflow_limits`.
- [x] 1.3 Define generated `workflow_limits` defaults with `sizing`, `routing`, `proposal_slicing_strategy`, `delivery_batch_strategy`, `stop_rules` and `preflight`.
- [x] 1.4 Add or update fixtures/golden files so generated `.opencode/project.yaml` contains `workflow_limits` and no top-level `loc_budget` or `delivery_strategy`.

## 2. Rescan and legacy-field behavior

- [x] 2.1 Update `lufy-ai init --rescan` merge logic to preserve user overrides under `workflow_limits`.
- [x] 2.2 Ensure rescan does not treat top-level `loc_budget` or `delivery_strategy` as canonical workflow-limit overrides.
- [x] 2.3 Add actionable reporting for legacy top-level workflow fields detected during rescan.
- [x] 2.4 Add or update tests covering fresh generation, force generation, rescan override preservation and legacy-field detection.

## 3. Workflow consumers and documentation

- [x] 3.1 Update `sdd-router` guidance/result contracts to read/report sizing, routing and proposal slicing from `workflow_limits` only.
- [x] 3.2 Update `orchestrator` guidance/result contracts to reference `workflow_limits` as the source of truth.
- [x] 3.3 Update delivery policy/role guidance to use `workflow_limits.delivery_batch_strategy`, `workflow_limits.preflight` and `workflow_limits.stop_rules` without treating proposal slicing as delivery batching.
- [x] 3.4 Update human-facing docs/examples that mention top-level `loc_budget` or `delivery_strategy` so they use `workflow_limits` paths.

## 4. Validation

- [x] 4.1 Run `openspec validate "unify-workflow-limits-config" --strict` after implementation changes.
- [x] 4.2 Run the repository's applicable grouped validation command for this scope, or document why no toolchain applies.
- [x] 4.3 Manually review generated examples/docs/contracts to confirm no canonical references to top-level `loc_budget` or `delivery_strategy` remain.
