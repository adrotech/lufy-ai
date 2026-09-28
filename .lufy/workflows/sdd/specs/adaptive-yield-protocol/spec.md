# adaptive-yield-protocol Specification

### Requirement: Advisory assignments use fenced leases

LUFY SHALL create adaptive assignments only from current advisory recommendations and SHALL bind them to pseudonymous owner, policy/snapshot fingerprint, expected projection version and a lease token digest with expiry.

#### Scenario: Current recommendation is confirmed

- **WHEN** a caller submits matching owner, recommendation, expected version and lease digest, the recommendation is still the last projection event, and global capacity plus actor budget remain available
- **THEN** one assignment event is durably recorded and budget is consumed in the adaptive projection.

#### Scenario: Stale or mismatched confirmation is submitted

- **WHEN** owner, recommendation fingerprint, expected version, lease digest or expiry does not match current state
- **THEN** assignment rejects/conflicts without consuming budget or overwriting prior state.

#### Scenario: Recommendation was displaced or capacity changed

- **WHEN** an intervening event advanced the projection after recommendation, or global capacity or accumulated actor budget is exhausted at confirmation
- **THEN** assignment rejects/conflicts in the authoritative compare-and-swap even if the caller supplied the ledger's current version.
### Requirement: Yield checkpoint is content-free and reusable

LUFY SHALL represent yield with bounded reason, assignment/lease identity, artifact/evidence refs, hypothesis/failed-attempt digests and successor capability tokens, without free-form work content.

#### Scenario: Blocked agent yields reusable context

- **GIVEN** an active assignment and valid lease
- **WHEN** yield supplies supported checkpoint refs and successor capabilities
- **THEN** the durable checkpoint identifies reusable context and next needs without containing prompt, response, summary, command output, raw path or hypothesis text.

#### Scenario: Yield contains forbidden content

- **WHEN** payload contains an unknown free-text field, raw path, secret or oversized references
- **THEN** decoding rejects before persistence and does not reflect the sensitive value.
### Requirement: Release happens only after durable checkpoint

LUFY SHALL release adaptive lease and budget only after the yield event is durably recorded or confirmed as an equivalent duplicate.

#### Scenario: Yield append succeeds

- **WHEN** Run Ledger returns `recorded` or same-fingerprint `duplicate_noop`
- **THEN** projection marks lease released, restores budget and requeues/completes/escalates demand according to checkpoint state.

#### Scenario: Persistence fails or conflicts

- **WHEN** ledger is unavailable or the idempotency key conflicts
- **THEN** operation does not report release and current assignment remains active with retry/escalation recovery.

#### Scenario: Expired lease receives an ordinary yield

- **WHEN** a caller submits a yield for an expired lease with any ordinary reason or next status
- **THEN** operation rejects/conflicts and the projection keeps the assignment active.

#### Scenario: Expired lease is explicitly recovered

- **WHEN** a caller submits exact assignment, actor, lease fencing and expected version with reason `lease_expiring` and `next_status: waiting`
- **THEN** the durable checkpoint may release and requeue the adaptive resources without treating wall-clock expiry as an implicit event.
### Requirement: Yield is idempotent under concurrency

LUFY SHALL use event idempotency, expected projection version and lease fencing so concurrent yield/assign operations cannot release twice or transfer stale work.

#### Scenario: Equivalent yield is retried

- **WHEN** the same idempotency key and canonical checkpoint are submitted again
- **THEN** outcome is `duplicate_noop` and no second budget release or event is produced.

#### Scenario: Two writers race with different checkpoints

- **WHEN** concurrent writers target the same active assignment from the same prior version
- **THEN** at most one advances projection and the other receives conflict/stale recovery.
### Requirement: Disabled mode permits safe cleanup only

Disabling adaptive routing SHALL prevent new assignments while allowing explicit fenced yield of pre-existing leases so rollback does not orphan budget.

#### Scenario: Feature is disabled with no existing lease

- **WHEN** assign or yield is requested
- **THEN** new assignment is blocked and no synthetic release is reported.

#### Scenario: Feature is disabled with a valid existing lease

- **WHEN** an explicit yield presents the matching owner, lease and expected version
- **THEN** cleanup may record the checkpoint and release adaptive resources without changing Result Contract ownership.
### Requirement: Projection is rebuildable from append-only events

LUFY SHALL derive assignment, budget and waiting state from validated Run Ledger events and SHALL treat cached status as replaceable derived data.

#### Scenario: Projection is missing or stale

- **WHEN** adaptive status detects a source digest mismatch
- **THEN** it rebuilds in memory or through explicit repair without rewriting source events.

#### Scenario: Historical events lack adaptive metadata

- **WHEN** older Run Ledger events are read
- **THEN** they remain valid, adaptive state is unavailable/empty as appropriate and no synthetic assignments are backfilled.
