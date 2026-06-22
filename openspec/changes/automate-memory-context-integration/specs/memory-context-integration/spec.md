## ADDED Requirements

### Requirement: OpenCode memory lifecycle integration

LUFY SHALL provide best-effort OpenCode lifecycle integration for memory orientation and validation without treating private memory note contents as managed assets.

#### Scenario: OpenCode session creation orients memory
- **WHEN** OpenCode loads project-local plugins and creates a session
- **THEN** the LUFY plugin SHALL run memory orientation best-effort using the installed memory hook

#### Scenario: Memory edits trigger validation
- **WHEN** OpenCode reports an edited file under `.lufy/memory/`
- **THEN** the LUFY plugin SHALL run memory validation best-effort using the installed validation hook

#### Scenario: Missing lifecycle support is diagnosable
- **WHEN** `lufy-ai doctor` or `lufy-ai verify --deep` runs
- **THEN** it SHALL report whether memory hooks and the OpenCode lifecycle plugin are present, and SHALL provide `lufy-ai sync --tool opencode --scope project` as recovery when they are missing

### Requirement: Context integration health is visible

LUFY SHALL surface context graph integration health through diagnostic commands instead of relying only on agent/skill policy.

#### Scenario: Doctor reports context recovery
- **WHEN** `lufy-ai doctor` runs and the context graph is `stale` or `not_available`
- **THEN** it SHALL report the status and recovery command `lufy-ai context build`

#### Scenario: Deep verify reports context recovery
- **WHEN** `lufy-ai verify --deep` runs and the context graph is `stale` or `not_available`
- **THEN** it SHALL report the status and recovery command `lufy-ai context build` without mutating context artifacts
