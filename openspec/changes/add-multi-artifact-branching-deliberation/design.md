## Context

El harness actual ya distingue T1/T2/T3, usa `workflow_limits` como fuente canonica, permite slices de revision y soporta `parallel_execution` para trabajo independiente con validacion agrupada despues del join. Lo que falta es un protocolo de deliberacion para artifacts OpenSpec cuando el riesgo o la incertidumbre hacen util comparar alternativas antes de comprometer un unico plan.

La capability debe operar sin crear roles nuevos. `sdd-router` detecta complejidad/riesgo/incertidumbre y recomienda si ramificar; `orchestrator` coordina candidates, aislamiento, join y escalacion; `implementer` o solution writer redacta alternatives; `reviewer` compara calidad/riesgo; el humano decide cuando las diferencias impliquen trade-offs no objetivos.

## Goals / Non-Goals

**Goals:**
- Definir un protocolo MVP para generar maximo 2 candidates por etapa.
- Priorizar branching de `proposal`; permitir branching de `design` solo si quedan decisiones tecnicas reales; permitir branching de `tasks` solo de forma rara y explicita.
- Exigir join/merge antes de design/tasks/implementation segun la etapa, dejando un unico artifact set canonico.
- Alinear branching paralelo con `workflow_limits` y `parallel_execution`: artifacts independientes, paths aislados, merge plan y validacion agrupada despues del join.
- Mantener compatibilidad adapter-neutral: OpenSpec es el adapter/metodologia actual, pero el core describe estados y gates transferibles.

**Non-Goals:**
- No crear agentes nuevos como `tech-lead`, `arbitrator` o equivalentes.
- No implementar runtime, CLI, root agents, skills, renderers ni managed assets en esta propuesta.
- No paralelizar delivery ni Git/GH.
- No resolver en paralelo decisiones de contratos publicos, seguridad o producto cuando el trade-off requiere criterio humano.

## Decisions

### Decision 1: Branching recomendado por router, coordinado por orchestrator

`sdd-router` SHALL reportar `artifact_branching` cuando detecte T1 o T2 multi-risk con alta incertidumbre, incluyendo `stage`, `candidate_count`, `reason`, `parallel_allowed`, `requires_join` y riesgos de escalacion. `orchestrator` SHALL decidir la ejecucion concreta, crear handoffs aislados, solicitar comparacion y bloquear avance hasta el join.

Alternativa considerada: crear un rol nuevo de arbitraje. Se descarta porque el usuario confirmo que el MVP debe usar roles existentes y porque la separacion router/orchestrator/reviewer ya cubre recomendacion, coordinacion y comparacion.

### Decision 2: Estado explicito de deliberacion

El flujo sistemico agrega estos estados conceptuales:

1. `routed`: tier, riesgos y limites evaluados.
2. `branching_candidate_generation`: candidates aislados se generan secuencialmente o en paralelo autorizado.
3. `join/decision`: `orchestrator` compara, pide review o escala al humano si hay trade-offs no objetivos.
4. `canonical_artifact_ready`: existe un unico proposal/design/tasks set canonico.
5. `implementation-ready`: el artifact canonico cumple apply requirements y puede pasar a implementacion.

Cada transicion debe registrar evidencia en Result Contract o handoff equivalente. Si falta join, el estado no puede avanzar a `implementation-ready`.

### Decision 3: Branching por etapa con colapso obligatorio

El MVP limita `candidate_count` a 2. La etapa primaria es `proposal` porque ahi se comparan enfoques de producto/scope/arquitectura antes de invertir en detalle. `design` MAY ramificarse despues de una proposal canonica si aun quedan decisiones tecnicas sustantivas. `tasks` SHOULD colapsar a un unico plan; solo MAY tener 2 candidates cuando exista riesgo explicito de estrategia de implementacion y el orchestrator lo documente.

Alternativa considerada: permitir branching libre en todas las etapas. Se descarta por costo cognitivo, riesgo de drift y porque `workflow_limits` exige slicing proporcional, no multiplicacion indefinida de artifacts.

### Decision 4: Aislamiento y merge plan

Cuando candidates se generen en paralelo, cada candidate MUST escribir artifacts aislados, por ejemplo bajo `openspec/changes/<change>/candidates/<stage>/<candidate-id>/`, o en rutas equivalentes del adapter efectivo. Cada candidate MUST incluir un merge plan con supuestos, decisiones, riesgos y elementos reutilizables. El join produce o actualiza los artifacts canonicos (`proposal.md`, `design.md`, `tasks.md`, `specs/**/spec.md`) antes de cualquier downstream implementation.

### Decision 5: Escalacion humana para trade-offs no objetivos

`reviewer` puede comparar calidad, riesgo, completitud y coherencia. `orchestrator` MUST escalar al humano cuando candidates difieran en contrato publico, seguridad, postura de producto, UX significativa, costo/beneficio no objetivo o cualquier decision que no pueda resolverse con criterios previamente acordados.

### Decision 6: Adapter-neutral core, adapter-specific rendering

El core define estados, limites, ownership y gates. Cada methodology/tool adapter decide como materializar paths y comandos. OpenSpec usa proposal/design/tasks/spec deltas; otros adapters podran renderizar artifacts equivalentes sin cambiar la semantica: candidates aislados, merge plan, join obligatorio y un canonico downstream.

## Risks / Trade-offs

- [Riesgo] Duplicar artifacts aumenta carga cognitiva. → [Mitigacion] Limite maximo de 2 candidates y desactivacion cuando hay solucion dominante/bajo riesgo.
- [Riesgo] Candidates paralelos pueden divergir o tocar los mismos paths. → [Mitigacion] Paths aislados, merge plan obligatorio y validacion agrupada solo despues del join.
- [Riesgo] El reviewer podria convertirse en arbitro de producto. → [Mitigacion] Reviewer compara calidad/riesgo; humano decide trade-offs no objetivos.
- [Riesgo] Adapters sin subagentes nativos no pueden paralelizar realmente. → [Mitigacion] El protocolo permite generacion secuencial y conserva los mismos gates.

## Migration Plan

Esta propuesta solo crea artifacts OpenSpec. La implementacion posterior debera actualizar roles/skills/renderers o documentacion gestionada que correspondan, validar assets si cambian superficies instalables y mantener el wrapper/CLI sin delivery no autorizado. No hay rollback runtime; para revertir la propuesta, eliminar o archivar el cambio antes de sync.

## Open Questions

- ¿El implementation slice posterior representara candidates aislados como subdirectorios `candidates/` dentro del cambio OpenSpec o como metadata en el Result Contract hasta el join?
- ¿Que formato exacto tendra `artifact_branching` en el handoff de `sdd-router` para minimizar cambios en structs/CLI?
