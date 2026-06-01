## Purpose
Definir `/lufy.timereport` como comando y skill local para generar Developer Impact Reports HTML offline sobre tiempo, actividad, decisiones, aprendizajes y ROI a partir de métricas estructurales locales, preservando privacidad por defecto.

## Requirements
### Requirement: Slash command and skill entrypoint
The system SHALL provide a `/lufy.timereport` OpenCode command that delegates to a local `lufy.timereport` skill for generating a time and ROI report.

#### Scenario: Command is discoverable
- **WHEN** a user inspects `.opencode/commands/lufy.timereport.md`
- **THEN** the command describes `/lufy.timereport`, its expected inputs, output behavior and its delegation to the `lufy.timereport` skill

#### Scenario: Skill is discoverable
- **WHEN** a user or agent inspects `.opencode/skills/lufy.timereport/`
- **THEN** a `SKILL.md` exists and describes purpose, usage, inputs, privacy boundaries, data sources, validation and outputs for the `lufy.timereport` skill

### Requirement: Offline self-contained Notion-style HTML report
The system SHALL generate a single offline self-contained HTML report with a Notion-inspired content-first style, embedded styles and no network access or remote assets.

#### Scenario: Default output path is temporary
- **WHEN** the user runs `/lufy.timereport` without specifying an output path
- **THEN** the generated report is written to a temporary path such as `/tmp/lufy-timereport-<timestamp>.html` and the path is reported to the user

#### Scenario: Configurable output path is honored
- **WHEN** the user runs `/lufy.timereport` with an explicit output path
- **THEN** the generated report is written to that path if the path is writable and no network resource is required to view it

#### Scenario: HTML contains required summary sections
- **WHEN** the report is generated successfully
- **THEN** it contains sections for task properties, executive summary, daily impact, wall-clock time, AI working time, human active time, step-by-step task context, learnings and pivots, net LOC, commits, tool calls, top tools, subagents, skills, phases, stack and data-source limitations

#### Scenario: Report uses Notion-inspired visual language
- **WHEN** the HTML report is generated
- **THEN** it uses a light warm background, subtle borders, page-property style metadata, callouts and database-like tables without external fonts, images, CDNs or remote resources

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

### Requirement: Task-scoped report by default
The system SHALL report the original user task by default, including its subagents, and SHALL require an explicit repo scope for all-repository aggregation.

#### Scenario: Default scope selects latest task root
- **WHEN** the user runs `/lufy.timereport` without `--scope`
- **THEN** the report uses `task` scope and includes the latest root OpenCode session for the target repository plus its descendant sessions by `parent_id`

#### Scenario: Child session anchor resolves to task root
- **WHEN** the user provides `--session-id` for a subagent or child session
- **THEN** the report climbs to the available root session and includes that root plus all descendant sessions for the task

#### Scenario: Repository-wide report is explicit
- **WHEN** the user provides `--scope repo`
- **THEN** the report may aggregate all sessions for the target repository and clearly labels the scope as `repo`

#### Scenario: Task scope reports workflow contract
- **WHEN** the report is generated in task scope
- **THEN** it includes scope mode, root session, anchor session, included session count, inferred time window, tier when provided or conservatively inferred, and methodology/spec when provided or detectable

#### Scenario: Git range follows task scope by default
- **WHEN** task scope is used without explicit `--from` or `--to`
- **THEN** Git commit and LOC metrics are limited to the inferred task time window instead of the full repository history

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

#### Scenario: Step timeline includes AI and human time
- **WHEN** task activity has structural user, assistant, tool or event timestamps
- **THEN** the report includes a sanitized step-by-step timeline with what type of work occurred, why it occurred, first/last activity, active duration, AI time and human time for each step

#### Scenario: Dates are readable by end users
- **WHEN** the report renders task windows, Git windows or step timestamps
- **THEN** it displays localized human-readable date/time labels instead of raw ISO timestamps

#### Scenario: Step timeline preserves privacy
- **WHEN** the report summarizes step-by-step activity
- **THEN** it uses structural labels such as exploration, implementation, validation, workflow/spec and delivery-readiness without including prompts, responses, tool arguments, tool outputs, file contents or diffs

#### Scenario: Learnings and pivots avoid fabricated intent
- **WHEN** the report includes learnings or pivots
- **THEN** it derives learnings from structural local signals and marks pivots as unavailable or partial unless an explicit structured signal exists

### Requirement: Git secondary metrics
The system SHALL use read-only Git metadata as a secondary source for commits and net LOC when Git is available for the target repository.

#### Scenario: Commits and LOC are reported for selected range
- **WHEN** the target directory is a Git repository and a report range is selected
- **THEN** the report includes commit count and net LOC for that range using read-only Git commands or equivalent local metadata

#### Scenario: Git unavailable degrades explicitly
- **WHEN** the target directory is not a Git repository or Git cannot be queried
- **THEN** commits and net LOC are marked `No disponible` and other non-Git report sections can still be generated

### Requirement: Stack metadata degrades gracefully
The system SHALL use `.lufy/project.yaml` for stack metadata when present and SHALL degrade gracefully when it is absent or invalid.

#### Scenario: Project YAML stack is used
- **WHEN** `.lufy/project.yaml` exists and contains stack metadata
- **THEN** the report includes the configured stack/tooling information in the stack section

#### Scenario: Project YAML is absent
- **WHEN** `.lufy/project.yaml` does not exist
- **THEN** the report uses safe heuristic stack detection when possible or reports stack as `No configurado` without failing report generation

#### Scenario: Project YAML is invalid
- **WHEN** `.lufy/project.yaml` exists but cannot be parsed
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
- **WHEN** validation runs with fixtures or setup where OpenCode DB, Git or `.lufy/project.yaml` are missing
- **THEN** the report explicitly marks the affected sections unavailable while preserving available sections

#### Scenario: Validation commands are read-only
- **WHEN** maintainers validate `lufy.timereport`
- **THEN** documented validation commands avoid network access and avoid mutating Git history, OpenCode databases or project configuration
