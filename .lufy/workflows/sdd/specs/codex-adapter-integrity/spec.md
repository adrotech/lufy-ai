# codex-adapter-integrity Specification

### Requirement: Explicit Codex selection is persisted consistently
The system SHALL persist the effective tool and methodology selection in project config and install state when a mutating command applies a Codex installation.

#### Scenario: Fresh Codex install persists one effective selection
- **WHEN** a user installs with `--tool codex` and explicit methodology tier overrides
- **THEN** `.lufy/config/project.yaml` and `.lufy/managed-state/install-state.json` record the same tool and methodology selections
- **AND** commands invoked later without repeated tool flags resolve Codex.

#### Scenario: Existing project preferences are preserved
- **WHEN** install updates tool or methodology in an existing valid project config
- **THEN** stack, surface, workflow limit, memory, context and unknown user-managed fields remain unchanged unless explicitly targeted.

### Requirement: Harness resolution has explicit precedence and provenance
The system SHALL resolve HarnessConfig with deterministic precedence and SHALL expose the source of each effective decision in diagnostics when ambiguity matters.

#### Scenario: Explicit command flags win
- **WHEN** a command supplies a valid explicit tool or methodology override
- **THEN** that value wins over project config, install state and defaults for the command
- **AND** a mutating command persists the resulting supported selection where its contract requires persistence.

#### Scenario: Installed repository has contradictory sources
- **WHEN** project config and install state disagree and no explicit flag resolves the intent
- **THEN** mutating asset commands fail without changing files
- **AND** report both values and a non-destructive recovery action.

### Requirement: Verify and doctor reconcile intent with installed reality
The system SHALL compare project config selection with install state selection in verification and diagnostics.

#### Scenario: Tool mismatch is a failure
- **WHEN** project config says `opencode` and install state says `codex`
- **THEN** `verify` reports a failure identifying expected and actual tool
- **AND** `doctor` does not report the installation as fully healthy.

#### Scenario: Methodology mismatch is a failure
- **WHEN** any tier differs between project config and install state
- **THEN** diagnostics identify the tier and both selections
- **AND** recovery preserves user-managed config fields and managed asset rollback data.

### Requirement: Project config and install state writes are failure-safe
The system SHALL avoid accepting a partially updated configuration pair as a successful installation.

#### Scenario: Install fails after project config merge
- **WHEN** an error prevents catalog application or install state persistence
- **THEN** the command returns a recovery path for the previous coherent state
- **AND** post-install verify cannot report success for the partial state.
