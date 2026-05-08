## ADDED Requirements

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
