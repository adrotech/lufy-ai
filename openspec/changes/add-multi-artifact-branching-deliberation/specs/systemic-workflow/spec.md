## ADDED Requirements

### Requirement: Branching state progression
The systemic workflow SHALL model artifact deliberation as explicit states from routing through canonical readiness.

#### Scenario: Branching states are observable
- **WHEN** a workflow enters multi-artifact deliberation
- **THEN** its state progression SHALL be observable as `routed`, `branching_candidate_generation`, `join/decision`, `canonical_artifact_ready` and `implementation-ready`

#### Scenario: State handoff captures evidence
- **WHEN** a branching state completes
- **THEN** the Result Contract or equivalent handoff SHALL capture the candidates generated, comparison or join evidence, remaining risks and next owner/action

### Requirement: Join before implementation
The systemic workflow SHALL require a join or decision step before implementation consumes artifacts that were branched.

#### Scenario: Canonical artifacts gate implementation
- **WHEN** a workflow has more than one candidate for proposal, design or tasks
- **THEN** implementation SHALL NOT start until `orchestrator` marks one canonical artifact set as selected or merged

#### Scenario: Downstream roles receive one source of truth
- **WHEN** the join completes
- **THEN** downstream `implementer`, `validator`, `reviewer` and `delivery` handoffs SHALL reference one canonical artifact set
- **AND** candidate artifacts SHALL remain context only unless explicitly promoted by the join

### Requirement: Human decision escalation in deliberation
The systemic workflow SHALL pause for human decision when candidate differences exceed objective quality or risk criteria.

#### Scenario: Non-objective trade-off pauses workflow
- **WHEN** candidate artifacts differ in public contract, security, product direction, significant UX or another non-objective trade-off
- **THEN** `orchestrator` MUST escalate to the human before selecting or merging the canonical artifact

#### Scenario: Clear objective comparison does not require human escalation
- **WHEN** candidates can be compared using already agreed objective criteria such as completeness, risk, validation clarity and coherence with specs
- **THEN** `reviewer` MAY provide a comparison and `orchestrator` MAY complete the join without additional human decision
