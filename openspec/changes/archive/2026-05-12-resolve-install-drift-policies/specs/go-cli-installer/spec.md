## ADDED Requirements

### Requirement: CLI scope flag for managed operations
The CLI SHALL expose a `--scope` flag for managed install/sync/verify/status operations where scope affects target resolution.

#### Scenario: Invalid scope rejected
- **WHEN** the user passes an unsupported `--scope` value
- **THEN** the CLI exits non-zero with allowed values `project`, `global` and `both`

#### Scenario: Scope shown in dry-run
- **WHEN** the user runs install or sync with `--dry-run` and a scope value
- **THEN** the output identifies the effective scope and root paths that would be written

### Requirement: CLI merge command
The CLI SHALL expose `lufy-ai merge <path>` for policy-driven drift resolution where ancestor and `.lufy-new` data exist.

#### Scenario: Help includes merge
- **WHEN** the user requests CLI help
- **THEN** the output lists `merge` as the command for reconciling `.lufy-new` files with local edits

#### Scenario: Merge does not write before tool succeeds
- **WHEN** the merge tool cannot be started or exits unsuccessfully
- **THEN** the CLI preserves the original target, ancestor and `.lufy-new` files

### Requirement: CLI restore discovery
The CLI SHALL support backup discovery in addition to restoring from explicit manifest paths.

#### Scenario: Restore list mode is non-mutating
- **WHEN** the user asks restore to list backups
- **THEN** the CLI reads backup manifests and prints available backups without writing target files

#### Scenario: Explicit manifest remains supported
- **WHEN** the user runs `restore --backup <manifest-or-dir>`
- **THEN** the CLI preserves the existing restore behavior for that explicit backup reference

### Requirement: CLI reports drift actions consistently
The CLI SHALL use consistent action names and JSON fields for policy-driven drift handling.

#### Scenario: Plan includes lufy-new action
- **WHEN** install or sync plans a no-replace drift resolution
- **THEN** human and JSON plan output identify an action for writing `.lufy-new` rather than a destructive update

#### Scenario: Verify and status share policy fields
- **WHEN** verify or status emits JSON
- **THEN** each relevant asset result includes policy, scope, target path and recommended action when drift is detected
