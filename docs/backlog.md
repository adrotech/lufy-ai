# Backlog estratégico de `lufy-ai`

Este backlog consolida el archivo externo `/Users/adrianrojas/Downloads/lufy-ai-backlog.md` con el estado actual del repositorio. La dirección de producto queda explícita: `lufy-ai` debe seguir siendo un harness generalista para repositorios existentes, no un harness hardcodeado a Go.

El foundation técnico del próximo ciclo es un scanner inicial (`lufy-ai init`) que detecta el stack del proyecto destino y genera `.opencode/project.yaml`. Agentes, skills, hooks, review y telemetría deben leer ese archivo en vez de asumir comandos fijos como `go test`, `gofmt` o `golangci-lint`.

## Decisiones de alcance

- Las mejoras nuevas deben ser genéricas por stack: Go, TypeScript/JavaScript/React/Next.js, Python y Java/Kotlin son el soporte v1.
- Otros stacks pueden detectarse como `supported: false` con placeholders editables; no bloquean `init`.
- Se preservan `/opsx-*` como comandos canónicos del workflow OpenSpec.
- Los comandos propios nuevos del kit usan namespace `/lufy.*`.
- No se integra Jira lifecycle en este backlog.
- No se migra `orchestrator` de agent a skill; se mantiene como agent OpenCode.
- Las features grandes deben convertirse en proposals OpenSpec separadas antes de implementación.

## Estado base actual

Capacidades ya existentes que afectan la planificación:

| Área | Estado actual | Impacto sobre backlog |
| --- | --- | --- |
| CLI Go | Existe en `tools/lufy-cli-go` con `install`, `verify`, `backup`, `restore`, `sync`, `status`, `merge`, `upgrade` y `version`. | Los items de CLI deben extender lo existente, no rediseñarlo desde cero. |
| Drift Resolution | Ya incluye policies por asset, `.lufy-new`, ancestors, `merge-block`, `--scope`, restore por ID/listado. | `LUFY-8`, `LUFY-9` y `LUFY-10` se implementan como refinamiento de planner/governance. |
| OpenSpec core v2 | Ya incluye config action-based, deltas, scenarios, `/opsx-sync`, `UPSTREAM.json` y resolver PATH/cache/embedded. | `LUFY-12` debe ser documental y de naming para comandos nuevos, no renombre de `/opsx-*`. |
| Result contracts | Ya hay templates y reglas base. | `LUFY-15` es normalización/enforcement, no introducción desde cero. |
| Templates por stack | No instalables hoy. | `LUFY-17` queda P3 y debe esperar `project.yaml`. |

## Prioridades

- **P0**: bloquea adopción seria o desbloquea la mayoría del backlog.
- **P1**: diferenciador competitivo con impacto directo en calidad, UX o gobernanza.
- **P2**: madurez operativa, documentación y consistencia.
- **P3**: largo plazo, expansión por dominio o stack.

Effort estimado:

| Tamaño | Rango |
| --- | --- |
| XS | Menos de medio día |
| S | Medio día a 1 día |
| M | 1 a 3 días |
| L | 3 a 7 días |
| XL | Más de 1 semana |

## P0 - Foundation stack-aware

### LUFY-0 - `lufy-ai init` y `.opencode/project.yaml`

**Problema:** la disciplina operativa del harness no puede depender de Go. El repo destino debe declarar reglas detectadas y editables para test, lint, format, coverage, observabilidad y anti-patrones.

**Alcance:**

- Agregar comando `lufy-ai init [--target <path>] [--force] [--rescan]`.
- Detectar stacks por archivos raíz: `go.mod`, `package.json`, `tsconfig.json`, `pyproject.toml`, `requirements.txt`, `setup.py`, `pom.xml`, `build.gradle`, `build.gradle.kts` y equivalentes conocidos.
- Detectar frameworks TS/JS: React, Next.js, Remix, Vue y Svelte; Vue/Svelte pueden quedar como placeholders si no hay reglas completas.
- Detectar test runners, linters, formatters, static analysis, CI y librerías de observabilidad.
- Generar `.opencode/project.yaml` con `schema_version`, `detected_at`, `stacks`, `ci`, `tdd` y `workflow_limits`.
- No sobrescribir un archivo existente sin `--force`.
- `--rescan` debe preservar overrides manuales y agregar stacks nuevos sin borrar configuración previa.

**Acceptance:**

