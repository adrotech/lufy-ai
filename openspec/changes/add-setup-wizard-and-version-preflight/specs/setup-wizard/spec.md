## ADDED Requirements

### Requirement: Setup starts with version preflight

`lufy-ai setup` SHALL check whether a newer stable LUFY AI release is available before planning or applying project configuration, unless the user explicitly disables the check.

#### Scenario: New version is available
- **WHEN** the user runs `lufy-ai setup --target <dir>` and the latest stable release is newer than the local version
- **THEN** setup reports the local version, latest version and recommended `lufy-ai upgrade --to <version>` command before continuing with configuration planning

#### Scenario: Local version is current
- **WHEN** the user runs `lufy-ai setup --target <dir>` and the local version is equal to or newer than the latest stable release
- **THEN** setup reports that LUFY AI is up to date and continues with configuration planning

#### Scenario: Require latest blocks mutation
- **WHEN** the user runs `lufy-ai setup --target <dir> --require-latest` and a newer version is available
- **THEN** setup exits non-zero before applying install, memory, context or verify actions

#### Scenario: Version check can be skipped
- **WHEN** the user runs `lufy-ai setup --target <dir> --skip-version-check`
- **THEN** setup skips the network version preflight and continues with configuration planning

### Requirement: Setup plans configurable features

`lufy-ai setup` SHALL build an explicit plan of configurable features and show whether each feature will apply, skip, or recommend user action.

#### Scenario: Dry run is non-mutating
- **WHEN** the user runs `lufy-ai setup --target <dir> --dry-run`
- **THEN** setup prints the version preflight and feature plan without writing project files

#### Scenario: Layout is planned after version preflight
- **WHEN** the target needs `.lufy` layout migration or README creation
- **THEN** setup reports the layout action as a feature after the version preflight and does not print layout output before the version result

#### Scenario: Missing install is planned
- **WHEN** the target lacks `.lufy/managed-state/install-state.json`
- **THEN** setup plans installation of managed assets using the existing install service

#### Scenario: Missing memory is planned
- **WHEN** the target has project config defaults but `.lufy/memory` is not initialized
- **THEN** setup plans memory initialization using the existing memory service

#### Scenario: Stack and methodology are visible
- **WHEN** setup builds a feature plan
- **THEN** setup includes explicit feature rows for stack/profile detection and SDD methodology state, with skip/apply status and recovery guidance

#### Scenario: Versioned feature metadata is exposed
- **WHEN** setup emits a human or JSON plan
- **THEN** configurable features include stable IDs and version metadata indicating when the feature became available

#### Scenario: Context graph missing or stale is planned
- **WHEN** the context graph status is not `ready`
- **THEN** setup plans `context build` using the existing context graph service

### Requirement: Setup applies selected defaults safely

`lufy-ai setup` SHALL apply its planned configuration only when the user passes an explicit confirmation flag in non-interactive/scripted mode.

#### Scenario: Apply requires confirmation
- **WHEN** setup detects actions that mutate files and the user did not pass `--yes` or `--dry-run`
- **THEN** setup exits non-zero with an actionable message to rerun with `--dry-run` or `--yes`

#### Scenario: Apply with yes
- **WHEN** setup runs with `--yes`
- **THEN** setup applies pending install, memory and context actions and runs verify using existing services

#### Scenario: JSON output is parseable
- **WHEN** setup runs with `--json`
- **THEN** setup emits a single JSON report containing version preflight, feature plan and applied status

#### Scenario: JSON without confirmation is safe
- **WHEN** setup runs with `--json` without `--yes` and the plan contains mutating actions
- **THEN** setup emits the JSON report and exits non-zero instead of silently succeeding

#### Scenario: Interactive checklist
- **WHEN** setup runs in an interactive terminal without `--yes`, `--dry-run` or `--json`
- **THEN** setup shows a Bubble Tea checklist of pending actions and applies only the selected actions after confirmation

### Requirement: Upgrade recommends setup for new features

`lufy-ai upgrade` SHALL guide users to run setup after a successful binary upgrade so new configurable features can be discovered.

#### Scenario: Upgrade success suggests setup
- **WHEN** `lufy-ai upgrade --to <version>` completes successfully
- **THEN** the output recommends `lufy-ai setup --target . --check-new-features`
