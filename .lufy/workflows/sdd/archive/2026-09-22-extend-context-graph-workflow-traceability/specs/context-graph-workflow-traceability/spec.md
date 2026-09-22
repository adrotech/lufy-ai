# context-graph-workflow-traceability Specification Delta

## Intent

Define trazabilidad ejecutable y carga de revisión observable sobre el Context Graph, preservando privacidad, freshness y autoridad de los archivos/comandos/roles reales.

## ADDED Requirements

### Requirement: Semantic workflow nodes

LUFY SHALL extract deterministic nodes for changes, specs, requirements, scenarios, tasks, decisions, tests, failures, runs, evidence, reviews, issues and pull requests from supported local sources.

#### Scenario: Structured SDD artifact is indexed

- **GIVEN** a valid Lufy SDD proposal, spec, tasks or verification artifact
- **WHEN** the context graph is built
- **THEN** semantic nodes use stable IDs, source paths, spans when available and explicit provenance.

#### Scenario: Ambiguous text is not promoted

- **GIVEN** prose that only resembles a task, test, PR or scenario
- **WHEN** the source is extracted
- **THEN** it remains lexical context and does not become a workflow evidence node without recognized structure.

### Requirement: Explicit workflow edges

LUFY SHALL support `implements`, `verifies`, `depends_on`, `caused_by`, `reviewed_by` and `supersedes` edges only from recognized structure or explicit bounded references.

#### Scenario: Valid explicit reference creates an edge

- **GIVEN** a supported artifact with a valid `lufy:` trace marker or structured reference
- **WHEN** both endpoints are resolvable
- **THEN** the graph contains one deterministic typed edge with reason and provenance.

#### Scenario: Broken reference remains a gap

- **GIVEN** an explicit reference whose target is absent or invalid
- **WHEN** the graph is built
- **THEN** no evidence edge is created and a bounded diagnostic identifies the unresolved reference.

#### Scenario: Similar names do not prove traceability

- **GIVEN** a test and scenario with similar labels but no explicit relationship
- **WHEN** extraction runs
- **THEN** LUFY does not create `verifies` or `implements` edges by fuzzy matching.

### Requirement: Workflow coverage query

LUFY SHALL expose a bounded coverage query that distinguishes covered, gap and unknown states for scenario-to-task and scenario-to-test relationships.

#### Scenario: Scenario without task or test is reported

- **GIVEN** a ready graph containing a scenario without incoming `implements` or `verifies` edges
- **WHEN** `lufy-ai context coverage` runs
- **THEN** the result reports the scenario ID, missing relationship, source and actionable recovery.

#### Scenario: Explicit task and test satisfy coverage

- **GIVEN** a scenario with resolvable task and test references
- **WHEN** coverage is queried
- **THEN** the scenario is reported covered with the supporting edge IDs and no inferred evidence.

### Requirement: Diff traceability assessment

LUFY SHALL compare changed files with explicit workflow paths and report every untraced changed file.

#### Scenario: Changed file has no workflow path

- **GIVEN** a changed source file with no `implements` or `verifies` path to a task/scenario
- **WHEN** review assessment runs
- **THEN** the file is listed as `untraced_file` and traceability cannot be reported complete.

#### Scenario: Changed file is explicitly traced

- **GIVEN** a changed file connected through explicit workflow edges
- **WHEN** review assessment runs
- **THEN** the result includes the supporting path and marks that file traced.

### Requirement: Typed review workload limits

`.lufy/config/project.yaml` SHALL define review workload limits only under `workflow_limits.review`, including maximum files per slice, maximum churn lines per slice, maximum concurrent slices and minimum evidence items.

#### Scenario: New config receives review defaults

- **WHEN** LUFY generates a project config
- **THEN** it emits `max_files_per_slice: 8`, `max_churn_lines_per_slice: 800`, `max_concurrent_slices: 3` and `min_evidence_items: 2` under `workflow_limits.review`.

#### Scenario: Rescan preserves review overrides

