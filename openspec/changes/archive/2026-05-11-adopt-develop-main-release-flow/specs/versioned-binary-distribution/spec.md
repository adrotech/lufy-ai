## MODIFIED Requirements

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
