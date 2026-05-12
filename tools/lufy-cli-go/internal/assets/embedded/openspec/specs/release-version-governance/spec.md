# release-version-governance Specification

## Purpose
Define how merged promotion PRs resolve release version bumps, skip semantics, tag-race handling and sanitized release metadata.

## Requirements
### Requirement: Auto-tag supports explicit release bump labels
The auto-tag workflow SHALL choose the next semantic version from release labels on the merged PR to `main`.

#### Scenario: Patch bump label
- **WHEN** a PR to `main` is merged with `release:patch`
- **THEN** the auto-tag workflow increments the patch component of the latest stable `v*` tag

#### Scenario: Minor bump label
- **WHEN** a PR to `main` is merged with `release:minor`
- **THEN** the auto-tag workflow increments the minor component and resets patch to zero

#### Scenario: Major bump label
- **WHEN** a PR to `main` is merged with `release:major`
- **THEN** the auto-tag workflow increments the major component and resets minor and patch to zero

#### Scenario: Default bump remains patch
- **WHEN** a PR to `main` is merged without a release bump label or skip label
- **THEN** the auto-tag workflow uses a patch bump and reports that default explicitly in logs

### Requirement: Auto-tag supports release skip
The auto-tag workflow SHALL allow explicitly skipping tag creation for a merged PR to `main`.

#### Scenario: Skip label prevents tag
- **WHEN** a PR to `main` is merged with `release:skip`
- **THEN** the auto-tag workflow exits successfully without creating or pushing a release tag

#### Scenario: Skip conflicts with bump labels
- **WHEN** a PR has `release:skip` and any bump label
- **THEN** the auto-tag workflow fails with an actionable message instead of guessing behavior

### Requirement: Auto-tag handles tag races safely
The auto-tag workflow SHALL avoid overwriting tags and SHALL retry safely when another process creates the next tag concurrently.

#### Scenario: Existing calculated tag is not overwritten
- **WHEN** the calculated next tag already exists locally or on origin
- **THEN** the workflow recalculates or exits without force-pushing or replacing the existing tag

#### Scenario: Retry after remote race
- **WHEN** pushing the calculated tag fails because the remote tag appeared after calculation
- **THEN** the workflow fetches tags again, recalculates the next tag and retries up to a bounded limit with backoff

### Requirement: Release metadata is sanitized
The auto-tag and release workflows SHALL sanitize human-provided text before including it in tag annotations or release notes.

#### Scenario: PR title sanitized for tag annotation
- **WHEN** a PR title contains newlines, control characters or excessive length
- **THEN** the tag annotation stores a normalized single-line summary and preserves the PR number and merge commit separately

#### Scenario: Release notes generated from trusted fields
- **WHEN** release notes are generated automatically
- **THEN** they use trusted GitHub metadata such as PR number, title, labels and commits without executing or interpreting user-provided content as shell commands
