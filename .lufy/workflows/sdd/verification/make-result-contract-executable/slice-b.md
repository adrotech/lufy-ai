# Verificación Slice B — Roles, evidencia y transiciones

## Alcance

El Slice B hace ejecutables las decisiones de rol, evidencia y transición sin incorporar todavía ownership, leases, joins ni idempotencia durable, que pertenecen al Slice C.

## Evidencia TDD

- **RED**: `go test ./internal/resultcontract` falló inicialmente por la ausencia intencional de `RolePolicy`, `EvidencePolicy`, `StateVector`, `result-transition/v1` y `EvaluateTransition`.
- **GREEN**: se implementaron políticas inyectables de rol/evidencia, vector derivado, tabla explícita de 81 pares y evaluación pura de transiciones con CAS por versión y fingerprint.
- **TRIANGULATE**: la matriz cubre 8 roles por 9 estados, rol desconocido fail-closed, evidencia material, los 81 pares de transición, recuperación y estado terminal.
- **REFACTOR**: se separaron claims, estado y transición en unidades adapter-neutral; `gofmt` y `go vet` quedaron limpios.

## Comandos ejecutados

| Comando | Resultado | Nota |
| --- | --- | --- |
| `go test ./internal/resultcontract -count=10` | passed | GREEN estable. |
| `go test -timeout 120s ./internal/resultcontract -count=1` | passed | Validación independiente del boundary. |
| `go test -race -timeout 120s ./internal/resultcontract -count=1` | passed | Sin carreras detectadas. |
| `go vet ./internal/resultcontract` | passed | Sin diagnósticos. |
| `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c ./internal/resultcontract` | passed | Compatibilidad de compilación Windows. |
| `git diff --check origin/develop` | passed | Sin errores de whitespace en el rango de integración. |

## Criterios observados

- Los roles desconocidos y los estados no permitidos se rechazan de forma cerrada.
- `validated`, `delivered` y `closed` requieren políticas de evidencia explícitas; texto libre, nombres de schema y placeholders no satisfacen evidencia material.
- `blocked` y `escalated` preservan dimensiones previas de delivery y sync.
- `closed` es terminal.
- Los conflictos de versión o fingerprint producen una decisión `conflict` recuperable sin mutar estado.
- Una transición aceptada incrementa la versión y deriva un fingerprint canónico determinista.

## Riesgo residual y siguiente slice

La evaluación aún no autentica ownership ni leases y no registra receipts durables. Esas garantías quedan deliberadamente bloqueadas hasta completar el Slice C.
