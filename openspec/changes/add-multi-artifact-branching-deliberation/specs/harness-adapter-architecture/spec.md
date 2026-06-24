## ADDED Requirements

### Requirement: Adapter-neutral artifact branching
The harness core SHALL define artifact branching semantics independently of any specific tool or methodology adapter.

#### Scenario: Core branching avoids adapter-specific paths
- **WHEN** the core model defines artifact branching, candidate isolation, join states or canonical artifact readiness
- **THEN** it SHALL NOT require OpenCode-specific paths, OpenSpec-only commands or any single tool surface as the universal implementation mechanism

#### Scenario: Adapter renders compatible artifacts
- **WHEN** a methodology or tool adapter supports artifact branching
- **THEN** it SHALL render candidate artifacts, merge plans and canonical artifact output using that adapter's supported artifact model while preserving the core states and gates

### Requirement: Adapter fallback preserves join semantics
Adapters without native parallelism or subagent isolation SHALL preserve the same branching gates through sequential or inline execution.

#### Scenario: Adapter lacks parallel candidate execution
- **WHEN** the effective adapter cannot safely run isolated candidates in parallel
- **THEN** the workflow MAY generate candidates sequentially
- **AND** it MUST still require isolated candidate artifacts, merge plan and join before canonical readiness

#### Scenario: Adapter cannot support safe isolation
- **WHEN** the effective adapter cannot provide isolated candidate artifacts or equivalent non-overwriting storage
- **THEN** `orchestrator` SHALL disable artifact branching for that adapter execution
- **AND** it SHALL report the adapter limitation and proceed with candidate_count 1 or escalate for human direction
