## ADDED Requirements

### Requirement: Numeric workload guard
The routing harness SHALL make workload decisions observable from estimated LOC and file count using canonical `workflow_limits` when available.

#### Scenario: LOC budget requires workload decision
- **GIVEN** `.opencode/project.yaml` exists and defines `workflow_limits.sizing.loc_budget`
- **WHEN** `sdd-router` estimates `estimated_loc` greater than `workflow_limits.sizing.loc_budget`
- **THEN** it SHALL emit `workload_decision_needed: true` and recommend a workload decision before implementation continues

#### Scenario: Five or more files trigger escalation or slicing
- **WHEN** `sdd-router` estimates `estimated_files >= 5`
- **THEN** it SHALL either escalate the tier or propose bounded slices appropriate to the risk and scope

#### Scenario: Missing sizing config is not invented
- **GIVEN** `.opencode/project.yaml` is missing or `workflow_limits.sizing.loc_budget` is not available
- **WHEN** `sdd-router` evaluates estimated workload
- **THEN** it SHALL report the sizing source as `not_available` and SHALL NOT use legacy top-level `loc_budget` or invented defaults

### Requirement: Canonical workflow limits propagation
The router and orchestrator SHALL read and propagate workflow-limit decisions from `.opencode/project.yaml` top-level `workflow_limits` paths when available, and SHALL report unavailable paths explicitly.

#### Scenario: Router reports all relevant workflow limit paths
- **WHEN** `sdd-router` evaluates a non-trivial request for a project
- **THEN** its output SHALL report availability for `workflow_limits.sizing`, `workflow_limits.routing`, `workflow_limits.proposal_slicing_strategy`, `workflow_limits.delivery_batch_strategy`, `workflow_limits.preflight`, and `workflow_limits.stop_rules`

#### Scenario: Orchestrator preserves routing decision
- **WHEN** `orchestrator` delegates after `sdd-router` classified a request
- **THEN** it SHALL propagate the workflow decision fields, source paths, workload decision, review slices, preflight status, stop-rule status, and delivery batching guidance needed by the receiving role

#### Scenario: Legacy top-level fields are ignored
- **GIVEN** `.opencode/project.yaml` contains top-level `loc_budget` or top-level `delivery_strategy`
- **WHEN** `sdd-router` or `orchestrator` computes workflow limits
- **THEN** it SHALL NOT consume those fields as canonical sizing, routing, slicing, batching, preflight, stop-rule, authorization, or closure inputs

### Requirement: Proposal slicing remains separate from delivery batching
The routing harness SHALL derive `review_slices` from proposal/review slicing configuration only, and SHALL keep delivery batching advisory until explicitly authorized delivery.

#### Scenario: Review slices use proposal slicing strategy
- **GIVEN** `workflow_limits.proposal_slicing_strategy` is available
- **WHEN** estimated file count, LOC, risk, or tier requires splitting before implementation or review
- **THEN** `sdd-router` SHALL derive `review_slices` from `workflow_limits.proposal_slicing_strategy`

#### Scenario: Delivery batching does not create review slices
- **GIVEN** `workflow_limits.delivery_batch_strategy` is available
- **WHEN** `sdd-router` creates or omits `review_slices`
- **THEN** it SHALL NOT use `workflow_limits.delivery_batch_strategy` as the source for proposal or review slicing decisions

#### Scenario: Delivery batching remains authorization-gated
- **WHEN** delivery batching guidance is present in a result or handoff
- **THEN** the workflow SHALL keep it separate from delivery authorization and SHALL NOT perform Git/GH operations without explicit user authorization

### Requirement: Chain strategy routing metadata
The routing harness SHALL treat `chain_strategy` as optional routing metadata that can be propagated without requiring a CLI struct change in this slice.

#### Scenario: Top-level auto-chain is propagated
- **GIVEN** `.opencode/project.yaml` defines top-level `chain_strategy: auto-chain`
- **WHEN** `sdd-router` classifies a request and risk is not high
- **THEN** it SHALL report the chain strategy and `orchestrator` SHALL propagate it to the next handoff without asking the user again

#### Scenario: Routing nested chain strategy is propagated
- **GIVEN** `.opencode/project.yaml` defines `workflow_limits.routing.chain_strategy: auto-chain`
- **WHEN** top-level `chain_strategy` is absent and `sdd-router` classifies a request
- **THEN** it SHALL report the nested chain strategy and `orchestrator` SHALL propagate it when no high-risk or authorization gate applies

#### Scenario: Missing chain strategy is explicit
- **GIVEN** neither top-level `chain_strategy` nor `workflow_limits.routing.chain_strategy` exists
- **WHEN** `sdd-router` reports workflow decision fields
- **THEN** it SHALL report chain strategy as `not_available` and SHALL NOT invent auto-chain behavior

#### Scenario: Auto-chain stops for high risk or authorization
- **GIVEN** `chain_strategy: auto-chain` is available
- **WHEN** a request triggers high risk, delivery, Git/GH work, protected branch policy, missing required information, or a configured stop rule
- **THEN** `orchestrator` SHALL pause for the appropriate role or explicit user authorization instead of chaining silently
