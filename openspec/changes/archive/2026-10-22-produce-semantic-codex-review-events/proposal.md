# Bloque 4: productor explícito de eventos de review

## Why
Las métricas ya reconocen review.started/completed/rework y defect.reopened, pero lifecycle no conoce la semántica de una review. Inferirla de SubagentStart o de un role hint inventaría evidencia.

## What Changes
- Añadir `lufy-ai review start|complete|rework|reopen` como operaciones explícitas del host sobre Run Ledger existente.
- Usar IDs opacos hasheados, timestamps del productor, idempotencia y CAS del ledger, sin persistir cuerpos, findings ni valores de entorno.
- Exigir Result Contract estricto y claim permitido para reviewer antes de registrar complete/rework/reopen; el host declara el outcome explícito y no se avanza ningún gate.
- Documentar invocación por host/user con permisos; no ampliar permisos de reviewer/orchestrator ni convertir hooks genéricos en reviews.

## Capabilities
### New Capabilities
- codex-semantic-review-events: operaciones explícitas, content-free e idempotentes.

## Aceptación y workload
WHEN se inicia/completa una review explícita THEN métricas deriva un par con TaskRef único sin duplicar retries.
WHEN falta inicio, evidencia válida, hay outcome contradictorio o conflicto THEN no se registra una conclusión falsa.
WHEN ocurre sólo lifecycle genérico THEN no aparecen eventos semánticos.
WHEN se registra review metadata THEN checkpoints/gates/estado terminal permanecen independientes.

Slice T1 por nueva CLI y contrato operacional que alimenta métricas. Una unidad de servicio, CLI, tests, docs y spec/mirror, con validación agrupada; dividir productor y consumo observable dejaría contrato incompleto. `workflow_limits` no configurado; graph del worktree not_available, fallback a archivos dirigidos. Overview HTML not_available. No branching de artifacts: la semántica queda explícita en esta propuesta; no se introduce inferencia alternativa.
