# Run Ledger Producer Contract

This contract defines the adapter-neutral boundary for producers of Lufy causal run metadata.

## Portable surface

- Producers submit the typed `lufy-run-event/v1` draft through `lufy-ai run record --idempotency-key <key>` or use `lufy-ai run checkpoint` for a typed workflow checkpoint.
- Every producer supplies a stable idempotency key. Repeated equivalent delivery is `duplicate_noop`; the same key with different typed metadata is a conflict and never overwrites the original event.
- Root/subagent relationships use local `run_id`, `parent_run_id` and optional `caused_by_event_id`. External identifiers must be pseudonymized before entering the domain.
- The allow-list excludes prompts, responses, messages, transcripts, tool arguments, diffs, file contents, secrets and arbitrary metadata maps.
- Ledger failure is advisory for lifecycle integrations: emit a compact sanitized warning and preserve the primary workflow result and gate state.

## Adapter responsibilities

An automatic adapter maps only documented stable lifecycle metadata into `EventDraft`, derives retry-stable idempotency keys and records explicit metric availability. It must not open transcript paths or infer completion of SDD, validation, delivery or archive gates.

Codex is the initial automatic producer for `SessionStart`, `SubagentStart`, `SubagentStop`, `Stop` and `SessionEnd`. Other adapters remain supported through the portable CLI/internal boundary; automatic instrumentation for them is intentionally not required in this phase.

Result Contract v1 may carry the optional `ledger` references defined in `result-contract.md`. Those references correlate evidence but never make the ledger authoritative for the handoff status.