- En repo Go puro, `init` genera stack `go` con comandos Go y thresholds configurables.
- En repo TS/Next, `init` genera stack `typescript` con frameworks `react` y `next`.
- En monorepo Go + TS, `stacks` contiene ambas entradas.
- En repo Rust, detecta `rust` como `supported: false` con placeholders editables.
- Cambiar manualmente `coverage_threshold` y correr `--rescan` preserva el override.

**Effort:** L.

### LUFY-1 - Agent `test-writer` parametrizado por `project.yaml`

**Problema:** el ciclo TDD debe ser consistente y verificable sin asumir lenguaje.

**Alcance:**

- Crear `.opencode/agents/test-writer.md`.
- Definir ciclo RED -> GREEN -> TRIANGULATE -> REFACTOR.
- Cargar comandos y anti-patrones desde `.opencode/project.yaml`.
- Modificar `implementer.md` para delegar tests sustantivos a `test-writer` en T1/T2.
- Modificar `validator.md` para bloquear T1/T2 si falta evidencia de triangulación cuando aplica.
- Registrar evidencia TDD en `apply-progress.md` o result contract del bloque.

**Acceptance:**

- En Go usa comando de test del stack Go y rechaza anti-patrones Go configurados.
- En TS usa runner detectado y rechaza anti-patrones TS configurados.
- En Python usa runner detectado y rechaza anti-patrones Python configurados.
- El threshold de coverage proviene del stack correspondiente.

**Effort:** M.

### LUFY-2 - Reviewer L1-L5 ponderado y HTML capability-aware

**Problema:** el reviewer debe producir evaluación consistente, exportable y adaptada al stack.

**Alcance:**

- Reemplazar o extender `.opencode/agents/reviewer.md` con scoring ponderado.
- Pesos base: Architecture 20%, Code Quality 15%, Simplicity 15%, Testing 20%, Observability 15% y PR Template gate.
- Aprobar solo con score >=80% y cero hallazgos L1/L2.
- Exigir desk-check de al menos 8 escenarios para cambios T1/T2 relevantes.
- Cargar anti-patrones, coverage y observability libs desde `.opencode/project.yaml`.
- Crear skill opcional `.opencode/skills/lufy.pr.review/` para HTML autocontenido en `/tmp/pr-review-<id>.html`.

**Acceptance:**

- En PR TS no busca librerías Go; usa `@opentelemetry/*`, `pino`, `winston` o lo declarado.
- En PR Go usa observability libs Go declaradas.
- El HTML contiene secciones scored, severidades L1-L5 y resultado final.

**Effort:** M.

## P1 - Diferenciadores competitivos

### LUFY-3 - `/lufy.timereport` con HTML autocontenido

**Estado:** entregado por PR #66; spec sincronizada y change archivado post-merge el 2026-05-24.

**Alcance:** crear skill `.opencode/skills/lufy.timereport/SKILL.md` que lea sesiones JSONL de OpenCode, `git log` y `.opencode/project.yaml` para generar KPIs de ROI y timeline de trabajo.

**Acceptance:** HTML offline con wall-clock, AI working time, tiempo humano activo, LOC neto, commits, tool calls, top tools, subagents, skills, fases y stack detectado.

**Effort:** L.

### LUFY-4 - `/lufy.onboard` con dry-run y demo stack-aware

**Alcance:** crear skill `.opencode/skills/lufy.onboard/SKILL.md` con validación de instalación y modo `--demo` que genera un T3 dummy adaptado al stack detectado.

**Acceptance:** usuario nuevo en repo TS/Go/Python puede ejecutar demo y entender el flujo en menos de 10 minutos. Si falta `project.yaml`, sugiere `lufy-ai init`.

**Effort:** S.

### LUFY-5 - Stop Rules numéricas en `orchestrator`

**Alcance:** agregar reglas explícitas a `.opencode/agents/orchestrator.md`: 4-file rule, 20-tool-calls rule, multi-file write rule y long-session rule.

**Acceptance:** una feature estimada en 5+ archivos escala tier o se divide antes de edición amplia.

**Effort:** S.

### LUFY-6 - Workload Guard config-driven

**Alcance:** extender `sdd-router.md` y `orchestrator.md` para leer `workflow_limits.sizing`, `workflow_limits.routing`, `workflow_limits.proposal_slicing_strategy`, `workflow_limits.delivery_batch_strategy`, `workflow_limits.preflight`, `workflow_limits.stop_rules` y `chain_strategy` desde `.opencode/project.yaml`.