- **GIVEN** a partial or customized `workflow_limits.review`
- **WHEN** rescan completes missing defaults
- **THEN** valid overrides and unknown nested extras are preserved.

#### Scenario: Canonical review limits are unavailable

- **GIVEN** project config or `workflow_limits.review` is absent
- **WHEN** a workload result is produced
- **THEN** every unavailable budget is reported as `not_available` and top-level legacy fields are not consumed.

### Requirement: Deterministic review workload decision

LUFY SHALL evaluate observed file count, churn, concurrent slices, evidence count and traceability against canonical review limits without authorizing delivery.

#### Scenario: Slice fits all available budgets

- **WHEN** observations are within limits, evidence is sufficient and traceability is complete
- **THEN** the assessment recommends `proceed` with the evaluated rules and evidence.

#### Scenario: Slice exceeds a splittable budget

- **WHEN** file count or churn exceeds its configured limit
- **THEN** the assessment recommends `split` and identifies the exact observed and allowed values.

#### Scenario: Concurrency or evidence creates workflow risk

- **WHEN** concurrent slices exceed the limit, required evidence is missing or traceability is unknown
- **THEN** the assessment recommends `escalate` and does not report delivery readiness.

### Requirement: Freshness-aware safe fallback

Every coverage or review query SHALL declare graph freshness and SHALL NOT treat stale or missing graph data as primary evidence.

#### Scenario: Ready graph supports traceability

- **WHEN** graph inputs and extractor version match the manifest
- **THEN** coverage and trace paths may be evaluated from the graph while retaining their source provenance.

#### Scenario: Stale graph degrades review safely

- **WHEN** the graph is stale
- **THEN** direct Git/config observations remain available, graph-dependent coverage becomes `unknown`, and recovery recommends an explicit rebuild.

#### Scenario: Missing graph never passes coverage

- **WHEN** no graph is available
- **THEN** coverage returns `unknown/not_available` rather than an empty successful result.

### Requirement: Content-free run projection

LUFY SHALL project workflow runs and causal relationships from validated Run Ledger metadata without adding raw runtime sources to generic discovery.

#### Scenario: Run metadata creates causal nodes

- **GIVEN** valid content-free Run Ledger events or projections
- **WHEN** workflow projection runs
- **THEN** run, task, evidence and `caused_by` relationships are added using bounded metadata and deterministic IDs.

#### Scenario: Runtime content stays excluded

- **WHEN** the context graph and reports are persisted
- **THEN** they contain no prompts, outputs, summaries, secrets, event bodies or raw runtime paths.

#### Scenario: Ledger unavailable does not break static graph

- **WHEN** Run Ledger is disabled, missing or invalid
- **THEN** static graph build continues and run-derived coverage/metrics are marked unavailable with recovery.

### Requirement: Review outcome metrics

LUFY SHALL derive review duration, rework count and reopened defect count only from recognized content-free event metadata and SHALL report metric availability.

#### Scenario: Complete review events produce metrics

- **GIVEN** paired review start/completion events plus rework or reopened-defect events
- **WHEN** `lufy-ai context metrics` runs
- **THEN** it returns deterministic durations and counts with source availability.

#### Scenario: Historical data is incomplete

- **WHEN** required events or timestamps are absent
- **THEN** metrics are `partial` or `unavailable` and missing values are not synthesized.

### Requirement: Harness consumes graph as secondary evidence

Router, orchestrator, reviewer, validator and delivery guidance SHALL consume workflow coverage and review workload results while preserving role authority and explicit delivery authorization.

#### Scenario: Reviewer receives bounded gaps

- **WHEN** a review assessment contains gaps or budget violations
- **THEN** reviewer output prioritizes affected slice, files, scenarios and recovery without broad repository rediscovery.

#### Scenario: Graph cannot advance a gate

- **WHEN** coverage or workload assessment reports success
- **THEN** validation, review score, remote checks, delivery authorization and closure remain separately required.

#### Scenario: Installed assets remain synchronized

- **WHEN** workflow traceability behavior changes
- **THEN** root instructions, embedded assets, specs and user-facing documentation describe the same contracts and fallback semantics.
