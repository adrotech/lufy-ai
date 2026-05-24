## ADDED Requirements

### Requirement: Numeric stop rules for workload guard
The workflow SHALL apply explicit numeric stop rules to prevent oversized, under-scoped, or low-evidence sessions from continuing silently.

#### Scenario: Four-file rule pauses for workload decision
- **WHEN** a routed block is estimated to touch four or more significant files or implementation discovers the block touches four or more significant files
- **THEN** the orchestrator SHALL require a workload decision, tier escalation, or review-slice plan before continuing beyond the current safe boundary

#### Scenario: Twenty-tool-calls rule pauses long routing
- **WHEN** a coherent routing or implementation block exceeds twenty tool calls without reaching a resumable state
- **THEN** the orchestrator SHALL pause for a concise state summary and decide whether to continue, compact context, escalate, or split the work

#### Scenario: Multi-file write rule requires plan
- **WHEN** a step proposes or attempts writes across multiple non-trivial files
- **THEN** the workflow SHALL verify that a scoped plan or review slice exists and SHALL avoid broad multi-file mutation without observable acceptance criteria

#### Scenario: Long-session rule requires resumable handoff
- **WHEN** a session becomes long enough that evidence, decisions, or next actions are no longer easily resumable
- **THEN** the workflow SHALL create or request a handoff summary before continuing with implementation, validation, or delivery routing

### Requirement: Stop-rule evidence in Result Contract
The workflow SHALL report triggered or evaluated numeric stop rules in Result Contract envelope v1 so downstream roles can continue without rediscovering the decision.

#### Scenario: Stop rule triggers blocked or escalated state
- **WHEN** a numeric stop rule is triggered and requires a decision before continuing
- **THEN** the result SHALL report `status: blocked` or `status: escalated`, the exact rule, the evidence that triggered it, and the recommended next owner/action

#### Scenario: Stop rules are clear at boundary
- **WHEN** a routed block reaches an implementation or validation boundary without triggering numeric stop rules
- **THEN** the result SHALL report `stop_rule_status: clear` or a proportional equivalent in the workflow decision fields

#### Scenario: Stop rules unavailable are explicit
- **GIVEN** `.opencode/project.yaml` or `workflow_limits.stop_rules` is not available
- **WHEN** a result reports workflow-limit fields
- **THEN** it SHALL report stop-rule configuration as `not_available` while still applying repository-level default guardrails from agent instructions
