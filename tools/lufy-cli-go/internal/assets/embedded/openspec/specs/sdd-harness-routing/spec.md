## Purpose

Define the proportional SDD harness used to route development work through the smallest safe workflow, including tier classification, lightweight router contracts, SDD Lite artifacts, review workload slicing, local-first skill resolution, installer asset synchronization, and delivery policy separation.
## Requirements
### Requirement: Tier classification
The system SHALL classify development requests into T1 Full SDD, T2 SDD Lite, or T3 Express before selecting a workflow when the request is non-trivial or ambiguous.

#### Scenario: Complex feature is classified as T1
- **WHEN** a request introduces a new capability, cross-cutting behavior, architecture decision, public contract change, security concern, delivery policy change, or high uncertainty
- **THEN** the system SHALL classify the request as T1 Full SDD

#### Scenario: Bounded functional change is classified as T2
- **WHEN** a request changes behavior in a bounded area, fixes a relevant bug, updates an agent or skill, or performs a controlled refactor with medium risk
- **THEN** the system SHALL classify the request as T2 SDD Lite

#### Scenario: Trivial request is classified as T3
- **WHEN** a request is small, mechanical, documentary, local, and has no meaningful behavior risk
- **THEN** the system SHALL classify the request as T3 Express

### Requirement: Lightweight SDD router
The system SHALL provide a lightweight `sdd-router` subagent that classifies the request and recommends the minimum safe workflow before activating heavier subagents or OpenSpec flows.

#### Scenario: Router avoids heavy workflow for simple task
- **WHEN** a request is clearly T3
- **THEN** the `sdd-router` SHALL recommend Express workflow without OpenSpec and without broad exploration

#### Scenario: Router escalates uncertainty
- **WHEN** a request has insufficient information to classify safely
- **THEN** the `sdd-router` SHALL return a low-confidence classification with a recommended clarification or focused exploration step

### Requirement: Router output contract
The `sdd-router` SHALL return a structured output that includes tier, confidence, reason, execution mode, recommended workflow, required subagents, required permissions, context slice, review workload, skill status, and stop reason when blocked.

#### Scenario: Router emits actionable handoff
- **WHEN** the `sdd-router` completes classification
- **THEN** its output SHALL identify the next agent or action and the minimum context that action needs

#### Scenario: Router blocks unsafe routing
- **WHEN** the request requires authorization or missing information before proceeding
- **THEN** the `sdd-router` SHALL set a stop reason instead of recommending implementation

#### Scenario: Router declares execution mode
- **WHEN** the `sdd-router` recommends a workflow
- **THEN** its output SHALL declare an execution mode such as Full SDD, SDD Lite, Express, Clarify, Explore Only, Verify Only, or Delivery Pending

### Requirement: Result contract
The system SHALL define a minimal result contract for agent handoffs and final responses that includes objective, actions, evidence, risks, status, and recommended next action.

#### Scenario: Agent returns verifiable result
- **WHEN** a subagent completes a routed step
- **THEN** the result SHALL include evidence appropriate to the step and SHALL identify whether the workflow is ready, blocked, escalated, or pending validation

#### Scenario: Result supports context recovery
- **WHEN** a workflow resumes after context compaction or handoff
- **THEN** the result contract SHALL provide enough state to continue without replaying the full prior conversation

### Requirement: Proportional context
The system SHALL pass only the minimum relevant context to each subagent based on the selected tier and role.

#### Scenario: T2 passes focused context
- **WHEN** a request is classified as T2
- **THEN** the next subagent SHALL receive the user intent, classification reason, known scope, acceptance criteria draft, and focused files or questions if known

#### Scenario: T3 avoids unnecessary context
- **WHEN** a request is classified as T3
- **THEN** the next subagent SHALL NOT receive broad repository context unless needed to safely complete the change

### Requirement: Permission minimization
The system SHALL select the smallest permission set needed for the current workflow step.

