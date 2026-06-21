## ADDED Requirements

### Requirement: Native context graph schema

LUFY SHALL define a local context graph artifact using schema `lufy-context-graph` with deterministic nodes, edges, sources, health, communities, important nodes and derived manifest information.

#### Scenario: Graph artifact declares schema version
- **WHEN** `lufy-ai context build` writes `.lufy/context/graph.json`
- **THEN** the artifact SHALL include `schema: "lufy-context-graph"`, stable workspace-relative node ids, deterministic edge ordering and enough metadata to verify staleness

#### Scenario: Graph avoids local-only absolute ids
- **WHEN** nodes or edges are serialized into `graph.json`
- **THEN** their primary ids SHALL be stable across machines by using workspace-relative paths or logical symbols instead of absolute local paths

### Requirement: Context graph persisted artifacts

LUFY SHALL use `.lufy/config/project.yaml` as the canonical configuration source for context graph, memory and vault settings, and SHALL persist graph outputs as derived local artifacts that can be inspected, rebuilt and validated.

#### Scenario: Project config owns graph and vault settings
- **WHEN** `lufy-ai init` or rescan writes `.lufy/config/project.yaml`
- **THEN** it SHALL include canonical `context_graph` settings and `memory.vault` without requiring separate graph, vault or memory configuration files

#### Scenario: Derived artifacts are not canonical config
- **WHEN** `lufy-ai context build` writes `manifest.json`, cache files or report files
- **THEN** those files SHALL be treated as derived/regenerable state, not as canonical configuration

#### Scenario: Build writes graph and summary
- **WHEN** `lufy-ai context build` completes successfully
- **THEN** `graph.json`, `graph-summary.md` and `GRAPH_REPORT.md` SHALL exist under `context_graph.root` and describe the same graph schema and source set

#### Scenario: Build remains idempotent
- **WHEN** input files, extractor versions and build options have not changed since the previous build
- **THEN** manifest or cache information SHALL allow the command to avoid unnecessary content churn in `.lufy/context/` artifacts

### Requirement: Deterministic initial extractors

LUFY SHALL provide deterministic initial extractors for Go, Markdown, YAML and JSON without relying on LLM or semantic embedding services by default.

#### Scenario: Go extractor uses parser and AST
- **WHEN** a Go source file is scanned
- **THEN** the extractor SHALL use Go parser/AST APIs to emit packages, imports, types, functions, methods and test-related nodes where present

#### Scenario: Structured text extractors emit conservative relationships
- **WHEN** Markdown, YAML or JSON files are scanned
- **THEN** the extractors SHALL emit ordered nodes and conservative references based on headings, keys, structure and explicit relative links or paths

### Requirement: Context CLI command suite

LUFY SHALL expose context graph operations through the Go CLI as `lufy-ai context scan/status/build/query/path/explain/diff` with bounded outputs designed to reduce broad file reads.

#### Scenario: Status reports graph availability
- **WHEN** `lufy-ai context status` runs in a workspace with no readable valid graph
- **THEN** it SHALL report `not_available` with a recovery hint such as `lufy-ai context build`

#### Scenario: Query returns deterministic matches
- **WHEN** `lufy-ai context query <term>` runs against a valid graph
- **THEN** it SHALL return ranked deterministic matches with node ids, labels, types, reasons, scores, bounded neighboring context and a token-savings summary

#### Scenario: Path and explain provide traceability
- **WHEN** `lufy-ai context path <from> <to>` or `lufy-ai context explain <node-or-path>` runs against a valid graph
- **THEN** the output SHALL explain the selected path, node or edge using source files, spans or extractor reasons when available

### Requirement: Diff impact analysis

LUFY SHALL support `lufy-ai context diff --base <ref>` to map changed files from a Git diff to graph nodes, neighbors and potentially affected specs, skills, agents or CLI areas.

#### Scenario: Diff impact maps changed files to graph neighborhoods
- **WHEN** `lufy-ai context diff --base origin/develop` runs with a valid graph and Git diff available
- **THEN** it SHALL report changed graph nodes, directly connected neighbors, impacted communities and explainable impact hints derived from structural edges

#### Scenario: Diff impact degrades without graph
- **WHEN** `lufy-ai context diff --base origin/develop` runs without a valid graph
- **THEN** it SHALL report `not_available` for graph-derived impact and SHALL NOT block normal Git diff inspection by other tools or agents

### Requirement: Context search skills for agent tooling

LUFY SHALL provide an OpenCode skill `lufy.context-search` and, when required by the installed catalog, an equivalent Codex/local-agent skill under `.agents` for querying the native context graph.

#### Scenario: OpenCode skill returns compact hints
- **WHEN** an OpenCode agent invokes `lufy.context-search` with a query, path or diff base
- **THEN** the skill SHALL call the appropriate `lufy-ai context` command and return compact hints rather than full graph dumps

#### Scenario: Equivalent local-agent skill follows same fallback
- **WHEN** the Codex/local-agent catalog requires a `.agents` skill for context graph search
- **THEN** the equivalent skill SHALL expose the same query/path/explain/diff behavior and the same `not_available` fallback semantics

### Requirement: Agent hint integration with graceful fallback

LUFY SHALL integrate optional context graph hints into `explorer`, `sdd-router` and `reviewer` while preserving their existing responsibilities and fallback workflows.

#### Scenario: Explorer uses hints before broad search
- **WHEN** `explorer` starts a non-trivial investigation and a valid graph is available
- **THEN** it MAY gather compact `context_graph_hints` before broader file search while still verifying conclusions against repository files

#### Scenario: Router does not depend on graph availability
- **WHEN** `sdd-router` classifies a request and the graph is missing or stale
- **THEN** it SHALL record graph hints as `not_available` and continue routing from explicit request, repository policy and available read-only evidence

#### Scenario: Reviewer treats graph as secondary evidence
- **WHEN** `reviewer` uses graph hints for impact or path analysis
- **THEN** it SHALL keep scoring and recommendations grounded in diff, tests, specs and static review rather than treating graph hints as authoritative evidence

### Requirement: Semantic analysis remains optional future phase

LUFY SHALL keep LLM, embedding or semantic ranking features out of the default `lufy-context-graph` implementation and reserve them for an explicit future opt-in phase.

#### Scenario: Default build avoids semantic services
- **WHEN** `lufy-ai context build` runs with default options
- **THEN** it SHALL NOT call LLMs, embedding APIs, remote semantic services or network-dependent ranking providers

#### Scenario: Future semantic data is isolated
- **WHEN** a future proposal adds optional semantic enrichment
- **THEN** it SHALL store any additional data under explicit extension fields or separate artifacts without changing the default deterministic behavior of `lufy-context-graph`

### Requirement: Context graph must be functionally useful

LUFY SHALL NOT consider the context graph ready if it only serializes a lexical index without reducing exploration cost for agents.

#### Scenario: Report guides first reads
- **WHEN** `lufy-ai context build` generates the derived report
- **THEN** the report SHALL include health, important nodes, communities, suggested questions and an audit trail sufficient to guide targeted first reads

#### Scenario: Cache avoids full re-extraction
- **WHEN** unchanged files are processed after a previous build
- **THEN** the build SHALL reuse hash-matched extractor results where available and report cache hit/miss counts
