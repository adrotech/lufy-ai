## ADDED Requirements

### Requirement: CLI rescan drift reporting
The CLI Go SHALL expose `lufy-ai init --rescan` as the stack-aware project rescan mode that reports drift between `.opencode/project.yaml` and current repository evidence.

#### Scenario: Help describes rescan drift behavior
- **WHEN** the user requests help for `lufy-ai init`
- **THEN** the output describes `--rescan` as refreshing stack evidence, preserving user overrides and reporting drift without destructive cleanup

#### Scenario: Rescan delegates outside main
- **WHEN** `cmd/lufy-ai/main.go` receives `init --rescan`
- **THEN** it delegates scanning, drift comparison, merge planning, reporting and writing logic to internal packages instead of implementing that logic in `main.go`

#### Scenario: Rescan reports clean idempotent state
- **WHEN** the user runs `lufy-ai init --target <dir> --rescan` twice without target or config changes between runs
- **THEN** the second run exits successfully, reports no drift and does not create backups or modify unrelated install state

### Requirement: CLI rescan validation coverage
The implementation of `lufy-ai init --rescan` SHALL be validated with Go tests and confined filesystem fixtures for drift, stale detection and idempotency.

#### Scenario: Fixture tests cover rescan drift categories
- **WHEN** Go tests run for the CLI packages
- **THEN** fixtures verify at least no-drift, new stack drift, tooling drift, CI drift, stale stack detection, invalid existing config and unknown field preservation

#### Scenario: Validation command covers rescan
- **WHEN** `scripts/validate.sh` runs after this change is implemented
- **THEN** the Go validation includes tests for `lufy-ai init --rescan` and still validates existing install, sync, verify and init behavior