#### Scenario: Router remains read-only
- **WHEN** `sdd-router` is invoked
- **THEN** it SHALL NOT edit files, run mutating shell commands, commit, push, create PRs, or invoke delivery operations

#### Scenario: Delivery remains explicitly authorized
- **WHEN** a workflow reaches a delivery step
- **THEN** Git/GH operations SHALL remain blocked unless the user explicitly authorized delivery

### Requirement: T2 SDD Lite artifact
The system SHALL define T2 SDD Lite as a compact professional artifact containing intent, current behavior, target behavior, scope, acceptance criteria, tasks, validation, and risks.

#### Scenario: T2 produces verifiable criteria
- **WHEN** a request is classified as T2
- **THEN** the workflow SHALL produce or maintain acceptance criteria with observable WHEN and THEN outcomes before implementation completes

#### Scenario: T2 escalates to T1
- **WHEN** T2 exploration or implementation reveals cross-cutting impact, unresolved architecture trade-offs, or high risk
- **THEN** the workflow SHALL recommend escalation to T1 Full SDD

### Requirement: Artifact store by tier
The system SHALL persist artifacts proportionally to the selected tier.

#### Scenario: T1 persists Full SDD artifacts
- **WHEN** a request is classified as T1
- **THEN** the workflow SHALL use OpenSpec artifacts as the source of truth before implementation

#### Scenario: T2 persists compact artifact when needed
- **WHEN** a request is classified as T2 and has behavior or risk that needs traceability
- **THEN** the workflow SHALL maintain a compact SDD Lite artifact or equivalent structured handoff before implementation completes

#### Scenario: T3 avoids unnecessary artifacts
- **WHEN** a request is classified as T3 and no durable decision or behavior contract is needed
- **THEN** the workflow SHALL NOT require creating persistent OpenSpec artifacts

### Requirement: Tier workflow mapping
The system SHALL map each tier to a proportional workflow.

#### Scenario: T1 uses Full SDD
- **WHEN** a request is classified as T1
- **THEN** the recommended workflow SHALL include OpenSpec proposal, design/spec/tasks when applicable, implementation, verification, and optional review/delivery gates

#### Scenario: T2 uses SDD Lite
- **WHEN** a request is classified as T2
- **THEN** the recommended workflow SHALL include focused exploration when needed, compact specification, implementation, and grouped validation

#### Scenario: T3 uses Express
- **WHEN** a request is classified as T3
- **THEN** the recommended workflow SHALL allow direct bounded implementation and proportional validation without mandatory OpenSpec artifacts

### Requirement: Skill resolution and optional bootstrap
The system SHALL resolve local project skills before recommending external skill bootstrap, and SHALL treat external bootstrap as optional and user-authorized.

#### Scenario: Local skills are sufficient
- **WHEN** local `.opencode/skills` cover the requested workflow or stack
- **THEN** the `sdd-router` SHALL prefer local skills and SHALL NOT recommend external bootstrap

#### Scenario: Skill coverage is missing
- **WHEN** local skills are missing or insufficient for the detected stack or workflow
- **THEN** the `sdd-router` MAY recommend a dry-run bootstrap command such as `npx autoskills --dry-run` and SHALL require explicit user authorization before any mutating install command

#### Scenario: External bootstrap remains subordinate
- **WHEN** external skills are installed or proposed
- **THEN** `AGENTS.md`, repository policies, and local `.opencode/skills` SHALL remain higher priority than externally bootstrapped skills

### Requirement: Subagent isolation and review workload
The system SHALL isolate subagent context and size review workload according to risk, tier, and role.

#### Scenario: Subagent receives role-scoped context
- **WHEN** the router delegates to a subagent
- **THEN** the handoff SHALL include only the context needed for that role and SHALL omit unrelated conversation history or broad repository dumps

#### Scenario: Review workload is proportional
- **WHEN** the router recommends review
- **THEN** it SHALL classify review workload as none, focused, or full based on tier, changed surface, and risk

