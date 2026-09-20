# Verification: add-agentic-run-ledger — Slice A

Fecha: 2026-09-19

## Alcance verificado

- Contrato tipado `lufy-run-event/v1`, canonicalización y fingerprint estable.
- Store filesystem append-only con un archivo atómico por evento.
- Idempotencia `recorded`, `duplicate_noop` y `conflict` sin overwrite.
- Reloj de Lamport, secuencia local, parent/cause y 24 writers concurrentes.
- Lock por run con lease, timeout y recuperación conservadora de owner vencido.
- Recovery cuando el evento queda publicado antes del receipt.
- Verify de eventos inválidos/parciales, receipts faltantes/huérfanos/inconsistentes y causalidad rota.
- Allow-list estricta y ausencia en disco de refs externas, paths y claves idempotentes en claro.

## Evidencia TDD

| Etapa | Comando | Resultado |
| --- | --- | --- |
| RED | `go test ./internal/runledger` | falló por API/contrato aún inexistente (`undefined`) |
| GREEN | `go test ./internal/runledger` | passed |
| Race + cobertura | `go test -race -coverprofile=/tmp/runledger.cover ./internal/runledger` | passed; 82.0% statements |
| Static | `go vet ./internal/runledger` | passed |
| Repo completo | `scripts/validate.sh` | passed; cobertura global 80.4%; build passed |

## Observaciones

- El primer `scripts/validate.sh` alcanzó 79.8% global porque el paquete nuevo tenía 66.5%; se agregaron casos contractuales y de corrupción reales hasta 82.0% del paquete y 80.4% global.
- `shellcheck` no está disponible localmente; el script lo reportó como notice y no existen cambios shell en este slice.
- El PR guard informó cero archivos porque los artefactos y el paquete siguen untracked hasta una futura autorización de delivery; la inspección de whitespace se complementa con revisión específica del árbol nuevo.
- Las proyecciones e índices de consulta se implementarán en Slice B; Slice A deja eventos y receipts como fuente verificable y reconstruible.
