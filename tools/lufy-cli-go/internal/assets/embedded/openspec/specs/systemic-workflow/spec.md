# systemic-workflow Specification

## Purpose
Define the systemic working model for agents: initial context analysis, bounded rereads, final grouped validation and evidence-based reporting.
## Requirements
### Requirement: Analisis inicial sistemico
El workflow SHALL analizar el sistema al inicio de una propuesta o bloque antes de implementar, identificando archivos existentes relevantes, componentes, dependencias, interconexiones y riesgos de comportamiento.

#### Scenario: Analisis antes de implementar
- **WHEN** una propuesta OpenSpec o bloque de trabajo requiere cambios sobre codigo, configuracion, agentes o documentacion existente
- **THEN** el agente responsable realiza una inspeccion inicial dirigida que cubre el contexto necesario para planificar sin depender de relecturas repetidas durante la implementacion normal

#### Scenario: Vision holistica del cambio
- **WHEN** el analisis inicial identifica varios componentes relacionados, como agentes, skills, politica, tests, documentacion, APIs, base de datos o servicios
- **THEN** el reporte o handoff explica como interactuan las partes relevantes y que dependencias condicionan la implementacion

### Requirement: Implementacion sin relecturas repetidas innecesarias
El workflow SHALL evitar volver a leer archivos viejos ya analizados durante la implementacion normal, excepto cuando el archivo haya sido modificado, exista conflicto, aparezca nueva evidencia, el alcance cambie, se detecte un bloqueo o el riesgo requiera confirmacion.

#### Scenario: Archivo viejo no modificado
- **WHEN** un archivo existente fue revisado en el analisis inicial y no fue modificado ni afectado por nueva evidencia
- **THEN** el agente no lo relee de forma rutinaria durante cada tarea de implementacion

#### Scenario: Relectura justificada
- **WHEN** un archivo viejo fue modificado, entra en conflicto con cambios concurrentes, participa en una falla, o nueva evidencia invalida el analisis inicial
- **THEN** el agente lo relee de forma dirigida y reporta la razon si afecta el flujo o la evidencia final

### Requirement: Revisión final acotada de archivos viejos modificados
El workflow SHALL incluir una revision final acotada de archivos viejos modificados o afectados antes de cerrar la implementacion o ejecutar validacion final.

#### Scenario: Cierre de bloque con archivos existentes modificados
- **WHEN** una propuesta modifica archivos existentes revisados previamente
- **THEN** el agente revisa al final esos archivos o diffs para comprobar coherencia con el analisis inicial, dependencias y comportamiento esperado

#### Scenario: Sin archivos viejos modificados
- **WHEN** la propuesta solo crea archivos nuevos y no afecta archivos existentes
- **THEN** no se requiere una relectura final de archivos viejos, salvo que una dependencia o riesgo detectado lo justifique

### Requirement: Validacion final agrupada con tests y coverage
El workflow SHALL ejecutar tests, coverage y validacion completa al final de todas las tareas de una propuesta cuando esos comandos existan y apliquen al alcance.

#### Scenario: Propuesta con tareas completas
- **WHEN** todas las tareas de implementacion de una propuesta estan finalizadas
- **THEN** el agente ejecuta la validacion agrupada disponible, incluyendo tests y coverage cuando existan para el toolchain real del alcance, y reporta comandos exactos y resultados

#### Scenario: Validacion no disponible
- **WHEN** no existe toolchain, comando de coverage o suite aplicable al alcance
- **THEN** el agente declara la limitacion y reporta evidencia estatica, documental o manual real sin afirmar que tests o coverage pasaron

### Requirement: Excepciones para feedback temprano
El workflow SHALL permitir relectura o validacion temprana solo cuando exista bloqueo, cambio riesgoso, diagnostico de falla, incertidumbre que afecte seguridad/correctness, o necesidad de feedback para autorregular el sistema.

#### Scenario: Bloqueo o falla durante implementacion
- **WHEN** una tarea queda bloqueada o falla por una causa no entendida
- **THEN** el agente puede releer archivos relevantes o ejecutar validacion enfocada antes del final para diagnosticar y destrabar el trabajo

#### Scenario: Cambio riesgoso
- **WHEN** el cambio afecta areas de alto impacto como contratos publicos, autenticacion, persistencia, instalador, release o delivery
- **THEN** el agente puede solicitar o ejecutar validacion temprana enfocada antes de continuar, manteniendo la validacion completa para el final

### Requirement: Relacion estructura-comportamiento
El workflow SHALL evaluar que el comportamiento esperado del sistema emerge de la estructura modificada y de sus relaciones, no solo de archivos individuales.

#### Scenario: Cambio transversal
- **WHEN** una propuesta cambia reglas compartidas, agentes, skills, politica o documentacion que afecta varias fases
- **THEN** la revision final verifica coherencia entre estructura estatica, dependencias y comportamiento dinamico esperado del workflow

### Requirement: Block-scoped proportional validation
The workflow SHALL run validation/testing proportionally at the end of a task, coherent block, proposal block, or review slice, SHALL apply any relevant `workflow_limits.preflight` and `workflow_limits.stop_rules`, and SHALL avoid constant test loops for individual micro-checkboxes unless an exception gate applies.

