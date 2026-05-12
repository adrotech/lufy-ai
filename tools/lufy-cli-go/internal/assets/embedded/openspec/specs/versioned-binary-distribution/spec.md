# versioned-binary-distribution Specification

## Purpose
Define versioned binary distribution through GitHub Releases, including supported platform artifacts, checksums, version metadata, upgrade behavior and release tag automation.

## Requirements
### Requirement: Versioned release artifacts
The system SHALL publish versioned `lufy-ai` binary artifacts for supported OS/arch targets via GitHub Releases only from stable `v*` tags whose commits are reachable from `origin/main`.

#### Scenario: Release contains platform artifacts
- **WHEN** a maintainer creates an authorized release tag `v*` on a commit reachable from `origin/main`
- **THEN** GitHub Releases contains one packaged `lufy-ai` binary artifact per supported OS/arch target with deterministic names that include version, OS and architecture

#### Scenario: Tag not on main is blocked
- **WHEN** the release workflow runs for a `v*` tag whose commit is not reachable from `origin/main`
- **THEN** the workflow fails before publishing GitHub Release assets

#### Scenario: Develop does not publish stable release
- **WHEN** changes exist only on `develop` and have not been promoted to `main`
- **THEN** no stable GitHub Release assets are published from those commits

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

### Requirement: Automatic patch release tags from main PR merges
The system SHALL automatically create the next stable patch semver tag when a pull request targeting `main` is closed with `merged == true`.

#### Scenario: Merged PR to main creates next patch tag
- **WHEN** a pull request targeting `main` is closed with `merged == true` and at least one valid `vMAJOR.MINOR.PATCH` tag exists
- **THEN** the system creates and pushes an annotated tag `vMAJOR.MINOR.(PATCH+1)` based on the highest existing simple semver tag

#### Scenario: First automatic release tag
- **WHEN** a pull request targeting `main` is closed with `merged == true` and no valid `vMAJOR.MINOR.PATCH` tags exist
- **THEN** the system creates and pushes annotated tag `v0.1.0`

#### Scenario: Non-merged or non-main PR does not tag
- **WHEN** a pull request is closed without being merged or targets a branch other than `main`
- **THEN** the system does not create or push a release tag

### Requirement: Automatic release tag target safety
The system SHALL create automatic release tags only on the final merge commit that is reachable from `origin/main`.

#### Scenario: Tag points to merge commit
- **WHEN** an automatic release tag is created for a merged PR to `main`
- **THEN** the tag points to the PR merge commit SHA from the GitHub event

#### Scenario: Merge commit must be reachable from main
- **WHEN** the PR merge commit is not reachable from `origin/main`
- **THEN** the system fails before creating or pushing a tag with a clear policy message

### Requirement: Automatic release tag idempotency
The system SHALL avoid overwriting or recreating existing release tags during automatic tag creation.

#### Scenario: Calculated tag already exists
- **WHEN** the next calculated `vMAJOR.MINOR.PATCH` tag already exists locally or on the remote
- **THEN** the system exits without creating, moving or pushing that tag and reports an explicit no-op message

#### Scenario: Created automatic tag dispatches release workflow
- **WHEN** the automatic tag is pushed successfully
- **THEN** the automatic tag workflow invokes the release workflow explicitly with `workflow_dispatch` for that tag

#### Scenario: Existing calculated tag does not dispatch duplicate release
- **WHEN** the next calculated `vMAJOR.MINOR.PATCH` tag already exists locally or on the remote
- **THEN** the system does not invoke the release workflow automatically for that existing tag

### Requirement: Release workflow supports manual and dispatched tags safely
The system SHALL keep release publication centralized in `.github/workflows/release.yml` and SHALL support both manual/human tag pushes and explicit workflow dispatch for an existing `v*` tag.

#### Scenario: Human tag push publishes through release workflow
- **WHEN** a `v*` tag is pushed by a human or integration capable of triggering tag push workflows
- **THEN** the release workflow validates the tag format and main reachability before building and publishing release artifacts

#### Scenario: Workflow dispatch publishes explicit tag
- **WHEN** the release workflow is invoked with `workflow_dispatch` and input `tag` set to an existing `v*` tag
- **THEN** the release workflow checks out that tag, validates the tag format and main reachability, and then builds and publishes release artifacts

### Requirement: Release artifacts include verification material
The system SHALL publish verification material alongside versioned `lufy-ai` binary artifacts so consumers can validate integrity, authenticity and provenance.

#### Scenario: Release contains signatures provenance and SBOM
- **WHEN** a stable GitHub Release is published for a `v*` tag
- **THEN** the release assets include binary archives, checksums, signatures or signing bundles, provenance/attestation data and an SBOM

#### Scenario: Checksum workflow remains compatible
- **WHEN** verification material is added to a release
- **THEN** existing checksum verification for binary archives remains available and documented as an integrity check

### Requirement: Release verification guidance exists
The system SHALL document how a maintainer or consumer can verify release artifacts using the published checksums, signatures and provenance.

#### Scenario: Verification commands documented
- **WHEN** release documentation describes downloading a binary artifact
- **THEN** it also describes the expected checksum and signature/provenance verification flow without requiring repository source checkout

#### Scenario: Verification limitations are explicit
- **WHEN** a verification tool such as `cosign` is not installed locally
- **THEN** documentation states the missing tool rather than implying verification has occurred
