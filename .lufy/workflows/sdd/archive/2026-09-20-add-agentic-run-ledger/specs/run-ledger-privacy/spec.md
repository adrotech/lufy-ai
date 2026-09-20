# run-ledger-privacy Specification Delta

## ADDED Requirements

### Requirement: Sensitive and conversational content is excluded by construction
The system SHALL persist only fields explicitly permitted by the run event schema and SHALL have no generic content body.

#### Scenario: Producer attempts to submit conversational content
- **WHEN** input contains a prompt, response, message, transcript, tool arguments or arbitrary metadata field
- **THEN** the write is rejected before persistence
- **AND** the diagnostic does not reproduce the submitted value.

#### Scenario: Artifact or tool produced rich content
- **WHEN** an event correlates an artifact, diff, command or tool invocation
- **THEN** the ledger stores only permitted type, outcome and digest references
- **AND** stores neither the content nor complete tool payload/output.

### Requirement: External identities and paths are pseudonymized locally
The system SHALL avoid persisting raw external session, turn, agent and absolute path identifiers.

#### Scenario: Hook contains external identifiers
- **WHEN** the adapter maps documented hook metadata
- **THEN** external identifiers are transformed into stable local references before persistence
- **AND** the raw values are absent from event and diagnostic files.

#### Scenario: Event references a filesystem artifact
- **WHEN** an artifact or evidence path is recorded
- **THEN** the event stores a path digest and safe logical category
- **AND** does not store the absolute path.

### Requirement: Ledger data remains local and outside version control
The system SHALL store runtime data under a Git-ignored local directory and SHALL perform no automatic network export.

#### Scenario: Project installs or upgrades managed assets
- **WHEN** Lufy updates its managed catalog
- **THEN** `.lufy/runtime/` remains user/runtime-owned and ignored by Git
- **AND** install, uninstall or upgrade does not upload or overwrite ledger history.

#### Scenario: Ledger records an event
- **WHEN** a local append succeeds
- **THEN** no network request is required or initiated by the ledger.

### Requirement: Observability failure does not block primary work
The system SHALL isolate best-effort lifecycle recording failures from workflow outcomes while making the failure visible.

#### Scenario: Hook cannot acquire storage or write an event
- **WHEN** a lifecycle producer encounters permissions, timeout, unavailable disk or invalid runtime state
- **THEN** it emits a compact sanitized warning with recovery guidance
- **AND** preserves the primary hook/workflow exit behavior and does not advance a gate.

#### Scenario: Explicit verification detects the same failure
- **WHEN** an operator invokes `run verify` against unavailable or corrupt storage
- **THEN** the command exits non-zero with the affected integrity category
- **AND** does not claim a valid run.

### Requirement: Retention is bounded, explicit and content-safe
The system SHALL support conservative age, count and byte limits without exposing deleted content.

#### Scenario: Terminal runs exceed policy
- **WHEN** explicit prune is executed after preview and terminal runs exceed configured limits
- **THEN** only eligible terminal run directories are removed in deterministic order
- **AND** the report includes identifiers/counts/bytes but no event content.

#### Scenario: Retention configuration is absent
- **WHEN** no project-specific retention values exist
- **THEN** documented conservative defaults apply
- **AND** missing configuration never means unlimited automatic deletion.

### Requirement: Privacy properties are continuously testable
The system SHALL include synthetic fixtures and assertions that prove prohibited values are absent from durable output.

#### Scenario: Privacy fixture exercises all producers
- **WHEN** tests run Codex and portable producers with canary secrets and conversational values
- **THEN** recursive inspection of the runtime fixture finds none of the canary values
- **AND** accepted digest and category metadata remains sufficient to correlate the run.
