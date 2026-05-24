## ADDED Requirements

### Requirement: Workflow limits canonical consumption
Agents, workflow documentation, delivery policy and result contracts SHALL consume and report project workflow limits from `.opencode/project.yaml` top-level `workflow_limits` only.

#### Scenario: Router reads canonical workflow limits
- **WHEN** `sdd-router` evaluates sizing, routing, slicing or escalation inputs for a project with `.opencode/project.yaml`
- **THEN** it reads those inputs from `workflow_limits` and MUST NOT read top-level `loc_budget` or top-level `delivery_strategy` as valid sources

#### Scenario: Orchestrator reports canonical workflow limits
- **WHEN** `orchestrator` reports routing rationale, handoff constraints or result contract fields that depend on project workflow limits
- **THEN** it references `workflow_limits` paths as the source of truth and MUST NOT report legacy top-level fields as canonical

#### Scenario: Delivery policy reads canonical workflow limits
- **WHEN** delivery guidance needs batching, preflight or stop-rule limits from project config
- **THEN** delivery policy and delivery role instructions consume those limits from `workflow_limits` only

### Requirement: Proposal slicing is separate from delivery batching
The workflow SHALL treat `workflow_limits.proposal_slicing_strategy` and `workflow_limits.delivery_batch_strategy` as different controls with different lifecycle phases.

#### Scenario: Proposal slicing before implementation or review
- **WHEN** a proposal or review workload needs to be split into smaller coherent slices
- **THEN** the workflow uses `workflow_limits.proposal_slicing_strategy` to decide implementation/review slices before delivery authorization

#### Scenario: Delivery batching after validation readiness
- **WHEN** validated or delivery-ready changes need to be grouped for Git/GH delivery
- **THEN** the workflow uses `workflow_limits.delivery_batch_strategy` and MUST NOT reinterpret proposal slicing rules as delivery batching authorization

### Requirement: Workflow preflight and stop rules
The workflow SHALL apply `workflow_limits.preflight` and `workflow_limits.stop_rules` as project-local gates for pausing, escalating or requiring evidence before continuing.

#### Scenario: Preflight gate before a bounded workflow phase
- **WHEN** a workflow phase has configured preflight checks under `workflow_limits.preflight`
- **THEN** the responsible role verifies or reports those checks before moving to the next state that depends on them

#### Scenario: Stop rule forces escalation
- **WHEN** an active task reaches a configured condition under `workflow_limits.stop_rules`
- **THEN** the responsible role pauses, reports the blocking condition and escalates to the appropriate role or user decision instead of continuing silently

## MODIFIED Requirements

### Requirement: Block-scoped proportional validation
The workflow SHALL run validation/testing proportionally at the end of a task, coherent block, proposal block, or review slice, SHALL apply any relevant `workflow_limits.preflight` and `workflow_limits.stop_rules`, and SHALL avoid constant test loops for individual micro-checkboxes unless an exception gate applies.

#### Scenario: Validation waits for coherent block boundary
- **WHEN** an agent completes an internal micro-step that is part of a larger coherent task or block
- **THEN** the workflow SHALL NOT require full validation/testing for that micro-step and SHALL defer grouped validation to the coherent block boundary

#### Scenario: Validation runs before validated state
- **WHEN** a task, coherent block, proposal block, or review slice is ready to move from `implemented` to `validated`
- **THEN** the workflow SHALL run the real applicable validation commands or document proportional static/manual evidence and SHALL report exact evidence before using the `validated` state

#### Scenario: Exception allows early validation
- **WHEN** a blocker, risky change, feedback loop, or failure diagnosis requires earlier evidence
- **THEN** the workflow MAY run focused validation before the block boundary while preserving grouped final validation for the block when applicable

#### Scenario: Configured workflow limits affect validation gate
- **WHEN** `workflow_limits.preflight` or `workflow_limits.stop_rules` define additional validation, pause or escalation conditions for the current block
- **THEN** the workflow applies those configured limits before reporting the block as `validated` or delivery-ready

## REMOVED Requirements

### Requirement: Legacy workflow limit source consumption
**Reason**: Top-level `loc_budget` and `delivery_strategy` are no longer valid workflow-limit sources and would create conflicting behavior across router, orchestrator, delivery policy and result contracts.

**Migration**: Update workflow consumers and human-facing documentation to read and report `workflow_limits` paths only.

#### Scenario: Legacy fields no longer drive workflow behavior
- **WHEN** `.opencode/project.yaml` contains top-level `loc_budget` or top-level `delivery_strategy`
- **THEN** agents, policy and result contracts MUST NOT use those fields to decide sizing, routing, slicing, delivery batching, stop rules or preflight behavior
