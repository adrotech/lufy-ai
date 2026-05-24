## ADDED Requirements

### Requirement: Result Contract envelope v1
The system SHALL define and use a canonical Result Contract envelope v1 for substantive routed agent handoffs and final workflow results.

#### Scenario: Routed agent emits envelope
- **WHEN** a routed local agent completes a substantive workflow step
- **THEN** its result includes `schema_version`, `status`, `executive_summary`, `artifacts`, `evidence`, `risks`, `next_recommended` and `skill_resolution`

#### Scenario: Envelope identifies resumable state
- **WHEN** a workflow resumes after handoff, compaction or session interruption
- **THEN** the envelope identifies whether the step is `ready`, `implemented`, `validated`, `delivery_pending`, `sync_pending`, `blocked`, `escalated`, `delivered` or `closed`

#### Scenario: Legacy output is normalized
- **WHEN** a third-party, historical or interrupted output does not provide Result Contract envelope v1
- **THEN** the orchestrator MAY normalize it into a minimal envelope with explicit `legacy_fallback: true` and any missing evidence marked as `not_available`

### Requirement: Workflow-limit decision output
The router and orchestrator SHALL expose workflow-limit-driven decisions as structured output derived from `.opencode/project.yaml` top-level `workflow_limits` when that file is available.

#### Scenario: Router reports workload decision inputs
- **WHEN** `sdd-router` evaluates a non-trivial request with `.opencode/project.yaml` available
- **THEN** it reports the `workflow_limits` paths considered, estimated workload inputs, tier decision, confidence, and whether `workload_decision_needed` is true

#### Scenario: Router proposes review slices from configured slicing limits
- **WHEN** estimated scope, file count, risk or configured routing limits require splitting before implementation or review
- **THEN** `sdd-router` uses `workflow_limits.proposal_slicing_strategy` to propose `review_slices` with objective, expected files, acceptance criteria, validation, risk and PR guidance

#### Scenario: Orchestrator carries workflow decisions forward
- **WHEN** orchestrator delegates to another agent after routing
- **THEN** the handoff includes the workflow decision fields needed by that role and does not require the receiving agent to rediscover the same limits from conversation history

### Requirement: Delivery batching remains authorization-gated
The workflow SHALL report delivery batching guidance separately from delivery authorization.

#### Scenario: Delivery batching guidance is advisory
- **WHEN** validated work has delivery grouping guidance from `workflow_limits.delivery_batch_strategy`
- **THEN** the result contract reports the recommended grouping but keeps delivery state as `delivery_pending` until the user explicitly authorizes Git/GH delivery

#### Scenario: Delivery role receives batching context
- **WHEN** delivery is explicitly authorized
- **THEN** the delivery role receives the relevant batching, preflight and stop-rule context from the Result Contract envelope or current `.opencode/project.yaml`
