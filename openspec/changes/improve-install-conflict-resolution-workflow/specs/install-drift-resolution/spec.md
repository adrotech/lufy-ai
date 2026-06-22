## ADDED Requirements

### Requirement: Structured conflict planning

The CLI SHALL expose a read-only conflict plan for unmanaged or drifted managed assets before applying install mutations.

#### Scenario: Conflict plan groups files
- **WHEN** an install target contains unmanaged conflicting files
- **THEN** the conflict plan groups each file into a useful category such as `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/templates`, `openspec/specs`, or `root/config`

#### Scenario: Conflict plan recommends safe actions
- **WHEN** a conflict plan item is emitted
- **THEN** it includes path, status, risk, reason, recommended action, and available actions
- **AND** destructive actions remain recommendations only until explicitly confirmed by a separate supported workflow

#### Scenario: JSON plan is automation-friendly
- **WHEN** the user runs the conflict plan command with JSON output
- **THEN** the CLI emits parseable JSON without human logs mixed in and without mutating target files

#### Scenario: Interactive setup presents the same plan
- **WHEN** `lufy-ai setup` runs interactively and detects install conflicts
- **THEN** setup presents a Bubble Tea read-only conflict review using the same grouped conflict plan contract
- **AND** setup does not apply install mutations until conflicts are resolved by an explicit supported workflow

### Requirement: Legacy asset cleanup remains explicit

The CLI SHALL distinguish active managed assets with legacy names from deprecated unused paths.

#### Scenario: Active lufy-ia harness is not deleted
- **WHEN** the target or source contains `lufy-ia.harness.md`
- **THEN** the CLI treats it as an active managed harness asset unless a future migration explicitly replaces it

#### Scenario: Deprecated layout paths are reported safely
- **WHEN** legacy `.lufy-ai/*` paths are detected
- **THEN** the conflict plan reports them as deprecated layout paths and recommends `lufy-ai migrate-layout --target <dir> --dry-run` before any cleanup
