## Purpose
Definir el workflow OpenSpec core v2 instalado por `lufy-ai`, incluyendo acciones explícitas, baseline local, specs delta, scenarios testables y sincronización de deltas antes de archivar cambios.

## Requirements
### Requirement: OpenSpec config uses action-based core v2 schema
The installed OpenSpec configuration SHALL declare core workflow actions explicitly instead of relying on implicit command conventions.

#### Scenario: Core actions are declared
- **WHEN** a target is installed or synced with the OpenSpec core v2 assets
- **THEN** `openspec/config.yaml` declares actions for explore, propose, apply, verify, sync and archive workflows

#### Scenario: Action schema is machine-readable
- **WHEN** an agent or command reads `openspec/config.yaml`
- **THEN** it can resolve the action name, description and expected artifacts without parsing human documentation

### Requirement: Change specs use delta markers
OpenSpec change specs SHALL use explicit delta sections to distinguish new, modified and removed requirements from main specs.

#### Scenario: Added requirement uses ADDED marker
- **WHEN** a change introduces a new capability requirement
- **THEN** the change spec places it under `## ADDED Requirements`

#### Scenario: Modified requirement uses MODIFIED marker
- **WHEN** a change updates an existing requirement
- **THEN** the change spec places the complete updated requirement under `## MODIFIED Requirements`

#### Scenario: Removed requirement uses REMOVED marker
- **WHEN** a change removes an existing requirement
- **THEN** the change spec places it under `## REMOVED Requirements` with reason and migration guidance

### Requirement: Scenarios are testable
OpenSpec requirements SHALL include scenarios that can be reviewed or tested through clear conditions and expected outcomes.

#### Scenario: Scenario uses explicit condition and outcome
- **WHEN** a requirement is added or modified in a change spec
- **THEN** it includes at least one `#### Scenario:` with `WHEN` and `THEN` clauses

#### Scenario: GIVEN is supported for setup context
- **WHEN** a scenario requires preconditions to be unambiguous
- **THEN** it may include a `GIVEN` clause before `WHEN` and `THEN`

### Requirement: Opsx sync applies validated deltas
The installed workflow SHALL provide `/opsx-sync` and `openspec-sync` to apply validated change deltas to main specs before archive.

#### Scenario: Sync applies delta specs
- **WHEN** a change contains valid delta specs and the user runs `/opsx-sync`
- **THEN** the workflow applies those deltas to the corresponding main specs without archiving the change

#### Scenario: Sync rejects invalid deltas
- **WHEN** a change spec lacks required delta markers or testable scenarios
- **THEN** `/opsx-sync` fails with an actionable error and does not mutate main specs

### Requirement: Active post-spec verification
The installed workflow SHALL actively verify expected OpenSpec artifacts after proposal generation and sync operations before moving to implementation, verification, archive, or delivery.

#### Scenario: Proposal generation verifies expected files
- **WHEN** `/opsx-propose` or `openspec-propose` creates a change
- **THEN** the workflow reads the expected files under `openspec/changes/<change>/`, verifies required artifacts are non-empty, verifies delta markers and testable scenarios, and stops with an actionable blocked result if any expected file is missing or invalid

#### Scenario: Proposal generation surfaces optional HTML overview
- **WHEN** the OpenSpec propose action completes successfully through any tool adapter
- **THEN** the harness workflow surfaces the optional HTML overview command `lufy-ai opsx render --change <change> --format html --theme notion-dark`, asks `¿Quieres que genere ahora el reporte HTML offline de los artifacts con tema Notion dark?`, and records `offered_pending` while the user has not answered

#### Scenario: Tool adapters preserve proposal overview outcome
- **WHEN** a tool adapter summarizes a completed OpenSpec propose action from a command, skill, subagent, or other methodology adapter
- **THEN** the adapter preserves the optional HTML overview prompt/outcome unless the proposal is `blocked`

#### Scenario: Proposal overview outcome states are explicit
- **WHEN** a successful proposal response reports the optional HTML overview outcome
- **THEN** the outcome SHALL be one of `generated`, `offered_pending`, `skipped_by_user` or `not_available`
- **AND** `skipped_by_user` SHALL only be used after an explicit user decline
- **AND** bare `skipped` SHALL NOT be used for a pending or unanswered offer except when documenting legacy output

#### Scenario: Sync verifies target specs
- **WHEN** `/opsx-sync` or `openspec-sync` applies deltas to main specs
- **THEN** the workflow reads every affected `openspec/specs/<capability>/spec.md` and verifies the added, modified or removed requirement titles match the planned delta before recommending verify or archive

#### Scenario: Failed post-spec verification blocks continuation
- **WHEN** active post-spec verification cannot find an expected artifact or synced requirement
- **THEN** the workflow reports `status: blocked`, names the missing path or requirement, and does not continue to apply, verify, archive or delivery guidance until the issue is repaired

### Requirement: Baseline metadata is installed
The installed OpenSpec workflow SHALL include `UPSTREAM.json` describing the effective baseline used by the local workflow assets.

#### Scenario: Baseline file exists
- **WHEN** a target is installed or synced with OpenSpec core v2 assets
- **THEN** `openspec/UPSTREAM.json` exists and records baseline version, profile and source metadata

#### Scenario: Baseline is offline-readable
- **WHEN** the user is offline
- **THEN** local commands can still read `openspec/UPSTREAM.json` to report the installed baseline

### Requirement: Opsx version reports effective workflow version
The installed workflow SHALL provide `opsx-version` reporting the effective OpenSpec workflow version and baseline source.

#### Scenario: Version report includes baseline
- **WHEN** the user runs `opsx-version` through the installed workflow
- **THEN** the output includes the effective OpenSpec version, profile, baseline source and whether it is local embedded metadata

#### Scenario: Missing baseline is actionable
- **WHEN** `openspec/UPSTREAM.json` is missing or invalid
- **THEN** `opsx-version` reports an actionable failure instead of inventing a version

### Requirement: Baseline participates in stay-updated resolution
The installed OpenSpec workflow SHALL use `openspec/UPSTREAM.json` as the local baseline input for stay-updated resolution.

#### Scenario: Baseline includes resolver metadata
- **WHEN** a target is installed or synced with stay-updated assets
- **THEN** `openspec/UPSTREAM.json` includes enough metadata to compare effective version, minimum compatible version and source type

#### Scenario: Baseline remains offline-readable
- **WHEN** the user is offline and no cache or PATH source is available
- **THEN** the installed baseline remains sufficient for local workflow commands to report the fallback version

### Requirement: Opsx version reports resolved source
The installed workflow SHALL report the resolved OpenSpec source layer, not only static baseline metadata.

#### Scenario: Version report shows resolver layer
- **WHEN** the user runs `opsx-version` after stay-updated support is installed
- **THEN** the output identifies `PATH`, cache or embedded baseline as the effective source

#### Scenario: Resolver failures are actionable
- **WHEN** resolver metadata, cache or baseline files are invalid
- **THEN** `opsx-version` reports the failing layer and the next recovery action instead of inventing a version

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
