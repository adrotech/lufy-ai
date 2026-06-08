## Purpose
Definir `.lufy/project.yaml` como configuración project-local editable para stacks detectados, metadata operacional y preferencias preservadas por `lufy-ai init`.

## Requirements
### Requirement: Project stack configuration file
`lufy-ai init` SHALL create `.lufy/project.yaml` as the editable project-local configuration file for detected stacks, project surfaces and operational rules.

#### Scenario: Project config created for empty target config
- **WHEN** the user runs `lufy-ai init --target <dir>` and `<dir>/.lufy/project.yaml` does not exist
- **THEN** the CLI creates `.lufy/project.yaml` with `schema_version`, `detected_at`, `project_profile`, `stacks`, `ci`, `tdd` and `workflow_limits`

### Requirement: Project surface profile
`.lufy/project.yaml` SHALL include a project surface profile that tells agents what product lens to apply independently from the technical stack.

#### Scenario: Surface profile is generated from detectable evidence
- **WHEN** `lufy-ai init --target <dir>` detects frontend, backend, mobile, CLI, infra, library or fullstack evidence
- **THEN** `.lufy/project.yaml` includes `project_profile.surfaces` entries with `id`, `type`, `roots`, `stacks`, `frameworks` and `agent_lens`
- **AND** `agent_lens` includes `structural_expectations` when the surface has expected folder, layer or boundary conventions

#### Scenario: Surface profile can be adjusted interactively
- **WHEN** the user runs `lufy-ai init --target <dir> --interactive` or `lufy-ai scan --target <dir>` in an interactive terminal
- **THEN** the CLI prompts for the primary project surface and writes the selected `agent_lens`
- **AND** when the selected surface is `backend`, the CLI writes `architecture.preferred`, `architecture.options` and `architecture.structural_expectations` for the selected backend architecture

#### Scenario: Frontend profile records feature-driven structure
- **WHEN** `lufy-ai init` or `lufy-ai scan` writes a `frontend` or frontend side of a `fullstack` surface
- **THEN** the surface records structural expectations for feature-driven colocation such as `src/features/<feature>/components`, `hooks`, `services`, `types.ts`, `index.ts` public barrels and routing/layout pages outside feature internals

#### Scenario: Backend profile records selected architecture structure
- **WHEN** `lufy-ai init` or `lufy-ai scan` writes a `backend` surface
- **THEN** the surface records the selected backend architecture as `controller_service_repository`, `clean_architecture` or `hexagonal`
- **AND** `architecture.structural_expectations` records the concrete layer or boundary checks that implementer, validator and reviewer must audit before approval

#### Scenario: Surface profile is automation-safe
- **WHEN** the CLI runs in a non-interactive environment
- **THEN** it preserves the automatically detected surface profile and does not block waiting for input
- **THEN** the CLI MUST NOT create top-level `loc_budget` or top-level `delivery_strategy`

#### Scenario: Existing project config is not overwritten by default
- **WHEN** the user runs `lufy-ai init --target <dir>` and `<dir>/.lufy/project.yaml` already exists
- **THEN** the CLI exits non-zero with an actionable message and MUST NOT overwrite the file unless `--force` or `--rescan` is used

#### Scenario: Force replaces generated project config
- **WHEN** the user runs `lufy-ai init --target <dir> --force` and `<dir>/.lufy/project.yaml` already exists
- **THEN** the CLI writes a freshly detected configuration to `.lufy/project.yaml` with `workflow_limits` as the only workflow-limit block
- **THEN** the freshly detected configuration MUST NOT include top-level `loc_budget` or top-level `delivery_strategy`

### Requirement: Canonical workflow limits block
`.lufy/project.yaml` SHALL define workflow sizing, routing, slicing, delivery batching, stop rules and preflight controls only under top-level `workflow_limits`.

#### Scenario: Workflow limits generated for new project config
- **WHEN** the user runs `lufy-ai init --target <dir>` and `<dir>/.lufy/project.yaml` does not exist
- **THEN** the generated `.lufy/project.yaml` contains top-level `workflow_limits` with `sizing`, `routing`, `proposal_slicing_strategy`, `delivery_batch_strategy`, `stop_rules` and `preflight`

#### Scenario: Proposal slicing and delivery batching are distinct
- **WHEN** `.lufy/project.yaml` is generated or rescanned
- **THEN** `workflow_limits.proposal_slicing_strategy` defines proposal/review-slice splitting behavior and `workflow_limits.delivery_batch_strategy` defines post-validation delivery grouping behavior as separate fields

### Requirement: Legacy workflow limit fields are not accepted as canonical
Top-level `loc_budget` and `delivery_strategy` in `.lufy/project.yaml` SHALL NOT be valid workflow-limit sources.

#### Scenario: Legacy top-level fields detected during rescan
- **GIVEN** `.lufy/project.yaml` contains top-level `loc_budget` or `delivery_strategy`
- **WHEN** the user runs `lufy-ai init --target <dir> --rescan`
- **THEN** the command reports that legacy workflow-limit fields are unsupported and MUST NOT treat them as canonical workflow-limit overrides

