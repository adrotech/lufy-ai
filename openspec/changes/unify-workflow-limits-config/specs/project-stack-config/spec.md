## ADDED Requirements

### Requirement: Canonical workflow limits block
`.opencode/project.yaml` SHALL define workflow sizing, routing, slicing, delivery batching, stop rules and preflight controls only under top-level `workflow_limits`.

#### Scenario: Workflow limits generated for new project config
- **WHEN** the user runs `lufy-ai init --target <dir>` and `<dir>/.opencode/project.yaml` does not exist
- **THEN** the generated `.opencode/project.yaml` contains top-level `workflow_limits` with `sizing`, `routing`, `proposal_slicing_strategy`, `delivery_batch_strategy`, `stop_rules` and `preflight`

#### Scenario: Proposal slicing and delivery batching are distinct
- **WHEN** `.opencode/project.yaml` is generated or rescanned
- **THEN** `workflow_limits.proposal_slicing_strategy` defines proposal/review-slice splitting behavior and `workflow_limits.delivery_batch_strategy` defines post-validation delivery grouping behavior as separate fields

### Requirement: Legacy workflow limit fields are not accepted as canonical
Top-level `loc_budget` and `delivery_strategy` in `.opencode/project.yaml` SHALL NOT be valid workflow-limit sources after this change.

#### Scenario: Legacy top-level fields detected during rescan
- **GIVEN** `.opencode/project.yaml` contains top-level `loc_budget` or `delivery_strategy`
- **WHEN** the user runs `lufy-ai init --target <dir> --rescan`
- **THEN** the command reports that legacy workflow-limit fields are unsupported and MUST NOT treat them as canonical workflow-limit overrides

#### Scenario: Fresh generated config omits legacy fields
- **WHEN** the user runs `lufy-ai init --target <dir>` and a new `.opencode/project.yaml` is written
- **THEN** the generated config contains no top-level `loc_budget` and no top-level `delivery_strategy`

## MODIFIED Requirements

### Requirement: Project stack configuration file
`lufy-ai init` SHALL create `.opencode/project.yaml` as the editable project-local configuration file for detected stacks and operational rules.

#### Scenario: Project config created for empty target config
- **WHEN** the user runs `lufy-ai init --target <dir>` and `<dir>/.opencode/project.yaml` does not exist
- **THEN** the CLI creates `.opencode/project.yaml` with `schema_version`, `detected_at`, `stacks`, `ci`, `tdd` and `workflow_limits`
- **THEN** the CLI MUST NOT create top-level `loc_budget` or top-level `delivery_strategy`

#### Scenario: Existing project config is not overwritten by default
- **WHEN** the user runs `lufy-ai init --target <dir>` and `<dir>/.opencode/project.yaml` already exists
- **THEN** the CLI exits non-zero with an actionable message and MUST NOT overwrite the file unless `--force` or `--rescan` is used

#### Scenario: Force replaces generated project config
- **WHEN** the user runs `lufy-ai init --target <dir> --force` and `<dir>/.opencode/project.yaml` already exists
- **THEN** the CLI writes a freshly detected configuration to `.opencode/project.yaml` with `workflow_limits` as the only workflow-limit block
- **THEN** the freshly detected configuration MUST NOT include top-level `loc_budget` or top-level `delivery_strategy`

### Requirement: Rescan preserves user overrides
`lufy-ai init --rescan` SHALL merge newly detected stack evidence into an existing `.opencode/project.yaml` without discarding user-managed preferences under supported editable boundaries.

#### Scenario: Coverage override preserved
- **WHEN** `.opencode/project.yaml` contains `coverage_threshold: 70` for stack `go` and the user runs `lufy-ai init --rescan`
- **THEN** the resulting config preserves `coverage_threshold: 70` for stack `go`

#### Scenario: Workflow limits override preserved
- **WHEN** `.opencode/project.yaml` contains user-managed overrides under `workflow_limits` and the user runs `lufy-ai init --rescan`
- **THEN** the resulting config preserves those `workflow_limits` overrides while refreshing detected stack, tooling or CI evidence as applicable

#### Scenario: New stack added on rescan
- **WHEN** an existing config contains only stack `go` and the target later gains `package.json` and `tsconfig.json`
- **THEN** `lufy-ai init --rescan` preserves the Go stack and adds a TypeScript stack

#### Scenario: Removed stack is deprecated not deleted
- **WHEN** an existing config contains stack `python` but Python marker files are no longer present
- **THEN** `lufy-ai init --rescan` preserves the stack entry and marks it deprecated or otherwise reports it without deleting user overrides

## REMOVED Requirements

### Requirement: Top-level legacy workflow limit fields
**Reason**: Top-level `loc_budget` and `delivery_strategy` create competing workflow-limit sources and are replaced by the canonical `workflow_limits` block.

**Migration**: Move any intentional workflow-limit overrides into `workflow_limits.sizing`, `workflow_limits.routing`, `workflow_limits.proposal_slicing_strategy`, `workflow_limits.delivery_batch_strategy`, `workflow_limits.stop_rules` or `workflow_limits.preflight` before relying on the config.

#### Scenario: Legacy workflow limits removed from generated schema
- **WHEN** `.opencode/project.yaml` is generated after this change
- **THEN** top-level `loc_budget` and top-level `delivery_strategy` are not part of the generated schema
