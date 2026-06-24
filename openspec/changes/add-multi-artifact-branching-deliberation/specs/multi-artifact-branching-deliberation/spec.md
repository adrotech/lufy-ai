## ADDED Requirements

### Requirement: Multi-artifact branching protocol
The harness SHALL provide a bounded protocol for generating, comparing, joining and collapsing multiple artifact candidates before implementation.

#### Scenario: High-uncertainty work enters branching
- **WHEN** a T1 request or multi-risk T2 request has high uncertainty across scope, architecture, product outcome, validation strategy or review workload
- **THEN** the workflow MAY enter `branching_candidate_generation` for planning artifacts
- **AND** candidate_count SHALL NOT exceed 2

#### Scenario: Dominant low-risk solution avoids branching
- **WHEN** the router or orchestrator identifies one dominant low-risk solution with clear acceptance criteria and no meaningful unresolved trade-off
- **THEN** candidate_count SHALL be 1
- **AND** deliberation or artifact branching SHALL be reported as `not_needed`

### Requirement: Branching lifecycle states
The harness SHALL make multi-artifact deliberation stateful and SHALL NOT advance to implementation until a canonical artifact set exists.

#### Scenario: Candidate lifecycle reaches canonical readiness
- **WHEN** artifact branching is selected
- **THEN** the workflow SHALL progress through `routed`, `branching_candidate_generation`, `join/decision`, `canonical_artifact_ready` and `implementation-ready` before implementation begins

#### Scenario: Join is mandatory before downstream work
- **WHEN** two proposal, design or task candidates exist
- **THEN** `orchestrator` MUST complete a join or decision step before any downstream design, task or implementation step uses those artifacts

### Requirement: Proposal-first branching
The harness SHALL treat proposal branching as the primary MVP branching stage and SHALL collapse selected or merged proposals before deeper planning.

#### Scenario: Two proposals require join before design
- **WHEN** two proposal candidates exist
- **THEN** `orchestrator` MUST join, select or merge them into one canonical `proposal.md` and aligned spec intent before design, tasks or implementation proceeds

#### Scenario: Proposal selected before optional design branching
- **WHEN** a proposal has been selected or merged
- **AND** substantial technical decisions still remain unresolved
- **THEN** `orchestrator` MAY request up to 2 design candidates with isolated artifacts and an explicit merge plan

### Requirement: Rare tasks branching
The harness SHALL collapse tasks to a single canonical implementation plan unless a documented implementation-strategy risk justifies explicit task candidates.

#### Scenario: Design join collapses tasks to one plan
- **WHEN** design has been selected or merged
- **THEN** tasks SHALL collapse to one canonical `tasks.md` plan by default

#### Scenario: Explicit implementation strategy risk permits task candidates
- **WHEN** an explicit implementation-strategy risk remains after design join
- **THEN** `orchestrator` MAY request up to 2 task candidates
- **AND** the handoff MUST record why one canonical task plan is not yet safe

### Requirement: Isolated parallel candidates
The harness SHALL isolate candidate artifacts and require merge plans when candidates are generated in parallel.

#### Scenario: Parallel candidates write isolated artifacts
- **WHEN** candidates are generated in parallel
- **THEN** each candidate MUST write isolated artifacts that do not overwrite the canonical artifact set or another candidate
- **AND** each candidate MUST include a merge plan describing assumptions, reusable decisions, risks and expected join inputs

#### Scenario: Parallel validation waits for join
- **WHEN** candidate generation completes
- **THEN** final validation SHALL be grouped after the join against the canonical artifact set rather than treating individual candidate checks as completion evidence

### Requirement: Canonical downstream artifact set
The harness SHALL require all downstream implementation to use a single canonical artifact set after deliberation.

#### Scenario: Join produces one canonical OpenSpec change
- **WHEN** join completes
- **THEN** downstream implementation MUST use one canonical OpenSpec change and artifact set
- **AND** non-selected candidate artifacts SHALL be treated as supporting deliberation evidence, not implementation source of truth

#### Scenario: Missing canonical artifact blocks implementation
- **WHEN** candidates exist but no canonical artifact set has been selected or merged
- **THEN** implementation SHALL be blocked
- **AND** `orchestrator` SHALL route to join, reviewer comparison or human decision before continuing

### Requirement: Human escalation for non-objective trade-offs
The harness SHALL escalate unresolved non-objective candidate differences to the human instead of inventing a decision.

#### Scenario: Public contract security or product differences escalate
- **WHEN** candidates differ on public contract, security posture, product behavior, significant UX, irreversible migration or another trade-off not objectively decided by existing criteria
- **THEN** `orchestrator` MUST escalate to the human with the candidate comparison and recommended decision points

#### Scenario: Objective quality comparison can proceed
- **WHEN** candidates differ only in objective quality, completeness, risk reduction or validation clarity under existing criteria
- **THEN** `reviewer` MAY compare them and `orchestrator` MAY select or merge based on the documented evidence
