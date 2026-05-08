## ADDED Requirements

### Requirement: CI validates supply-chain release artifacts
The CI/release validation system SHALL verify that supply-chain metadata for release artifacts is produced and internally consistent before publishing or accepting release workflow changes.

#### Scenario: Release smoke includes verification metadata
- **WHEN** release artifact smoke validation runs
- **THEN** it confirms expected signature/provenance/SBOM files exist for the generated release artifact set or reports the missing files as a release blocker

#### Scenario: PR validation covers workflow syntax and local gates
- **WHEN** a pull request changes release or auto-tag workflows
- **THEN** local/CI validation includes syntax checks and the repository grouped validation command where applicable

### Requirement: CI fails on unpinned release-sensitive actions
The validation system SHALL detect release-sensitive workflow changes that use floating third-party action refs.

#### Scenario: Floating actions detected
- **WHEN** release-sensitive workflows contain `uses:` references to third-party actions without a commit SHA
- **THEN** validation fails with the workflow path and offending action reference

#### Scenario: Local validation can reproduce action pinning check
- **WHEN** a maintainer runs the grouped validation locally for a workflow change
- **THEN** the same action pinning check can run without requiring GitHub-hosted secrets or release publication