#### Scenario: Validation waits for coherent block boundary
- **WHEN** an agent completes an internal micro-step that is part of a larger coherent task or block
- **THEN** the workflow SHALL NOT require full validation/testing for that micro-step and SHALL defer grouped validation to the coherent block boundary

#### Scenario: Validation runs before validated state
- **WHEN** a task, coherent block, proposal block, or review slice is ready to move from `implemented` to `validated`
- **THEN** the workflow SHALL run the real applicable validation commands or document proportional static/manual evidence and SHALL report exact evidence before using the `validated` state

#### Scenario: Exception allows early validation
- **WHEN** a blocker, risky change, feedback loop, or failure diagnosis requires earlier evidence
- **THEN** the workflow MAY run focused validation before the block boundary while preserving grouped final validation for the block when applicable

#### Scenario: Configured workflow limits affect validation gate
- **WHEN** `workflow_limits.preflight` or `workflow_limits.stop_rules` define additional validation, pause or escalation conditions for the current block
- **THEN** the workflow applies those configured limits before reporting the block as `validated` or delivery-ready

### Requirement: Workflow limits canonical consumption
Agents, workflow documentation, delivery policy and result contracts SHALL consume and report project workflow limits from `.lufy/project.yaml` top-level `workflow_limits` only.

#### Scenario: Router reads canonical workflow limits
- **WHEN** `sdd-router` evaluates sizing, routing, slicing or escalation inputs for a project with `.lufy/project.yaml`
- **THEN** it reads those inputs from `workflow_limits` and MUST NOT read top-level `loc_budget` or top-level `delivery_strategy` as valid sources

#### Scenario: Orchestrator reports canonical workflow limits
- **WHEN** `orchestrator` reports routing rationale, handoff constraints or result contract fields that depend on project workflow limits
- **THEN** it references `workflow_limits` paths as the source of truth and MUST NOT report legacy top-level fields as canonical

#### Scenario: Delivery policy reads canonical workflow limits
- **WHEN** delivery guidance needs batching, preflight or stop-rule limits from project config
- **THEN** delivery policy and delivery role instructions consume those limits from `workflow_limits` only

### Requirement: Proposal slicing is separate from delivery batching
The workflow SHALL treat `workflow_limits.proposal_slicing_strategy` and `workflow_limits.delivery_batch_strategy` as different controls with different lifecycle phases.

#### Scenario: Proposal slicing before implementation or review
- **WHEN** a proposal or review workload needs to be split into smaller coherent slices
- **THEN** the workflow uses `workflow_limits.proposal_slicing_strategy` to decide implementation/review slices before delivery authorization

#### Scenario: Delivery batching after validation readiness
- **WHEN** validated or delivery-ready changes need to be grouped for Git/GH delivery
- **THEN** the workflow uses `workflow_limits.delivery_batch_strategy` and MUST NOT reinterpret proposal slicing rules as delivery batching authorization

### Requirement: Workflow preflight and stop rules
The workflow SHALL apply `workflow_limits.preflight` and `workflow_limits.stop_rules` as project-local gates for pausing, escalating or requiring evidence before continuing.

#### Scenario: Preflight gate before a bounded workflow phase
- **WHEN** a workflow phase has configured preflight checks under `workflow_limits.preflight`
- **THEN** the responsible role verifies or reports those checks before moving to the next state that depends on them

#### Scenario: Stop rule forces escalation
- **WHEN** an active task reaches a configured condition under `workflow_limits.stop_rules`
- **THEN** the responsible role pauses, reports the blocking condition and escalates to the appropriate role or user decision instead of continuing silently

### Requirement: Observable workflow-limit gates
The workflow SHALL make configured `workflow_limits.preflight` and `workflow_limits.stop_rules` observable in result contracts before a block advances to validation-ready, delivery-ready, delivered or closed states.

#### Scenario: Preflight status is reported before state advance
- **WHEN** a workflow block reaches a boundary that depends on configured preflight checks
- **THEN** the result contract reports each relevant preflight check as `passed`, `not_applicable`, `not_available` or `blocked` before advancing state

#### Scenario: Stop rule blocks silent continuation
- **WHEN** a configured stop rule is triggered by estimated file count, LOC, tool calls, session length, risk, validation failure or delivery condition
- **THEN** the workflow reports `status: blocked` or `status: escalated` with the exact stop rule and recovery path instead of continuing silently

### Requirement: Proportional Result Contract detail
The workflow SHALL scale Result Contract detail by tier while preserving the canonical envelope fields required for context recovery.

#### Scenario: T3 uses compact envelope
- **WHEN** work is classified as T3 Express and has no meaningful handoff or delivery risk
- **THEN** the result may use compact field values and `not_applicable` entries while still preserving envelope identity, status, evidence and next action

#### Scenario: T1 or multi-risk T2 uses full evidence
- **WHEN** work is classified as T1 or a multi-risk T2 review slice
- **THEN** the result contract includes explicit acceptance criteria status, validation evidence, workflow-limit decisions, risks and follow-ups sufficient for review without replaying the full conversation

