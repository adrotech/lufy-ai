# codex-contract-parity Specification Delta

## ADDED Requirements

### Requirement: Codex-only installation is contract complete
The system SHALL install every mandatory policy, template, reference and script required by Codex-visible Lufy skills without requiring `.opencode` assets.

#### Scenario: Delivery skill resolves canonical policy
- **WHEN** a delivery agent loads the Codex git delivery skill in a Codex-only installation
- **THEN** the canonical delivery invariants are readable from an installed tool-neutral or Codex-local path
- **AND** missing OpenCode paths do not weaken branch, validation or authorization gates.

#### Scenario: PR reviewer renders the required report
- **WHEN** a Codex-only installation executes the PR reviewer skill
- **THEN** its report template and review framework are locally available or fully embedded in the skill package
- **AND** the output contract remains equivalent to the OpenCode core contract.

### Requirement: Core role semantics remain equivalent across adapters
The system SHALL preserve the same Lufy role boundaries and Result Contract semantics for OpenCode and Codex while allowing tool-native execution surfaces.

#### Scenario: Codex exposes native Lufy roles
- **WHEN** runtime discovery exposes the exact custom role
- **THEN** orchestration reports `agent_execution_mode: native`
- **AND** preserves the role sandbox and delivery boundary.

#### Scenario: Exact role is unavailable
- **WHEN** only generic Codex roles or no subagent tooling are exposed
- **THEN** the workflow reports `emulated` or `inline` explicitly
- **AND** never claims native isolated execution.

#### Scenario: Result Contract crosses adapter boundaries
- **WHEN** a substantive role completes a routed phase
- **THEN** its result preserves status, evidence, risks, next action, adapter context and workflow decision fields required by Lufy.

### Requirement: Shared contracts have one maintained source
The system SHALL avoid manually maintaining divergent complete copies of tool-neutral role, policy and workflow contracts.

#### Scenario: Shared invariant changes
- **WHEN** a shared delivery, gate or Result Contract invariant is updated
- **THEN** generated or rendered OpenCode and Codex assets receive the same invariant
- **AND** adapter-specific overlays remain limited to native surface behavior.

#### Scenario: Catalog renders Codex methodology skills
- **WHEN** the effective catalog selects Codex with OpenSpec or Lufy SDD
- **THEN** it includes only the selected methodology contracts plus neutral core skills
- **AND** every included dependency is present in the same effective catalog.

### Requirement: Codex skills use native progressive disclosure
The system SHALL package Codex skills with concise discovery metadata and complete selected instructions/resources.

#### Scenario: Skill is discovered implicitly or explicitly
- **WHEN** Codex matches a skill description or the user invokes it
- **THEN** the selected `SKILL.md` and required local resources provide the complete workflow
- **AND** unrelated large contracts are not loaded into initial context.