#### Scenario: Fresh generated config omits legacy fields
- **WHEN** the user runs `lufy-ai init --target <dir>` and a new `.lufy/project.yaml` is written
- **THEN** the generated config contains no top-level `loc_budget` and no top-level `delivery_strategy`

### Requirement: Supported stack detection
`lufy-ai init` SHALL detect supported v1 stacks from repository files without executing project toolchains.

#### Scenario: Go stack detected
- **WHEN** the target contains `go.mod`
- **THEN** `.lufy/project.yaml` includes a stack with `id: go`, `supported: true`, package manager `go modules`, Go test, formatter, static analysis and default coverage settings

#### Scenario: TypeScript React Next stack detected
- **WHEN** the target contains `package.json`, `tsconfig.json` and dependencies for `react` and `next`
- **THEN** `.lufy/project.yaml` includes a stack with `id: typescript`, `supported: true`, frameworks including `react` and `next`, and detected JavaScript package manager, test runner, linter and formatter when present

#### Scenario: JavaScript stack detected without TypeScript
- **WHEN** the target contains `package.json` and does not contain `tsconfig.json`
- **THEN** `.lufy/project.yaml` includes a stack with `id: javascript`, `supported: true` and JavaScript tooling inferred from `package.json`

#### Scenario: Python stack detected
- **WHEN** the target contains `pyproject.toml`, `requirements.txt` or `setup.py`
- **THEN** `.lufy/project.yaml` includes a stack with `id: python`, `supported: true` and Python test, lint, format and static analysis commands inferred from available files

#### Scenario: Java or Kotlin stack detected
- **WHEN** the target contains `pom.xml`, `build.gradle` or `build.gradle.kts`
- **THEN** `.lufy/project.yaml` includes a supported JVM stack with Maven or Gradle commands inferred from the build file

### Requirement: Unsupported stack placeholders
`lufy-ai init` SHALL report known but unsupported stacks as editable placeholders instead of failing initialization.

#### Scenario: Rust stack detected as unsupported
- **WHEN** the target contains `Cargo.toml` and no supported stack evidence is required for success
- **THEN** `.lufy/project.yaml` includes a stack with `id: rust`, `supported: false`, placeholder commands and notes that official support is pending

#### Scenario: Unknown files do not block init
- **WHEN** the target contains files for stacks not recognized by `lufy-ai init`
- **THEN** the CLI completes detection for known stacks and MUST NOT fail solely because unknown technology files exist

### Requirement: Tooling and CI metadata
`lufy-ai init` SHALL include detected test runner, formatter, linter, static analysis, observability and CI metadata in `.lufy/project.yaml`.

#### Scenario: JavaScript tooling inferred from package manifest
- **WHEN** `package.json` declares `vitest`, `jest`, `mocha`, `eslint` or `prettier` in dependencies, devDependencies or scripts
- **THEN** the generated stack includes corresponding `test_runner`, `linter` and `formatter` commands when they can be inferred

#### Scenario: CI provider detected
- **WHEN** the target contains `.github/workflows`, `.gitlab-ci.yml`, `Jenkinsfile` or `.circleci/config.yml`
- **THEN** `.lufy/project.yaml` includes `ci.detected: true`, provider metadata and known workflow paths

#### Scenario: Observability libraries detected
- **WHEN** known observability libraries are present in stack manifests
- **THEN** the generated stack includes those libraries in `observability_libs`

### Requirement: Stack-aware format dispatch hook
The installed harness SHALL provide a local format-dispatch hook that uses `.lufy/project.yaml` formatter and linter autofix metadata for changed files without assuming a fixed language toolchain.

#### Scenario: Hook formats supported stack files
- **WHEN** `.opencode/hooks/format-dispatch.sh` receives a changed `.go`, `.ts`, `.tsx` or `.py` file whose extension matches a supported stack formatter in `.lufy/project.yaml`
- **THEN** it runs the configured formatter command for that file and runs configured linter `auto_fix` when present

#### Scenario: Hook ignores unknown or unsupported files quietly
- **WHEN** the hook receives a file with an unmatched extension, a file from an unsupported stack, a missing `.lufy/project.yaml`, or an empty/TODO formatter command
- **THEN** it exits with code 0 without noisy output

#### Scenario: Hook stays confined to the project root
- **WHEN** the hook receives an absolute or relative file path outside the configured project root
- **THEN** it exits with code 0 without formatting that file

### Requirement: Rescan preserves user overrides
`lufy-ai init --rescan` SHALL merge newly detected stack and surface evidence into an existing `.lufy/project.yaml` without discarding user-managed preferences.

#### Scenario: Coverage override preserved
- **WHEN** `.lufy/project.yaml` contains `coverage_threshold: 70` for stack `go` and the user runs `lufy-ai init --rescan`
- **THEN** the resulting config preserves `coverage_threshold: 70` for stack `go`

#### Scenario: Workflow limits override preserved
- **WHEN** `.lufy/project.yaml` contains user-managed overrides under `workflow_limits` and the user runs `lufy-ai init --rescan`
- **THEN** the resulting config preserves those `workflow_limits` overrides while refreshing detected stack, tooling or CI evidence as applicable

