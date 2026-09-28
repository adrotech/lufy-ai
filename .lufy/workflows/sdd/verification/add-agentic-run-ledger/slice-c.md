# Verification: add-agentic-run-ledger — Slice C

Fecha: 2026-09-20

## Alcance verificado

- `codexlifecycle` acepta únicamente metadata estable modelada (`session_id`, `turn_id`, `agent_id`, `agent_type`) e ignora `transcript_path` y campos desconocidos.
- `SessionStart`, `SubagentStart`, `SubagentStop`, `Stop` y `SessionEnd` producen eventos causales con reintentos idempotentes.
- Bindings `lufy-run-binding/v1` correlacionan referencias externas pseudonimizadas con runs locales sin persistir IDs externos, mensajes ni transcripts.
- Fallos de configuración/storage emiten warning sanitizado, conservan `continue=true` y no avanzan gates.
- `.codex/hooks.json` y su asset embebido registran `SubagentStart` y mantienen JSON válido.
- Result Contract v1 admite un bloque `ledger` completamente opcional; payloads legacy sin el bloque siguen siendo válidos.
- `run-ledger-producer.md` define la frontera portable para otros adapters sin exigir instrumentación automática.
- Catálogo root/embedded, hashes calculados y contratos gestionados permanecen en paridad.

## Evidencia TDD y validación

| Etapa | Comando | Resultado |
| --- | --- | --- |
| RED lifecycle | `go test -timeout 30s ./internal/codexlifecycle ./internal/assets` | falló: no se creaban runs/bindings, no había warning best-effort y faltaba `ledger`/`SubagentStart` en assets |
| GREEN focalizado | `go test -timeout 45s ./internal/codexlifecycle ./internal/assets ./internal/harnesscatalog ./internal/cli` | passed |
| Race integrado | `go test -race -timeout 60s ./internal/codexlifecycle ./internal/assets ./internal/harnesscatalog ./internal/cli ./internal/runledger` | passed |
| Cobertura focalizada | `go test -race -timeout 60s -cover ./internal/codexlifecycle ./internal/assets ./internal/harnesscatalog ./internal/cli ./internal/runledger` | passed; codexlifecycle 76.5%, assets 81.6%, harnesscatalog 91.3%, cli 77.0%, runledger 79.1% |
| Static | `go vet ./internal/codexlifecycle ./internal/assets ./internal/harnesscatalog ./internal/cli ./internal/runledger` | passed |
| Hooks JSON | `python3 -m json.tool` sobre hooks root y embedded | passed |
| Privacidad fixture | inspección recursiva de `.lufy/runtime` sintético | passed; ausentes session/turn/agent IDs, transcript, unknown payload y assistant message canaries |
| Repo completo | `scripts/validate.sh` | passed; catálogo/harness coupling OK, cobertura global 80.2%, build passed |

## Observaciones

- La pseudonimización incorpora provenance local del repositorio y namespace por tipo; solo se persisten hashes SHA-256 y IDs locales derivados.
- `Stop` registra un checkpoint observable sin declarar un estado de gate; los estados terminales del run provienen de `SubagentStop` y `SessionEnd`, no de inferencias SDD.
- Si un stop llega sin start durable, el adapter reconstruye primero el binding/start mínimo y luego registra el terminal, conservando causalidad.
- La integración automática sigue limitada a Codex. Otros adapters pueden usar el contrato tipado y la CLI portable.
