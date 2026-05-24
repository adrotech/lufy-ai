## 1. OpenSpec artifacts

- [x] 1.1 Create proposal artifact for `add-numeric-stop-rules-workload-guard`.
- [x] 1.2 Create design artifact covering canonical workflow limits, numeric workload guard, stop rules and optional `chain_strategy` handling.
- [x] 1.3 Create delta spec for `sdd-harness-routing` with valid ADDED requirements and WHEN/THEN scenarios.
- [x] 1.4 Create delta spec for `systemic-workflow` with valid ADDED requirements and WHEN/THEN scenarios.

## 2. Agent implementation (future)

- [x] 2.1 Update root `.opencode/agents/orchestrator.md` with 4-file rule, 20-tool-calls rule, multi-file write rule, long-session rule, `chain_strategy` propagation and Result Contract fields.
- [x] 2.2 Update embedded/catalog copy of `orchestrator.md` so installed assets match root behavior.
- [x] 2.3 Update root `.opencode/agents/sdd-router.md` to read/propagate `workflow_limits.sizing`, `workflow_limits.routing`, `workflow_limits.proposal_slicing_strategy`, `workflow_limits.delivery_batch_strategy`, `workflow_limits.preflight`, `workflow_limits.stop_rules` and optional `chain_strategy`.
- [x] 2.4 Update embedded/catalog copy of `sdd-router.md` so installed assets match root behavior.
- [x] 2.5 Ensure router emits `workload_decision_needed: true` when `estimated_loc > workflow_limits.sizing.loc_budget` and escalates/proposes slices when `estimated_files >= 5`.
- [x] 2.6 Ensure `review_slices` derive from `workflow_limits.proposal_slicing_strategy`, not `workflow_limits.delivery_batch_strategy`.
- [x] 2.7 Ensure missing `.opencode/project.yaml` or missing paths are reported as `not_available` without reading legacy top-level `loc_budget` / `delivery_strategy`.

## 3. Delivery/config compatibility (future, minimal if needed)

- [x] 3.1 Decide during implementation whether any minimal delivery policy wording update is required; keep delivery authorization separate from batching guidance.
- [x] 3.2 If assets/catalog are touched, update installer managed asset catalog/manifest consistently.
- [x] 3.3 Do not change CLI Go `WorkflowLimits` structs for `chain_strategy` in this slice unless implementation discovers a required structural validation path and records the decision.

## 4. Validation

- [x] 4.1 Run `openspec validate add-numeric-stop-rules-workload-guard --strict`.
- [x] 4.2 Run `openspec status --change "add-numeric-stop-rules-workload-guard"`.
- [x] 4.3 Run `scripts/validate.sh` after agent/assets/catalog implementation if the scope touches embedded assets or CLI Go validation paths.
- [x] 4.4 Verify parity between root agent files and embedded/catalog assets before reporting implementation ready.
