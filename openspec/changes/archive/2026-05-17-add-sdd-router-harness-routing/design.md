## Context

Lufy AI ya cuenta con roles separados: `orchestrator`, `explorer`, `implementer`, `validator`, `reviewer` y `delivery`. También cuenta con OpenSpec para SDD completo y una política de delivery con permisos y gates claros.

El problema no es la falta de roles. El problema es decidir cuándo usarlos. Un harness profesional debe limitar herramientas, contexto y permisos según el trabajo. La clasificación debe ocurrir antes de activar agentes caros, permisos de edición o OpenSpec completo.

La metodología se apoya en prácticas de SDD, BDD y harness engineering:

- SDD: intención y requisitos antes de implementación.
- BDD: Discovery, Formulation y Automation mediante ejemplos observables.
- Gherkin: escenarios con WHEN y THEN verificables.
- ADR: decisiones solo cuando hay trade-offs reales.
- Harness engineering: herramientas, contexto, permisos y verificación como entorno controlado alrededor del modelo.
- Skill registry: resolución explícita de skills locales antes de considerar bootstrap externo.

## Goals / Non-Goals

**Goals:**

- Crear un `sdd-router` liviano, read-only y de bajo contexto.
- Clasificar pedidos en T1 Full SDD, T2 SDD Lite o T3 Express.
- Definir cuándo aplicar OpenSpec completo, mini-spec o edición directa.
- Pasar a cada subagente solo el contexto necesario para su rol.
- Reducir permisos y tool calls en tareas simples.
- Mantener rutas de escalado cuando un pedido resulta más riesgoso de lo esperado.
- Definir execution modes, contratos de resultado y artifact store mínimo para handoffs reproducibles.
- Permitir bootstrap opcional de skills con AutoSkills cuando falten skills locales relevantes y el usuario lo autorice.
- Mantener la documentación del proyecto y los assets embebidos sincronizados con el harness instalado.
- Diseñar features/propuestas pensando en el reviewer humano mediante slices revisables, validables y entregables por partes cuando el alcance lo justifique.
- Generar un binario local del installer para validar que los assets embebidos evolucionados compilan e instalan correctamente.
- Aclarar la frontera entre policy compartida de delivery y comportamiento operativo del subagente `delivery`.

**Non-Goals:**

- No reemplazar el `orchestrator` existente.
- No automatizar commits, pushes, PRs o delivery sin autorización explícita.
- No crear un agente de testing separado en esta propuesta.
- No cambiar el instalador, la CLI Go ni la política de releases.
- No obligar a usar OpenSpec completo para T2 o T3.
- No convertir AutoSkills en dependencia obligatoria del flujo.
- No ejecutar `npx autoskills` ni instalar skills externas sin autorización explícita del usuario.
- No publicar una release, crear tag, push o PR como parte de esta propuesta.
- No mover toda la política de delivery dentro del subagente `delivery` si otras piezas del sistema la necesitan como contrato compartido.

## Decisions

### Decision: agregar `sdd-router` como subagente separado

El `orchestrator` seguirá coordinando, pero delegará la clasificación inicial a un subagente liviano.

Alternativa considerada: agregar toda la lógica al `orchestrator`. Se descarta porque aumenta contexto y responsabilidad en el agente primario, y dificulta mantener permisos mínimos.

### Decision: salida estructurada del router

El `sdd-router` devolverá un contrato estable con tier, confianza, razón, flujo recomendado, subagentes necesarios, permisos requeridos y contexto mínimo para el siguiente paso.

Alternativa considerada: salida narrativa libre. Se descarta porque dificulta automatizar routing y aumenta ambigüedad entre agentes.

### Decision: execution mode y result contract explícitos

Cada handoff del router declarará un `execution_mode` proporcional al tier: `full_sdd`, `sdd_lite`, `express`, `clarify`, `explore_only`, `verify_only` o `delivery_pending`. Los agentes que reciban el handoff deberán responder con un result contract mínimo: objetivo, acciones realizadas, evidencia, riesgos, estado y siguiente acción recomendada.

