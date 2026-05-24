## Context

El repositorio ya define `workflow_limits` como fuente canónica para sizing, routing, proposal slicing, delivery batching, preflight y stop rules. También mantiene copias instalables/embebidas de assets OpenCode, por lo que cualquier cambio posterior en agentes raíz debe contemplar paridad con los assets instalados.

El objetivo de esta propuesta es especificar el contrato antes de modificar agentes: `sdd-router` toma decisiones de routing/workload con límites numéricos observables y `orchestrator` propaga esas decisiones sin re-leer ni reinterpretar configuraciones en cada handoff.

## Goals / Non-Goals

### Goals
- Documentar reglas numéricas concretas: 4-file rule, 20-tool-calls rule, multi-file write rule y long-session rule.
- Hacer explícito que `workflow_limits` es la única fuente canónica para límites de workflow y que campos legacy top-level no se consumen.
- Definir reporting `not_available` cuando `.opencode/project.yaml` o una ruta esperada no existe.
- Mantener separadas las fases de proposal/review slicing y delivery batching.
- Definir cómo `chain_strategy: auto-chain` se propaga desde routing a orchestrator sin preguntar salvo riesgo alto.

### Non-Goals
- No implementar todavía cambios en `.opencode/agents/*`.
- No modificar el CLI Go ni estructuras internas de `WorkflowLimits` en este slice salvo que una implementación posterior lo declare necesario.
- No autorizar delivery, commits, PRs ni GitHub Project sync.

## Decisions

### Fuente canónica de límites
- `sdd-router` y `orchestrator` SHALL leer/reportar límites desde `.opencode/project.yaml` top-level `workflow_limits` cuando el archivo exista.
- Rutas esperadas:
  - `workflow_limits.sizing`
  - `workflow_limits.routing`
  - `workflow_limits.proposal_slicing_strategy`
  - `workflow_limits.delivery_batch_strategy`
  - `workflow_limits.preflight`
  - `workflow_limits.stop_rules`
- Si falta el archivo o una ruta, el resultado SHALL usar `not_available` para esa ruta, no defaults inventados.
- Top-level `loc_budget` y `delivery_strategy` SHALL permanecer legacy/no canónicos y no se usarán para sizing, routing, slicing, batching, preflight, stop rules, autorización o cierre.

### Workload guard numérico
- Si `estimated_loc > workflow_limits.sizing.loc_budget`, `sdd-router` SHALL emitir `workload_decision_needed: true` y recomendar decisión de carga/slice antes de implementación.
- Si `estimated_files >= 5`, `sdd-router` SHALL escalar tier o proponer `review_slices` según riesgo y estrategia de slicing disponible.
- `review_slices` SHALL derivarse de `workflow_limits.proposal_slicing_strategy`; `delivery_batch_strategy` no puede usarse para justificar slicing de propuesta/review.

### Stop rules explícitas del orchestrator
- 4-file rule: si el alcance estimado o real toca 4+ archivos significativos, el orchestrator SHALL requerir una pausa de workload/review y decidir si se mantiene el tier, se escala o se divide en slices.
- 20-tool-calls rule: si una sesión/ruta excede 20 tool calls para un bloque coherente, el orchestrator SHALL pausar y pedir resumen/decisión de continuación, escalamiento o compactación.
- Multi-file write rule: si un paso intenta escribir múltiples archivos no triviales, el orchestrator SHALL confirmar que existe plan/slice y evidencia de alcance antes de continuar.
- Long-session rule: si la sesión se vuelve larga o pierde evidencia resumible, el orchestrator SHALL pedir handoff/resumen antes de seguir con implementación o validación.

### `chain_strategy`
- Para este slice, `chain_strategy` se modela como metadata opcional de routing, no como cambio obligatorio de la estructura CLI `WorkflowLimits`.
- Orden de lectura recomendado para agentes Markdown:
  1. `.opencode/project.yaml` top-level `chain_strategy`.
  2. `workflow_limits.routing.chain_strategy` si existe.
  3. `not_available` si falta.
- Con `chain_strategy: auto-chain`, el orchestrator SHALL propagar la estrategia en handoffs y puede encadenar al siguiente rol sin preguntar cuando el riesgo no sea alto ni requiera autorización explícita.
- Riesgos altos, delivery, Git/GH, cambios protegidos o stop rules disparadas SHALL seguir requiriendo pausa/autorización según política.

## Validation Plan

- Validar la propuesta con `openspec validate add-numeric-stop-rules-workload-guard --strict`.
- Revisar status con `openspec status --change "add-numeric-stop-rules-workload-guard"`.
- En implementación posterior, validar paridad raíz/embebidos/catalog y ejecutar `scripts/validate.sh` si se tocan assets o CLI Go.

## Open Questions / Risks

- `chain_strategy` aparece como backlog/config deseada pero puede no existir en la estructura actual del CLI Go. La decisión de este diseño evita romper compatibilidad al tratarlo como lectura opcional por agentes Markdown; una propuesta futura puede tiparlo en Go si se necesita validación estructural.
- Las reglas numéricas necesitan wording operativo preciso en agentes para no bloquear cambios T3 triviales de forma excesiva.
