# Tasks: extend-context-graph-workflow-traceability

## Review Workload Harness

| Slice | Objetivo | Archivos/áreas esperadas | Criterio de salida | Validación agrupada | Riesgo | PR guidance |
| --- | --- | --- | --- | --- | --- | --- |
| A | Dominio y extracción semántica | `contextgraph/domain`, `extractors` | nodos/edges deterministas solo con refs explícitas | unit + property/golden + privacy canaries | falsos enlaces | same_pr |
| B | Coverage, trace y freshness | `contextgraph/application`, `cli` | gaps observables y fallback no optimista | service + CLI integration | grafo stale tratado como evidencia | same_pr |
| C | Budgets de review | `projectconfig`, Git adapter, evaluator | limits 8/800/3/2 y decisión reproducible | config/rescan + decision table + cross-platform | bloqueo excesivo o legacy source | same_pr |
| D | Runs, métricas y harness | `runledger` port, roles, contracts, docs/assets | metadata content-free y consumidores sincronizados | ledger/CLI + coupling/parity + E2E | privacidad y drift de assets | same_pr |

`workload_decision_needed=true`: el alcance estimado supera 4 archivos y probablemente el LOC budget default. Se mantiene un PR único porque los slices comparten schemas y CLI; se debe reevaluar si un slice crece o toca un contrato público no previsto.

- [x] 1. Slice A — Contrato y RED de trazabilidad
  - [x] 1.1 Definir node/edge kinds, IDs, provenance, diagnostics y límites bounded.
  - [x] 1.2 Escribir RED para specs, requirements, scenarios, tasks, decisions, tests y refs GitHub.
  - [x] 1.3 Escribir RED de markers `implements/verifies/depends_on/caused_by/reviewed_by/supersedes`, refs rotas y orden determinista.
  - [x] 1.4 Añadir canaries de secretos, prompts, outputs y runtime paths.

- [x] 2. Slice A — Extractores y grafo semántico
  - [x] 2.1 Implementar extracción semántica sin fuzzy evidence.
  - [x] 2.2 Implementar normalización/allowlist de workflow edges y diagnostics de refs inválidas.
  - [x] 2.3 Versionar extractor/cache y preservar schema compatible.
  - [x] 2.4 Ejecutar validación agrupada Slice A y registrar TDD.

- [x] 3. Slice B — Coverage, trace y freshness
  - [x] 3.1 Escribir RED para scenario sin task/test, archivo diff sin trace y cobertura completa.
  - [x] 3.2 Centralizar preflight `ready/stale/not_available` antes de queries probatorias.
  - [x] 3.3 Implementar `context coverage` y trace paths con resultados bounded.
  - [x] 3.4 Implementar fallback seguro: diff directo disponible, coverage `unknown`, recovery explícito.
  - [x] 3.5 Añadir CLI human/JSON, exit codes y pruebas Slice B.

- [x] 4. Slice C — `workflow_limits.review`
  - [x] 4.1 Escribir RED de defaults, YAML parcial, rescan, extras y rechazo/diagnóstico de valores inválidos.
  - [x] 4.2 Tipar `max_files_per_slice`, `max_churn_lines_per_slice`, `max_concurrent_slices` y `min_evidence_items`.
  - [x] 4.3 Preservar overrides user-owned y mantener top-level legacy como no canónico.
  - [x] 4.4 Extender Git adapter con numstat bounded y portabilidad CRLF/path.
  - [x] 4.5 Implementar evaluator `proceed/split/escalate` y `context review`.
  - [x] 4.6 Ejecutar tabla de decisiones y validación agrupada Slice C.

- [x] 5. Slice D — Runs y métricas content-free
  - [x] 5.1 Escribir RED de run/task/evidence/causalidad y availability de métricas.
  - [x] 5.2 Implementar port/projection mínima de Run Ledger sin escanear `.lufy/runtime/**`.
  - [x] 5.3 Derivar review time, rework y reopened defects desde event names/content-free metadata.
  - [x] 5.4 Implementar `context metrics` con `available/partial/unavailable`.
  - [x] 5.5 Probar ausencia durable de prompts, summaries, outputs, secretos y paths runtime.

- [x] 6. Slice D — Harness, contratos y documentación
  - [x] 6.1 Integrar assessment en router/orchestrator/reviewer/validator/delivery sin delegar autoridad al grafo.
  - [x] 6.2 Actualizar Result Contract, `AGENTS.md.template`, delivery contract y docs.
  - [x] 6.3 Mantener root/embedded assets y specs en paridad.
  - [x] 6.4 Ejecutar coupling/catalog/parity y E2E de Slice D.

- [ ] 7. Verify, sync y delivery readiness
  - [x] 7.1 Ejecutar suite, build, race focalizado y `scripts/validate.sh`.
  - [x] 7.2 Verificar proposal/design/spec, privacidad, fallback y cross-platform.
  - [x] 7.3 Registrar evidencia en `verification/extend-context-graph-workflow-traceability/` y strict validate.
  - [x] 7.4 Sincronizar deltas validados.
  - [x] 7.5 Preparar trazabilidad con issue `#221`; no iniciar Fase 5 antes de merge/cierre.
  - [ ] 7.6 No archivar ni cerrar sin delivery, checks, sync y gates completos.
