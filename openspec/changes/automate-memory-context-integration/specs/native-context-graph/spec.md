## MODIFIED Requirements

### Requirement: Context CLI command suite

LUFY SHALL expose context graph operations through the Go CLI as `lufy-ai context scan/status/build/query/path/explain/diff` with bounded outputs designed to reduce broad file reads.

#### Scenario: Status reports graph availability
- **WHEN** `lufy-ai context status` runs in a workspace with no readable valid graph
- **THEN** it SHALL report `not_available` with a recovery hint such as `lufy-ai context build`

#### Scenario: Query returns deterministic matches
- **WHEN** `lufy-ai context query <term>` runs against a valid graph
- **THEN** it SHALL return ranked deterministic matches with node ids, labels, types, reasons, scores, bounded neighboring context and a token-savings summary

#### Scenario: Internal managed state is excluded by default
- **WHEN** `lufy-ai context build` discovers workspace files with default configuration
- **THEN** it SHALL exclude `.lufy/managed-state/backups/**` and `.lufy/managed-state/ancestors/**` from indexed sources and query results

#### Scenario: Project config can tune graph exclusions
- **WHEN** `.lufy/config/project.yaml` defines `context_graph.exclude`
- **THEN** context graph discovery SHALL apply those project-level exclusion patterns before extracting supported files
