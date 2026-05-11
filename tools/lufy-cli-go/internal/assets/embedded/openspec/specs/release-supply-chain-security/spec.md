# release-supply-chain-security Specification

## Purpose
Define the release supply-chain controls for signed artifacts, provenance, SBOM publication, pinned actions and minimum workflow permissions.

## Requirements
### Requirement: Release artifacts are keylessly signed
The release workflow SHALL produce keyless signatures for every published release artifact and checksum file using GitHub Actions OIDC identity.

#### Scenario: Signatures attached to release
- **WHEN** a stable `v*` release is published from a commit reachable from `origin/main`
- **THEN** each downloadable artifact and the checksum file has a corresponding signature or bundle attached to the GitHub Release

#### Scenario: No long-lived signing secret required
- **WHEN** the release workflow signs artifacts
- **THEN** it uses OIDC/keyless signing and does not require a repository secret containing a private signing key

### Requirement: Release provenance is published
The release workflow SHALL generate provenance for published release artifacts so consumers can verify what workflow and commit produced them.

#### Scenario: Provenance references release commit
- **WHEN** release provenance is generated
- **THEN** it identifies the release tag, resolved commit, workflow identity and artifact subjects

#### Scenario: Provenance is release-scoped
- **WHEN** artifacts are uploaded to GitHub Releases
- **THEN** the provenance file or attestation is uploaded with the same release rather than only retained as an expiring workflow artifact

### Requirement: Release SBOM is published
The release workflow SHALL generate and publish an SBOM for the `lufy-ai` release artifacts.

#### Scenario: SBOM attached to release
- **WHEN** release artifacts are built
- **THEN** the release includes an SBOM file in a standard machine-readable format such as SPDX or CycloneDX

#### Scenario: SBOM covers Go module dependencies
- **WHEN** the SBOM is generated
- **THEN** it includes the Go module dependency graph for `tools/lufy-cli-go`

### Requirement: Sensitive workflows pin third-party actions
Workflows that build, tag, sign, attest or publish release artifacts SHALL pin third-party GitHub Actions to immutable commit SHAs.

#### Scenario: Actions are pinned by SHA
- **WHEN** `release.yml`, `auto-release-tag.yml` or release-relevant CI uses a third-party action
- **THEN** the `uses:` reference points to a commit SHA, with a nearby comment or naming convention preserving the human-readable upstream action version

#### Scenario: Floating action tags are blocked in review
- **WHEN** a workflow change introduces `uses:` references like `@v4`, `@v5` or branch names for release-sensitive jobs
- **THEN** validation or review treats the change as blocked until the action is pinned

### Requirement: Release workflows use minimum permissions
Release-sensitive workflows SHALL declare the minimum GitHub token permissions needed for their specific operations.

#### Scenario: Release workflow permissions are explicit
- **WHEN** the release workflow publishes artifacts, signatures, provenance or SBOM
- **THEN** permissions are declared explicitly and include only required scopes such as `contents: write` and OIDC/provenance scopes when needed

#### Scenario: Auto-tag permissions are constrained
- **WHEN** the auto-tag workflow creates a tag and dispatches the release workflow
- **THEN** it does not request unrelated permissions beyond those needed to read PR metadata, write tags and dispatch workflows
