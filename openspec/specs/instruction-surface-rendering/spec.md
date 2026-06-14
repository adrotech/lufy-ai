# instruction-surface-rendering Specification

## Purpose
TBD - created by archiving change abstract-harness-tool-methodology-adapters. Update Purpose after archive.
## Requirements
### Requirement: Neutral role definitions
El sistema SHALL mantener definiciones neutrales de roles Lufy para agentes principales y subagentes antes de renderizarlas a una tool concreta.

#### Scenario: Primary roles are neutral
- **WHEN** se define `orchestrator`, `router` o `delivery` como rol core
- **THEN** la definición SHALL describir responsabilidades, permisos, gates y outputs sin depender de `.opencode`, OpenCode, OpenSpec, `/opsx-*` u otra tool/metodología concreta

#### Scenario: Subagent roles are neutral
- **WHEN** se define `explorer`, `implementer`, `test-writer`, `validator` o `reviewer` como rol core
- **THEN** la definición SHALL describir comportamiento del rol y fallback inline sin asumir que la tool soporta subagentes nativos

### Requirement: Instruction renderer composes bindings
El renderer de instrucciones SHALL componer role core, tool binding y methodology binding para generar assets operativos.

#### Scenario: OpenCode OpenSpec rendering
- **WHEN** el preset efectivo es `tool=opencode` y `methodology=openspec`
- **THEN** el renderer MAY producir `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, templates y policies compatibles con la instalación actual

#### Scenario: Methodology none rendering
- **WHEN** una metodología efectiva es `none`
- **THEN** el renderer SHALL omitir instrucciones que obliguen a crear proposal/spec/tasks persistentes para ese tier y SHALL conservar instrucciones de Result Contract y validación proporcional

### Requirement: Textual leak checks
El sistema SHALL validar que assets neutrales y adapters futuros no contengan referencias indebidas a tools o metodologías no seleccionadas.

#### Scenario: Core role leak check
- **WHEN** se validan roles core neutrales
- **THEN** el check SHALL fallar si encuentra referencias a `.opencode`, `opencode.json`, `.claude`, `.codex`, `openspec/` o `/opsx-*`

#### Scenario: None methodology leak check
- **WHEN** se validan assets renderizados para metodología `none`
- **THEN** el check SHALL fallar si esos assets requieren `openspec/`, `/opsx-*` o artefactos metodológicos formales

#### Scenario: Tool adapter leak check
- **WHEN** se validan assets de un adapter no OpenCode
- **THEN** el check SHALL fallar si esos assets dependen de `.opencode` u `opencode.json` salvo que el adapter declare compatibilidad OpenCode explícita

### Requirement: Rendered output golden tests
El renderer SHALL contar con golden tests o evidencia equivalente para demostrar que el preset OpenCode/OpenSpec inicial conserva comportamiento.

#### Scenario: OpenCode default remains equivalent
- **WHEN** se renderizan assets para el preset default
- **THEN** la salida SHALL compararse contra fixtures esperados o hashes controlados para detectar cambios accidentales en agentes, subagentes, commands, skills, templates y policies

#### Scenario: Role agent golden fixtures
- **WHEN** el renderer genera agentes OpenCode desde roles neutrales
- **THEN** la salida SHALL incluir path de destino, contexto de adapter, permisos, responsabilidades, boundaries, skills directos y contrato de salida, y SHALL compararse contra fixtures golden versionados

### Requirement: Embedded assets stay synchronized
Cuando el renderer o los assets operativos cambien, los assets embebidos del binario SHALL actualizarse o generarse de forma verificable.

#### Scenario: Installable asset changes
- **WHEN** cambia un agente, subagente, command, skill, template o policy instalable
- **THEN** `tools/lufy-cli-go/internal/assets/embedded` y el catálogo efectivo SHALL reflejar el cambio antes de reportar readiness

### Requirement: Codex PR review skill preserves Lufy HTML contract
El skill visible para Codex `pr-reviewer` SHALL preserve the observable Lufy PR review contract instead of degrading to chat-only findings.

#### Scenario: Codex PR review generates HTML artifact
- **WHEN** un usuario en Codex pide un PR review o PR audit
- **THEN** `.agents/skills/pr-reviewer/SKILL.md` SHALL require creating `pr_review/` when missing
- **AND** it SHALL require writing `pr_review/pr-review-<number>-<yyyyMMdd-HHmm>.html` or a slug equivalent for non-numbered reviews
- **AND** it SHALL require reporting the generated path and `open pr_review/pr-review-<...>.html` in the final response

#### Scenario: Codex PR review includes full review contract
- **WHEN** `.agents/skills/pr-reviewer/SKILL.md` is installed or synced
- **THEN** it SHALL require scoring, severity-ordered findings, desk check and simulation, evidence, limitations, action items, and final recommendation
- **AND** it SHALL reference the canonical `.opencode/skills/pr.reviewer/SKILL.md` contract when that file exists
- **AND** it SHALL require the canonical HTML template or equivalent Notion-dark markers when available
