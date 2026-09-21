# Verificación Slice D — CLI y normalización

## Alcance

Este sublímite implementa `result validate`, `result normalize` y `result transition`, junto con sus formatos versionados, exit codes, ayuda y command palette. Lifecycle, bridge Run Ledger y paridad de assets permanecen en el siguiente sublímite del Slice D.

## Evidencia TDD

- **RED**: los tests CLI fallaron por la ausencia de constantes `ExitResult*`, dispatcher, comandos y entradas de palette.
- **GREEN**: se implementaron inputs explícitos, policies de rol/evidencia, normalización legacy estrecha, request de transición tipado y store durable de receipts.
- **TRIANGULATE**: se cubrieron stdin/file, human/JSON, roles, payloads ambiguos, read-only, recording, idempotencia y categorías de error.
- **REFACTOR**: el dominio conserva adapters separados; filesystem y ayuda pública quedan en CLI.

## Comandos ejecutados

| Comando | Resultado | Nota |
| --- | --- | --- |
| `go test ./internal/cli -run '^TestResult' -count=10` | passed | Flujos CLI estables. |
| `go test -timeout 120s ./internal/resultcontract ./internal/cli ./internal/tui/commandpalette -count=1` | passed | Validación independiente de los tres paquetes. |
| `go test -race -timeout 180s ./internal/resultcontract ./internal/cli ./internal/tui/commandpalette -count=1` | passed | Sin data races. |
| `go vet ./internal/resultcontract ./internal/cli ./internal/tui/commandpalette` | passed | Sin diagnósticos. |
| `GOCACHE=/private/tmp/lufy-sliced-windows-cache GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c -o /private/tmp/lufy-cli-sliced.test.exe ./internal/cli` | passed | Cross-compilación Windows de la superficie CLI. |

## Criterios observados

- `--stdin` y `--file` son explícitos y mutuamente excluyentes.
- JSON devuelve un único documento versionado y content-free.
- Normalización acepta solo `lufy-result-legacy/v1` allow-listed y bounded; no interpreta texto libre ni claims avanzados.
- La transición es read-only por defecto; `--record` usa lock y escritura atómica y guarda solo `TransitionReceipt`.
- Exit codes: usage `2`, invalid `3`, rejected `4`, conflict `5`, unavailable `6`.

## Riesgo residual

El receipt store de CLI es local y mínimo. La correlación con Run Ledger, su policy de disponibilidad y la validación real del lifecycle se completan en tasks 6.1–6.6.
