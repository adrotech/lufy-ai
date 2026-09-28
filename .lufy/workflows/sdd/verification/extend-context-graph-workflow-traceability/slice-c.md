# Slice C — `workflow_limits.review` y workload assessment

## Estado

- Configuración, observaciones Git, evaluator y CLI: implementados y validados.
- El fallback de Slice B queda completo: aun sin grafo usable, el diff directo se conserva y la decisión escala.
- Gate del change: `implemented`; faltan Slice D, validación final, sync y delivery.

## Contratos

- La única fuente canónica es `workflow_limits.review`; campos top-level homónimos no se consumen.
- Los defaults de una configuración nueva son `8/800/3/2`; una carga parcial conserva ceros como límites no disponibles y `rescan` completa defaults sin perder overrides ni extras.
- `context review --base <ref>` observa archivos, adiciones, eliminaciones y churn con `git diff --numstat --no-renames -z --end-of-options`.
- `proceed` exige límites disponibles, evidencia suficiente, concurrencia dentro del límite y trazabilidad completa.
- Exceder archivos o churn recomienda `split`; falta de evidencia, exceso de concurrencia o trazabilidad `unknown/incomplete` recomienda `escalate` y tiene precedencia.
- Alcanzar exactamente un máximo permanece dentro del presupuesto; solamente excederlo viola el límite.
- Totales cubren el diff completo y el detalle se limita a 128 elementos.

## TDD

### RED

- `projectconfig` no exponía `WorkflowReviewLimits` ni validaba defaults, parciales, rescan o negativos.
- el adapter Git no exponía estadísticas numstat bounded.
- application no tenía `Review`, tabla de decisiones ni fallback conservador.
- CLI rechazaba `context review`.

### GREEN / TRIANGULATE / REFACTOR

- defaults, YAML parcial, extras, top-level legacy ignorado y cuatro negativos;
- paths con espacios, CRLF/NUL, binarios, ref inválida y 140 archivos;
- combinaciones independientes de config/grafo/trazabilidad/concurrencia/evidencia/files/churn;
- grafo missing/stale conserva observaciones directas y nunca aprueba;
- human/JSON/help/usage, enteros no negativos y semántica de exit codes;
- auditoría posterior ajustó el boundary de máximos a comparación estricta y agregó `--end-of-options`.

## Evidencia

| Comando | Resultado |
| --- | --- |
| `go test ./internal/projectconfig ./internal/contextgraph/... ./internal/cli -count=1` | passed |
| `git diff --check` | passed |

## Riesgo residual

Cuando el diff supera 128 archivos, los totales permanecen completos pero no puede probarse trazabilidad para todo el detalle; el resultado queda `unknown` y escala de forma conservadora.
