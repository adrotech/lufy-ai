# scored-stack-aware-reviewer Specification

## Purpose
TBD - created by archiving change add-scored-stack-aware-reviewer. Update Purpose after archive.
## Requirements
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
`reviewer` SHALL use `.lufy/config/project.yaml` when available to adapt anti-pattern, coverage and observability expectations to the affected stack.

#### Scenario: Project config exists
- **GIVEN** `.lufy/config/project.yaml` declares stack-specific anti-patterns, coverage thresholds or observability libraries
- **WHEN** reviewer evaluates a change affecting that stack
- **THEN** it uses those declarations instead of assuming Go, TypeScript, Python or any other fixed stack

#### Scenario: Project config missing
- **WHEN** `.lufy/config/project.yaml` is missing or lacks relevant stack fields
- **THEN** reviewer marks stack-specific guidance as `not_available` and does not invent project-specific requirements

### Requirement: Desk-check scenarios for substantive review
`reviewer` SHALL include at least eight desk-check scenarios for T1 and T2 changes where behavior, workflow, validation or release risk is substantive.

#### Scenario: T1 or T2 substantive review
- **WHEN** reviewer evaluates a T1 or T2 change with meaningful behavior or workflow impact
- **THEN** it includes at least eight named desk-check scenarios covering happy path, failure path, edge cases, validation and release risk

#### Scenario: T3 review
- **WHEN** reviewer evaluates a trivial T3 change
- **THEN** it may mark the eight-scenario desk-check as `not_applicable` with a concise reason
