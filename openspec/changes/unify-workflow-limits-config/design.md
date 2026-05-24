## Context

`.opencode/project.yaml` is generated and rescanned by `lufy-ai init` as the project-local source for stack metadata and operational preferences. Current generated config exposes workflow controls through legacy top-level `loc_budget` and `delivery_strategy`, while upcoming agent and policy behavior needs one stable place for limits that affect sizing, SDD routing, proposal slicing, delivery batching, stop conditions and preflight checks.

The user decision is a clean break: old top-level fields are not preserved as valid workflow-limit sources and are replaced by `workflow_limits` as the only canonical block.

## Goals / Non-Goals

**Goals:**
- Generate `.opencode/project.yaml` with top-level `workflow_limits` and without top-level `loc_budget` or `delivery_strategy`.
- Preserve user overrides already inside `workflow_limits` during `--rescan`.
- Define a stable logical shape for `workflow_limits` covering `sizing`, `routing`, `proposal_slicing_strategy`, `delivery_batch_strategy`, `stop_rules` and `preflight`.
- Align agent workflow contracts so `sdd-router`, `orchestrator`, delivery policy and result contracts read/report the same canonical source.
- Make proposal slicing separate from delivery batching.

**Non-Goals:**
- Implement code in this proposal phase.
- Preserve backward compatibility by silently accepting top-level `loc_budget` or `delivery_strategy` as valid canonical inputs.
- Define every final numeric threshold; implementation may use existing defaults as long as they are nested under the required keys and documented consistently.
- Change stack detection, CI detection, auth, ports or delivery authorization rules beyond the config-source unification.

## Decisions

1. **Use `workflow_limits` as a top-level canonical block.**
   - Rationale: workflow limits are project-level operational policy, not per-stack metadata, and must be easy for agents and delivery policy to locate.
   - Alternative considered: keep `loc_budget`/`delivery_strategy` and add aliases. Rejected because the requested break clean requires removing multiple valid sources.

2. **Group workflow concerns by consumer-facing intent.**
   - Required subkeys:
     - `sizing`: size/LOC/complexity budget inputs used for work sizing.
     - `routing`: T1/T2/T3 or equivalent routing thresholds and escalation cues.
     - `proposal_slicing_strategy`: rules for splitting proposals/review slices before implementation or review.
     - `delivery_batch_strategy`: rules for delivery grouping after validation/delivery authorization.
     - `stop_rules`: limits that force pausing, escalation or handoff.
     - `preflight`: checks required before implementation, validation or delivery stages.
   - Alternative considered: one flat limits map. Rejected because it would keep delivery batching and proposal slicing ambiguous.

3. **Treat legacy top-level fields as invalid workflow-limit sources, not migration inputs.**
   - Rationale: this prevents silent config drift and ensures docs/agents/policy converge on one source.
   - Implementation may emit actionable guidance when legacy fields are detected, but MUST NOT merge them into `workflow_limits` as if they were trusted overrides.

4. **Preserve only overrides inside `workflow_limits` on rescan.**
   - Rationale: users need stable editable boundaries, but the boundary is the new canonical block.
   - Unknown non-workflow fields remain governed by existing rescan preservation requirements; this change only removes canonical status from legacy workflow fields.

## Risks / Trade-offs

- **Breaking existing local configs** → Mitigation: fail/report actionably when legacy top-level fields are encountered and document the required manual move to `workflow_limits`.
- **Ambiguous key semantics** → Mitigation: specs require separate `proposal_slicing_strategy` and `delivery_batch_strategy`, plus docs/policy updates.
- **Partial consumer migration** → Mitigation: tasks include agents, policy, docs and result contracts so no consumer continues to read legacy fields.
- **Rescan accidental overwrite** → Mitigation: specs require preserving overrides under `workflow_limits` and no-mutation behavior for invalid configs.

## Migration Plan

1. Update config model/default generation to write `workflow_limits` and remove generated top-level `loc_budget`/`delivery_strategy`.
2. Update rescan merge logic to preserve `workflow_limits` user overrides and to stop treating legacy top-level fields as canonical.
3. Update docs, agents, delivery policy and result contracts to read/report `workflow_limits` only.
4. Validate with strict OpenSpec and targeted implementation tests once code changes are authorized.

Rollback would require reintroducing the legacy fields as canonical sources, which is intentionally out of scope for this change.

## Open Questions

- Exact default threshold values can be confirmed during implementation by reusing current generated defaults where they exist, nested under `workflow_limits`.
