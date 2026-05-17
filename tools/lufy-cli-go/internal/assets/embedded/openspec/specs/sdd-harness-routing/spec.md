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
