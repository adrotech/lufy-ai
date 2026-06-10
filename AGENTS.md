# AGENTS.md

Guía operativa para agentes que trabajan en este repositorio `lufy-ai`.

## Snapshot del proyecto

- **Repositorio**: configuración local de OpenCode y flujo SDD/OpenSpec para `lufy-ai`.
- **CLI del producto**: la CLI Go vive en `tools/lufy-cli-go`; no asumir una CLI legacy fuera de esa ruta.
- **Instalador**: `scripts/install.sh` es un wrapper estricto del CLI Go y no debe reintroducir fallback legacy.
- **Tooling raíz**: no hay `package.json` ni `tsconfig*.json` en la raíz; no asumir comandos Node/TS globales.
- **Tooling `.opencode`**: `.opencode/package.json` contiene dependencias del plugin TUI, no una suite de validación del producto.
- **Validación real**: normalmente estática/documental salvo que la tarea indique un toolchain específico. Siempre reportar comandos ejecutados y resultados reales.
- **Workflow limits**: `.lufy/config/project.yaml` usa `workflow_limits` como única fuente canónica; no consumir `loc_budget` ni `delivery_strategy` top-level como límites válidos.
- **Result Contract envelope v1**: handoffs y resultados sustantivos deben usar el envelope YAML canónico con estado, evidencia, riesgos, siguiente acción y decisión de workflow cuando aplique.
- **Workflow sistémico**: analizar archivos existentes, dependencias e interconexiones al inicio; evitar relecturas repetidas durante implementación; releer al final solo archivos viejos modificados/afectados o casos justificados.
- **Idioma**: respuestas, documentación humana, PRs y comentarios en español; preservar identificadores técnicos, rutas, flags y nombres de comandos.
- **Ramas y releases**: `develop` es la base normal de integración; `main` es productiva/estable; los releases estables se publican solo desde tags `v*` sobre commits alcanzables desde `main`.

## Estructura relevante

- `.opencode/agents/`: definiciones de agentes (`orchestrator`, `sdd-router`, `explorer`, `implementer`, `test-writer`, `validator`, `reviewer`, `delivery`).
- `.opencode/commands/`: slash commands del flujo OpenSpec (`opsx-explore`, `opsx-propose`, `opsx-apply`, `opsx-verify`, `opsx-sync`, `opsx-archive`) y comandos LUFY (`lufy.close`, `lufy.pr-review`, `lufy.onboard`, `lufy.timereport`).
- `.opencode/skills/sdd-workflow/`: skills para explorar, proponer, aplicar, verificar, sincronizar y archivar cambios OpenSpec; skills LUFY transversales viven en `.opencode/skills/lufy.*`.
- `.opencode/plugins/agent-observatory.tsx`: plugin TUI local Agent Observatory.
- `.opencode/policies/delivery.md`: fuente canónica para delivery, branch safety, validación y gates de cambios completos.
- `openspec/`: propuestas, especificaciones y tareas del flujo OpenSpec.
- `tools/lufy-cli-go/`: implementación actual de la CLI Go usada por el instalador.
- `scripts/install.sh`: wrapper estricto hacia `tools/lufy-cli-go`, sin fallback legacy.
- `docs/`: documentación del proyecto cuando exista.
- `AGENTS.md.template`: plantilla genérica; este archivo es la guía real del repo.

## Comandos disponibles y límites

Ejecutar desde la raíz salvo que se indique otra ruta.