**Acceptance:** si `estimated_loc > workflow_limits.sizing.loc_budget`, el router emite `workload_decision_needed: true` y propone `review_slices` según `workflow_limits.proposal_slicing_strategy`; delivery agrupa con `workflow_limits.delivery_batch_strategy`; con `auto-chain`, propaga estrategia sin preguntar salvo riesgo alto.

**Effort:** M.

### LUFY-7 - Hook PostToolUse de formato dinámico

**Alcance:** crear `.opencode/hooks/format-dispatch.sh` que lea `.opencode/project.yaml`, matchee extensión editada y ejecute formatter/linter auto-fix del stack.

**Acceptance:** formatea `.go`, `.ts/.tsx` y `.py` según configuración; archivos desconocidos salen con código 0 sin ruido.

**Effort:** S.

### LUFY-8 - CLI `merge` 3-way refinado

**Estado:** ya existe `lufy-ai merge` y Drift Resolution con `.lufy-new`/ancestors.

**Alcance restante:** consolidar motor text 3-way, UX no interactiva (`--accept-theirs`, `--accept-ours`) y decidir si se adopta TUI con dependencia explícita o se mantiene zero-deps.

**Acceptance:** conflicto entre asset local y catalog crea sidecar seguro; `merge --accept-theirs` resuelve sin TUI y registra estado coherente.

**Effort:** M-L según decisión de TUI.

### LUFY-9 - CLI governance: `pin`, `unpin`, `doctor`, `info`, `status`

**Estado:** `status` ya existe.

**Alcance restante:** agregar `pin`, `unpin`, `doctor` e `info`; extender `status` para stacks, drift, conflicts pending y frozen assets.

**Acceptance:** asset pinned no es tocado por `sync`; `doctor` valida preflight, manifest y `project.yaml`; `info` muestra catalog version, assets y stacks.

**Effort:** M.

### LUFY-10 - Planner 8-state

**Alcance:** extender planner actual a `Skip | Create | UpdateManaged | Conflict | Frozen | Template | Merge | Remove`.

**Acceptance:** assets removidos del catálogo se eliminan solo si no fueron modificados; assets con drift se preservan con warning; templates se renderizan una vez con variables de `project.yaml`.

**Effort:** L.

## P2 - Madurez operativa

### LUFY-11 - Lessons learned versionado

**Alcance:** crear `docs/lessons/lufy-ai.md` con bugs históricos, decisiones rechazadas y decisiones no-ADR.

**Acceptance:** archivo existe con al menos 3 entradas seed y README lo enlaza.

**Effort:** S.

### LUFY-12 - Namespace dual `/opsx-*` + `/lufy.*`

**Alcance:** documentar que `/opsx-*` se preserva para OpenSpec y `/lufy.*` se usa para extras propios.

**Acceptance:** README y docs distinguen ambos namespaces; no se renombra ningún `/opsx-*` existente.

**Effort:** XS.

### LUFY-13 - README walkthrough end-to-end

**Alcance:** agregar walkthrough desde instalación, `lufy-ai init`, `/lufy.onboard --demo`, primer T3 y `/lufy.timereport`.

**Acceptance:** usuario nuevo llega a primera feature demo en menos de 10 minutos siguiendo solo README.

**Effort:** S.

### LUFY-14 - Verificación activa post-spec

**Alcance:** agregar al orchestrator una regla de lectura/verificación del archivo esperado después de invocar comandos de spec/sync; si Engram está habilitado, validar registro de delta cuando aplique.

**Acceptance:** falla simulada de generación de spec produce STOP con error accionable.

**Effort:** S.

### LUFY-15 - Result Contract envelope v1 unificado

**Estado:** validado localmente mediante `standardize-result-contract-workflow-decisions`; envelope v1 definido para agentes locales, conectado a decisiones `workflow_limits` y sincronizado a specs principales.

**Alcance restante:** cerrar delivery autorizado y archive del change.

**Acceptance:** cada agent produce YAML válido con `schema_version`, `status`, `executive_summary`, `artifacts`, `next_recommended`, `risks` y `skill_resolution`.

**Effort:** M.

### LUFY-16 - `init --rescan` y drift de stack

**Dependencia:** LUFY-0.

**Alcance:** detectar stack drift, stacks nuevos, stacks removidos y edad de `detected_at`; `doctor --check-stack-drift` advierte si el scan está viejo o cambió la estructura del repo.