### Requirement: TDD delegation for substantive T1 and T2 changes
The workflow SHALL route substantive test design or test implementation for T1 and T2 changes through `test-writer` when a TDD cycle is applicable.

#### Scenario: Implementer delegates test work
- **WHEN** `implementer` is assigned a T1 or T2 change with substantive test creation or revision needs
- **THEN** `implementer` delegates the test-focused portion to `test-writer` or records why TDD delegation is not applicable in the Result Contract envelope

#### Scenario: T3 change does not require delegation
- **WHEN** a T3 Express change is trivial, mechanical or documentation-only and does not require substantive test behavior
- **THEN** the workflow does not require `test-writer` delegation and may record TDD evidence as `not_applicable`

### Requirement: Validator gates required TDD evidence
The workflow SHALL require validator review of TDD evidence for T1 and T2 changes where TDD delegation or equivalent TDD evidence is required.

#### Scenario: Required TDD evidence is present
- **WHEN** `validator` evaluates a T1 or T2 change that required TDD evidence
- **THEN** it verifies that RED, GREEN, TRIANGULATE and REFACTOR evidence is present or explicitly marked `not_applicable` with reasons before reporting the block as `validated`

#### Scenario: Required TDD evidence is missing
- **WHEN** `validator` evaluates a T1 or T2 change that required TDD evidence but the evidence is absent or incomplete
- **THEN** it reports `blocked` or `escalated` with the missing evidence and next owner instead of reporting `validated`

### Requirement: Weighted review gate for substantive changes
The workflow SHALL use weighted reviewer output as a quality gate for T1 and T2 changes that require independent review.

#### Scenario: Review gate passes
- **WHEN** a T1 or T2 change has reviewer score at least 80%, zero L1/L2 findings and proportional validation evidence
- **THEN** the workflow may treat review as approval-ready while still requiring explicit delivery authorization for Git/GH actions

#### Scenario: Review gate blocks
- **WHEN** reviewer score is below 80% or any L1/L2 finding exists
- **THEN** the workflow reports the block with findings, score, affected categories and next owner instead of advancing to delivery-ready

### Requirement: Review remains separate from validation and delivery
The workflow SHALL keep reviewer qualitative scoring separate from validator command evidence and delivery authorization.

#### Scenario: Reviewer lacks command evidence
- **WHEN** reviewer identifies missing tests or validation evidence
- **THEN** it escalates to `validator` for command evidence rather than claiming commands passed

#### Scenario: Review is approval-ready
- **WHEN** reviewer reports approval-ready
- **THEN** Git/GH delivery remains blocked until the user explicitly authorizes `delivery`

### Requirement: Numeric stop rules for workload guard
The workflow SHALL apply explicit numeric stop rules to prevent oversized, under-scoped, or low-evidence sessions from continuing silently.

#### Scenario: Four-file rule pauses for workload decision
- **WHEN** a routed block is estimated to touch four or more significant files or implementation discovers the block touches four or more significant files
- **THEN** the orchestrator SHALL require a workload decision, tier escalation, or review-slice plan before continuing beyond the current safe boundary

#### Scenario: Twenty-tool-calls rule pauses long routing
- **WHEN** a coherent routing or implementation block exceeds twenty tool calls without reaching a resumable state
- **THEN** the orchestrator SHALL pause for a concise state summary and decide whether to continue, compact context, escalate, or split the work

#### Scenario: Multi-file write rule requires plan
- **WHEN** a step proposes or attempts writes across multiple non-trivial files
- **THEN** the workflow SHALL verify that a scoped plan or review slice exists and SHALL avoid broad multi-file mutation without observable acceptance criteria

#### Scenario: Long-session rule requires resumable handoff
- **WHEN** a session becomes long enough that evidence, decisions, or next actions are no longer easily resumable
- **THEN** the workflow SHALL create or request a handoff summary before continuing with implementation, validation, or delivery routing

### Requirement: Stop-rule evidence in Result Contract
The workflow SHALL report triggered or evaluated numeric stop rules in Result Contract envelope v1 so downstream roles can continue without rediscovering the decision.

#### Scenario: Stop rule triggers blocked or escalated state
- **WHEN** a numeric stop rule is triggered and requires a decision before continuing
- **THEN** the result SHALL report `status: blocked` or `status: escalated`, the exact rule, the evidence that triggered it, and the recommended next owner/action

#### Scenario: Stop rules are clear at boundary
- **WHEN** a routed block reaches an implementation or validation boundary without triggering numeric stop rules
- **THEN** the result SHALL report `stop_rule_status: clear` or a proportional equivalent in the workflow decision fields

#### Scenario: Stop rules unavailable are explicit
- **GIVEN** `.lufy/project.yaml` or `workflow_limits.stop_rules` is not available
- **WHEN** a result reports workflow-limit fields
- **THEN** it SHALL report stop-rule configuration as `not_available` while still applying repository-level default guardrails from agent instructions
