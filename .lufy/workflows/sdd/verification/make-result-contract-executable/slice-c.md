# Verificación Slice C — Ownership, leases, joins e idempotencia

## Alcance

El Slice C agrega controles de ownership y lease, fencing por versión, receipts atómicos content-free e idempotentes y validación explícita de joins. Los adapters persistentes y el wiring CLI/lifecycle pertenecen al Slice D.

## Evidencia TDD

- **RED**: los tests nuevos fallaron únicamente por símbolos Slice C ausentes: ownership/lease, receipts, outcomes, joins y puertos asociados.
- **GREEN**: se implementaron los guards y puertos mínimos sin modificar los tests.
- **TRIANGULATE**: se cubrieron expiración, mismatches, retries concurrentes, conflictos sin overwrite y todas las fallas de join declaradas.
- **REFACTOR**: `EvaluateTransition` conserva dependencias externas inyectadas mediante clock, `ReceiptStore` y `JoinResolver`; el receipt mantiene una allow-list cerrada de seis campos.

## Comandos ejecutados

| Comando | Resultado | Nota |
| --- | --- | --- |
| `go test ./internal/resultcontract` antes de producción | failed (esperado) | RED por símbolos Slice C ausentes. |
| `go test ./internal/resultcontract -count=10` | passed | GREEN estable, incluido retry concurrente. |
| `go test -timeout 120s ./internal/resultcontract -count=1` | passed | Validación independiente del boundary. |
| `go test -race -timeout 120s ./internal/resultcontract -count=1` | passed | Sin data races. |
| `go vet ./internal/resultcontract` | passed | Sin diagnósticos. |
| `GOCACHE=/private/tmp/lufy-slicec-windows-cache GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c -o /private/tmp/resultcontract-slicec-independent.test.exe ./internal/resultcontract` | passed | Cross-compilación aislada para Windows. |
| `git diff --check origin/develop` | passed | Sin errores de whitespace en el rango de integración. |

La primera ejecución Windows en paralelo con tests nativos falló por contaminación del cache compartido del toolchain. La repetición secuencial con `GOCACHE` aislado pasó; no se observó un defecto de código.

## Criterios observados

- Owner y rol deben coincidir con ownership vigente; referencias y tokens se aceptan solo como digests SHA-256.
- El lease debe coincidir y permanecer vigente; `expected_version` actúa como fencing antes de evaluar el lease.
- `ReceiptStore.Record` es atómico: 16 workers sobre la misma key producen una aceptación y 15 `duplicate_noop`.
- Reutilizar la key con otro intent produce `conflict` sin sobrescribir el receipt original.
- El receipt no guarda envelope, prompt, output, paths ni idempotency key cruda.
- Un join exige el conjunto exacto declarado, children terminales `closed`, causalidad al padre, fingerprints coincidentes y evidencia agrupada mediante digest.

## Riesgo residual y siguiente slice

Los puertos aún no están conectados a almacenamiento durable, CLI, lifecycle ni Run Ledger. El Slice D debe cablearlos preservando fail-open solo en hooks observacionales y fail-closed en gates de transición.
