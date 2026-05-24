## Why

El harness SDD ya declara `workflow_limits` como fuente canónica, pero las reglas numéricas de pausa/escalamiento y la propagación de configuración aún no están completamente especificadas para `orchestrator` y `sdd-router`. Esto permite decisiones inconsistentes de workload, slicing, batching o chaining cuando el proyecto define límites explícitos o cuando `.opencode/project.yaml` no existe.

## What Changes

- Agregar reglas explícitas de stop/workload para `orchestrator`: 4-file rule, 20-tool-calls rule, multi-file write rule y long-session rule.
- Extender el contrato de routing para que `sdd-router` lea y propague `workflow_limits.sizing`, `workflow_limits.routing`, `workflow_limits.proposal_slicing_strategy`, `workflow_limits.delivery_batch_strategy`, `workflow_limits.preflight`, `workflow_limits.stop_rules` y `chain_strategy` desde `.opencode/project.yaml` cuando esté disponible.
- Definir que los estimados `estimated_loc` y `estimated_files` activan decisiones observables: `estimated_loc > workflow_limits.sizing.loc_budget` emite `workload_decision_needed: true`; `estimated_files >= 5` escala tier o propone slices.
- Separar explícitamente proposal/review slicing de delivery batching: `review_slices` se derivan de `workflow_limits.proposal_slicing_strategy`, mientras `delivery_batch_strategy` sigue siendo advisory y requiere autorización de delivery.
- Definir comportamiento seguro cuando `.opencode/project.yaml` falta: reportar `not_available`, no inventar valores y no leer campos legacy top-level `loc_budget` / `delivery_strategy`.
- Resolver `chain_strategy` como metadata opcional de routing sin exigir cambios de CLI en este slice: leer top-level `chain_strategy` o `workflow_limits.routing.chain_strategy` si existen; reportar `not_available` si falta.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `sdd-harness-routing`: agrega decisiones numéricas de workload, consumo/propagación de límites canónicos y estrategia de chaining en el contrato router/orchestrator.
- `systemic-workflow`: agrega reglas explícitas de pausa/escalamiento por número de archivos, tool calls, escrituras multiarchivo y sesiones largas, con reporting observable en Result Contract.

## Impact

- Archivos de agentes raíz y assets embebidos esperados en implementación posterior: `.opencode/agents/orchestrator.md`, `.opencode/agents/sdd-router.md`, `tools/lufy-cli-go/internal/assets/embedded/**` y catálogo/manifest de assets si corresponde.
- Configuración leída: `.opencode/project.yaml` top-level `workflow_limits` y `chain_strategy` opcional.
- Validación esperada: `openspec validate add-numeric-stop-rules-workload-guard --strict`, `openspec status --change "add-numeric-stop-rules-workload-guard"` y `scripts/validate.sh` cuando la implementación toque assets/catalog o CLI Go.