#### Scenario: Reviewer receives bounded slices
- **WHEN** a T1 request or a multi-risk T2 request is routed
- **THEN** the router SHOULD recommend review slices that identify objective, expected files, acceptance criteria, validation, risk, and PR split guidance when useful

#### Scenario: Trivial work is not fragmented
- **WHEN** a request is classified as T3 Express
- **THEN** the workflow SHALL NOT require review slices or multiple PR recommendations unless new risk appears

### Requirement: Review Workload Harness
The system SHALL help authors shape features and proposals into reviewer-friendly deliverables when doing so reduces cognitive load, risk, or PR size.

#### Scenario: Feature is split into reviewable subproblems
- **WHEN** a feature contains separable subproblems, independent risk areas, or multiple validation concerns
- **THEN** the workflow SHALL recommend small reviewable slices with their own objective, acceptance criteria, validation evidence, and delivery boundary

#### Scenario: Slices remain proportional
- **WHEN** a change can be reviewed safely as one small unit
- **THEN** the workflow SHALL keep it as one deliverable instead of forcing artificial micro-slices

#### Scenario: PR guidance remains advisory
- **WHEN** review slices are produced
- **THEN** the workflow SHALL treat PR split guidance as advisory until delivery is explicitly authorized

### Requirement: Documentation and installer asset synchronization
The system SHALL keep public documentation, local OpenCode assets, and embedded installer assets synchronized when harness capabilities change.

#### Scenario: Documentation reflects installed harness
- **WHEN** a harness capability such as `sdd-router`, SDD Lite templates, result contracts, or skill bootstrap guidance is added
- **THEN** user-facing documentation and operational documentation SHALL describe the installed behavior without presenting roadmap-only features as current capabilities

#### Scenario: Installer embeds current assets
- **WHEN** local managed assets change under `.opencode`, `AGENTS.md.template`, `openspec`, or templates included in the catalog
- **THEN** `tools/lufy-cli-go/internal/assets/embedded` and the installer catalog SHALL be updated so a standalone binary installs the current harness

#### Scenario: Local installer is generated for validation
- **WHEN** embedded assets or installer catalog entries change
- **THEN** the workflow SHALL build a local `tools/lufy-cli-go/bin/lufy-ai` binary and run grouped validation before reporting readiness

### Requirement: Delivery policy separation
The system SHALL keep shared delivery invariants in `.opencode/policies/delivery.md` and delivery-agent execution details in `.opencode/agents/delivery.md`.

#### Scenario: Shared policy remains reusable
- **WHEN** orchestrator, validator, reviewer, implementer, or delivery need branch, validation, release, or authorization rules
- **THEN** they SHALL treat `.opencode/policies/delivery.md` as the canonical shared policy

#### Scenario: Delivery agent remains operational
- **WHEN** the `delivery` subagent performs authorized Git/GH delivery work
- **THEN** it SHALL use its own agent definition as the operational runbook while enforcing the shared delivery policy

### Requirement: Task/block delivery gate
The harness SHALL evaluate task completion at the level of a task, coherent block, or review slice, and SHALL NOT treat individual micro-checkboxes as independent closure boundaries unless they are explicitly declared as the coherent delivery unit.

#### Scenario: Micro-checkbox does not trigger closure
- **GIVEN** a `tasks.md` item contains nested implementation micro-checkboxes
- **WHEN** one nested micro-checkbox is completed
- **THEN** the harness SHALL keep the parent task or block open until the coherent block gate has implementation, proportional validation, and required delivery state evidence

#### Scenario: Coherent block reaches gate
- **WHEN** all implementation work for a task, coherent block, or review slice is finished
- **THEN** the harness SHALL require a state handoff that distinguishes implementation evidence, validation evidence, and delivery status before reporting the unit as closed

### Requirement: Explicit task/block states
The harness SHALL use explicit task/block states that distinguish `implemented`, `validated`, `delivery_pending`, `delivered`, and `closed` or documented equivalents with the same semantics.