- OpenSpec/OpenCode: usar `/opsx-explore`, `/opsx-propose`, `/opsx-apply`, `/opsx-verify`, `/opsx-sync`, `/opsx-archive` cuando corresponda; usar `/lufy.close` para cierre transversal del workflow con PR/branch cleanup; usar `/lufy.pr-review` para reviews HTML de PR en español.
- Observatory TUI: `/observatory`, `/observatory-agents`, `/observatory-subagents`, `/observatory-cost`.
- Validación agrupada local: `scripts/validate.sh` ejecuta el whitespace check con rango/base de PR y la validación Go disponible.
- Git inspección: `git status --short`, `git diff`, `git diff --check`, `git diff --check origin/develop`, `git diff --check origin/develop...HEAD`, `git log` según permisos del rol.
- No inventar `npm test`, `npm run typecheck`, `tsc` u otros comandos si el toolchain no existe para el alcance actual.
- En proyectos frontend con `pnpm` configurado, declarar validaciones no mutantes en `.lufy/config/project.yaml` bajo `validation.allowed_commands.implementer` para que `implementer` las herede sin hardcodearlas en el agente; siguen sujetas a validación agrupada y a que el toolchain exista.
- Respetar la preferencia de validación agrupada: no correr tests constantemente; agrupar tests, coverage y validación completa al final de todas las tareas de un bloque/proposal salvo bloqueo, cambio riesgoso o diagnóstico.
- Evaluar gates por task, bloque coherente o review slice; los micro-checkboxes internos no implican cierre, archive-ready ni delivery por sí solos.
- Para cambios que terminarán en PR contra `develop`, el chequeo de whitespace debe reproducir el rango del PR: usar `git diff --check origin/develop...HEAD` sobre commits ya preparados y `git diff --check origin/develop` cuando haya cambios pendientes en worktree. No basta `git diff --check` local, porque puede omitir whitespace introducido en commits anteriores de la rama.
- Si se requiere validación no disponible, reportar la limitación y la evidencia estática/manual realizada.

## Reglas de arquitectura y workflow

1. Mantener cambios enfocados y mínimos.
2. No revertir ni sobrescribir trabajo local no relacionado.
3. Mantener handlers/controllers delgados; servicios contienen reglas de negocio.
4. No exponer entidades de persistencia como contratos HTTP/API.
5. Usar inyección por constructor donde aplique.
6. Mantener scopes transaccionales estrechos.
7. No cambiar puertos, defaults de auth, esquema de base de datos ni contratos públicos salvo autorización explícita.
8. Añadir o actualizar pruebas/documentación solo cuando estén ligadas al cambio.
9. Nunca afirmar validación exitosa sin evidencia de comando o revisión manual concreta.
10. Preferir lectura y edición específicas sobre exploración amplia.
11. En handoffs y gestión de contexto, resumir decisiones, evitar dumps largos y preservar solo la evidencia mínima útil.
12. Mantener `scripts/install.sh` como wrapper estricto de `tools/lufy-cli-go`; no reintroducir rutas legacy.
13. Aplicar pensamiento sistémico: entender el todo, interconexiones, dependencias, bucles de feedback y cómo la estructura estática produce comportamiento dinámico.
14. Durante una propuesta, concentrar el análisis de código viejo al inicio y la revisión final en archivos viejos modificados/afectados; no releer archivos ya analizados salvo conflicto, bloqueo, nueva evidencia, cambio de alcance o riesgo explícito.
15. Usar routing proporcional T1/T2/T3 para propuestas, funcionalidades y tareas: elegir el flujo más pequeño que resuelva el pedido con seguridad.
16. Mantener aislamiento de subagentes: pasar contexto mínimo, permisos mínimos y contrato de salida claro.
17. Resolver skills local-first desde `.opencode/skills`; AutoSkills puede sugerirse solo como bootstrap/fallback opcional con `npx autoskills --dry-run` y autorización explícita antes de comandos mutantes.
18. Aplicar Review Workload Harness en T1 y T2 con varios riesgos: pensar en el reviewer humano, dividir features grandes en slices revisables y entregar por partes cuando reduzca carga cognitiva; no fragmentar T3 artificialmente.

## Routing SDD proporcional

