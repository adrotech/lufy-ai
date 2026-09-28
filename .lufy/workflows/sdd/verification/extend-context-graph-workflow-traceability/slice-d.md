# Slice D — Run Ledger, métricas y harness

## Estado

- Proyección content-free, métricas, CLI, roles, contratos, specs y assets: implementados y validados.
- Gate del change: `implemented`; resta validación final, sync y delivery readiness.

## Contratos

- Solo el adapter explícito `LoadWorkflowProjection` consulta el Run Ledger mediante su API validada; `.lufy/runtime/**` sigue fuera del discovery genérico.
- La proyección bounded conserva IDs, secuencia local, causalidad, task/evidence refs, event name y timestamps. No expone cuerpos, prompts, outputs, summaries, secretos ni paths runtime.
- El digest de la proyección participa en freshness: cambios del ledger dejan stale al graph persistido hasta rebuild.
- `context metrics` reconoce `review.started`, `review.completed`, `review.rework` y `defect.reopened`; pares completos producen duración y datos incompletos retornan `partial/unavailable` sin síntesis.
- Router, orchestrator, reviewer, validator y delivery consumen el assessment como evidencia secundaria. `proceed` no autoriza ni sustituye gates.

## TDD

### RED

- no existían el adapter/projection content-free, los nodos causales, `MetricsResult` ni `context metrics`;
- CLI y application fallaron inicialmente por símbolos ausentes.

### GREEN / TRIANGULATE / REFACTOR

- runs parent/child, event causality, task/evidence refs y digest determinista;
- metrics available, partial y unavailable con timestamps ausentes;
- proyección en graph y freshness tras un evento nuevo;
- canaries de contenido privado y runtime path;
- human/JSON/help/usage del comando;
- paridad exacta de assets root/embedded y spec activa sincronizada.

## Evidencia

| Comando | Resultado |
| --- | --- |
| `go test ./internal/contextgraph/... ./internal/projectconfig ./internal/runledger ./internal/cli ./internal/assets ./internal/harnesscatalog -count=1` | passed |
| `git diff --check` | pendiente de gate final agrupado |

## Riesgos residuales

- La proyección limita el detalle a 4096 eventos; al truncar reporta `partial` y recovery.
- Eventos históricos sin nombres/timestamps reconocidos permanecen `partial/unavailable`; no se realiza backfill inferido.
