## Context

El harness ya separa roles (`implementer`, `validator`, `delivery`, `orchestrator`) y exige autorización explícita para Git/GH. Sin embargo, la política de cierre OpenSpec enumera pasos de delivery como parte de “task complete” y puede confundirse con marcar micro-checkboxes, dejar tareas como completas antes de validación agrupada, o tratar “tasks complete” como equivalente a `archive-ready`/`closed`.

La propuesta debe actualizar artefactos de workflow, no producto: políticas, agentes, skills y comandos que guían `/opsx-apply`, `/opsx-verify`, `/opsx-archive` y delivery.

## Goals / Non-Goals

**Goals:**

- Definir un gate por task/bloque coherente o slice revisable, no por micro-checkbox.
- Formalizar estados de progreso (`implemented`, `validated`, `delivery_pending`, `delivered`, `closed`) y transiciones esperadas.
- Mantener la separación estricta de roles: implementación, validación read-only y delivery autorizado.
- Alinear validación/testing con la filosofía de validación agrupada proporcional.
- Evitar que “tasks complete” se interprete como `closed`, `archive-ready`, commit/push o PR automático.

**Non-Goals:**

- No automatizar commits, push, PRs ni GitHub Projects desde `implementer` o `validator`.
- No cambiar el formato core v2 de deltas más allá de documentación/semántica del workflow instalado.
- No modificar la implementación del producto, el instalador Go, puertos, auth ni esquemas de datos.
- No archivar cambios OpenSpec ni ejecutar delivery como parte de esta propuesta.

## Decisions

### Gate por bloque coherente, no por micro-checkbox

El cierre se evalúa en el nivel mínimo que tenga sentido operacional: task de `tasks.md`, bloque coherente, o review slice. Los micro-checkboxes pueden representar subtareas internas, pero no disparan validación completa ni delivery por sí solos.

Alternativa considerada: correr tests/commit/push tras cada micro-checkbox. Se rechaza porque rompe la filosofía de validación agrupada, aumenta ruido de commits y mezcla responsabilidades de subagentes.

### Estados explícitos de cierre

El workflow usará estados semánticos:

- `implemented`: cambios aplicados por `implementer`; aún falta validación proporcional.
- `validated`: evidencia proporcional registrada por `validator` o por el rol permitido para el bloque; aún falta delivery si aplica.
- `delivery_pending`: el bloque está validado, pero Git/GH requiere autorización explícita o está pendiente.
- `delivered`: delivery autorizado ejecutó commit/push y PR/sync externo cuando aplica.
- `closed`: el bloque/cambio tiene implementación, validación, delivery requerido y artefactos sincronizados para el estado declarado.

Alternativa considerada: conservar solo `blocked`/`sync_pending`. Se rechaza porque no distingue avance real de falta de autorización de delivery.

### Delivery sigue siendo exclusivo y autorizado

Ningún tier ni task completion autoriza Git/GH. `delivery` mantiene ownership de commit, push, PR y sincronización externa, con autorización explícita del usuario.

Alternativa considerada: permitir que subagentes con scope hagan delivery automático. Se rechaza por seguridad, trazabilidad y separación de roles.

### Archive requiere semántica de cierre, no solo checkboxes

`/opsx-archive` debe tratar tareas marcadas como señal necesaria pero no suficiente: debe existir verificación de que el cambio está cerrado o que el flujo permitido no requiere delivery pendiente. Si falta delivery autorizado o sync requerido, el estado debe ser `delivery_pending`, `sync_pending` o `blocked`.

Alternativa considerada: archivar con checkboxes completos y nota manual. Se rechaza porque perpetúa la ambigüedad entre completion local y cierre operacional.

## Risks / Trade-offs

- Riesgo: demasiados estados aumentan carga cognitiva. → Mitigación: documentarlos como transiciones simples y usar equivalentes solo si preservan la semántica.
- Riesgo: tareas documentales T3 podrían parecer sobregestionadas. → Mitigación: el gate es proporcional; puede cerrarse con revisión estática y sin PR si el usuario no requiere delivery, pero debe declarar el estado real.
- Riesgo: comandos/skills/agentes queden desalineados. → Mitigación: implementar en policy canónica primero y propagar referencias a agentes, skills y comandos afectados.
- Riesgo: cambios de assets `.opencode` requieran sincronización de instalador/embedded assets en un cambio posterior. → Mitigación: incluirlo como tarea de implementación/verificación si el catálogo de assets lo requiere.