Alternativa considerada: inferir modo y resultado desde texto libre. Se descarta porque complica continuidad, revisión y recuperación de contexto.

### Decision: T2 usa SDD Lite, no Full SDD recortado

T2 tendrá un artefacto compacto con intención, comportamiento actual, comportamiento objetivo, scope, criterios de aceptación, tareas y validación. No requiere `proposal.md`, `design.md` y specs delta completas salvo que el router escale a T1.

Alternativa considerada: usar siempre OpenSpec completo para todo cambio funcional. Se descarta porque introduce fricción innecesaria y contradice el principio de flujo proporcional.

### Decision: T3 puede evitar OpenSpec y explorer

T3 puede ir directo a `implementer` cuando el cambio es claro, pequeño y de bajo riesgo. Si aparece incertidumbre, escala a T2.

Alternativa considerada: exigir exploración para todo cambio. Se descarta porque encarece tareas simples y aumenta permisos/contexto sin aportar seguridad proporcional.

### Decision: artifact store mínimo por tier

T1 seguirá usando artefactos OpenSpec completos. T2 usará un mini-spec compacto cuando el cambio requiera comportamiento verificable. T3 podrá documentarse solo en el result contract final cuando no haga falta artefacto persistente.

Alternativa considerada: persistir siempre un artefacto por cada pedido. Se descarta porque agrega ruido a tareas triviales y contradice YAGNI.

### Decision: skill registry local primero y bootstrap externo opcional

El router deberá reportar `skill_status`: skills locales relevantes, stack detectado si aplica, cobertura suficiente o faltante, y recomendación de bootstrap. Si faltan skills locales, podrá sugerir `npx autoskills --dry-run` como primer paso, pero nunca instalar ni ejecutar comandos mutantes sin autorización explícita.

Alternativa considerada: instalar AutoSkills automáticamente cuando falten skills. Se descarta por seguridad, licenciamiento, ruido potencial y porque este repositorio ya tiene skills propias.

### Decision: aislamiento de subagentes y review workload proporcional

Cada subagente recibirá solo el contexto necesario para su rol y tier. El router podrá recomendar `review_workload` como `none`, `focused` o `full` según riesgo, alcance y tier.

Alternativa considerada: compartir todo el contexto de conversación con todos los agentes. Se descarta porque aumenta costo, filtración accidental de contexto y riesgo de decisiones fuera de rol.

### Decision: Review Workload Harness con slices revisables

Las propuestas y features T1, y los T2 con más de un eje de riesgo, deberán pensar explícitamente en la persona que revisa. El router y los templates podrán recomendar `review_slices`: cortes pequeños con objetivo, archivos esperados, criterios `WHEN`/`THEN`, validación, riesgo y sugerencia de PR cuando convenga. Esto no obliga a micro-PRs para todo: T3 no se fragmenta y T2 solo se divide si reduce carga cognitiva o riesgo de revisión.

Alternativa considerada: agregar una fase obligatoria de slicing para todo pedido. Se descarta porque complejiza tareas simples y contradice el principio proporcional. El slicing es una herramienta para reducir el costo humano de revisión, no un ritual fijo.

### Decision: delivery permanece separado

El router puede indicar que delivery será necesario, pero no autoriza ni ejecuta Git/GH. La autorización explícita del usuario sigue siendo obligatoria.

Alternativa considerada: permitir que el flujo por tier autorice delivery automáticamente. Se descarta por seguridad y por política vigente del repositorio.

### Decision: policy de delivery compartida, agente delivery operativo

`.opencode/policies/delivery.md` seguirá siendo la fuente canónica de invariantes compartidas: branch safety, autorización explícita, gates de validación, reglas de release, cierre de tareas y estados `blocked`/`sync_pending`. `.opencode/agents/delivery.md` contendrá el runbook operativo del agente: cuándo actuar, qué comandos puede ejecutar, cómo aplicar la política y cómo reportar evidencia.

Alternativa considerada: mover todo el contenido de delivery al subagente `delivery`. Se descarta porque `orchestrator`, `validator`, `reviewer` e `implementer` también necesitan conocer límites de delivery antes de delegar o reportar readiness. Duplicar esa política en un solo agente aumentaría drift y haría menos visible el contrato compartido.

