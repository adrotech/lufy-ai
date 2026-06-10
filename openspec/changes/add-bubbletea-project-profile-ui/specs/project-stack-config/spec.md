## MODIFIED Requirements

### Requirement: Project surface profile
`.lufy/config/project.yaml` SHALL include a project surface profile that tells agents what product lens to apply independently from the technical stack.

#### Scenario: Surface profile is generated from detectable evidence
- **WHEN** `lufy-ai init --target <dir>` detects frontend, backend, mobile, CLI, infra, library or fullstack evidence
- **THEN** `.lufy/config/project.yaml` includes `project_profile.surfaces` entries with `id`, `type`, `roots`, `stacks`, `frameworks` and `agent_lens`

#### Scenario: Surface profile can be adjusted with a TUI
- **WHEN** the user runs `lufy-ai init --target <dir> --interactive` or `lufy-ai scan --target <dir>` in an interactive terminal
- **THEN** the CLI opens a Bubble Tea/Charm TUI for reviewing detected surfaces before writing `.lufy/config/project.yaml`
- **THEN** the user can select a surface and adjust its `type` among supported surface types
- **THEN** the written surface uses the selected `type` and matching `agent_lens`

#### Scenario: TUI cancellation is non-mutating
- **WHEN** the user cancels the project profile TUI before confirming
- **THEN** the CLI exits non-zero with an actionable cancellation message
- **THEN** the CLI MUST NOT write or partially rewrite `.lufy/config/project.yaml`

#### Scenario: Surface profile is automation-safe
- **WHEN** the CLI runs in a non-interactive environment
- **THEN** it preserves the automatically detected surface profile and does not block waiting for input
- **THEN** the CLI MUST NOT create top-level `loc_budget` or top-level `delivery_strategy`

## ADDED Requirements

### Requirement: Project profile TUI adapter boundary
The interactive project profile UI SHALL be implemented as a CLI adapter over `projectconfig.ProfilePrompt` and SHALL NOT move Bubble Tea dependencies into `internal/projectconfig`.

#### Scenario: Projectconfig remains UI-independent
- **WHEN** project profile TUI support is added
- **THEN** `internal/projectconfig` imports no Bubble Tea, Bubbles or Lip Gloss packages
- **THEN** `internal/projectconfig` remains usable by non-interactive tests and automation without terminal dependencies

#### Scenario: CLI wires TUI through ProfilePrompt
- **WHEN** `lufy-ai init --interactive` or interactive `lufy-ai scan` needs user input
- **THEN** `internal/cli` wires a TUI-backed `projectconfig.ProfilePrompt` into `projectconfig.Service`
- **THEN** the service continues to coordinate scan, rescan merge, optional profile prompting and persistence through the existing port
