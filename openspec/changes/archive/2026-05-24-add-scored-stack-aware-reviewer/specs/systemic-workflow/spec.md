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
