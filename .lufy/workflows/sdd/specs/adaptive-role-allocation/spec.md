# adaptive-role-allocation Specification

### Requirement: Demand and capability contracts are strict and bounded

LUFY SHALL decode versioned `DemandSignal` and `CapabilityProfile` inputs through strict allow-listed schemas with bounded identifiers, lists, numeric ranges and sanitized errors.

#### Scenario: Valid allocation snapshot is decoded

- **GIVEN** one demand and bounded capability profiles with supported schema versions
- **WHEN** adaptive evaluation starts
- **THEN** inputs are canonicalized to deterministic identities and retain only documented metadata.

#### Scenario: Unknown or sensitive content is supplied

- **WHEN** an input contains unknown keys, duplicate YAML keys, oversized lists, raw prompts, outputs, paths or secrets
- **THEN** decoding rejects before scoring and diagnostics identify only the safe field/reason.

### Requirement: Roles are temporary capability hints

LUFY SHALL keep pseudonymous actor identity, current capabilities and `role_hint` as separate concepts; an assignment SHALL NOT grant permissions absent from the canonical Role Contract.

#### Scenario: Same actor is eligible for another function

- **GIVEN** an actor profile with capabilities matching a demand outside its prior role hint
- **WHEN** candidates are ranked
- **THEN** the actor may be recommended with a new temporary role hint while its identity, permissions and history remain unchanged.

#### Scenario: Role hint implies a protected permission

- **WHEN** a recommendation would require delivery, security or another undeclared permission
- **THEN** the result escalates and no permission or owner mutation is inferred from the hint.

### Requirement: Deterministic scoring is explainable

LUFY SHALL rank eligible profiles using the versioned integer policy `deterministic-v1`, expose every score term and resolve ties with stable keys.

#### Scenario: Same snapshot is evaluated repeatedly

- **WHEN** identical canonical demand, profiles, policy and allocation state are evaluated
- **THEN** ranking, score breakdown, exclusions, decision and fingerprint are identical across runs and operating systems.

#### Scenario: Demand or capacity changes

- **WHEN** priority, capability match, available budget, risk gap or coordination cost changes
- **THEN** the output identifies changed terms and any resulting recommendation change.

#### Scenario: Required capability is absent

- **WHEN** a profile lacks a required capability or sufficient budget
- **THEN** it is marked ineligible with a bounded reason and cannot win through other score terms.

### Requirement: Protected boundaries preempt allocation

LUFY SHALL evaluate protected boundaries before scoring and SHALL require human/orchestrator escalation for delivery, security, public contracts, database schema or destructive migrations.

#### Scenario: Protected work is requested

- **WHEN** a demand includes any protected boundary
- **THEN** decision is `escalate`, no candidate wins, no lease is created and recovery identifies the required authority.

#### Scenario: Non-protected work is requested

- **WHEN** the demand contains no protected boundary
- **THEN** normal eligibility and scoring may proceed without granting delivery or gate authority.

### Requirement: Adaptive routing is disabled by default

`.lufy/config/project.yaml` SHALL expose `adaptive_routing` with explicit `enabled`, `mode`, policy, lease and bounded pool settings, and missing config SHALL behave as disabled.

#### Scenario: New config is generated

- **WHEN** LUFY creates a project config
- **THEN** `adaptive_routing.enabled` is false, mode is `shadow` and bounded defaults are emitted.

#### Scenario: Partial config is rescanned

- **GIVEN** valid user overrides and unknown nested extras
- **WHEN** rescan completes missing defaults
- **THEN** overrides/extras are preserved and the feature is not activated implicitly.

#### Scenario: Feature is disabled

- **WHEN** recommendation or assignment is requested with adaptive routing disabled
- **THEN** no new adaptive assignment or lease is recorded and current deterministic routing remains authoritative.

### Requirement: Shadow and advisory modes have distinct effects

LUFY SHALL support only `shadow` and `advisory` modes in v1 and SHALL report `gate_advanced: false` for every adaptive decision.

#### Scenario: Shadow recommendation is evaluated

- **WHEN** mode is `shadow`
- **THEN** a ranked recommendation may be returned or observed but no lease, budget consumption, ownership change or gate transition occurs.

#### Scenario: Advisory recommendation is explicitly confirmed

- **WHEN** mode is `advisory` and a caller confirms a current recommendation
- **THEN** an adaptive assignment may be recorded with lease and budget projection while Result Contract ownership and gates remain unchanged.

### Requirement: Waiting pool is bounded and starvation-visible

LUFY SHALL project unassigned/yielded demand into a bounded deterministic waiting pool and SHALL expose starvation risk without silently elevating priority or authority.

#### Scenario: Multiple demands are waiting

- **WHEN** status is projected
- **THEN** items are ordered by priority, logical waiting age and stable demand ID with explicit truncation metadata.

#### Scenario: Demand waits beyond configured cycles

- **WHEN** the starvation threshold is reached
- **THEN** status reports a starvation warning and recommended review, but does not autoassign protected work or mutate priority.
