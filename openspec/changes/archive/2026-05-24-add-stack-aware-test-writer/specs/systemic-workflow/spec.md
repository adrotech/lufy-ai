## ADDED Requirements

### Requirement: TDD delegation for substantive T1 and T2 changes
The workflow SHALL route substantive test design or test implementation for T1 and T2 changes through `test-writer` when a TDD cycle is applicable.

#### Scenario: Implementer delegates test work
- **WHEN** `implementer` is assigned a T1 or T2 change with substantive test creation or revision needs
- **THEN** `implementer` delegates the test-focused portion to `test-writer` or records why TDD delegation is not applicable in the Result Contract envelope

#### Scenario: T3 change does not require delegation
- **WHEN** a T3 Express change is trivial, mechanical or documentation-only and does not require substantive test behavior
- **THEN** the workflow does not require `test-writer` delegation and may record TDD evidence as `not_applicable`

### Requirement: Validator gates required TDD evidence
The workflow SHALL require validator review of TDD evidence for T1 and T2 changes where TDD delegation or equivalent TDD evidence is required.

#### Scenario: Required TDD evidence is present
- **WHEN** `validator` evaluates a T1 or T2 change that required TDD evidence
- **THEN** it verifies that RED, GREEN, TRIANGULATE and REFACTOR evidence is present or explicitly marked `not_applicable` with reasons before reporting the block as `validated`

#### Scenario: Required TDD evidence is missing
- **WHEN** `validator` evaluates a T1 or T2 change that required TDD evidence but the evidence is absent or incomplete
- **THEN** it reports `blocked` or `escalated` with the missing evidence and next owner instead of reporting `validated`
