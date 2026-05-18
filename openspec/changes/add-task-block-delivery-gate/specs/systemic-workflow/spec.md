## ADDED Requirements

### Requirement: Block-scoped proportional validation
The workflow SHALL run validation/testing proportionally at the end of a task, coherent block, proposal block, or review slice, and SHALL avoid constant test loops for individual micro-checkboxes unless an exception gate applies.

#### Scenario: Validation waits for coherent block boundary
- **WHEN** an agent completes an internal micro-step that is part of a larger coherent task or block
- **THEN** the workflow SHALL NOT require full validation/testing for that micro-step and SHALL defer grouped validation to the coherent block boundary

#### Scenario: Validation runs before validated state
- **WHEN** a task, coherent block, proposal block, or review slice is ready to move from `implemented` to `validated`
- **THEN** the workflow SHALL run the real applicable validation commands or document proportional static/manual evidence and SHALL report exact evidence before using the `validated` state

#### Scenario: Exception allows early validation
- **WHEN** a blocker, risky change, feedback loop, or failure diagnosis requires earlier evidence
- **THEN** the workflow MAY run focused validation before the block boundary while preserving grouped final validation for the block when applicable
