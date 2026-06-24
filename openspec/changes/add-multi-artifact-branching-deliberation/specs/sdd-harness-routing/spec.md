## ADDED Requirements

### Requirement: Artifact branching recommendation
The `sdd-router` SHALL evaluate whether high-risk or high-uncertainty T1/T2 work needs artifact branching before implementation planning continues.

#### Scenario: Multi-risk work recommends artifact branching
- **WHEN** a T1 request or multi-risk T2 request has high uncertainty across scope, architecture, product outcome, validation strategy or review workload
- **THEN** `sdd-router` MUST recommend `artifact_branching` when branching is needed to compare credible alternatives
- **AND** the recommendation SHALL include candidate_count with a maximum of 2

#### Scenario: Dominant solution disables branching
- **WHEN** a request has a dominant low-risk solution, clear acceptance criteria and no meaningful unresolved trade-off
- **THEN** `sdd-router` SHALL recommend candidate_count 1
- **AND** it SHALL report artifact branching or deliberation as `not_needed`

#### Scenario: Router preserves workflow limits
- **WHEN** `sdd-router` recommends artifact branching
- **THEN** the recommendation SHALL cite relevant `workflow_limits` paths and `parallel_execution` constraints when available
- **AND** it SHALL NOT treat proposal slicing or delivery batching as delivery authorization

### Requirement: Branching handoff fields
The routing handoff SHALL carry enough artifact-branching metadata for `orchestrator` and downstream roles to proceed without rediscovering the decision.

#### Scenario: Router emits branching metadata
- **WHEN** artifact branching is recommended
- **THEN** the router output SHALL include stage, candidate_count, reason, parallel_allowed, requires_join, candidate_isolation guidance, merge_plan_required and human_escalation_triggers

#### Scenario: Branching not applicable is explicit
- **WHEN** artifact branching is not needed for a routed request
- **THEN** the router output SHALL explicitly report artifact_branching as `not_needed` or an equivalent falsey structured value
- **AND** downstream roles SHALL continue with one canonical artifact path

### Requirement: Parallel candidate constraints
The routing harness SHALL allow parallel candidate generation only when artifacts are independent and join planning is explicit.

#### Scenario: Parallel candidates require independent artifacts
- **WHEN** `parallel_execution.enabled` is true and branching candidates are generated in parallel
- **THEN** `sdd-router` and `orchestrator` SHALL require independent candidate artifact paths, a merge plan and grouped validation after join

#### Scenario: Unsafe parallelism is blocked
- **WHEN** delivery, Git/GH work, unresolved public contracts, unresolved security decisions or shared mutable canonical artifacts would be parallelized
- **THEN** the harness SHALL NOT run those steps in parallel
- **AND** it SHALL route to sequential join or human decision before continuing
