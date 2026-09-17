## ADDED Requirements

### Requirement: Deterministic active surface resolution
Lufy SHALL resolve an execution surface from explicit selection, changed files and configured surface roots using deterministic precedence and explainable decisions.

#### Scenario: Specific root selects one surface
- **WHEN** all changed files belong to the most specific root of one configured leaf surface
- **THEN** the execution plan SHALL select that surface and record the matching evidence

#### Scenario: Connected surfaces compose a fullstack flow
- **WHEN** changed files affect connected frontend and backend surfaces
- **THEN** the execution plan SHALL select both leaves, activate their fullstack composition and report composed mode

#### Scenario: Ambiguity remains actionable
- **WHEN** multiple surfaces match and no declared composition resolves them
- **THEN** planning SHALL fail with an actionable error listing explicit surface choices

### Requirement: Typed validation plan
Lufy SHALL translate active surface expectations, stack commands, change signals and optional capabilities into deduplicated typed validation rules.

#### Scenario: Frontend UI change
- **WHEN** an active frontend surface contains UI file changes
- **THEN** the plan SHALL require frontend validation and browser evidence without executing any command

#### Scenario: Cross-surface contract change
- **WHEN** a changed file signals an API or shared contract and frontend and backend are connected
- **THEN** the plan SHALL include contract compatibility and end-to-end evidence in addition to surface-specific rules

#### Scenario: Contract changes elevate connected consumers
- **WHEN** an automatically resolved frontend or backend change modifies a declared API or shared contract
- **THEN** the execution plan SHALL activate the connected fullstack composition so producer and consumers are validated together

#### Scenario: Realtime persistent application
- **WHEN** a surface declares realtime, rendering, offline or persistent-state capabilities
- **THEN** the plan SHALL add capability-specific evidence without introducing game-specific branches in domain resolution

#### Scenario: Stack commands remain scoped to their surface
- **WHEN** a composed plan includes different toolchains for frontend and backend
- **THEN** each surface-specific validation rule SHALL only suggest commands from the stacks linked to that surface

### Requirement: Interactive capability detection
Lufy SHALL detect reusable interactive application capabilities from project evidence while preserving explicit user capabilities during rescan.

#### Scenario: Web game with desktop shell and persistence
- **WHEN** JavaScript project dependencies or files identify rendering/realtime, Tauri, offline and persistence technologies
- **THEN** `init` or `scan` SHALL add the corresponding generic capabilities to the detected surface and its fullstack composition

#### Scenario: Rescan preserves and enriches capabilities
- **WHEN** a configured surface contains manual capabilities and rescan detects additional capabilities
- **THEN** the merged project profile SHALL preserve both sets without duplicates

### Requirement: Read-only CLI contract
Lufy SHALL expose execution planning through a read-only `plan` command with equivalent human and JSON representations.

#### Scenario: JSON automation output
- **WHEN** `lufy-ai plan --json` succeeds
- **THEN** stdout SHALL contain a single `surface-execution-plan/v1` document with selected surfaces, decisions, contracts and validation rules

#### Scenario: Explicit surface override
- **WHEN** `--surface` identifies a configured surface ID or an unambiguous surface type
- **THEN** that explicit selection SHALL take precedence over automatic diff routing

#### Scenario: Planning does not mutate target
- **WHEN** any valid `lufy-ai plan` invocation runs
- **THEN** it SHALL only read project configuration and Git state and SHALL NOT write project files or run validation commands
