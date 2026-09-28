# result-contract-ledger-correlation Specification Delta

## ADDED Requirements

### Requirement: Decisions correlate with Run Ledger without content

The system SHALL optionally record only allow-listed causal metadata and digests for validation/transition decisions.

#### Scenario: Accepted transition is recorded
- **WHEN** valid run, cause and idempotency refs are supplied and recording succeeds
- **THEN** event correlates schema, decision, status, version and fingerprints
- **AND** excludes contract body, summary, command output, raw paths and arbitrary metadata.

#### Scenario: Privacy canary appears in contract
- **WHEN** fields contain synthetic prompt, secret, path or output canaries
- **THEN** recursive ledger inspection finds none
- **AND** permitted digests remain sufficient.

### Requirement: Ledger outcomes preserve idempotency and conflicts

The system SHALL expose actual ledger receipt outcome without promoting it to workflow truth.

#### Scenario: Ledger records or deduplicates
- **WHEN** bridge receives new decision or equivalent retry
- **THEN** it reports recorded or duplicate_noop
- **AND** references durable event when available.

#### Scenario: Ledger reports conflict
- **WHEN** same key has different allow-listed metadata
- **THEN** bridge reports conflict
- **AND** prior event/status is not overwritten.

### Requirement: Failure policy distinguishes observation from mutation

The system SHALL apply explicit policy when ledger is unavailable.

#### Scenario: Best-effort lifecycle cannot record
- **WHEN** lifecycle validates output but storage is unavailable
- **THEN** hook continues with sanitized warning and ledger unavailable
- **AND** no gate advances automatically.

#### Scenario: Mutating transition requires durability
- **WHEN** transition requires ledger durability and storage is unavailable
- **THEN** it does not report accepted/delivered/closed
- **AND** returns retry/escalation recovery.

### Requirement: Lifecycle parses contracts instead of trusting substrings

The system SHALL replace substring-only detection with bounded parse/validation for supported substantive output.

#### Scenario: Text mentions schema but envelope is invalid
- **WHEN** assistant message mentions v1 but cannot decode/validate
- **THEN** lifecycle reports invalid contract
- **AND** does not accept the substring as handoff.

#### Scenario: Valid envelope omits optional ledger
- **WHEN** complete v1 is valid with no ledger block
- **THEN** lifecycle accepts structural compatibility
- **AND** does not require or fabricate ledger result.
