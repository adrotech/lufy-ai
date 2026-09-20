# Tasks: add-agentic-run-ledger

## Review Workload Harness

| Slice | Objetivo | Criterio de salida | Validación agrupada | Guía de PR |
| --- | --- | --- | --- | --- |
| A | Core causal, storage e idempotencia | append/retry/concurrency/recovery verificados | unit + race/concurrency + corruption fixtures | separar del wiring de adapters |
| B | CLI, config, proyecciones y retención | comandos y JSON contract estables | CLI integration + golden tests | incluir ejemplos de operación |
| C | Codex, Result Contract y assets | root/subagent trazables sin contenido privado | lifecycle + catalog/contract tests | reviewer verifica límites adapter-neutral |
| D | Privacidad, degradación, docs y E2E | no data leakage y workflow no bloqueado | privacy scan + full validation + OS matrix | cierre transversal y evidencia |

- [x] 1. Slice A — Contract-first y pruebas RED
  - [x] 1.1 Definir tipos, enums, límites y errores sanitizados de `lufy-run-event/v1` sin mapas arbitrarios.
  - [x] 1.2 Escribir pruebas RED para canonicalización, fingerprint, IDs, causalidad, reloj de Lamport y secuencia local.
  - [x] 1.3 Escribir pruebas RED para idempotencia: record, duplicate-noop y conflict sin overwrite.
  - [x] 1.4 Escribir pruebas RED para concurrencia, publicación atómica, crash windows y lock recovery cross-platform.
  - [x] 1.5 Escribir pruebas negativas de privacidad para prompts, mensajes, transcript, argumentos, diffs, secretos y contenido de archivos.

- [x] 2. Slice A — Dominio y filesystem store
  - [x] 2.1 Implementar el dominio adapter-neutral, validación estricta y canonicalización determinista.
  - [x] 2.2 Implementar `RunStore` y el layout bajo `.lufy/runtime/runs/` usando paths seguros y escrituras atómicas.
  - [x] 2.3 Implementar lock por run, asignación de Lamport clock y secuencia dentro de la sección crítica.
  - [x] 2.4 Implementar receipts y fingerprint con semántica `recorded`, `duplicate_noop` y `conflict`.
  - [x] 2.5 Implementar recovery/verify de eventos parciales, receipts huérfanos e índices/proyecciones reconstruibles.
  - [x] 2.6 Ejecutar y registrar validación agrupada del Slice A antes de integrar consumidores.

- [x] 3. Slice B — Proyecciones y CLI
  - [x] 3.1 Implementar projector determinista para árbol causal, status, blockers, evidence, artifacts y métricas disponibles/ausentes.
  - [x] 3.2 Implementar `lufy-ai run record` con input tipado, clave idempotente explícita y salida JSON.
  - [x] 3.3 Implementar `lufy-ai run checkpoint` para estados/gates/next owner permitidos.
  - [x] 3.4 Implementar `lufy-ai run status` y `run summary` con formatos humano/JSON y selección por root/subagent.
  - [x] 3.5 Implementar `lufy-ai run verify` read-only y `--repair` explícito para derivados, nunca para reescribir eventos.
  - [x] 3.6 Implementar `lufy-ai run prune --dry-run` y ejecución explícita, preservando runs activos.
  - [x] 3.7 Añadir configuración compatible para habilitación y retención con defaults documentados y preservación de campos user-owned.
  - [x] 3.8 Agregar `.lufy/runtime/` a `.gitignore` sin convertirlo en asset administrado.
  - [x] 3.9 Ejecutar y registrar validación agrupada del Slice B, incluyendo golden JSON y límites de retención.

- [x] 4. Slice C — Codex y contratos portables
  - [x] 4.1 Extender el input de `codexlifecycle` solo con metadata estable de hooks, ignorando explícitamente `transcript_path` y campos desconocidos.
  - [x] 4.2 Registrar `SubagentStart` en `.codex/hooks.json` y catálogo/renderizado correspondiente.
  - [x] 4.3 Implementar bindings pseudonimizados entre session/turn/agent refs y runs Lufy.
  - [x] 4.4 Mapear SessionStart/SubagentStart/SubagentStop/Stop/SessionEnd a eventos causales idempotentes.
  - [x] 4.5 Mantener hooks best-effort: warning compacto ante fallo, sin mutar gates ni cambiar el exit principal.
  - [x] 4.6 Extender Result Contract v1 de forma compatible con referencias opcionales `run_id`, `event_id` y ledger status.
  - [x] 4.7 Documentar la superficie portable para otros adapters sin exigir integración automática en esta fase.
  - [x] 4.8 Actualizar assets, hashes/manifest y tests de integridad donde corresponda.
  - [x] 4.9 Ejecutar y registrar validación agrupada del Slice C con fixtures lifecycle y tests de contratos.

- [x] 5. Slice D — Hardening, documentación y E2E
  - [x] 5.1 Probar que el storage no contiene valores prohibidos con fixtures sintéticos y escaneo de salida.
  - [x] 5.2 Probar degradación por permisos, disco/rename, lock timeout, corrupción e input sobredimensionado.
  - [x] 5.3 Probar dos writers y jerarquías root/subagent con causalidad, retry y métricas parciales.
  - [x] 5.4 Documentar schema, CLI, privacidad, retención, recovery, límites y relación con OpenTelemetry/Observatory.
  - [x] 5.5 Ejecutar `go test ./...`, build, race donde aplique, `scripts/validate.sh` y matriz cross-platform disponible.
  - [x] 5.6 Registrar evidencia final bajo `verification/add-agentic-run-ledger/` y refrescar `change-overview.html`.

- [x] 6. Verify, sync y delivery readiness
  - [x] 6.1 Verificar implementación contra proposal, design y los tres deltas; resolver gaps sin ampliar alcance silenciosamente.
  - [x] 6.2 Ejecutar validación SDD estricta y revisión privacy/security específica.
  - [x] 6.3 Sincronizar deltas validados a specs canónicas antes de archive.
  - [x] 6.4 Preparar delivery en slices revisables según el harness y registrar comandos/resultados reales.
  - [x] 6.5 No archivar ni reportar `closed` hasta satisfacer validación, sync, delivery y checks remotos requeridos.
