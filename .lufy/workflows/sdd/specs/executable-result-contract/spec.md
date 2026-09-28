# executable-result-contract Specification

### Requirement: Result Contract v1 is strictly parseable and bounded

The system SHALL decode `result-contract/v1` from supported YAML or JSON into a typed allow-listed model with bounded input and sanitized diagnostics.

#### Scenario: Canonical v1 envelope is decoded
- **WHEN** a caller supplies a documented v1 envelope within limits
- **THEN** the decoder returns the typed contract and preserves documented optional blocks
- **AND** does not require absent optional ledger, overview, diagnostics or structural acceptance blocks.

#### Scenario: Ambiguous or unsupported input is supplied
- **WHEN** input contains duplicate or unknown keys, unsupported schema/enums, aliases/tags, invalid encoding or values over limits
- **THEN** decoding fails before a transition decision
- **AND** the diagnostic identifies safe field path/reason without echoing values.

### Requirement: Equivalent contracts have deterministic canonical identity

The system SHALL canonicalize semantically equivalent supported envelopes to stable JSON and SHA-256 independent of source formatting or OS.

#### Scenario: Equivalent YAML and JSON are compared
- **WHEN** documents differ only by source format, key order, indentation or LF/CRLF
- **THEN** canonical representation and fingerprint are identical.

#### Scenario: Contract meaning changes
- **WHEN** a canonical status, evidence result or reference changes
- **THEN** the fingerprint changes deterministically.

### Requirement: Role and claim validation do not fabricate evidence

The system SHALL validate role allowed statuses, envelope coherence and minimum evidence categories without treating declarations as proof.

#### Scenario: Role emits unsupported status
- **WHEN** a role submits a status outside registered `allowed_status`
- **THEN** validation rejects it and no gate advances.

#### Scenario: Advanced claim lacks required evidence
- **WHEN** a contract claims validated, delivered or closed without policy-required evidence categories
- **THEN** validation returns missing-evidence recovery and next owner
- **AND** does not synthesize a passed command or gate.

### Requirement: Legacy normalization is explicit and conservative

The system SHALL normalize supported legacy output only through an explicit operation preserving provenance and uncertainty.

#### Scenario: Supported legacy payload is normalized
- **WHEN** explicit normalization is requested
- **THEN** output is v1 with `legacy_fallback: true`
- **AND** missing evidence is unavailable/not-run rather than passed.

#### Scenario: Legacy text is ambiguous
- **WHEN** multiple candidates, conflicting status or unbounded text prevents deterministic normalization
- **THEN** normalization rejects instead of guessing.

### Requirement: Stable CLI surfaces expose validation decisions

The system SHALL expose human-readable and versioned JSON commands for validation, normalization and transition evaluation.

#### Scenario: JSON validation is requested
- **WHEN** a supported result command uses JSON output
- **THEN** stdout contains one versioned decision document
- **AND** excludes decoration and private input content.

#### Scenario: Invalid input is evaluated
- **WHEN** validation or normalization rejects input
- **THEN** command exits non-zero by documented category
- **AND** emits sanitized recovery.