### Decision: docs y assets embebidos deben evolucionar juntos

La documentación raíz y los assets instalables deben describir el mismo sistema. Cuando se agregan agentes, templates o políticas, deben actualizarse tanto los archivos raíz como `tools/lufy-cli-go/internal/assets/embedded/` para que un binario standalone instale el harness vigente.

Alternativa considerada: actualizar solo docs raíz y dejar assets embebidos para otro cambio. Se descarta porque generaría un installer nuevo que no refleja lo documentado y rompería la promesa de instalación standalone.

### Decision: nuevo installer local, no release publicada

La implementación generará un binario local `tools/lufy-cli-go/bin/lufy-ai` con assets embebidos actualizados y validará instalación/verificación. La publicación de un release estable queda fuera de este cambio porque requiere promoción a `main`, tag `v*` y delivery autorizado.

Alternativa considerada: cambiar versión estable documentada o publicar release directamente. Se descarta porque la política de release exige tags sobre commits alcanzables desde `main` y autorización explícita de delivery.

## Risks / Trade-offs

- [Riesgo] El router clasifica mal un pedido ambiguo. → Mitigación: incluir `confidence`, razón explícita y regla de escalar al tier superior cuando haya duda relevante.
- [Riesgo] T2 se vuelva demasiado informal. → Mitigación: exigir mini-spec con criterios observables y validación esperada.
- [Riesgo] T1 se use de más. → Mitigación: regla explícita de elegir el flujo más pequeño que resuelva el pedido con seguridad.
- [Riesgo] Contratos JSON demasiado rígidos. → Mitigación: mantener campos mínimos obligatorios y permitir notas libres acotadas.
- [Riesgo] Duplicación con `explorer`. → Mitigación: `sdd-router` decide flujo; `explorer` investiga impacto cuando el router lo solicita.
- [Riesgo] Skills externas contradigan reglas locales. → Mitigación: skills locales y `AGENTS.md` tienen prioridad; AutoSkills solo es bootstrap opcional con autorización.
- [Riesgo] El artifact store agregue ruido. → Mitigación: persistencia proporcional: completa en T1, compacta en T2, final/result-only en T3.
- [Riesgo] Documentación y assets embebidos diverjan. → Mitigación: actualizar ambos en el mismo cambio y validar el binario local.
- [Riesgo] Duplicación entre policy y agente delivery. → Mitigación: policy contiene invariantes compartidas; agente contiene ejecución y reporte.
- [Riesgo] Review slices generen burocracia o demasiados PRs. → Mitigación: aplicarlos proporcionalmente; T3 no se fragmenta, T2 solo cuando hay varios riesgos y T1 los usa para bajar carga cognitiva.

## Migration Plan

1. Agregar definición del subagente `sdd-router`.
2. Actualizar `orchestrator` para invocarlo antes de flujos no triviales o ambiguos.
3. Documentar T1/T2/T3 y sus rutas.
4. Agregar template de T2 SDD Lite y result contract.
5. Documentar skill registry local-first y bootstrap opcional con AutoSkills.
6. Verificar que T3 pueda seguir siendo directo sin OpenSpec cuando corresponda.
7. Actualizar documentación pública y operativa del proyecto.
8. Sincronizar assets embebidos de la CLI Go y catálogo instalable.
9. Construir el binario local del installer y validar con el flujo agrupado disponible.
10. Agregar Review Workload Harness al router, templates y documentación, incluyendo `review_slices` proporcionales.

No requiere migración de datos ni cambios de compatibilidad externa.

## Open Questions

- El artefacto T2 Lite vivirá inicialmente como template documental o dentro de una skill dedicada?
- El `sdd-router` debe ejecutarse siempre para pedidos de implementación o solo cuando el `orchestrator` detecte ambigüedad/no trivialidad?
- La compatibilidad con `.agents/skills` de AutoSkills se documentará solo como bootstrap externo o se mapeará también a `.opencode/skills` en un cambio posterior?
- La versión pública siguiente se definirá al preparar delivery/release; esta propuesta no fija el próximo tag.
