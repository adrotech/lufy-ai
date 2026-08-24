## MODIFIED Requirements

### Requirement: T2 SDD Lite artifact
The system SHALL define T2 SDD Lite as a compact professional artifact containing intent, current behavior, target behavior, scope, acceptance criteria, tasks, validation, and risks.

#### Scenario: T2 produces verifiable criteria
- **WHEN** a request is classified as T2
- **THEN** the workflow SHALL produce or maintain acceptance criteria with observable WHEN and THEN outcomes before implementation completes

#### Scenario: T2 captures LLM execution contract
- **WHEN** a request uses `lufy-sdd/lite`
- **THEN** the Lite artifact SHALL capture the LLM objective, bounded scope, explicit constraints, relevant environment assumptions, acceptance criteria, validation commands or static checks, and risks before implementation completes
- **AND** it SHALL remain compact enough to fit in `proposal.md` and `tasks.md` without requiring `design.md` or spec deltas

#### Scenario: T2 includes compact diagrams when useful
- **WHEN** a T2 change has a non-obvious workflow, component interaction, external integration, or data touchpoint
- **THEN** the Lite artifact SHOULD include a compact Mermaid diagram or structured flow that reduces ambiguity
- **AND** it SHALL NOT require database or architecture diagrams for purely local mechanical changes

#### Scenario: T2 surfaces optional overview outcome
- **WHEN** a T2 SDD Lite specification or structured handoff is ready
- **THEN** the workflow SHALL surface the optional overview/render outcome when the selected methodology and tool adapter provide one, or record `not_available` when no render surface exists

#### Scenario: T2 escalates to T1
- **WHEN** T2 exploration or implementation reveals cross-cutting impact, unresolved architecture trade-offs, or high risk
- **THEN** the workflow SHALL recommend escalation to T1 Full SDD

### Requirement: Lufy SDD Full artifact depth
The system SHALL define Lufy SDD Full as a prescriptive methodology artifact that guides LLM planning, architecture, implementation boundaries, validation and delivery for T1 work.

#### Scenario: Full captures LLM goals and constraints
- **WHEN** a request uses `lufy-sdd/full`
- **THEN** the Full artifacts SHALL capture measurable LLM objectives, desired outcomes, non-goals, explicit development restrictions, affected surfaces, environment assumptions, and required validation evidence

#### Scenario: Full captures architecture and patterns
- **WHEN** a request uses `lufy-sdd/full`
- **THEN** `design.md` SHALL include sections for architecture, component responsibilities, design patterns, data or persistence impact when applicable, security and privacy, operational concerns, tradeoffs, and migration or rollback notes when relevant

#### Scenario: Full includes diagrams as source artifacts
- **WHEN** a request uses `lufy-sdd/full`
- **THEN** `design.md` SHALL provide Mermaid placeholders or completed diagrams for workflow flow, component relationships, architecture context, and data model or persistence when applicable
- **AND** diagrams SHALL live in Markdown source artifacts rather than in the derived `change-overview.html`

#### Scenario: Full tasks express coherent gates
- **WHEN** a request uses `lufy-sdd/full`
- **THEN** `tasks.md` SHALL organize work into coherent blocks for analysis, design, implementation, tests, validation, sync, delivery readiness, and documentation
- **AND** micro-checkboxes SHALL NOT be treated as closure without block-level evidence
