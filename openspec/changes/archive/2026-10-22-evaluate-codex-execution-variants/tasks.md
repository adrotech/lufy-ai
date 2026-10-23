# Tasks — bloque 4

- [x] 0.1 Definir comparación, aceptación B4-AC1–AC5, decisión de workload y resultado negativo permitido.
- [x] 0.2 Validar propuesta estricta y registrar baseline B2/B3 disponible sin inventar datos.
- [x] 1.1 Implementar/ejecutar comparación acotada de entrega inline/referencia con controles de fidelidad y métricas observadas.
- [x] 1.2 Verificar recuperación Lite/checkpoints e idempotencia sobre mecanismos existentes.
- [x] 1.3 Evaluar alcance de autorización y estrategia TDD; identificar qué es desk check y qué tiene datos pareados.
- [x] 2.1 Validar y documentar decisión de adopción/no adopción con límites, sin claims de ahorro no medidos.
- [x] 2.2 Sync de delta y mirror tras aceptación.
- [x] 3.1 Publicación autorizada en PR #243 con cinco checks iniciales exitosos; elegibilidad de archive comprobada. HEAD con archive y merge final siguen gates explícitos en verification.md/result.yaml.

Decisión validada: keep_baseline, adopted=false; primera muestra baseline timeout180s y parada inmediata, candidato no ejecutado. Sin claims de mejora. Checkpoint recorded/duplicate_noop y run verify saludable. TDD comparado como desk check. Evidencia en docs/codex-execution-experiments-2026-10-22.md; unittest68/68 passed.