**Acceptance:** agregar `package.json` a repo Go puro propone stack TS sin tocar config Go manual.

**Effort:** S.

## P3 - Largo plazo

### LUFY-17 - Templates stack-specific por capability

**Alcance:** paquetes opcionales activables por `.opencode/project.yaml` para Go, TS/Next, Python/FastAPI, Java/Spring y Rust.

**Acceptance:** cada paquete incluye assets reales, categoría en manifest, validación y activación condicional por stack.

**Effort:** XL por stack.

### LUFY-18 - Domain-specific subagents

**Alcance:** subagentes especializados como `crud-designer`, `api-contract-reviewer`, `schema-migration-validator` y `dependency-upgrade-reviewer`.

**Acceptance:** cada subagent tiene trigger claro, contrato de salida, validación y criterios para invocación condicional.

**Effort:** XL por subagent end-to-end.

## Dependencias

```text
LUFY-0 init + project.yaml
  -> LUFY-1 test-writer
  -> LUFY-2 reviewer stack-aware
  -> LUFY-3 timereport con stack label
  -> LUFY-4 onboard demo stack-aware
  -> LUFY-6 workload guard
  -> LUFY-7 format dispatch
  -> LUFY-9 doctor/info stack-aware
  -> LUFY-10 template variables
  -> LUFY-16 rescan/drift
  -> LUFY-17 templates stack-specific

LUFY-15 result contract
  -> LUFY-1 test-writer evidence
  -> LUFY-2 reviewer output
  -> LUFY-6 workload decisions

LUFY-8 merge refinado
  -> LUFY-9 pin/doctor/info
  -> LUFY-10 planner 8-state

LUFY-5 stop rules y LUFY-14 post-spec verification son independientes.
LUFY-11, LUFY-12 y LUFY-13 son documentales/operativos y paralelizables.
```

## Plan de implementación recomendado

### Release A - Foundation genérica (`v0.4.0` sugerido)

Objetivo: introducir configuración stack-aware sin cambiar aún todos los agentes.

| Slice | Items | Proposal sugerida | Riesgo | Validación mínima |
| --- | --- | --- | --- | --- |
| A1 | LUFY-0 núcleo | `add-stack-aware-project-init` | Alto: schema nuevo y persistencia config | Cubierto en repo; validar/limpiar change local stale antes de archive si reaparece activo |
| A2 | LUFY-16 rescan/drift | `add-project-rescan-stack-drift` | Medio: merge de overrides | Cubierto por PR #65 y archive correspondiente |
| A3 | LUFY-12 docs namespace | `document-lufy-command-namespace` | Bajo | Revisión documental y `git diff --check` |

### Release B - Agentes y gates (`v0.4.1` sugerido)

Objetivo: hacer que el workflow use `project.yaml` en TDD, review y routing.

| Slice | Items | Proposal sugerida | Riesgo | Validación mínima |
| --- | --- | --- | --- | --- |
| B1 | LUFY-15 | `standardize-result-contract-workflow-decisions` | Medio: coordinación entre agentes | Revisión de agentes, fixtures de outputs, validación documental |
| B2 | LUFY-1 | `add-stack-aware-test-writer` | Alto: cambio de flujo T1/T2 | Simulaciones Go/TS/Python y revisión de gates validator |
| B3 | LUFY-2 | `add-scored-stack-aware-reviewer` | Medio | PR dry-run o fixture de review, HTML si se incluye skill |
| B4 | LUFY-5 + LUFY-6 | `add-numeric-stop-rules-workload-guard` | Medio | Casos de router/orchestrator con estimated LOC y slices |
| B5 | LUFY-14 | `add-active-post-spec-verification` | Bajo-medio | Simulación de spec faltante y spec válido |

### Release C - UX y automatización (`v0.5.0` sugerido)

Objetivo: mejorar onboarding, telemetría y formateo automático con configuración dinámica.

| Slice | Items | Proposal sugerida | Riesgo | Validación mínima |
| --- | --- | --- | --- | --- |
| C1 | LUFY-7 | `add-stack-aware-format-dispatch-hook` | Medio: hooks y latencia | Smokes por extensión, archivos desconocidos, config ausente |
| C2 | LUFY-4 + LUFY-13 | `add-lufy-onboard-walkthrough` | Bajo-medio | Demo dry-run por stack y revisión README |
| C3 | LUFY-3 | `add-lufy-timereport` | Medio: parsing de sesiones | Entregado por PR #66; spec sincronizada y archivada post-merge |
| C4 | LUFY-11 | `add-lessons-learned-log` | Bajo | Archivo seed y enlace desde README/docs |