- **T1 Full SDD**: nuevas capabilities, impacto transversal, arquitectura, contratos públicos, seguridad, delivery policy o alta incertidumbre; usar OpenSpec completo.
- **T2 SDD Lite**: cambio funcional acotado, bug relevante, agente/skill o refactor controlado; usar mini-spec o handoff estructurado con criterios WHEN/THEN y validación agrupada.
- **T3 Express**: cambio trivial, mecánico, documental o local sin riesgo relevante; permitir implementación directa acotada y validación proporcional.
- **Fast path OpenSpec/docs-only**: si el programa global es T1 pero el siguiente micro-slice toca solo 1-2 artefactos OpenSpec/docs, no cambia runtime, no requiere delivery y tiene aceptación clara, clasificar el slice como T2/T3 con `fast_path_allowed: true`; validación esperada: `openspec validate "<change>" --strict` cuando aplique y revisión estática de archivos/checklists.
- Escalar T3 → T2 si aparece comportamiento incierto, criterios no observables o alcance mayor al previsto.
- Escalar T2 → T1 si aparecen decisiones de arquitectura, impacto transversal, contratos públicos, seguridad o alta incertidumbre.
- Para T1 y T2 con varios ejes de riesgo, definir `review_slices` con objetivo, archivos esperados, criterios WHEN/THEN, validación, riesgo y guía de PR.
- Para sizing/routing/slicing, leer `workflow_limits.sizing`, `workflow_limits.routing` y `workflow_limits.proposal_slicing_strategy`; no confundirlo con `workflow_limits.delivery_batch_strategy`.
- Para paralelismo, leer `parallel_execution`; recomendar agentes paralelos solo para `review_slices` independientes, archivos independientes, plan de merge y validación agrupada después del join. No paralelizar delivery, migraciones schema/db, contratos públicos compartidos, decisiones API no cerradas ni slices que tocan los mismos archivos.
- Delivery nunca queda autorizado por el tier; requiere autorización explícita del usuario y rol `delivery`.
- Estados de gate por bloque: `implemented` = cambios aplicados y validación pendiente; `validated` = evidencia proporcional registrada; `delivery_pending` = falta autorización/ejecución Git/GH, checks remotos existentes aún pendientes o sync; `delivered` = delivery autorizado ejecutado con checks remotos requeridos exitosos y evidenciados; `closed` = implementación, validación, delivery/checks remotos/sync requeridos y precondiciones satisfechas.

## Result Contract envelope v1

Usar este envelope para handoffs y resultados sustantivos de agentes locales. Para T3 simples, mantenerlo compacto con `not_applicable`; para salidas legacy/terceros, `orchestrator` puede normalizar con `legacy_fallback: true` y marcar evidencia faltante como `not_available`.

```yaml
schema_version: result-contract/v1
status: ready | implemented | validated | delivery_pending | sync_pending | blocked | escalated | delivered | closed
legacy_fallback: false
executive_summary: <1-3 lineas en espanol>
artifacts:
  changed:
    - <path or none>
  referenced:
    - <path/spec/PR or none>
evidence:
  commands:
    - command: <command or none>
      result: passed | failed | blocked | not_run
      notes: <key output or reason>
  static:
    - <manual/static evidence or not_applicable>
workflow_decision:
  tier: T1 | T2 | T3 | not_applicable
  program_tier: T1 | T2 | T3 | not_applicable
  slice_tier: T1 | T2 | T3 | not_applicable
  fast_path_allowed: true | false | not_applicable
  adapter_context:
    tool_id: opencode | not_applicable
    methodology_id: openspec | lufy-sdd | none | not_applicable
    methodology_mode: full | lite | none | not_applicable
    methodology_required: true | false | not_applicable
    execution_mode: full-sdd | sdd-lite | express | not_applicable
  workflow_limits_source: workflow_limits | not_available
  workflow_limits_paths:
    sizing: workflow_limits.sizing | not_available
    routing: workflow_limits.routing | not_available
    proposal_slicing: workflow_limits.proposal_slicing_strategy | not_available
    delivery_batching: workflow_limits.delivery_batch_strategy | not_applicable
    preflight: workflow_limits.preflight | not_available
    stop_rules: workflow_limits.stop_rules | not_available
  workload_decision_needed: true | false
  review_slices:
    - <slice summary or not_applicable>
  preflight_status: passed | blocked | not_applicable | not_available
  stop_rule_status: clear | triggered | not_applicable | not_available
  delivery_batching_guidance: <guidance or not_applicable>
risks:
  - <risk/follow-up or none>
next_recommended:
  owner: orchestrator | explorer | implementer | test-writer | validator | reviewer | delivery | user | none
  action: <next action>
skill_resolution:
  local_skills_used:
    - <skill or none>
  bootstrap_recommended: true | false
  notes: <notes or none>
```

## Roles de agentes

- `orchestrator`: coordina y enruta; no edita ni ejecuta shell.
- `sdd-router`: clasifica T1/T2/T3 en modo read-only/no-shell, recomienda execution mode, contexto mínimo, skill status y review workload; no ejecuta shell/Git/OpenSpec/validación y deriva a `explorer`, `validator` o `delivery` cuando se requiere estado, evidencia o Git/GH.
- `explorer`: investiga en modo read-only y produce handoff para implementación.
- `implementer`: implementa cambios acotados; no hace commit, push, PR ni sync de Projects.
- `test-writer`: escribe o ajusta pruebas TDD stack-aware para cambios T1/T2 sustantivos y reporta evidencia RED/GREEN/TRIANGULATE/REFACTOR; no hace delivery.
- `validator`: valida y diagnostica en modo read-only; no edita.
- `reviewer`: revisa calidad, riesgos y cobertura con scoring L1-L5 stack-aware; no edita.
- `delivery`: con autorización explícita, maneja Git/GH, PRs y trazabilidad siguiendo `.opencode/policies/delivery.md`.

