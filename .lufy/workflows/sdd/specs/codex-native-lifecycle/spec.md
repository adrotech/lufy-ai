# codex-native-lifecycle Specification

### Requirement: Codex installation provides a useful native lifecycle
The system SHALL install project-local Codex hooks that support Lufy orientation and validation without requiring OpenCode runtime assets.

#### Scenario: Session start ensures derived skill state
- **WHEN** Codex starts or resumes in a trusted Lufy project
- **THEN** the lifecycle checks or ensures the Codex skill registry best-effort
- **AND** reports compact recovery when the CLI or registry is unavailable.

#### Scenario: Optional context and memory are unavailable
- **WHEN** context graph or Obsidian memory is not initialized
- **THEN** SessionStart reports availability and recovery without blocking the session
- **AND** does not invent memory hints.

#### Scenario: Lifecycle protects private content
- **WHEN** hooks produce additional context or diagnostics
- **THEN** output excludes prompt bodies, secrets, environment values and memory note contents
- **AND** includes only bounded status, paths and recovery metadata.

### Requirement: Subagent and stop hooks preserve workflow gates
The system SHALL use Codex lifecycle events to diagnose workflow state without silently advancing gates.

#### Scenario: Subagent completes without usable payload
- **WHEN** SubagentStop receives an empty or invalid substantive result
- **THEN** the hook reports the invalid payload condition for orchestrator recovery
- **AND** does not mark implementation, validation, delivery or closure complete.

#### Scenario: Main turn stops with incomplete workflow
- **WHEN** Stop can identify an active change with missing required evidence
- **THEN** it reports current status and next action compactly
- **AND** does not mutate tasks, specs, delivery or archive state.

### Requirement: Codex rules enforce conservative command policy
The system SHALL install testable Codex rules for sensitive Git/GitHub and destructive command prefixes.

#### Scenario: Delivery mutation requires review
- **WHEN** Codex evaluates commit, push, PR creation, PR merge, tag or release command prefixes
- **THEN** the Lufy rule decision is at least `prompt`
- **AND** the decision does not represent user delivery authorization.

#### Scenario: Known destructive command is forbidden
- **WHEN** a command matches a narrowly defined destructive prefix that Lufy policy disallows
- **THEN** the rule returns `forbidden` with a safer recovery alternative.

#### Scenario: Read-only command does not inherit mutation policy accidentally
- **WHEN** a read-only Git/GitHub command does not match a sensitive prefix
- **THEN** inline `not_match` examples prove the mutation rule does not overmatch.

### Requirement: Deep diagnostics are adapter-aware
The system SHALL execute lifecycle and plugin diagnostics only for the active tool adapter.

#### Scenario: Codex deep verify does not request OpenCode assets
- **WHEN** `verify --deep --tool codex` runs on a Codex installation
- **THEN** it validates Codex hooks, rules, agents and skills
- **AND** does not warn about missing `.opencode` hooks or plugins.
