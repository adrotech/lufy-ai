## Context

The harness already has proportional SDD routing, review slices, local-first skills, and templates for result contracts. Recent workflow work also made `.opencode/project.yaml` top-level `workflow_limits` the canonical source for sizing, routing, proposal slicing, delivery batching, stop rules, and preflight.

The remaining gap is enforcement and observability: agent outputs still vary by role, workflow-limit decisions are not always explicit, and handoffs can require replaying context to understand whether a block is ready, blocked, escalated, validated, or delivery-pending.

## Goals / Non-Goals

**Goals:**

- Define one Result Contract envelope v1 that all substantive routed agent outputs can share.
- Make workflow-limit decisions explicit in router/orchestrator outputs and downstream handoffs.
- Preserve proportional routing: T3 remains lightweight, while T1/T2 get stronger structure and evidence.
- Keep delivery authorization separate from delivery batching guidance.
- Keep the change stack-agnostic and driven by `.opencode/project.yaml` when available.

**Non-Goals:**

- No new CLI command or public CLI flag.
- No database, network API, or external service integration.
- No replacement of OpenSpec artifacts or `/opsx-*` commands.
- No requirement that third-party or legacy subagents perfectly emit the envelope; orchestrator may normalize or summarize fallback outputs.

## Decisions

1. **Use an envelope, not role-specific schemas.**
   - Decision: define common top-level fields such as `schema_version`, `status`, `executive_summary`, `artifacts`, `evidence`, `workflow_decision`, `risks`, `next_recommended`, and `skill_resolution`.
   - Rationale: one envelope makes handoffs resumable while allowing role-specific details inside nested fields.
   - Alternative considered: keep separate contracts per agent. Rejected because it preserves the current normalization burden.

2. **Represent workflow-limit outcomes as structured decisions.**
   - Decision: router/orchestrator outputs include the `workflow_limits` source, exact paths considered, derived `workload_decision_needed`, stop/preflight status, review slices, and delivery batching guidance.
   - Rationale: the system should show why it escalated, paused, split, or deferred delivery.
   - Alternative considered: rely on prose notes. Rejected because prose is difficult to validate and easy to omit.

3. **Prefer strict current outputs, tolerate legacy fallback.**
   - Decision: local agents should emit envelope v1 for substantive routed work, but orchestrator may accept non-envelope outputs by wrapping them into a minimal fallback summary.
   - Rationale: this is a clean improvement without making every existing historical transcript or external subagent unusable.
   - Alternative considered: hard fail on every non-envelope response. Rejected because external agents and interrupted sessions may not comply.

4. **Keep delivery batching non-authoritative until delivery is explicitly authorized.**
   - Decision: `workflow_limits.delivery_batch_strategy` can recommend grouping after validation readiness, but cannot authorize Git/GH actions or mark work delivered.
   - Rationale: repository policy already requires explicit user authorization and remote check evidence.

## Risks / Trade-offs

- **Overly verbose contracts for T3** -> Mitigation: require the full envelope for substantive routed steps, but allow compact values and `not_applicable` fields for simple Express work.
- **Agents drift from the envelope over time** -> Mitigation: document the envelope in `AGENTS.md`/template and each role instruction that produces handoffs.
- **Workflow-limit interpretation becomes duplicated** -> Mitigation: make `sdd-router` the primary classifier and `orchestrator` the coordinator that carries decisions forward rather than inventing independent thresholds.
- **Embedded asset drift recurs** -> Mitigation: implementation tasks explicitly include embedded asset synchronization and `TestEmbeddedCatalogMatchesRepositoryAssets` validation when `.opencode` or specs change.

## Migration Plan

1. Update delta specs and local agent guidance to define envelope v1 and workflow decision fields.
2. Update root guidance/templates so human and installed instructions match.
3. Sync embedded managed assets for changed `.opencode`, `AGENTS.md.template`, and specs.
4. Validate OpenSpec, agent/documentation consistency, and embedded catalog parity.

Rollback is documentation/config-only: revert the changed agent/policy/spec/template files and regenerate embedded assets from the prior committed state if needed.

## Open Questions

- None currently blocking implementation. Field names can be refined during implementation if the final envelope remains testable and covers the specified requirements.
