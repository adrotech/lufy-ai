## Why

El harness actualmente puede mezclar “tasks complete” con “archive-ready” o “closed”, lo que permite interpretar que marcar micro-checkboxes equivale a cierre operativo. Este cambio formaliza un gate por task/bloque coherente que separa implementación, validación proporcional y delivery autorizado sin romper la separación de roles.

## What Changes

- Definir estados explícitos de cierre por task/bloque: `implemented`, `validated`, `delivery_pending`, `delivered` y `closed`.
- Aclarar que el gate se evalúa por task, bloque coherente o slice revisable, no por cada micro-checkbox interno.
- Mantener que `implementer` implementa, `validator` valida en read-only, `delivery` realiza Git/GH solo con autorización explícita y `orchestrator` coordina.
- Establecer que la validación/testing es proporcional y agrupada al cierre del bloque/slice, evitando loops constantes de tests salvo excepciones justificadas.
- Corregir la semántica OpenSpec para que “tasks complete” no implique por sí solo `archive-ready`, `closed`, commit, push o PR.
- Documentar que sin autorización explícita de delivery el estado queda en `delivery_pending`/`blocked`, no en `closed`.

## Capabilities

### New Capabilities

- Ninguna.

### Modified Capabilities

- `sdd-harness-routing`: formaliza el gate de cierre por task/bloque, los estados de handoff y la separación de roles hasta delivery autorizado.
- `systemic-workflow`: ajusta la regla de validación agrupada para que aplique al cierre de task/bloque coherente o slice, no a micro-checkboxes.
- `openspec-core-v2-workflow`: aclara la semántica de tasks, verificación y archive para que tareas marcadas no se confundan con cambio cerrado o listo para archivar sin delivery/sync requerido.

## Impact

- Documentación/política: `.opencode/policies/delivery.md`, `AGENTS.md` y posiblemente `AGENTS.md.template`.
- Agentes: `.opencode/agents/orchestrator.md`, `.opencode/agents/implementer.md`, `.opencode/agents/validator.md`, `.opencode/agents/delivery.md`, y referencias relacionadas si existen.
- Skills OpenSpec: `.opencode/skills/sdd-workflow/openspec-apply-change/`, `openspec-verify-change/`, `openspec-archive-change/`.
- Comandos: `.opencode/commands/opsx-apply.md`, `.opencode/commands/opsx-verify.md`, `.opencode/commands/opsx-archive.md`.
- Specs OpenSpec: deltas bajo `openspec/changes/add-task-block-delivery-gate/specs/`.
- No cambia contratos de producto, puertos, auth, esquema de datos ni implementación del instalador.