#### Scenario: Surface overrides are preserved
- **WHEN** `.lufy/project.yaml` contains user-managed `project_profile.surfaces` entries and the user runs `lufy-ai init --rescan`
- **THEN** the resulting config preserves existing surface entries and adds newly detected surface entries without overwriting manual `agent_lens` choices

#### Scenario: New stack added on rescan
- **WHEN** an existing config contains only stack `go` and the target later gains `package.json` and `tsconfig.json`
- **THEN** `lufy-ai init --rescan` preserves the Go stack and adds a TypeScript stack

#### Scenario: Removed stack is deprecated not deleted
- **WHEN** an existing config contains stack `python` but Python marker files are no longer present
- **THEN** `lufy-ai init --rescan` preserves the stack entry and marks it deprecated or otherwise reports it without deleting user overrides

### Requirement: Rescan drift validation
`lufy-ai init --rescan` SHALL compare an existing `.lufy/project.yaml` with the target repository's current stack, tooling and CI evidence before deciding whether to write changes.

#### Scenario: No drift keeps project config unchanged
- **GIVEN** `.lufy/project.yaml` matches the current detectable stack, tooling and CI evidence
- **WHEN** the user runs `lufy-ai init --target <dir> --rescan`
- **THEN** the CLI reports no drift and MUST NOT rewrite `.lufy/project.yaml`

#### Scenario: New stack evidence is reported and merged
- **GIVEN** `.lufy/project.yaml` contains stack `go`
- **WHEN** the target later contains TypeScript/Next marker files and the user runs `lufy-ai init --target <dir> --rescan`
- **THEN** the CLI reports a new stack drift item for `typescript` and writes a merged `.lufy/project.yaml` that preserves the existing Go stack and user-managed fields

#### Scenario: Tooling drift is reported per stack
- **GIVEN** `.lufy/project.yaml` contains a JavaScript stack without a detected test runner
- **WHEN** `package.json` later declares `vitest` and the user runs `lufy-ai init --target <dir> --rescan`
- **THEN** the CLI reports tooling drift for the JavaScript stack and records the newly detected test runner without removing unrelated stack preferences

### Requirement: Stale stack detection is non-destructive
`lufy-ai init --rescan` SHALL detect configured stacks whose marker files are no longer present and SHALL treat them as stale evidence rather than implicit deletion requests.

#### Scenario: Missing marker files mark a stack stale
- **GIVEN** `.lufy/project.yaml` contains stack `python` with user-managed preferences
- **WHEN** the target no longer contains `pyproject.toml`, `requirements.txt` or `setup.py` and the user runs `lufy-ai init --target <dir> --rescan`
- **THEN** the CLI reports the Python stack as stale or deprecated and MUST preserve the stack entry and its user-managed preferences

#### Scenario: Stale stack does not delete files
- **GIVEN** a stale stack is detected during `lufy-ai init --rescan`
- **WHEN** the command completes successfully
- **THEN** the CLI MUST NOT delete source files, configuration files, `.lufy/project.yaml` entries or unknown user files as part of stale handling

### Requirement: Rescan preserves editable configuration boundaries
`lufy-ai init --rescan` SHALL update only generated or detected fields that are safe to refresh and SHALL preserve editable user-managed fields by default.

#### Scenario: Unknown fields are preserved
- **GIVEN** `.lufy/project.yaml` contains an unknown top-level field and an unknown field under stack `go`
- **WHEN** the user runs `lufy-ai init --target <dir> --rescan`
- **THEN** the resulting `.lufy/project.yaml` preserves both unknown fields unless the command fails before writing

#### Scenario: User override wins over regenerated default
- **GIVEN** `.lufy/project.yaml` contains a user override such as `coverage_threshold: 70` for stack `go`
- **WHEN** current Go detection would otherwise use a different default threshold and the user runs `lufy-ai init --target <dir> --rescan`
- **THEN** the resulting `.lufy/project.yaml` preserves `coverage_threshold: 70` and reports that the user-managed override was preserved when reporting is verbose enough to include preserved decisions

### Requirement: Rescan report is structured and actionable
`lufy-ai init --rescan` SHALL emit a structured human-readable report that identifies drift decisions and suggested next actions.

#### Scenario: Report includes actionable drift fields
- **WHEN** `lufy-ai init --target <dir> --rescan` detects any stack, tooling, CI or stale drift
- **THEN** the report includes each item's category, severity, affected stack or field path, applied status and suggested action

#### Scenario: Report distinguishes detection from mutation
- **WHEN** `lufy-ai init --target <dir> --rescan` detects stale evidence that is not safe to remove automatically
- **THEN** the report marks the item as detected or skipped and MUST NOT imply that cleanup or deletion was performed

#### Scenario: Invalid existing config fails without mutation
- **WHEN** `.lufy/project.yaml` exists but is not parseable as the supported configuration format and the user runs `lufy-ai init --target <dir> --rescan`
- **THEN** the CLI exits non-zero, reports an actionable parse error and MUST NOT overwrite the existing file
