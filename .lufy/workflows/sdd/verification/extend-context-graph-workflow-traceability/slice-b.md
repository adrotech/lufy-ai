# Slice B — coverage, trace y freshness

## Estado

- Coverage/trace: implementado y validado.
- Fallback directo de diff: pendiente de Slice C junto con `context review`.
- Gate del change: `implemented`; no está listo para sync ni delivery.

## Contratos

- `context coverage` clasifica cada scenario como `covered` o `gap` y exige task `implements` más test `verifies` explícitos.
- `context trace <node>` recorre solamente edges forward `defines` y workflow allow-listed hasta un scenario.
- Un grafo `stale` o `not_available` devuelve `status: unknown`, `graph_status` exacto y recovery; nunca reconstruye ni aprueba implícitamente.
- Coverage conserva el total y limita el detalle a 128 scenarios.

## TDD

### RED

- application no tenía `Coverage`, `Trace` ni sus contratos.
- CLI rechazaba `coverage` y `trace` como subcomandos desconocidos.
- funciones Go normales ignoraban markers `lufy:implements`.

### GREEN / TRIANGULATE / REFACTOR

- fixtures covered, missing task y missing test;
- 140 scenarios verifican límite/truncation;
- missing/stale verifican degradación `unknown`;
- file → function → task → scenario verifica dirección y provenance;
- archivo sin camino retorna `untraced_file`;
- human/JSON/help/usage y ExitOK estructurado;
- preflight carga un snapshot y valida manifest/hash antes de usar evidencia.

## Evidencia

| Comando | Resultado |
| --- | --- |
| `go test ./internal/contextgraph/application -run 'TestCoverage|TestTrace' -count=1` | passed |
| `go test ./internal/cli -run 'TestRunContextCoverageAndTrace' -count=1` | passed |
| `go test ./internal/contextgraph/... -count=1` | passed |
| `go test ./internal/cli -count=1` | passed |
| `git diff --check` | passed |

## Riesgo residual

Existe un TOCTOU local mínimo si una fuente cambia inmediatamente después del hash de freshness. Cuando el preflight observa mismatch, no usa el grafo como evidencia.
