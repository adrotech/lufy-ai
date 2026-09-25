# Tasks: experiment-adaptive-role-allocation

## Review Workload Harness

| Slice | Objetivo | Archivos/áreas esperadas | Criterio de salida | Validación agrupada | Riesgo | PR guidance |
| --- | --- | --- | --- | --- | --- | --- |
| A | Dominio, config y scoring | `adaptive/domain`, `projectconfig` | contratos bounded, disabled-by-default y ranking determinista | decoder/property/table + config/rescan | score opaco o activación accidental | same_pr |
| B | Ledger, leases, pool y yield | `runledger`, `adaptive/adapters` | append/idempotency/fencing y release durable | concurrency/race/crash/golden/privacy | doble assignment o release parcial | same_pr |
| C | Application y CLI | `adaptive/application`, `cli` | shadow/advisory y comandos con JSON estable | service + CLI matrix + E2E | recomendación confundida con autoridad | same_pr |
| D | Harness, contratos y docs | roles, contracts, managed/embedded assets, docs | consumidores sincronizados sin autonomía protegida | coupling/parity/full suite | drift de assets o privilegios | same_pr |

`workload_decision_needed=true`: el alcance supera `workflow_limits.sizing` habitual y toca contratos transversales. Se mantiene un PR único porque los slices comparten schemas, ledger y CLI. Detener y reevaluar si un slice supera el límite de review, requiere storage paralelo o introduce un modo autónomo.

- [x] 1. Slice A — Contratos y pruebas RED
  - [x] 1.1 Definir schemas/bounds de `DemandSignal`, `CapabilityProfile`, `Assignment`, `WaitingItem` y `YieldCheckpoint`.
  - [x] 1.2 Escribir RED para unknown/duplicate/oversized, invalid UTF-8, rangos, refs SHA-256 y enums.
  - [x] 1.3 Escribir tabla RED de scoring, ineligibility, protected boundaries y tie-break estable.
  - [x] 1.4 Añadir canaries de prompts, outputs, summaries, secretos, paths e hipótesis/intentos en texto.

- [x] 2. Slice A — Dominio y configuración
  - [x] 2.1 Implementar `adaptive/domain` puro con canonicalización, policy `deterministic-v1` y breakdown completo.
  - [x] 2.2 Implementar `adaptive_routing` disabled-by-default con modes, TTL y bounds.
  - [x] 2.3 Preservar YAML parcial/extras y completar defaults mediante rescan sin activar el feature.
  - [x] 2.4 Consumir límites existentes como señales sin crear aliases ni autoridad nueva.
  - [x] 2.5 Ejecutar validación agrupada Slice A y registrar TDD.

- [x] 3. Slice B — Eventos y proyección adaptativa
  - [x] 3.1 Escribir RED de append, duplicate-noop, conflict, expected version y causalidad.
  - [x] 3.2 Extender Run Ledger con metadata adaptativa allow-listed y versionada.
  - [x] 3.3 Implementar proyección reconstruible de assignments activas, budget y waiting pool bounded.
  - [x] 3.4 Ordenar pool por prioridad/edad lógica/ID y reportar starvation sin autoelevar prioridad.
  - [x] 3.5 Probar compatibilidad con eventos históricos y ausencia de backfill inventado.

- [x] 4. Slice B — Lease y protocolo yield
  - [x] 4.1 Escribir RED de owner mismatch, lease inválida/vencida/stale y writers concurrentes.
  - [x] 4.2 Implementar assign advisory con lease digest, TTL, fencing e idempotencia.
  - [x] 4.3 Implementar checkpoint yield con artifacts/evidence/hypothesis/failed-attempt refs y successor capabilities.
  - [x] 4.4 Liberar lease/budget y reencolar solo tras receipt durable; preservar estado ante conflict/unavailable.
  - [x] 4.5 Ejecutar race, crash/retry, repair, golden y privacy tests del Slice B.

- [x] 5. Slice C — Application service y modes
  - [x] 5.1 Implementar `recommend` con disabled/shadow/advisory y `gate_advanced=false`.
  - [x] 5.2 Implementar `assign`, `yield` y `status` sobre ports inyectados.
  - [x] 5.3 Rechazar protected boundaries antes de scoring y devolver next owner humano/orchestrator.
  - [x] 5.4 Permitir cleanup explícito de leases existentes al deshabilitar el feature sin crear nuevas assignments.
  - [x] 5.5 Probar snapshots stale, no candidate, capacity exhaustion, starvation y ledger unavailable.

- [x] 6. Slice C — CLI y experiencia operativa
  - [x] 6.1 Implementar `lufy-ai adaptive recommend` con stdin/file, target y human/JSON.
  - [x] 6.2 Implementar `adaptive assign`, `adaptive yield` y `adaptive status` con exit codes estables.
  - [x] 6.3 Mantener read-only por default y exigir recording explícito para efectos durables cuando corresponda.
  - [x] 6.4 Actualizar help/palette y probar usage, diagnostics sanitizados y output scriptable.
  - [x] 6.5 Ejecutar E2E shadow y advisory recommend -> assign -> yield -> requeue.

- [x] 7. Slice D — Harness, seguridad y documentación
  - [x] 7.1 Integrar señales en orchestrator/router sin crear roles nuevos ni permitir avance automático de gates.
  - [x] 7.2 Documentar roles como capacidades temporales, yield seguro y fronteras protegidas.
  - [x] 7.3 Actualizar `AGENTS.md.template`, contratos, docs de arquitectura/CLI y assets managed/embedded.
  - [x] 7.4 Mantener paridad OpenCode/Codex y fallback cuando adaptive routing no esté soportado.
  - [x] 7.5 Ejecutar coupling/catalog/parity y auditoría recursiva de privacidad.

- [ ] 8. Verify, sync y delivery readiness
  - [x] 8.1 Ejecutar suite completa, build, race focalizado y `scripts/validate.sh`.
  - [x] 8.2 Verificar proposal/design/specs, state machine, rollback, cross-platform y no gate advancement.
  - [x] 8.3 Registrar evidencia en `verification/experiment-adaptive-role-allocation/` y strict validate.
  - [x] 8.4 Sincronizar deltas validados a specs activas.
  - [x] 8.5 Preparar trazabilidad con issue `#222`; no iniciar fase 6 antes de merge/cierre.
  - [x] 8.6 Revalidar recommendation vigente, capacidad global y budget del actor dentro del CAS de confirmación, con regresiones de stale state.
  - [x] 8.7 Añadir recuperación durable y fenced de leases vencidas sin liberación implícita por reloj.
  - [ ] 8.8 No archivar ni cerrar sin delivery, checks remotos, sync y gates completos.
