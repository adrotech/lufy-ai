# agentic-run-ledger Specification

### Requirement: Versioned event contract is minimal and deterministic
The system SHALL persist run events with a versioned, typed, allow-listed contract whose canonical representation is deterministic and bounded.

#### Scenario: Valid event is canonicalized
- **WHEN** a producer submits a supported typed event within configured limits
- **THEN** the ledger assigns or validates required identities and serializes one canonical `lufy-run-event/v1` envelope
- **AND** equivalent inputs produce the same fingerprint.

#### Scenario: Unknown or unbounded input is submitted
- **WHEN** a producer supplies an unknown field, unsupported enum, invalid identifier or value over its limit
- **THEN** the ledger rejects the event without persisting a partial record
- **AND** the diagnostic identifies the field category without echoing its value.

### Requirement: Events are append-only and atomically published
The system SHALL store each accepted event as an immutable, atomically published record under the run directory.

#### Scenario: Event append succeeds
- **WHEN** a valid new event is recorded
- **THEN** readers observe either the complete event or no event
- **AND** an existing event file is never rewritten in place.

#### Scenario: Process stops during publication
- **WHEN** a process terminates before an event or receipt is fully published
- **THEN** the prior event history remains readable
- **AND** verification reports or safely recovers the incomplete derived state without inventing an event.

### Requirement: Repeated delivery is idempotent and conflicts are explicit
The system SHALL evaluate an idempotency key and canonical fingerprint before producing a durable effect.

#### Scenario: Same key and same fingerprint are retried
- **WHEN** an already accepted key is submitted with the same fingerprint
- **THEN** the ledger returns `duplicate_noop` and the original `event_id`
- **AND** no second event is created.

#### Scenario: Same key is reused for different input
- **WHEN** an accepted key is submitted with a different fingerprint
- **THEN** the ledger returns a sanitized conflict
- **AND** preserves the original event and receipt unchanged.

### Requirement: Causal relationships survive concurrent execution
The system SHALL represent root, parent, direct cause and logical order independently from wall-clock ordering.

#### Scenario: Subagent event is linked to its parent
- **WHEN** a subagent start is observed with a resolvable parent binding
- **THEN** its run stores `parent_run_id` and the causing event reference
- **AND** status reconstruction places it below the correct parent.

#### Scenario: Concurrent writers append to one run
- **WHEN** two supported processes attempt to append events to the same run
- **THEN** each accepted event receives a unique local sequence and monotonically advancing Lamport clock
- **AND** neither writer produces a partially interleaved record.

### Requirement: Artifacts, evidence and checkpoints are content-free references
The system SHALL correlate work products using typed references and digests rather than storing their contents.

#### Scenario: Agent produces an artifact and validation evidence
- **WHEN** a checkpoint references a task, artifact and evidence
- **THEN** the event contains only allowed identifiers, categories, results and path/content digests
- **AND** the artifact body, command output and file contents are absent.

#### Scenario: Metrics are not supplied by the adapter
- **WHEN** duration, token or cost data is unavailable
- **THEN** the projection reports the metric as unavailable
- **AND** does not infer a zero value.

### Requirement: Derived state is verifiable and rebuildable
The system SHALL treat receipts, bindings, indexes and summaries as derived structures that can be checked against immutable events.

#### Scenario: Projection is stale or missing
- **WHEN** status or summary detects that a projection does not match its source event set
- **THEN** it rebuilds in memory or reports staleness according to command mode
- **AND** never treats the stale projection as stronger evidence than source events.

#### Scenario: Explicit repair is requested
- **WHEN** an operator runs verification with the supported repair option
- **THEN** only reconstructible derived structures are replaced atomically
- **AND** immutable source events are not modified.
