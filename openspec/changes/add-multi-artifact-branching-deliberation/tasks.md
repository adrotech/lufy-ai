## 1. Protocol contract and routing metadata

- [x] 1.1 Define the `artifact_branching` metadata shape for `sdd-router` handoffs, including stage, candidate_count, reason, parallel_allowed, requires_join, candidate_isolation, merge_plan_required and human_escalation_triggers.
- [x] 1.2 Update existing role instructions so `sdd-router` recommends branching only for T1 or multi-risk T2 uncertainty, caps candidates at 2 and reports `not_needed` for dominant low-risk solutions.
- [x] 1.3 Ensure `workflow_limits` and `parallel_execution` are the only config sources for slicing/parallel constraints and that delivery batching remains separate from branching authorization.

## 2. Candidate generation and isolation

- [x] 2.1 Update orchestrator guidance to coordinate `branching_candidate_generation` with isolated candidate artifacts and no new roles.
- [x] 2.2 Define the candidate artifact storage convention or adapter-neutral equivalent, including merge plan contents and non-overwrite guarantees for parallel generation.
- [x] 2.3 Update implementer/solution-writer guidance so candidates include assumptions, trade-offs, reusable decisions, validation considerations and join inputs.

## 3. Join, comparison and escalation

- [x] 3.1 Update orchestrator guidance to require join before design/tasks/implementation when multiple candidates exist.
- [x] 3.2 Update reviewer guidance to compare candidates by quality, risk, completeness, validation clarity and coherence without acting as product/security arbitrator.
- [x] 3.3 Add human escalation behavior for candidate differences in public contract, security, product direction, significant UX or non-objective trade-offs.
- [x] 3.4 Ensure downstream handoffs reference one canonical artifact set after join and treat non-selected candidates as context only.

## 4. Adapter and documentation alignment

- [x] 4.1 Document adapter-neutral branching semantics and adapter-specific rendering expectations for OpenSpec and non-parallel adapters.
- [x] 4.2 Update installed/managed documentation or renderer assets if the implementation changes user-visible harness behavior.
- [x] 4.3 Verify no new agent role is introduced and no delivery/Git/GH path is enabled by branching.

## 5. Validation and verification

- [x] 5.1 Add or update tests/checks if implementation touches executable routing/rendering code; otherwise document why validation is static/OpenSpec-only.
- [x] 5.2 Run grouped validation for the implementation slice, including `openspec validate "add-multi-artifact-branching-deliberation" --strict` and any real toolchain checks affected by changed assets.
- [x] 5.3 Verify all changed artifacts preserve Result Contract envelope semantics, workflow-limit reporting and structural acceptance gates before requesting review or delivery.
