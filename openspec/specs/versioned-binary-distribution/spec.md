# versioned-binary-distribution Specification

## Purpose
TBD - created by archiving change add-versioned-binary-release-installer. Update Purpose after archive.
## Requirements
### Requirement: Versioned release artifacts
The system SHALL publish versioned `lufy-ai` binary artifacts for supported OS/arch targets via GitHub Releases.

#### Scenario: Release contains platform artifacts
- **WHEN** a maintainer creates an authorized release tag
- **THEN** GitHub Releases contains one packaged `lufy-ai` binary artifact per supported OS/arch target with deterministic names that include version, OS and architecture

#### Scenario: Unsupported platform omitted explicitly
- **WHEN** an OS/arch target is not supported by the release matrix
- **THEN** no artifact is published for that target and installation tooling reports the platform as unsupported instead of guessing a fallback

### Requirement: Release checksums
The system SHALL publish SHA-256 checksums for every release artifact and make checksum verification part of release validation.

#### Scenario: Checksums file generated
- **WHEN** release artifacts are built
- **THEN** a checksum file is generated in the same release and contains one SHA-256 entry for every downloadable artifact

#### Scenario: Checksum validation in CI
- **WHEN** the release workflow packages artifacts
- **THEN** CI verifies that each artifact hash matches the generated checksums before publication or before marking the release job successful

### Requirement: CLI version metadata
The `lufy-ai` binary SHALL expose version metadata through a `version` command.

#### Scenario: Version output includes build metadata
- **WHEN** the user runs `lufy-ai version`
- **THEN** the output includes at least semantic version, git commit, build date, GOOS and GOARCH

#### Scenario: Development build is explicit
- **WHEN** the binary was built locally without release metadata
- **THEN** `lufy-ai version` reports a clearly marked development or unknown version instead of pretending to be a release

### Requirement: Distribution channels are layered
The system SHALL treat GitHub Releases with checksums as the source of truth for additional distribution channels.

#### Scenario: Secondary package manager uses release artifacts
- **WHEN** Homebrew, Scoop or similar package manager support is added
- **THEN** its formula or manifest references versioned GitHub Release artifacts and their checksums instead of rebuilding from unrelated sources

#### Scenario: Go install remains explicit about limitations
- **WHEN** `go install` is documented or supported as a channel
- **THEN** documentation states whether the resulting binary includes standalone assets and does not present it as clone-free install unless that is true
