# result-transition-validation Specification Delta

## ADDED Requirements

### Requirement: Result transitions use a versioned protocol and explicit state table

The system SHALL evaluate `result-transition/v1` intents against an explicit state table while preserving flat status as compatible projection.

#### Scenario: Allowed transition satisfies guards
- **WHEN** prior state, next contract, role, evidence and context satisfy an allowed edge
- **THEN** evaluator returns accepted with next version and fingerprint.

#### Scenario: Edge or terminal rule is violated
- **WHEN** edge is absent, leaves closed, or skips required validation/sync/delivery
- **THEN** evaluator returns rejected with recovery
- **AND** prior state is unchanged.

### Requirement: Version and fingerprint fence stale writers

The system SHALL require expected version and previous fingerprint for transitions from existing state.

#### Scenario: Caller observes current state
- **WHEN** expected version/fingerprint match
- **THEN** evaluation may continue to other guards.

#### Scenario: Caller holds stale state
- **WHEN** version or fingerprint differs
- **THEN** result is conflict
- **AND** recovery requires reload, not overwrite.

### Requirement: Ownership leases are combined with fencing

The system SHALL validate pseudonymous owner, role, lease expiry and token digest without using time as the only guard.

#### Scenario: Active owner submits current transition
- **WHEN** owner/role match, lease is active and fencing is current
- **THEN** ownership guard passes without exposing raw token.

#### Scenario: Lease or owner is invalid
- **WHEN** owner differs, lease expired or digest mismatches
- **THEN** transition rejects/conflicts
- **AND** no later version is produced.

### Requirement: Transition retries are idempotent

The system SHALL evaluate idempotency key and canonical intent fingerprint before a second durable decision.

#### Scenario: Identical transition is retried
- **WHEN** accepted key is submitted with same fingerprint
- **THEN** result is duplicate_noop with original decision identity/version.

#### Scenario: Key is reused for another intent
- **WHEN** accepted key has different fingerprint
- **THEN** result is conflict
- **AND** original receipt/state remain unchanged.

### Requirement: Joins require terminal children and grouped evidence

The system SHALL accept a join only for explicitly required children with terminal contracts, causal correlation and grouped evidence.

#### Scenario: Required children are complete
- **WHEN** each declared child has terminal fingerprint, valid causality and grouped evidence
- **THEN** join guard passes deterministically.

#### Scenario: Child or evidence is incomplete
- **WHEN** child is missing, non-terminal, blocked without recovery, unrelated or evidence absent
- **THEN** join rejects with incomplete categories
- **AND** silence/timeout is not completion.
