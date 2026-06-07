## MODIFIED Requirements

### Requirement: Project config implementation boundaries
La implementacion de `.lufy/project.yaml` SHALL keep model, scanning, rescan merge, persistence and CLI prompting responsibilities separated enough to preserve SOLID boundaries while keeping the public YAML schema stable.

#### Scenario: Service orchestrates without owning detector details
- **WHEN** `lufy-ai init`, `lufy-ai init --rescan` or `lufy-ai scan` builds a project config
- **THEN** the application service SHALL coordinate scanning, merge, optional profile prompting and persistence without embedding stack-specific detector logic in the service method

#### Scenario: Detectors are independently extensible
- **WHEN** a future stack or surface detector is added
- **THEN** it SHALL be possible to add it through a detector strategy or registry without rewriting unrelated detector implementations

#### Scenario: Public YAML remains compatible
- **WHEN** projectconfig internals are refactored
- **THEN** `.lufy/project.yaml` schema version 1, field names, defaults and preserved unknown fields SHALL remain compatible with the current behavior
