## ADDED Requirements

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
