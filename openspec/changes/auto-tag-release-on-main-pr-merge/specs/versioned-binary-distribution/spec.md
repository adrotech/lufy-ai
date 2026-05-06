## ADDED Requirements

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
