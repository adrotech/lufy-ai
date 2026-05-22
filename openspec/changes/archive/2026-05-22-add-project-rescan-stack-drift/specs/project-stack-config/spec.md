## ADDED Requirements

### Requirement: Rescan drift validation
`lufy-ai init --rescan` SHALL compare an existing `.opencode/project.yaml` with the target repository's current stack, tooling and CI evidence before deciding whether to write changes.

#### Scenario: No drift keeps project config unchanged
- **GIVEN** `.opencode/project.yaml` matches the current detectable stack, tooling and CI evidence
- **WHEN** the user runs `lufy-ai init --target <dir> --rescan`
- **THEN** the CLI reports no drift and MUST NOT rewrite `.opencode/project.yaml`

#### Scenario: New stack evidence is reported and merged
- **GIVEN** `.opencode/project.yaml` contains stack `go`
- **WHEN** the target later contains TypeScript/Next marker files and the user runs `lufy-ai init --target <dir> --rescan`
- **THEN** the CLI reports a new stack drift item for `typescript` and writes a merged `.opencode/project.yaml` that preserves the existing Go stack and user-managed fields

#### Scenario: Tooling drift is reported per stack
- **GIVEN** `.opencode/project.yaml` contains a JavaScript stack without a detected test runner
- **WHEN** `package.json` later declares `vitest` and the user runs `lufy-ai init --target <dir> --rescan`
- **THEN** the CLI reports tooling drift for the JavaScript stack and records the newly detected test runner without removing unrelated stack preferences

### Requirement: Stale stack detection is non-destructive
`lufy-ai init --rescan` SHALL detect configured stacks whose marker files are no longer present and SHALL treat them as stale evidence rather than implicit deletion requests.

#### Scenario: Missing marker files mark a stack stale
- **GIVEN** `.opencode/project.yaml` contains stack `python` with user-managed preferences
- **WHEN** the target no longer contains `pyproject.toml`, `requirements.txt` or `setup.py` and the user runs `lufy-ai init --target <dir> --rescan`
- **THEN** the CLI reports the Python stack as stale or deprecated and MUST preserve the stack entry and its user-managed preferences

#### Scenario: Stale stack does not delete files
- **GIVEN** a stale stack is detected during `lufy-ai init --rescan`
- **WHEN** the command completes successfully
- **THEN** the CLI MUST NOT delete source files, configuration files, `.opencode/project.yaml` entries or unknown user files as part of stale handling

### Requirement: Rescan preserves editable configuration boundaries
`lufy-ai init --rescan` SHALL update only generated or detected fields that are safe to refresh and SHALL preserve editable user-managed fields by default.

#### Scenario: Unknown fields are preserved
- **GIVEN** `.opencode/project.yaml` contains an unknown top-level field and an unknown field under stack `go`
- **WHEN** the user runs `lufy-ai init --target <dir> --rescan`
- **THEN** the resulting `.opencode/project.yaml` preserves both unknown fields unless the command fails before writing

#### Scenario: User override wins over regenerated default
- **GIVEN** `.opencode/project.yaml` contains a user override such as `coverage_threshold: 70` for stack `go`
- **WHEN** current Go detection would otherwise use a different default threshold and the user runs `lufy-ai init --target <dir> --rescan`
- **THEN** the resulting `.opencode/project.yaml` preserves `coverage_threshold: 70` and reports that the user-managed override was preserved when reporting is verbose enough to include preserved decisions

### Requirement: Rescan report is structured and actionable
`lufy-ai init --rescan` SHALL emit a structured human-readable report that identifies drift decisions and suggested next actions.

#### Scenario: Report includes actionable drift fields
- **WHEN** `lufy-ai init --target <dir> --rescan` detects any stack, tooling, CI or stale drift
- **THEN** the report includes each item's category, severity, affected stack or field path, applied status and suggested action

#### Scenario: Report distinguishes detection from mutation
- **WHEN** `lufy-ai init --target <dir> --rescan` detects stale evidence that is not safe to remove automatically
- **THEN** the report marks the item as detected or skipped and MUST NOT imply that cleanup or deletion was performed

#### Scenario: Invalid existing config fails without mutation
- **WHEN** `.opencode/project.yaml` exists but is not parseable as the supported configuration format and the user runs `lufy-ai init --target <dir> --rescan`
- **THEN** the CLI exits non-zero, reports an actionable parse error and MUST NOT overwrite the existing file