### Release D - Gobernanza avanzada de CLI (`v0.5.x` sugerido)

Objetivo: completar freeze/pin, planner granular, merge refinado y auditoría.

| Slice | Items | Proposal sugerida | Riesgo | Validación mínima |
| --- | --- | --- | --- | --- |
| D1 | LUFY-8 refinado | `refine-cli-three-way-merge` | Medio-alto | Tests merge, conflicts, accept ours/theirs |
| D2 | LUFY-9 | `add-cli-governance-commands` | Medio | Tests pin/unpin/doctor/info/status JSON |
| D3 | LUFY-10 | `extend-planner-eight-state` | Alto: planner e install-state | Golden tests plan, sync/remove/template/merge/frozen |

### Release E - Expansión por stack y dominio (`v0.6+` sugerido)

Objetivo: sumar valor especializado una vez estable el foundation genérico.

| Slice | Items | Proposal sugerida | Riesgo | Validación mínima |
| --- | --- | --- | --- | --- |
| E1 | LUFY-17 primer stack | `add-stack-specific-capability-pack-go` o `...-typescript-next` | Alto | Assets reales, manifest, smokes en repo fixture |
| E2 | LUFY-18 primer subagent | `add-api-contract-reviewer-agent` | Alto | Fixtures de OpenAPI/consumers y contrato de salida |

## Orden operativo sugerido

1. Empezar por `add-stack-aware-project-init`.
2. Cerrar `project.yaml` y `--rescan` antes de tocar agentes que dependan de esa config.
3. Estandarizar result contracts antes de `test-writer` y reviewer para que la evidencia sea parseable.
4. Implementar `test-writer` y reviewer en paralelo solo si el schema de `project.yaml` ya está estable.
5. Agregar stop rules/workload guard antes de features grandes para reducir PRs gigantes.
6. Implementar hooks, onboarding y timereport cuando el foundation esté consumible por usuarios.
7. Dejar planner 8-state para después de pin/merge refinado, porque toca install-state y sync.
8. Posponer templates stack-specific y domain agents hasta validar adopción real del flujo stack-aware.

## Open questions antes de implementar

| Pregunta | Recomendación inicial |
| --- | --- |
| Nombre del archivo de config | Usar `.opencode/project.yaml` para scope OpenCode; evitar `lufy.config.yaml` salvo necesidad externa. |
| Parser YAML en Go | Evaluar dependencia mínima o formato JSON. Si se mantiene YAML, justificar dependencia y validarla en supply chain. |
| TUI para merge | Mantener zero-deps inicialmente; adoptar Bubble Tea solo si la UX interactiva lo justifica. |
| Thresholds default | Go 85%, TS/Python/Java 80% como defaults editables. |
| Hooks automáticos | Deben ser opt-in o instalados disabled-by-default si hay riesgo de latencia en repos grandes. |
| `project.yaml` user-managed | Preservar overrides; no regenerar destructivamente. |

## Verificación transversal

Para cada proposal derivada:

- Ejecutar análisis sistémico inicial sobre archivos afectados y dependencias.
- Definir escenarios `WHEN/THEN` observables en OpenSpec.
- Agrupar validación al final del bloque, no en loops innecesarios.
- Usar `scripts/validate.sh` para cambios de CLI/assets cuando aplique.
- Usar `git diff --check origin/develop` o rango PR-aware antes de delivery.
- Reportar limitaciones reales si no hay toolchain disponible.
- No afirmar soporte de stack o comando si no existe asset, scanner, test o doc correspondiente.

## Resumen de esfuerzo

| Prioridad | Items | Effort total aproximado |
| --- | --- | --- |
| P0 | LUFY-0 a LUFY-2 | 5 a 13 días |
| P1 | LUFY-3 a LUFY-10 | 12 a 26 días |
| P2 | LUFY-11 a LUFY-16 | 4 a 9 días |
| P3 | LUFY-17 a LUFY-18 | 2+ semanas |

P0+P1+P2 equivale a 21 a 48 días secuenciales. Con proposals paralelizables después de LUFY-0/LUFY-15, el calendario real puede reducirse si se separan PRs por slice y se evita mezclar CLI, agentes y docs en una sola entrega.
