# causal-run-observability Specification Delta

## ADDED Requirements

### Requirement: Operators can inspect a run through stable CLI surfaces
The system SHALL expose run status, summary and verification in deterministic human-readable and machine-readable forms.

#### Scenario: Root run has subagents and checkpoints
- **WHEN** an operator requests status or summary for the root run
- **THEN** the output relates each subagent to its parent, task, artifacts, evidence, blockers and terminal state
- **AND** identifies unavailable metrics explicitly.

#### Scenario: Machine-readable output is requested
- **WHEN** a supported run command uses JSON output
- **THEN** it returns a versioned deterministic response with stable status and error codes
- **AND** excludes terminal decoration and private content.

### Requirement: Portable producers can record typed checkpoints
The system SHALL provide adapter-neutral CLI and internal interfaces for recording events and workflow checkpoints.

#### Scenario: Non-Codex adapter records a checkpoint
- **WHEN** a producer supplies a valid run reference, event kind, idempotency key and typed checkpoint
- **THEN** the same domain and store rules used by automatic producers are applied
- **AND** no adapter-specific payload is required.

#### Scenario: Result Contract references ledger evidence
- **WHEN** a substantive handoff includes supported optional ledger references
- **THEN** consumers can correlate its run and event without making the ledger authoritative for the handoff status
- **AND** legacy Result Contract payloads without those references remain valid.

### Requirement: Codex lifecycle metadata creates causal bindings
The system SHALL map documented Codex hook metadata to run events without consuming transcript content.

#### Scenario: Codex starts a subagent
- **WHEN** `SubagentStart` supplies session, turn, agent and agent type metadata
- **THEN** the adapter creates or resolves a pseudonymized child binding and records a causal start event
- **AND** does not open or persist a transcript path.

#### Scenario: Codex stops a subagent or session
- **WHEN** `SubagentStop`, `Stop` or `SessionEnd` is received more than once
- **THEN** the derived idempotency key makes subsequent equivalent delivery a no-op
- **AND** the terminal relationship remains consistent.

### Requirement: Verification distinguishes source integrity from derived freshness
The system SHALL verify event schema, identity, causal links, clocks, receipts and projections with actionable results.

#### Scenario: Run history is valid
- **WHEN** `run verify` examines a complete consistent run
- **THEN** it exits successfully and reports source integrity and projection freshness separately.

#### Scenario: Causal link or receipt is invalid
- **WHEN** verification finds a missing cause, invalid parent, duplicate sequence or mismatched receipt
- **THEN** it exits non-zero with a sanitized diagnostic and recovery classification
- **AND** does not silently alter source events.

### Requirement: Retention preserves active work
The system SHALL apply bounded retention only to terminal runs that satisfy the configured policy.

#### Scenario: Prune is previewed
- **WHEN** an operator runs prune in dry-run mode
- **THEN** the command reports eligible run identifiers, counts and bytes without deleting files.

#### Scenario: Active run exceeds an age or size threshold
- **WHEN** retention evaluates a run that is not terminal
- **THEN** the run and all of its causal data are preserved
- **AND** the summary identifies it as protected active state.
