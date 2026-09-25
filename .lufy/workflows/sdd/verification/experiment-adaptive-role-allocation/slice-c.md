# Verificación Slice C — application service y CLI adaptativa

## Estado

`validated` para el alcance del Slice C. Harness/contratos/documentación, sync, delivery y cierre continúan pendientes.

## Resultado funcional

- `adaptive/application` expone `recommend`, `assign`, `yield` y `status` mediante ports inyectados; no depende del filesystem ni de la implementación concreta de Run Ledger.
- `disabled` devuelve decisión explícita sin tocar ledger; `shadow` recomienda de forma no durable; `advisory` registra solamente cuando el caller solicita `--record` y aporta idempotency key/run.
- Toda decisión conserva `gate_advanced=false`; protected boundaries retornan `escalate` antes de scoring.
- `parallel_execution.max_parallel_agents` (solo si está enabled) y `workflow_limits.review.max_concurrent_slices` se consumen por sus paths canónicos como señales de capacidad. No se crean aliases ni autoridad nueva.
- Assignment exige recommendation vigente, actor/role/policy/fingerprint/score coincidentes, expected version y lease dentro de `adaptive_routing.lease_ttl_seconds`.
- Deshabilitar el feature bloquea nuevas assignments, pero un yield fenced explícito puede limpiar una lease preexistente.
- CLI disponible: `adaptive recommend|assign|yield|status`, stdin/file, human/JSON y categorías estables `0/2/4/5/6`.

## Evidencia TDD

- RED/GREEN application: modes, protected boundary, capacity exhaustion, no candidate, ledger unavailable, recording intent y cleanup disabled.
- RED/GREEN CLI: disabled-by-default, shadow sin eventos, advisory `recommend -> assign -> yield -> requeue`, status reconstruido, file input, output JSON y rechazo sin `--record`.
- Recommendation durable se incorpora como evento causal intermedio; assignment solo confirma la recommendation corriente.

## Comandos

| Comando | Resultado | Evidencia |
| --- | --- | --- |
| `go test ./internal/adaptive/... ./internal/cli ./internal/runledger ./internal/tui/commandpalette -count=1` | passed | service, commands, palette y E2E |
| `go test -race ./internal/adaptive/... ./internal/runledger -count=1` | passed | application/adapter/ledger sin race |
| `go test ./... -count=1` | passed fuera del sandbox | regresión completa con loopback para `httptest` |
| `go build ./...` | passed | build completo del módulo |

## Riesgos pendientes

- El harness todavía no consume las salidas adaptativas; Slice D debe documentarlas como evidencia no autoritativa y mantener escalamiento humano para fronteras protegidas.
- Contratos, assets managed/embedded y documentación todavía deben sincronizar la misma semántica disabled/shadow/advisory.
- No se ha ejecutado sync/archive/delivery del change; el estado global no es `closed`.