#### Scenario: Implementation is not closure
- **WHEN** `implementer` finishes code, documentation, configuration, or proposal edits for a block
- **THEN** the result SHALL report `implemented` or pending validation, not `closed`, unless validation and required delivery evidence already exist from the correct roles

#### Scenario: Validation is not delivery authorization
- **WHEN** validation evidence passes for a block but Git/GH delivery has not been explicitly authorized
- **THEN** the workflow SHALL report `delivery_pending` or `blocked` and SHALL NOT report the block as `closed`

#### Scenario: Delivery completes authorized Git/GH work
- **WHEN** the user explicitly authorizes delivery and `delivery` completes the required commit, push, PR, or external sync for the block
- **THEN** the workflow SHALL report `delivered` and MAY report `closed` only when no required implementation, validation, sync, or archive precondition remains

### Requirement: Role-separated gate execution
The harness SHALL keep gate responsibilities separated by role: `implementer` implements bounded changes, `validator` validates and diagnoses read-only, `delivery` performs authorized Git/GH and external sync, and `orchestrator` coordinates state transitions.

#### Scenario: Implementer stops before delivery
- **WHEN** `implementer` completes a task/block and delivery would be required to close it
- **THEN** `implementer` SHALL report readiness and the required next role instead of committing, pushing, creating PRs, or updating GitHub Projects

#### Scenario: Validator remains read-only
- **WHEN** `validator` verifies a task/block gate
- **THEN** `validator` SHALL provide validation evidence and next-state recommendation without editing files, committing, pushing, creating PRs, or updating GitHub Projects

#### Scenario: Orchestrator routes delivery pending work
- **WHEN** a block is validated but not delivered
- **THEN** `orchestrator` SHALL identify the state as `delivery_pending` or `blocked` and request explicit user authorization before routing to `delivery`

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
The router and orchestrator SHALL expose workflow-limit-driven decisions as structured output derived from `.lufy/project.yaml` top-level `workflow_limits` when that file is available.

#### Scenario: Router reports workload decision inputs
- **WHEN** `sdd-router` evaluates a non-trivial request with `.lufy/project.yaml` available
- **THEN** it reports the `workflow_limits` paths considered, estimated workload inputs, tier decision, confidence, and whether `workload_decision_needed` is true

#### Scenario: Router proposes review slices from configured slicing limits
- **WHEN** estimated scope, file count, risk or configured routing limits require splitting before implementation or review
- **THEN** `sdd-router` uses `workflow_limits.proposal_slicing_strategy` to propose `review_slices` with objective, expected files, acceptance criteria, validation, risk and PR guidance

#### Scenario: Orchestrator carries workflow decisions forward
- **WHEN** orchestrator delegates to another agent after routing
- **THEN** the handoff includes the workflow decision fields needed by that role and does not require the receiving agent to rediscover the same limits from conversation history

### Requirement: Planning-only fast path
The routing harness SHALL distinguish a broad program tier from the tier of the next micro-slice and SHALL allow a lightweight path for bounded planning-only or OpenSpec/docs-only work.

#### Scenario: T1 program has T2 or T3 planning slice
- **GIVEN** the broader program is T1
- **WHEN** the next micro-slice touches only 1-2 OpenSpec/docs artifacts, has no runtime/app file changes, no delivery request, no security impact and no public-contract impact
- **THEN** `sdd-router` SHALL classify the slice as T2 or T3 and report `program_tier: T1`, `slice_tier: T2 | T3` and `fast_path_allowed: true`

#### Scenario: Prior context is sufficient
- **WHEN** the prompt or previous handoff identifies the affected artifacts, task and acceptance criteria for a planning-only slice
- **THEN** `orchestrator` SHALL NOT launch an additional `explorer` only to formalize the same handoff
- **AND** it MAY route directly to `implementer` with the bounded context slice

