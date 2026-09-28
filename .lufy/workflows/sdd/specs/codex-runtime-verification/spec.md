# codex-runtime-verification Specification

### Requirement: Codex adapter has an end-to-end installation matrix
The system SHALL test Codex across each supported methodology combination and lifecycle stage that can change adapter behavior.

#### Scenario: Codex OpenSpec installation is exercised
- **WHEN** the E2E matrix installs Codex with OpenSpec Full/Lite and T3 none
- **THEN** install, idempotencia, skills, verify, doctor and sync behavior are asserted
- **AND** no OpenCode instruction surface leaks into the effective catalog.

#### Scenario: Codex Lufy SDD installation is exercised
- **WHEN** the E2E matrix installs Codex with Lufy SDD Full/Lite and T3 none
- **THEN** install, skill registry, SDD new/validate/sync/archive and automatic overview behavior are asserted
- **AND** OpenSpec assets are absent unless independently selected.

### Requirement: Runtime probes distinguish unavailable from invalid
The system SHALL treat optional Codex runtime probes as additional evidence without making local Codex installation mandatory for all CI jobs.

#### Scenario: Codex CLI is available
- **WHEN** validation finds a compatible Codex CLI
- **THEN** it verifies stable feature discovery, custom agent loading prerequisites, hook configuration and execpolicy rules
- **AND** records exact command evidence.

#### Scenario: Codex CLI is unavailable
- **WHEN** CI or a development environment lacks Codex CLI
- **THEN** runtime probes report `not_available` or skip with reason
- **AND** mandatory structural, unit and integration tests still run.

### Requirement: Verification asserts behavior, not only asset presence
The system SHALL verify cross-file invariants and command outcomes in addition to file existence and SHA-256.

#### Scenario: Structurally valid files contain inconsistent selections
- **WHEN** project config and install state are individually parseable but semantically contradictory
- **THEN** verification fails despite both files existing and matching their independent schemas.

#### Scenario: Empty lifecycle surface is installed
- **WHEN** an adapter claims hooks or rules support but installs no effective behavior
- **THEN** deep verification reports the capability as unconfigured or placeholder
- **AND** public status does not describe it as active parity.

### Requirement: Documentation follows executable capability evidence
The system SHALL describe Codex features according to current installed behavior and verified runtime support.

#### Scenario: Codex capability reaches production
- **WHEN** hooks, rules, contracts or runtime probes are implemented and validated
- **THEN** README, installation, architecture, status and roadmap agree on availability and limitations.

#### Scenario: Capability remains deferred
- **WHEN** plugin marketplace, Observatory or advanced reporting is not implemented
- **THEN** documentation keeps it explicitly out of current installable parity.
