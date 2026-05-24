## ADDED Requirements

### Requirement: Observable workflow-limit gates
The workflow SHALL make configured `workflow_limits.preflight` and `workflow_limits.stop_rules` observable in result contracts before a block advances to validation-ready, delivery-ready, delivered or closed states.

#### Scenario: Preflight status is reported before state advance
- **WHEN** a workflow block reaches a boundary that depends on configured preflight checks
- **THEN** the result contract reports each relevant preflight check as `passed`, `not_applicable`, `not_available` or `blocked` before advancing state

#### Scenario: Stop rule blocks silent continuation
- **WHEN** a configured stop rule is triggered by estimated file count, LOC, tool calls, session length, risk, validation failure or delivery condition
- **THEN** the workflow reports `status: blocked` or `status: escalated` with the exact stop rule and recovery path instead of continuing silently

### Requirement: Proportional Result Contract detail
The workflow SHALL scale Result Contract detail by tier while preserving the canonical envelope fields required for context recovery.

#### Scenario: T3 uses compact envelope
- **WHEN** work is classified as T3 Express and has no meaningful handoff or delivery risk
- **THEN** the result may use compact field values and `not_applicable` entries while still preserving envelope identity, status, evidence and next action

#### Scenario: T1 or multi-risk T2 uses full evidence
- **WHEN** work is classified as T1 or a multi-risk T2 review slice
- **THEN** the result contract includes explicit acceptance criteria status, validation evidence, workflow-limit decisions, risks and follow-ups sufficient for review without replaying the full conversation
