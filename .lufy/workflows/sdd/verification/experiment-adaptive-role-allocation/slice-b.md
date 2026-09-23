# Verificación Slice B — ledger, leases y yield adaptativo

## Estado

`validated` para el alcance del Slice B. El cambio global continúa activo: application service, CLI, harness, sync, delivery y cierre permanecen pendientes.

## Resultado funcional

- Run Ledger admite eventos `demand`, `recommendation`, `assignment` y `yield` con envelope `lufy-run-adaptive/v1`, campos por kind estrictamente allow-listed y referencias content-free.
- `ExpectedLocalSequence` implementa CAS bajo el lock durable del run; retries equivalentes producen `duplicate_noop` antes del fencing y payloads distintos con la misma key producen conflicto.
- La cadena durable `demand -> recommendation -> assignment -> yield` conserva causalidad explícita; el evento intermedio se completó al integrar application service en Slice C.
- La proyección `lufy-adaptive-status/v1` reconstruye assignments activas, presupuesto y waiting pool bounded; ordena por prioridad, edad lógica e ID, marca starvation sin mutar prioridad y tolera eventos históricos sin metadata adaptativa.
- Assignment IDs no pueden reutilizarse. Owner, lease digest/expiry y expected version se validan antes de yield.
- Yield libera lease/presupuesto y reencola/completa únicamente después de un receipt durable; conflicto, writer stale o storage unavailable preservan la assignment activa.

## Evidencia TDD

- RED: `go test ./internal/adaptive/... ./internal/runledger -count=1` falló inicialmente por ausencia de tipos adaptativos, CAS y adapter de ledger.
- GREEN: contratos, persistencia, proyección, fencing y protocolo yield implementados con tests focalizados.
- TRIANGULATE: duplicate/noop vs conflict, assign/yield concurrentes, owner/lease mismatch, lease vencida, writer stale, crash entre evento/receipt, replay/golden, storage unavailable y compatibilidad histórica.
- REFACTOR: estado mutable separado descartado; `Status` siempre deriva desde eventos append-only y el adapter depende del port mínimo `runledger.Store`.

## Comandos

| Comando | Resultado | Evidencia |
| --- | --- | --- |
| `go test ./internal/adaptive/... ./internal/runledger -count=1` | passed | dominio, ledger, adapter y regresión focalizada |
| `go test -race ./internal/adaptive/... ./internal/runledger -count=1` | passed | writers concurrentes sin race |
| `go test ./... -count=1` | passed fuera del sandbox | suite completa; los packages `upgrade` y `versioncheck` requieren loopback para `httptest` |
| `go build ./...` | passed | compilación completa del módulo |

## Fixtures y controles

- Golden: `internal/adaptive/adapters/testdata/adaptive-status.golden.json`.
- Crash/retry: evento adaptativo publicado antes del receipt se recupera como `duplicate_noop` y luego verifica healthy.
- Repair/rebuild: replay in-memory produce el mismo digest y JSON; no reescribe los eventos fuente.
- Privacy: scan recursivo de eventos rechaza prompt, response, summary, secretos e hipótesis/intentos en texto; artifacts, evidence, actor y lease se guardan como tokens/digests permitidos.

## Riesgos pendientes

- El adapter durable aún no está expuesto por application service o CLI; esa integración corresponde al Slice C.
- La policy efectiva todavía debe consumir `parallel_execution` y `workflow_limits.review` como señales sin aliases; task 2.4 permanece pendiente hasta ese join.
- Los comandos adaptativos no deben confundirse con autoridad sobre Result Contract ni avanzar gates; se verificará en Slice C/D.
