## ADDED Requirements

### Requirement: Weighted reviewer scoring
The harness SHALL define `reviewer` output with weighted scoring across Architecture, Code Quality, Simplicity, Testing, Observability and PR Template gate.

#### Scenario: Reviewer emits weighted score
- **WHEN** `reviewer` evaluates a completed T1 or T2 implementation
- **THEN** it reports category scores, category weights, total score percentage and merge/readiness recommendation in Result Contract envelope v1

#### Scenario: Score below approval threshold
- **WHEN** the total reviewer score is below 80%
- **THEN** reviewer reports the result as not approval-ready and includes the categories responsible for the score loss

### Requirement: L1-L5 severity gate
`reviewer` SHALL classify findings using severities L1 through L5 and SHALL block approval when L1 or L2 findings exist.

#### Scenario: Blocking finding exists
- **WHEN** reviewer finds an L1 or L2 issue
- **THEN** it reports `status: blocked` or equivalent non-ready state and explains release impact plus next owner

#### Scenario: Non-blocking findings only
- **WHEN** reviewer finds only L3, L4 or L5 issues and score is at least 80%
- **THEN** it may report approval-ready with residual risks and follow-up recommendations

### Requirement: Stack-aware review inputs
`reviewer` SHALL use `.opencode/project.yaml` when available to adapt anti-pattern, coverage and observability expectations to the affected stack.

#### Scenario: Project config exists
- **GIVEN** `.opencode/project.yaml` declares stack-specific anti-patterns, coverage thresholds or observability libraries
- **WHEN** reviewer evaluates a change affecting that stack
- **THEN** it uses those declarations instead of assuming Go, TypeScript, Python or any other fixed stack

#### Scenario: Project config missing
- **WHEN** `.opencode/project.yaml` is missing or lacks relevant stack fields
- **THEN** reviewer marks stack-specific guidance as `not_available` and does not invent project-specific requirements

### Requirement: Desk-check scenarios for substantive review
`reviewer` SHALL include at least eight desk-check scenarios for T1 and T2 changes where behavior, workflow, validation or release risk is substantive.

#### Scenario: T1 or T2 substantive review
- **WHEN** reviewer evaluates a T1 or T2 change with meaningful behavior or workflow impact
- **THEN** it includes at least eight named desk-check scenarios covering happy path, failure path, edge cases, validation and release risk

#### Scenario: T3 review
- **WHEN** reviewer evaluates a trivial T3 change
- **THEN** it may mark the eight-scenario desk-check as `not_applicable` with a concise reason
*** Add File: /Users/adrianrojas/Desktop/projects/lufy-ai/openspec/changes/add-scored-stack-aware-reviewer/specs/systemic-workflow/spec.md
## ADDED Requirements

### Requirement: Weighted review gate for substantive changes
The workflow SHALL use weighted reviewer output as a quality gate for T1 and T2 changes that require independent review.

#### Scenario: Review gate passes
- **WHEN** a T1 or T2 change has reviewer score at least 80%, zero L1/L2 findings and proportional validation evidence
- **THEN** the workflow may treat review as approval-ready while still requiring explicit delivery authorization for Git/GH actions

#### Scenario: Review gate blocks
- **WHEN** reviewer score is below 80% or any L1/L2 finding exists
- **THEN** the workflow reports the block with findings, score, affected categories and next owner instead of advancing to delivery-ready

### Requirement: Review remains separate from validation and delivery
The workflow SHALL keep reviewer qualitative scoring separate from validator command evidence and delivery authorization.

#### Scenario: Reviewer lacks command evidence
- **WHEN** reviewer identifies missing tests or validation evidence
- **THEN** it escalates to `validator` for command evidence rather than claiming commands passed

#### Scenario: Review is approval-ready
- **WHEN** reviewer reports approval-ready
- **THEN** Git/GH delivery remains blocked until the user explicitly authorizes `delivery`
