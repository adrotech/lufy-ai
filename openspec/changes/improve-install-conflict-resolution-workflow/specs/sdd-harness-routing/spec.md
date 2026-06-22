## ADDED Requirements

### Requirement: Parallel review groups from conflict plans

The harness SHALL allow conflict plan groups to become parallel review tasks when they are independent and a merge/join plan exists.

#### Scenario: Independent groups can run in parallel
- **WHEN** a conflict plan contains independent categories touching disjoint files
- **THEN** the orchestrator may request multiple specialist reviews in parallel, one per category, with grouped validation after the results join

#### Scenario: Shared files block parallelism
- **WHEN** conflict groups touch the same files, shared public contracts, delivery operations, or unresolved API decisions
- **THEN** the harness keeps execution sequential and reports the blocking reason
