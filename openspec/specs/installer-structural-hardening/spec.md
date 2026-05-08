# installer-structural-hardening Specification

## Purpose
TBD - created by archiving change hardening-structural-foundation. Update Purpose after archive.
## Requirements
### Requirement: Asset mirror parity gate
The installer SHALL provide a deterministic validation gate that detects drift between canonical managed assets in the repository root and the embedded assets used by standalone binaries.

#### Scenario: Root and embedded catalogs match
- **WHEN** the parity validation runs in development or CI
- **THEN** it compares the root catalog and embedded catalog by target path, kind, policy and source SHA-256, ignoring intentionally excluded paths such as `openspec/changes`

#### Scenario: Drift is reported before release
- **WHEN** a managed root asset differs from its embedded counterpart or is missing from either side
- **THEN** validation fails with the affected target path so the mirror can be updated before publishing a binary

### Requirement: Portable relative path safety
The installer SHALL reject relative paths that escape the allowed root regardless of slash style or platform separator semantics.

#### Scenario: Backslash traversal rejected
- **WHEN** a catalog, state or backup manifest path contains `..\` traversal or mixed separators that normalize outside the root
- **THEN** path validation rejects it before any filesystem read or write

#### Scenario: Safe relative path preserved
- **WHEN** a managed path is a normal relative path like `.opencode/agents/explorer.md`
- **THEN** path validation returns a clean relative path suitable for `SafeJoin`

### Requirement: Stable catalog fingerprint
The installer SHALL calculate a stable fingerprint for the effective managed asset catalog using deterministic asset metadata.

#### Scenario: Fingerprint from sorted assets
- **WHEN** a catalog is built from checkout or embedded assets
- **THEN** the fingerprint is derived from sorted file assets using at least `targetRel` and `sourceSHA256`

#### Scenario: Same content produces same fingerprint
- **WHEN** two catalogs contain the same managed file targets and source hashes regardless of traversal order
- **THEN** they produce the same fingerprint

### Requirement: Atomic managed file writes
The installer SHALL write managed file content atomically to avoid partially written target files on process interruption.

#### Scenario: Install and sync write atomically
- **WHEN** install or sync copies or updates a managed file
- **THEN** content is written to a temp file in the destination directory and then renamed into place

#### Scenario: Backup and restore write atomically
- **WHEN** backup captures a file or restore writes a captured file back to target
- **THEN** the destination file is written atomically with no direct partial `os.WriteFile` target write

