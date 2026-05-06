## ADDED Requirements

### Requirement: Release artifact CI
The CI system SHALL build and validate versioned release artifacts for the `lufy-ai` CLI.

#### Scenario: Matrix build for release artifacts
- **WHEN** the release workflow runs for an authorized tag or release event
- **THEN** it builds `lufy-ai` artifacts from `tools/lufy-cli-go/` for the supported OS/arch matrix without depending on root Node/TS tooling

#### Scenario: Version metadata injected
- **WHEN** the release workflow builds binaries
- **THEN** it injects version, commit and build date metadata consumed by `lufy-ai version`

### Requirement: Release smoke validation
The CI system SHALL validate release artifacts before they are published as installable artifacts.

#### Scenario: Artifact version smoke
- **WHEN** a release artifact is built for the runner platform or can be executed in CI
- **THEN** CI runs `lufy-ai version` from the packaged artifact and confirms the expected version metadata is present

#### Scenario: Install smoke from artifact
- **WHEN** a release artifact is executable in CI
- **THEN** CI runs at least `install --dry-run`, a temporary install and `verify --target <temp> --no-engram` using the packaged binary or unpacked release artifact

#### Scenario: Checksum smoke
- **WHEN** release artifacts and checksum files are generated
- **THEN** CI recalculates SHA-256 hashes and confirms they match the published checksum entries

### Requirement: Bootstrap CI validation
The CI system SHALL validate the bootstrap installer without requiring live mutation of user machines.

#### Scenario: Bootstrap dry-run
- **WHEN** bootstrap installer code changes
- **THEN** CI or local validation can run a dry-run or fixture-backed mode that resolves OS/arch, version and artifact URL without writing outside a temporary directory

#### Scenario: Bootstrap checksum failure test
- **WHEN** bootstrap validation runs against a fixture with an incorrect checksum
- **THEN** the test confirms installation is blocked and the binary is not executed
