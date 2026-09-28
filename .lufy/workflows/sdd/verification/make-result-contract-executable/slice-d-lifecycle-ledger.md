# Verificación Slice D — Lifecycle, Run Ledger y contratos

## Alcance

Este sublímite reemplaza la detección textual del lifecycle, agrega correlación content-free con Run Ledger, conecta `result transition --record` a durabilidad causal y mantiene contratos/assets root y embedded en paridad.

## Evidencia TDD

- **RED**: faltaban extractor bounded, bridge, wiring CLI/ledger y assets ejecutables.
- **GREEN**: lifecycle decodifica un documento puro o un único fence YAML/JSON; el bridge registra metadata allow-listed e idempotente; CLI falla cerrada si no obtiene durabilidad.
- **TRIANGULATE**: se cubrieron recorded, duplicate_noop, conflict, unavailable, ambigüedad, oversize y canaries de privacidad.
- **REFACTOR**: el bridge vive en un adapter dedicado; Result Contract y Run Ledger permanecen desacoplados en autoridad.

## Comandos ejecutados

| Comando | Resultado | Nota |
| --- | --- | --- |
| `go test -timeout 180s ./internal/codexlifecycle ./internal/resultcontractledger ./internal/cli ./internal/assets ./internal/harnesscatalog -count=1` | passed | Validación independiente focalizada. |
| `go test -race -timeout 240s ./internal/codexlifecycle ./internal/resultcontractledger ./internal/cli ./internal/assets ./internal/harnesscatalog -count=1` | passed | Sin data races. |
| `go vet ./internal/codexlifecycle ./internal/resultcontractledger ./internal/cli ./internal/assets ./internal/harnesscatalog` | passed | Sin diagnósticos. |
| `GOCACHE=/private/tmp/lufy-sliced2-windows-cache GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -exec=true ./internal/codexlifecycle ./internal/resultcontractledger ./internal/cli ./internal/assets ./internal/harnesscatalog` | passed | Cross-compilación Windows de paquetes afectados. |

## Criterios observados

- Un substring de schema, un fence incompleto, múltiples candidatos o un payload mayor a 64 KiB no se aceptan como handoff.
- `SubagentStop` mantiene `continue=true`, emite warning sanitizado y no modifica tasks ni gates.
- El bridge persiste schema, status, fingerprints, decisión, versión y referencias digeridas; nunca persiste summary, prompt, path u output.
- Reintentos equivalentes producen una única correlación durable; metadata distinta con la misma key produce conflicto sin overwrite.
- Ledger unavailable devuelve `unavailable` sin identidad/version/fingerprint de una aceptación.
- Los contratos y templates gestionados mantienen paridad byte a byte entre root y embedded.

## Riesgo residual

Receipt y Run Ledger son dos escrituras locales: una interrupción entre ambas se recupera mediante retry `duplicate_noop`, que completa la correlación faltante. La validación final del change debe comprobar la suite completa y la recuperación.
