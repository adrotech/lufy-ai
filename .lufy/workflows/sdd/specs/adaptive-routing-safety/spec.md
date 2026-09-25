# adaptive-routing-safety Specification

### Requirement: Adaptive decisions never advance workflow gates

LUFY SHALL treat adaptive recommendations, assignments and yields as scheduler evidence only; Result Contract, validation, review, delivery, sync and closure remain independently authoritative.

#### Scenario: Recommendation has a high score

- **WHEN** a candidate ranks first with all adaptive checks satisfied
- **THEN** output still reports `gate_advanced: false` and cannot claim validated, delivered or closed.

#### Scenario: Yield selects a successor role hint

- **WHEN** a checkpoint names successor capabilities or role hint
- **THEN** no Result Contract owner or role permission changes until the existing authoritative transition accepts it.
### Requirement: Runtime modes fail safely

LUFY SHALL distinguish disabled, shadow, advisory, unavailable, conflict and escalated outcomes in versioned human/JSON surfaces.

#### Scenario: Ledger is unavailable in shadow mode

- **WHEN** read-only scoring can complete but observation persistence cannot
- **THEN** recommendation remains explicitly non-durable and no effect is claimed.

#### Scenario: Ledger is unavailable for advisory mutation

- **WHEN** assign or yield requires persistence
- **THEN** command fails without partial state and returns bounded recovery.
### Requirement: Allocation telemetry is content-free

LUFY SHALL persist only allow-listed IDs, enums, integer scores, digests, logical versions and UTC timestamps for adaptive routing.

#### Scenario: Persistence is recursively inspected

- **WHEN** events, receipts, locks, projections and diagnostics are scanned with privacy canaries
- **THEN** no prompts, responses, summaries, secrets, command outputs, raw paths, diffs or hypothesis/attempt text are present.

#### Scenario: Metric is unavailable

- **WHEN** duration, token, cost, capability or context evidence is missing
- **THEN** availability is explicit and missing data is not treated as zero or fabricated success.
### Requirement: CLI exposes explicit safe operations

LUFY SHALL expose scriptable `adaptive recommend`, `adaptive assign`, `adaptive yield` and `adaptive status` commands with strict stdin/file selection, stable exit categories and sanitized diagnostics.

#### Scenario: JSON recommendation is requested

- **WHEN** a valid snapshot is evaluated with `--json`
- **THEN** stdout contains one versioned decision document with policy, mode, breakdown, evidence availability and no decoration.

#### Scenario: Mutating operation lacks explicit recording intent

- **WHEN** assign or yield would require a durable effect without its required explicit option/input
- **THEN** CLI remains read-only or returns usage/rejected rather than mutating implicitly.
### Requirement: Harness consumers preserve authority boundaries

Orchestrator, router, implementer, reviewer, validator and delivery guidance SHALL consume adaptive output only within their existing permissions and SHALL preserve manual escalation for protected work.

#### Scenario: Router consumes advisory recommendation

- **WHEN** adaptive routing recommends a capability/role hint
- **THEN** router may include it in planning but still applies SDD tier, isolation, tool availability and user scope.

#### Scenario: Delivery is recommended as successor

- **WHEN** successor capabilities intersect delivery or another protected boundary
- **THEN** harness requests explicit human authorization and does not execute Git/GH from the adaptive decision.

#### Scenario: Installed assets are synchronized

- **WHEN** adaptive routing contracts change
- **THEN** root instructions, managed/embedded assets, OpenCode/Codex guidance and documentation describe the same disabled/shadow/advisory and protected-boundary semantics.
