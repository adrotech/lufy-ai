## ADDED Requirements

### Requirement: OpenSpec task state semantics
The installed OpenSpec workflow SHALL distinguish task checkbox completion from operational closure, archive readiness, and delivery completion.

#### Scenario: Tasks complete is not archive-ready by itself
- **WHEN** every checkbox in `tasks.md` is marked complete
- **THEN** `/opsx-verify` and `/opsx-archive` SHALL still evaluate required validation, sync, delivery, and blocker evidence before declaring the change archive-ready or closed

#### Scenario: Delivery pending blocks closure
- **WHEN** a change has implementation and validation evidence but requires commit, push, PR, issue update, project sync, or other Git/GH delivery that has not been explicitly authorized or completed
- **THEN** the OpenSpec workflow SHALL report `delivery_pending`, `sync_pending`, or `blocked` instead of `closed` or archive-ready

#### Scenario: Archive rejects unresolved closure gates
- **WHEN** `/opsx-archive` is invoked for a change with unresolved delivery, sync, validation, or task/block gate state
- **THEN** the command SHALL refuse archive with an actionable recovery instruction and SHALL NOT treat checked tasks as sufficient confirmation

### Requirement: Opsx command gate language
The installed `/opsx-apply`, `/opsx-verify`, and `/opsx-archive` workflows SHALL use consistent gate language for task/block states and SHALL preserve role separation when recommending next actions.

#### Scenario: Apply reports implemented state
- **WHEN** `/opsx-apply` finishes an implementation block without final validation or delivery
- **THEN** it SHALL report `implemented` or pending validation and SHALL recommend `/opsx-verify` or validation by the appropriate role rather than reporting closure

#### Scenario: Verify reports delivery pending
- **WHEN** `/opsx-verify` validates a change that still needs authorized Git/GH delivery
- **THEN** it SHALL report `validated` with `delivery_pending` or `blocked` next action and SHALL NOT perform delivery itself

#### Scenario: Archive requires closed state
- **WHEN** `/opsx-archive` evaluates a completed change
- **THEN** it SHALL require evidence that the change is `closed` or explicitly does not need delivery/sync before applying deltas to archived state
