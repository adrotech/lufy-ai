# Adaptive Routing Contract

This contract defines the adapter-neutral safety boundary for adaptive role allocation. OpenCode, Codex, and future harness adapters may consume adaptive evidence, but they must not reinterpret it as workflow authority.

## Authority Model

- `actor_ref`, current capabilities, and `role_hint` are separate concepts. A `role_hint` is a temporary scheduling suggestion; it does not change identity, ownership, permissions, tool access, or canonical Role Contract boundaries.
- Adaptive recommendations, assignments, leases, yields, scores, and waiting-pool state are scheduler evidence only. They never advance Result Contract, validation, review, delivery, sync, archive, merge, or closure gates.
- Existing roles remain the only execution roles. Adaptive routing must not create synthetic roles or broaden the permissions of `orchestrator`, `sdd-router`, `explorer`, `implementer`, `test-writer`, `validator`, `reviewer`, or `delivery`.

## Modes And Fallback

- `disabled` is the default and prevents new adaptive assignments.
- `shadow` may compute and optionally record recommendations, but it never creates an assignment, consumes budget, changes ownership, or performs work.
- `advisory` may record an assignment or yield only after an explicit CLI/API mutation request with the required idempotency and lease inputs.
- Assignment confirmation must atomically verify that its recommendation is still the latest projection event and that global capacity plus accumulated actor budget remain available. A caller cannot revive a displaced recommendation by copying the current ledger version.
- No autonomous mode is defined by this contract.
- When adaptive configuration, CLI support, Run Ledger, or adapter integration is absent or unavailable, consumers must report `not_available` or `disabled` and continue with the existing deterministic SDD/role routing. They must not infer a recommendation, assignment, or successful release.

## Protected Boundaries

The following boundaries preempt scoring and require human/orchestrator escalation:

- delivery or Git/GitHub publication;
- security or authorization policy;
- public contracts;
- database schema;
- destructive migrations.

A recommendation for `delivery`, or any successor capability that intersects a protected boundary, is not authorization. The harness must obtain the same explicit approval and evidence required without adaptive routing.

## Safe Yield

- Yield means a safe, auditable release of adaptive lease and budget; it is not loss of identity, history, permissions, or workflow ownership.
- A checkpoint may contain only bounded enums/tokens and existing artifact/evidence references or SHA-256 digests for hypotheses and failed attempts. It must not contain prompts, responses, summaries, command output, secrets, raw paths, or free-form work content.
- Lease and budget are released only after Run Ledger returns a durable `recorded` receipt or an equivalent same-fingerprint `duplicate_noop`.
- Conflict, stale lease, owner mismatch, unavailable storage, or an ordinary yield after expiry leaves the assignment active and requires retry or escalation.
- Expiry never releases resources by wall clock alone. An expired lease may be recovered only by a durable, exact-fence checkpoint with reason `lease_expiring` and `next_status: waiting`.
- Disabling adaptive routing may still allow an explicit fenced yield of a pre-existing lease so rollback does not orphan adaptive resources.

## Harness Consumption

- `sdd-router` may include an adaptive recommendation in planning, but still applies tier, methodology, user scope, isolation, tool availability, review workload, and protected-boundary rules.
- `orchestrator` may use the recommendation to choose which existing capability to request. It must not spawn, reassign, mutate, deliver, or advance a gate solely because of adaptive output.
- Specialist roles operate within their existing permissions. A temporary hint never bypasses an implementation approval, validation evidence, review finding, delivery authorization, remote check, sync, or closure condition.
- Every human or machine-facing adaptive result must keep `gate_advanced=false` or the equivalent explicit non-authority statement.

## Privacy And Observability

- Persist only allow-listed, bounded, content-free metadata and pseudonymous/digested references.
- Diagnostics identify a safe field and reason without echoing rejected sensitive values.
- Run Ledger remains the append-only source of truth; projections are rebuildable and repair must not invent assignments or backfill history.