#### Scenario: Lightweight OpenSpec-only validation
- **WHEN** a fast-path slice modifies only OpenSpec/docs artifacts and no delivery is requested
- **THEN** validation SHALL default to `openspec validate "<change>" --strict` when a change ID exists plus static checkbox/file review
- **AND** dirty worktree state SHALL be treated as a delivery risk rather than a documentation-validation blocker unless there is concrete evidence of mixed runtime changes

#### Scenario: Fast path is not allowed
- **WHEN** the slice changes runtime/app files, affects security or public contracts, requires delivery, touches more than two artifacts, or has unclear acceptance criteria
- **THEN** `fast_path_allowed` SHALL be false and the workflow SHALL use the proportional T1/T2/T3 routing path for the actual risk

### Requirement: Delivery batching remains authorization-gated
The workflow SHALL report delivery batching guidance separately from delivery authorization.

#### Scenario: Delivery batching guidance is advisory
- **WHEN** validated work has delivery grouping guidance from `workflow_limits.delivery_batch_strategy`
- **THEN** the result contract reports the recommended grouping but keeps delivery state as `delivery_pending` until the user explicitly authorizes Git/GH delivery

#### Scenario: Delivery role receives batching context
- **WHEN** delivery is explicitly authorized
- **THEN** the delivery role receives the relevant batching, preflight and stop-rule context from the Result Contract envelope or current `.lufy/project.yaml`

### Requirement: Numeric workload guard
The routing harness SHALL make workload decisions observable from estimated LOC and file count using canonical `workflow_limits` when available.

#### Scenario: LOC budget requires workload decision
- **GIVEN** `.lufy/project.yaml` exists and defines `workflow_limits.sizing.loc_budget`
- **WHEN** `sdd-router` estimates `estimated_loc` greater than `workflow_limits.sizing.loc_budget`
- **THEN** it SHALL emit `workload_decision_needed: true` and recommend a workload decision before implementation continues

#### Scenario: Five or more files trigger escalation or slicing
- **WHEN** `sdd-router` estimates `estimated_files >= 5`
- **THEN** it SHALL either escalate the tier or propose bounded slices appropriate to the risk and scope

#### Scenario: Missing sizing config is not invented
- **GIVEN** `.lufy/project.yaml` is missing or `workflow_limits.sizing.loc_budget` is not available
- **WHEN** `sdd-router` evaluates estimated workload
- **THEN** it SHALL report the sizing source as `not_available` and SHALL NOT use legacy top-level `loc_budget` or invented defaults

### Requirement: Canonical workflow limits propagation
The router and orchestrator SHALL read and propagate workflow-limit decisions from `.lufy/project.yaml` top-level `workflow_limits` paths when available, and SHALL report unavailable paths explicitly.

#### Scenario: Router reports all relevant workflow limit paths
- **WHEN** `sdd-router` evaluates a non-trivial request for a project
- **THEN** its output SHALL report availability for `workflow_limits.sizing`, `workflow_limits.routing`, `workflow_limits.proposal_slicing_strategy`, `workflow_limits.delivery_batch_strategy`, `workflow_limits.preflight`, and `workflow_limits.stop_rules`

#### Scenario: Orchestrator preserves routing decision
- **WHEN** `orchestrator` delegates after `sdd-router` classified a request
- **THEN** it SHALL propagate the workflow decision fields, source paths, workload decision, review slices, preflight status, stop-rule status, and delivery batching guidance needed by the receiving role

#### Scenario: Legacy top-level fields are ignored
- **GIVEN** `.lufy/project.yaml` contains top-level `loc_budget` or top-level `delivery_strategy`
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
- **GIVEN** `.lufy/project.yaml` defines top-level `chain_strategy: auto-chain`
- **WHEN** `sdd-router` classifies a request and risk is not high
- **THEN** it SHALL report the chain strategy and `orchestrator` SHALL propagate it to the next handoff without asking the user again

#### Scenario: Routing nested chain strategy is propagated
- **GIVEN** `.lufy/project.yaml` defines `workflow_limits.routing.chain_strategy: auto-chain`
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
