# ci-quality-gates Specification

## Purpose
Define the local and CI validation gates that keep the Go CLI, shell scripts, workflows and release E2E paths measurable, portable and backed by real commands.

## Requirements
### Requirement: Go coverage gate
The validation system SHALL measure Go test coverage for `tools/lufy-cli-go` and enforce an initial documented threshold.

#### Scenario: Coverage profile generated
- **WHEN** CI or grouped local validation runs the Go quality gate
- **THEN** it generates a Go coverage profile from `tools/lufy-cli-go` using real Go tooling

#### Scenario: Coverage threshold enforced
- **WHEN** the measured coverage is below the configured threshold
- **THEN** validation fails with the measured percentage and configured threshold

### Requirement: Go lint gate
The validation system SHALL run a minimal Go lint/static analysis gate for the CLI implementation.

#### Scenario: Lint runs in CI
- **WHEN** CI validates changes affecting `tools/lufy-cli-go`
- **THEN** it runs the configured Go lint/static analysis command without depending on root Node/TS tooling

#### Scenario: Missing local lint tool is explicit
- **WHEN** local grouped validation cannot run an optional lint tool because it is not installed
- **THEN** it reports the limitation explicitly rather than claiming lint success

### Requirement: Shell script lint gate
The validation system SHALL lint repository shell scripts that participate in install, bootstrap, release or validation flows.

#### Scenario: ShellCheck validates critical scripts
- **WHEN** CI runs shell validation
- **THEN** it runs ShellCheck against `scripts/*.sh` and `tools/lufy-cli-go/scripts/*.sh`

#### Scenario: Shell lint scope excludes documentation snippets
- **WHEN** shell validation selects files
- **THEN** it does not lint Markdown/YAML embedded snippets as shell scripts

### Requirement: Multi-platform Go validation
The CI system SHALL validate Go tests/builds across supported development platforms.

#### Scenario: OS matrix runs tests and build
- **WHEN** pull request CI runs for CLI-relevant changes
- **THEN** it executes Go tests and build on Linux, macOS and Windows runners or an explicitly documented supported subset

#### Scenario: Platform-specific smokes are gated
- **WHEN** a smoke requires POSIX shell behavior or non-Windows paths
- **THEN** CI runs that smoke only on compatible runners and documents skipped platforms

### Requirement: Post-release E2E validation
The CI system SHALL provide a separate E2E validation path for published GitHub Release artifacts.

#### Scenario: E2E uses published release assets
- **WHEN** the post-release E2E workflow runs for a `v*` tag
- **THEN** it downloads the published artifact and checksum from GitHub Releases, verifies integrity and runs install/verify against a temporary target

#### Scenario: E2E is not required for normal PRs
- **WHEN** a normal pull request runs CI
- **THEN** it does not require a published release artifact to pass

### Requirement: Output and runtime regression tests
The CLI test suite SHALL include targeted regressions for user-visible plan output and command runtime wiring.

#### Scenario: Golden plan output covered
- **WHEN** installer plan output behavior changes
- **THEN** tests compare representative output to an approved fixture or structured expectation

#### Scenario: CLI runtime entrypoint covered
- **WHEN** Go tests run for `cmd/lufy-ai`
- **THEN** at least one test covers command dispatch/runtime behavior without mutating the user's real filesystem
