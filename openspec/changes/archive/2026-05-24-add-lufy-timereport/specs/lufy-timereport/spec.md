## ADDED Requirements

### Requirement: Slash command and skill entrypoint
The system SHALL provide a `/lufy.timereport` OpenCode command that delegates to a local `lufy.timereport` skill for generating a time and ROI report.

#### Scenario: Command is discoverable
- **WHEN** a user inspects `.opencode/commands/lufy.timereport.md`
- **THEN** the command describes `/lufy.timereport`, its expected inputs, output behavior and its delegation to the `lufy.timereport` skill

#### Scenario: Skill is discoverable
- **WHEN** a user or agent inspects `.opencode/skills/lufy.timereport/`
- **THEN** a `SKILL.md` exists and describes purpose, usage, inputs, privacy boundaries, data sources, validation and outputs for the `lufy.timereport` skill

### Requirement: Offline self-contained HTML report
The system SHALL generate a single offline self-contained HTML report containing embedded styles and any required client-side behavior without requiring network access or remote assets.

#### Scenario: Default output path is temporary
- **WHEN** the user runs `/lufy.timereport` without specifying an output path
- **THEN** the generated report is written to a temporary path such as `/tmp/lufy-timereport-<timestamp>.html` and the path is reported to the user

#### Scenario: Configurable output path is honored
- **WHEN** the user runs `/lufy.timereport` with an explicit output path
- **THEN** the generated report is written to that path if the path is writable and no network resource is required to view it

#### Scenario: HTML contains required summary sections
- **WHEN** the report is generated successfully
- **THEN** it contains sections for wall-clock time, AI working time, human active time, net LOC, commits, tool calls, top tools, subagents, skills, phases, stack and data-source limitations

### Requirement: OpenCode SQLite primary source
The system SHALL use the local OpenCode SQLite database as the primary source for session and activity metrics when the database is available.

#### Scenario: Sessions are filtered to the current repository
- **WHEN** the OpenCode SQLite database contains sessions for multiple directories
- **THEN** only sessions whose project/session directory matches the current repository or the user-selected target directory are included in report metrics

#### Scenario: Missing OpenCode database degrades explicitly
- **WHEN** the OpenCode SQLite database does not exist or cannot be opened read-only
- **THEN** the report generation does not fabricate OpenCode metrics and reports affected metrics as `No disponible` with an actionable limitation message

#### Scenario: Schema mismatch degrades explicitly
- **WHEN** required SQLite tables or columns for a metric are unavailable
- **THEN** only that metric or section is marked `No disponible` and the report includes which source was unavailable without exposing raw DB content

### Requirement: Privacy by default
The system SHALL include only sanitized structural metrics by default and MUST NOT include prompts, assistant outputs, full tool payloads, file contents, diffs or `session_diff` data in the generated HTML.

#### Scenario: Sensitive conversational content is excluded
- **WHEN** sessions contain user prompts or assistant outputs
- **THEN** the default report includes aggregated counts and timing metadata only, not the prompt or output text

#### Scenario: Sensitive diff content is excluded
- **WHEN** OpenCode or related sources contain `session_diff` or tool output with diffs/file contents
- **THEN** the default report excludes those diffs and file contents from the HTML

#### Scenario: Tool calls are summarized safely
- **WHEN** tool calls are present in OpenCode activity
- **THEN** the report includes tool names, counts and top-tool aggregates but not complete arguments, outputs or payload bodies by default

### Requirement: Time heuristics are explicit
The system SHALL compute and disclose deterministic heuristics for wall-clock time, AI working time and human active time.

#### Scenario: Wall-clock time is computed from activity bounds
- **WHEN** included sessions have timestamped activity
- **THEN** wall-clock time is calculated from the earliest included timestamp to the latest included timestamp for the selected range and the report identifies the selected range

#### Scenario: AI working time uses assistant and tool activity
- **WHEN** assistant messages, tool calls or OpenCode events have timestamped activity
- **THEN** AI working time is estimated from those intervals with documented gap caps so long idle periods are not counted as continuous AI work

#### Scenario: Human active time uses user interaction intervals
- **WHEN** user messages or human-triggered events have timestamped activity
- **THEN** human active time is estimated from interaction-adjacent intervals with documented gap caps so idle time is not counted as active work

#### Scenario: Missing timestamps prevent false precision
- **WHEN** timestamps required for a time metric are missing or ambiguous
- **THEN** the metric is marked `No disponible` or `Estimación parcial` with a limitation note instead of inventing exact time

### Requirement: Git secondary metrics
The system SHALL use read-only Git metadata as a secondary source for commits and net LOC when Git is available for the target repository.

#### Scenario: Commits and LOC are reported for selected range
- **WHEN** the target directory is a Git repository and a report range is selected
- **THEN** the report includes commit count and net LOC for that range using read-only Git commands or equivalent local metadata

#### Scenario: Git unavailable degrades explicitly
- **WHEN** the target directory is not a Git repository or Git cannot be queried
- **THEN** commits and net LOC are marked `No disponible` and other non-Git report sections can still be generated

### Requirement: Stack metadata degrades gracefully
The system SHALL use `.opencode/project.yaml` for stack metadata when present and SHALL degrade gracefully when it is absent or invalid.

#### Scenario: Project YAML stack is used
- **WHEN** `.opencode/project.yaml` exists and contains stack metadata
- **THEN** the report includes the configured stack/tooling information in the stack section

#### Scenario: Project YAML is absent
- **WHEN** `.opencode/project.yaml` does not exist
- **THEN** the report uses safe heuristic stack detection when possible or reports stack as `No configurado` without failing report generation

#### Scenario: Project YAML is invalid
- **WHEN** `.opencode/project.yaml` exists but cannot be parsed
- **THEN** the report marks configured stack metadata as unavailable, includes an actionable limitation and does not overwrite the file

### Requirement: Phase, subagent and skill summaries
The system SHALL summarize phases, subagents and skills from available structural OpenCode activity without depending on sensitive content.

#### Scenario: Subagents are summarized
- **WHEN** OpenCode activity identifies subagent usage
- **THEN** the report includes subagent names and counts/durations when available without including their prompts or outputs

#### Scenario: Skills are summarized
- **WHEN** OpenCode activity or command context identifies skill usage
- **THEN** the report includes skill names and counts/durations when available without including skill inputs or generated content

#### Scenario: Phases are inferred from structural events
- **WHEN** sessions include messages, tool calls and command activity that can be grouped into work phases
- **THEN** the report includes a phase timeline using documented heuristics such as exploration, implementation, validation and delivery-readiness

### Requirement: Fixture-based validation
The implementation SHALL include validation coverage using sanitized fixtures or synthetic data for OpenCode activity, Git metrics and optional stack metadata.

#### Scenario: Sanitized fixture validates required metrics
- **WHEN** validation runs against a fixture containing sessions, messages, parts, events and Git-like metadata
- **THEN** the generated report includes the required metric sections and excludes prompts, outputs, diffs and full tool payloads

#### Scenario: Degradation fixtures validate missing sources
- **WHEN** validation runs with fixtures or setup where OpenCode DB, Git or `.opencode/project.yaml` are missing
- **THEN** the report explicitly marks the affected sections unavailable while preserving available sections

#### Scenario: Validation commands are read-only
- **WHEN** maintainers validate `lufy.timereport`
- **THEN** documented validation commands avoid network access and avoid mutating Git history, OpenCode databases or project configuration