## OpenSpec workflow

- Explorar idea o impacto: `opsx-explore` / skill `openspec-explore`.
- Crear propuesta completa: `opsx-propose` / skill `openspec-propose`.
- Implementar tareas: `opsx-apply` / skill `openspec-apply-change`.
- Verificar implementación contra artefactos: `opsx-verify` / skill `openspec-verify-change`.
- Sincronizar deltas validados a specs principales: `opsx-sync` / skill `openspec-sync`.
- Archivar cambio completado: `opsx-archive` / skill `openspec-archive-change`.
- Cerrar/finalizar spec activa o cambio LUFY con gates de validación, sync, delivery, PR cerrado/merged y limpieza segura de rama: `/lufy.close` / skill `lufy.close`.
- Una tarea OpenSpec marcada en `tasks.md` no equivale por sí sola a `closed` ni `archive-ready`; solo se considera cerrada si cumple los gates de `.opencode/policies/delivery.md` con estado explícito.
- En `opsx-apply`, completar tareas por bloque sin test loops ni relecturas rutinarias; en `opsx-verify`, correr la validación final agrupada disponible, incluyendo tests/coverage solo si existen para el alcance real.
- Foco activo actual: `install-managed-assets-with-hash-idempotency` (assets gestionados, SHA-256, manifest, idempotencia, backup/restore y verify estructural).
- No archivar `migrate-installer-to-go-cli` mientras tenga tasks incompletas; tasks incompletas implican `blocked`, no archive.

## Workflow de memoria Obsidian

- Obsidian es la memoria canónica portable cuando `.lufy/config/project.yaml` declara `memory.provider: obsidian`; usar `lufy-ai memory status/search/validate` y los skills locales `lufy.mem-*` cuando el contexto histórico aporte.
- Para trabajo T1/T2 no trivial, y para T3 con contexto histórico probable, buscar en Obsidian con consultas cortas por issue/spec/ruta/concepto y resumir hallazgos como `memory_hints` compactos (path, línea, status, relevancia); no pasar dumps completos.
- Después de trabajo significativo, guardar en Obsidian solo aprendizajes durables: decisiones de arquitectura, reglas, flows, lessons, patrones reutilizables, cambios de configuración, gotchas, outcomes de delivery o resúmenes de sesión. No guardar ruido rutinario ni estados duplicados.
- Nunca tratar memoria como evidencia más fuerte que archivos, comandos o instrucciones explícitas. La trazabilidad durable de memoria se registra en Obsidian.

## Política de delivery

- Consultar `.opencode/policies/delivery.md` para validación por tiers, branch safety, PRs, sync y estados `blocked` / `sync_pending`.
- PR normal: ramas `feature/*`, `fix/*`, `chore/*` o equivalentes → `develop`.
- Promoción productiva: `develop` → `main` con autorización y evidencia de validación.
- `main` no es base de trabajo diario; se reserva para producción, release y hotfix explícitamente autorizado.
- Tags de release estable: `v*` creados desde commits alcanzables desde `origin/main`; no publicar releases desde `develop` sin promoción.
- No hacer commit, push, PR ni actualizar GitHub Projects sin autorización explícita del usuario y rol `delivery`.
- No crear PR desde ramas protegidas como `develop`, `main`, `master` o `development`, salvo promoción `develop` → `main` explícitamente autorizada.
- Al crear PR, `delivery` debe consultar/esperar checks remotos con evidencia (`gh pr checks <PR>` o `gh pr view ... statusCheckRollup/mergeStateStatus`) y no reportar `delivered`/`closed` si fallan, quedan pendientes o falta evidencia; usar `blocked` o `delivery_pending` con recovery.
- Nunca usar force push salvo solicitud explícita.

## Formato de reporte

- Incluir objetivo, cambios/evidencia, riesgos y estado listo/bloqueado.
- Mantener resúmenes concisos; usar rutas y líneas cuando ayuden.
- Si falta contexto o una decisión, pedirla o devolver el bloqueo exacto.
