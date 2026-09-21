# Tasks: make-result-contract-executable

## Review Workload Harness

| Slice | Objetivo | Criterio de salida | Validación agrupada |
| --- | --- | --- | --- |
| A | Schema, decoder, canonicalización | corpus v1 determinista | unit + fuzz/property + golden |
| B | Roles, evidencia y transiciones | tabla sin gates ambiguos | state table + role/evidence matrix |
| C | Ownership, leases, joins/idempotencia | CAS/fencing/retry seguros | race + concurrency + crash fixtures |
| D | CLI, lifecycle, ledger, assets/docs | adapters compatibles/content-free | CLI/lifecycle E2E + full suite |

- [x] 1. Slice A — Contract-first y pruebas RED
  - [x] 1.1 Inventariar fixtures reales de roles, docs y adapters.
  - [x] 1.2 Definir structs, enums y límites cerrados para bloques v1, preservando opcionales.
  - [x] 1.3 Escribir RED YAML/JSON para valid, unknown/duplicate, aliases/tags, invalid UTF-8 y oversized.
  - [x] 1.4 Escribir RED de canonicalización/fingerprint ante orden YAML, LF/CRLF y OS.
  - [x] 1.5 Añadir canaries de prompt/secreto/output y diagnósticos sanitizados.

- [x] 2. Slice A — Dominio y decoder
  - [x] 2.1 Implementar `internal/resultcontract` adapter-neutral con errores tipados.
  - [x] 2.2 Implementar decoder bounded/strict sin mapas arbitrarios.
  - [x] 2.3 Implementar canonical JSON y SHA-256 estable.
  - [x] 2.4 Implementar validación estructural/coherencia sin afirmar evidencia material.
  - [x] 2.5 Registrar RED/GREEN/TRIANGULATE/REFACTOR y validar Slice A.

- [x] 3. Slice B — Roles, evidencia y transiciones
  - [x] 3.1 Probar exhaustivamente `allowed_status` por rol.
  - [x] 3.2 Definir vector derivado y tabla explícita de edges/terminal/recovery.
  - [x] 3.3 Definir policies mínimas de evidencia con inyección SDD/delivery.
  - [x] 3.4 Implementar `result-transition/v1`, intent y decision.
  - [x] 3.5 Validar role/status, evidencia, previous fingerprint y expected version.
  - [x] 3.6 Probar que reflexión, texto final o schema substring no avanzan gates.
  - [x] 3.7 Ejecutar validación agrupada Slice B.

- [x] 4. Slice C — Ownership, leases, joins e idempotencia
  - [x] 4.1 Escribir RED de owner mismatch, lease válido/vencido, stale version y fencing.
  - [x] 4.2 Escribir RED same-key duplicate-noop y conflicting fingerprint.
  - [x] 4.3 Escribir RED join: terminal, missing, blocked, causal mismatch y grouped evidence.
  - [x] 4.4 Implementar owner/lease con refs pseudonimizadas y token digest.
  - [x] 4.5 Implementar receipts con semántica Run Ledger sin persistir envelope.
  - [x] 4.6 Implementar join validator con conjunto requerido explícito.
  - [x] 4.7 Ejecutar race/concurrency y crash/retry del Slice C.

- [x] 5. Slice D — CLI y normalización
  - [x] 5.1 Implementar `lufy-ai result validate` con stdin/file, role y human/JSON.
  - [x] 5.2 Implementar `result normalize` conservador con fallback/provenance.
  - [x] 5.3 Implementar `result transition` read-only por default y recording explícito.
  - [x] 5.4 Definir exit codes para valid, rejected, conflict, unavailable y usage.
  - [x] 5.5 Añadir palette/help/docs y pruebas CLI.

- [x] 6. Slice D — Lifecycle, ledger y contratos
  - [x] 6.1 Reemplazar substring de lifecycle por parse/validate bounded.
  - [x] 6.2 Mantener hook best-effort, warning sanitizado y `continue=true`.
  - [x] 6.3 Implementar bridge content-free y probar todos los receipt outcomes.
  - [x] 6.4 Actualizar contrato, producer, templates y assets embedded en paridad.
  - [x] 6.5 Probar ausencia durable de summary, prompts, paths, outputs y canaries.
  - [x] 6.6 Ejecutar coupling/catalog/lifecycle y validación Slice D.

- [x] 7. Verify, sync y delivery readiness
  - [x] 7.1 Ejecutar suite, build, race focalizado, fuzz bounded y `scripts/validate.sh`.
  - [x] 7.2 Verificar contra proposal/design/deltas y revisar privacidad/cross-platform.
  - [x] 7.3 Registrar evidencia en `verification/make-result-contract-executable/` y strict validate.
  - [x] 7.4 Sincronizar deltas validados.
  - [x] 7.5 Preparar PR único con cuatro review slices; no iniciar Loop Engine antes de merge/cierre.
  - [x] 7.6 No archivar ni cerrar sin delivery, checks, sync y gates completos.
