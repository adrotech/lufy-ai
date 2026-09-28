# Verification: add-agentic-run-ledger — Slice D

Fecha: 2026-09-20

## Alcance verificado

- Fixtures sintéticos de dominio, CLI y Codex rechazan o excluyen prompts, mensajes, transcripts, argumentos, diffs, contenido, secretos, IDs externos y claves idempotentes en claro.
- El store falla sin publicar ante runtime no disponible o read-only; lock timeout, lease recovery, crash entre evento/receipt, parciales y corrupción mantienen diagnósticos sanitizados.
- Payload sobredimensionado y campos desconocidos se rechazan antes de persistir.
- Dos writers hijos concurrentes preservan parent/cause, Lamport, terminalidad del árbol y disponibilidad parcial/ausente de métricas.
- La documentación operativa cubre schema, CLI, privacidad, idempotencia, retención, recovery, límites, OpenTelemetry y Agent Observatory.
- La matriz CI existente ejecuta tests y build Go en Ubuntu, macOS y Windows; la ejecución local cubre macOS y deja los otros OS para checks remotos de delivery.

## Evidencia agrupada

| Gate | Comando o evidencia | Resultado |
| --- | --- | --- |
| Unit/integration completo | `go test -timeout 120s ./...` | passed |
| Build | `go build ./cmd/lufy-ai` | passed |
| Race focalizado | `go test -race -timeout 90s ./internal/runledger ./internal/codexlifecycle ./internal/projectconfig ./internal/cli ./internal/assets` | passed |
| Degradación | tests de unavailable path, read-only runtime, active/stale lock, interrupted atomic file, receipt recovery y corrupción | passed; read-only se salta solo si el filesystem/proceso no aplica permisos |
| Privacidad | escaneo recursivo de runtime sintético y allow-list estricta | passed; canaries ausentes |
| Whitespace/static | `git diff --check origin/develop`, `go vet` focalizado | passed |
| Validación repo | `scripts/validate.sh` | passed; cobertura global 80.2%, build y catalog/harness coupling OK |
| Cross-platform declarada | `.github/workflows/go-cli-install.yml` matrix `ubuntu-latest`, `macos-latest`, `windows-latest` | disponible; ejecución remota pendiente de delivery |

## Artefactos de cierre del slice

- `docs/run-ledger.md`
- `.lufy/contracts/run-ledger-producer.md`
- `tools/lufy-cli-go/internal/runledger/`
- `tools/lufy-cli-go/internal/codexlifecycle/`
- `.lufy/workflows/sdd/changes/add-agentic-run-ledger/change-overview.html`

## Riesgos residuales

- Los checks Windows/Linux reales solo pueden evidenciarse al publicar el PR; localmente se validó macOS y se preservaron primitives portables sin `flock`.
- El ledger sigue deliberadamente local: no hay export OpenTelemetry, backend remoto ni consumo durable desde Observatory en esta fase.
