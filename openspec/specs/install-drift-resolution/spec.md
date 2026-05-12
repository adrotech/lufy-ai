# install-drift-resolution Specification

## Purpose
TBD - created by archiving change resolve-install-drift-policies. Update Purpose after archive.
## Requirements
### Requirement: Declarative asset policies
The installer SHALL assign every catalog entry a declarative update policy that controls how install and sync handle local drift.

#### Scenario: Supported policies are explicit
- **WHEN** the CLI builds the effective asset catalog
- **THEN** each asset has one of `managed`, `no-replace`, `merge-block`, `merge-json` or `metadata` as its policy

#### Scenario: Unknown policy is rejected
- **WHEN** an asset policy cannot be parsed from catalog or persisted state
- **THEN** install, sync and verify fail with an actionable unsupported-policy error before writing target files

### Requirement: No-replace assets preserve user drift
Assets with policy `no-replace` SHALL preserve the user-owned target file when local drift exists and write the new lufy version to a `.lufy-new` sibling path.

#### Scenario: Drifted no-replace asset gets new sibling
- **WHEN** a `no-replace` asset exists in the target, has drift from the last recorded target hash and the source version changed
- **THEN** install or sync writes the new source content to `<target>.lufy-new` and does not modify the original target file

#### Scenario: Clean no-replace asset can update safely
- **WHEN** a `no-replace` asset is recorded in install state, has no local drift and the source version changed
- **THEN** install or sync may update the original target after creating the required backup and ancestor records

### Requirement: Merge-block assets preserve user text
Assets with policy `merge-block` SHALL update only lufy-managed blocks delimited by `<!-- LUFY:BEGIN <id> -->` and `<!-- LUFY:END <id> -->` while preserving all text outside those blocks.

#### Scenario: Managed block updates in place
- **WHEN** a target file contains a well-formed lufy block and the source block content changed
- **THEN** install or sync replaces only the content inside the matching block markers and preserves content outside the block byte-for-byte except for required line-ending normalization documented by the implementation

#### Scenario: Missing block is inserted safely
- **WHEN** a merge-block target exists without a required source block and the insertion point is supported
- **THEN** install or sync inserts the missing lufy block without deleting user content

#### Scenario: Corrupt markers block writing
- **WHEN** a merge-block target has duplicate, nested or unclosed lufy markers
- **THEN** install or sync reports a conflict and does not write the target file

### Requirement: Ancestors are recorded for managed versions
The CLI SHALL record the last clean lufy-provided content for drift-resolvable assets under `.lufy-ai/ancestors/` using safe relative paths.

#### Scenario: Ancestor stored after successful write
- **WHEN** install or sync successfully writes or updates a policy-managed asset that supports drift resolution
- **THEN** the CLI stores the source content used for that write as the asset ancestor under `.lufy-ai/ancestors/`

#### Scenario: Ancestor path cannot escape target
- **WHEN** an asset target path would map to an unsafe ancestor path
- **THEN** the CLI rejects the operation before writing any ancestor or target content

### Requirement: Drift status is user-visible
The CLI SHALL expose policy-driven drift status in human and JSON outputs for verify and status.

#### Scenario: Verify reports lufy-new
- **WHEN** a target contains a `.lufy-new` file generated for a no-replace asset
- **THEN** `verify --json` includes the original path, new sibling path, policy and recommended next action

#### Scenario: Status reports merge-required
- **WHEN** a drifted asset needs manual merge or review
- **THEN** `status` reports the asset as requiring user action instead of reporting the whole installation as opaque failure

### Requirement: Scope-aware installation
The CLI SHALL support project, global and both scopes for installable assets while preserving project-only behavior when requested.

#### Scenario: Project scope preserves current behavior
- **WHEN** the user runs install or sync with `--scope=project`
- **THEN** OpenCode assets are planned under the project target as they are today

#### Scenario: Global scope uses OpenCode config directory
- **WHEN** the user runs install or sync with `--scope=global`
- **THEN** shared OpenCode assets are planned under the resolved OpenCode global config directory and project-only assets remain under the project target only when required by their catalog scope

#### Scenario: Both scope covers global and project assets
- **WHEN** the user runs install or sync with `--scope=both`
- **THEN** the plan includes global shared assets and project-local assets without duplicating entries in install state

### Requirement: Merge command uses ancestor user and new versions
The CLI SHALL provide a merge workflow for files with recorded ancestor, current user content and `.lufy-new` content.

#### Scenario: Merge validates inputs before tool invocation
- **WHEN** the user runs `lufy-ai merge <path>`
- **THEN** the CLI validates that the target file, recorded ancestor and `.lufy-new` file exist and are safe before invoking any external merge tool

#### Scenario: Merge tool is configurable
- **WHEN** `LUFY_MERGE_TOOL` is set
- **THEN** `lufy-ai merge <path>` uses that tool invocation instead of the default merge tool

### Requirement: Restore lists and selects backups
Restore UX SHALL allow users to discover backup IDs and restore a selected backup without requiring them to manually locate manifest paths.

#### Scenario: Restore lists backups
- **WHEN** the user runs restore in list mode for a target with backups
- **THEN** the CLI lists available backup IDs, timestamps and manifest paths without modifying files

#### Scenario: Restore selected backup by ID
- **WHEN** the user runs restore with a valid backup ID and required confirmation or dry-run flags
- **THEN** the CLI resolves that ID to a backup manifest and applies the existing confined restore rules
