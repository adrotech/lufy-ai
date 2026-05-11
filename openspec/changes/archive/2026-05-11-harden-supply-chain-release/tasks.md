## 1. Workflow Baseline And Pinning

- [x] 1.1 Inventory third-party `uses:` references in `.github/workflows/*.yml` and record current upstream action versions.
- [x] 1.2 Pin release-sensitive workflow actions to commit SHAs while preserving human-readable upstream version comments.
- [x] 1.3 Add a local validation check that fails on floating third-party actions in release-sensitive workflows.
- [x] 1.4 Wire the action-pinning check into `scripts/validate.sh` or an equivalent grouped validation path without requiring GitHub secrets.

## 2. Release Signing Provenance And SBOM

- [x] 2.1 Update `.github/workflows/release.yml` permissions for keyless signing/provenance/SBOM with minimum required scopes.
- [x] 2.2 Add keyless signing for release archives and checksum file, and upload signatures or signing bundles as release assets.
- [x] 2.3 Add provenance generation for release artifact subjects and upload the provenance/attestation with the GitHub Release.
- [x] 2.4 Add SBOM generation for `tools/lufy-cli-go` release artifacts and upload the SBOM with the GitHub Release.
- [x] 2.5 Extend release artifact smoke validation to assert expected signatures/provenance/SBOM files exist for the generated artifact set.

## 3. Auto-Tag Governance

- [x] 3.1 Reduce `.github/workflows/auto-release-tag.yml` token permissions to the minimum needed for PR metadata, tag creation and release dispatch.
- [x] 3.2 Implement release label resolution for `release:patch`, `release:minor`, `release:major` and `release:skip`.
- [x] 3.3 Block conflicting release labels with actionable errors and preserve default patch bump when no release label is present.
- [x] 3.4 Add bounded retry/backoff for tag races, with remote tag re-fetch before each retry and no force-push behavior.
- [x] 3.5 Sanitize PR title text before writing tag annotations or generated release notes.

## 4. Documentation And Specs

- [x] 4.1 Document release verification guidance for checksums, signatures, provenance and SBOM without requiring source checkout.
- [x] 4.2 Document release label policy and skip semantics for promotion PRs into `main`.
- [x] 4.3 Update embedded managed assets if changed root artifacts are part of the install catalog.
- [x] 4.4 Update OpenSpec tasks/specs as implementation details are finalized.

## 5. Validation And Delivery

- [x] 5.1 Run `scripts/validate.sh` from repository root.
- [x] 5.2 Run release artifact build/smoke validation locally where available without publishing a release.
- [x] 5.3 Validate workflow YAML syntax after workflow edits.
- [x] 5.4 Verify OpenSpec status for `harden-supply-chain-release` before implementation delivery.
