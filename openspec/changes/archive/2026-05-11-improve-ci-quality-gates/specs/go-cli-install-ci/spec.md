## ADDED Requirements

### Requirement: CI quality gates extend installer validation
The Go installer CI SHALL include quality gates beyond basic test/build while preserving the existing install and wrapper smoke coverage.

#### Scenario: Quality gates run before smokes
- **WHEN** CI validates a pull request affecting CLI, scripts, workflows or managed assets
- **THEN** coverage, lint/static checks and shell script lint run before expensive installer/wrapper smokes where practical

#### Scenario: Existing smoke coverage preserved
- **WHEN** quality gates are added to the workflow
- **THEN** dry-run install, real install, verify, idempotence, backup/restore and wrapper delegation smokes remain covered

### Requirement: Local validation remains reproducible
The repository SHALL expose local validation commands matching the CI quality gates as closely as practical.

#### Scenario: Grouped local validation includes quality gates
- **WHEN** a maintainer runs `scripts/validate.sh`
- **THEN** it runs PR-aware whitespace, action pinning, Go tests/build and available quality gates for the CLI scope

#### Scenario: Unavailable optional tools are reported
- **WHEN** a local machine lacks optional tools such as ShellCheck or golangci-lint
- **THEN** the validation output reports the missing tool instead of inventing success
