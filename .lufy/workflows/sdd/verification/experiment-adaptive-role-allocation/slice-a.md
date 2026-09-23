# Verificación Slice A — adaptive role allocation

## Estado

`implemented` con validación focalizada GREEN. El consumo de `parallel_execution` y `workflow_limits.review` queda pendiente para el application service (task 2.4), por lo que el Slice A todavía no declara cierre completo.

## Evidencia TDD

- RED: `go test ./internal/adaptive/domain ./internal/projectconfig -count=1` falló porque aún no existían los contratos, evaluator ni `AdaptiveRoutingConfig`.
- GREEN: el mismo scope pasó después de implementar decoder strict, validación bounded, scoring y config.
- TRIANGULATE: tabla de protected boundary, missing capability, budget insuficiente, disabled, tie-break y cambio de inputs.
- REFACTOR: scoring quedó como functional core sin I/O; config/rescan preserva extras y no activa el feature.

## Comandos

| Comando | Resultado | Evidencia |
| --- | --- | --- |
| `go test ./internal/adaptive/domain ./internal/projectconfig -count=1` | passed | contratos/scoring/config |
| `go test -race ./internal/adaptive/... -count=1` | passed | dominio libre de race |
| `go test ./internal/projectconfig ./internal/cli -count=1` | passed | config y regresión CLI |
| `git diff --check` | passed | sin errores de whitespace |

## Riesgos pendientes

- La policy efectiva todavía no combina límites existentes; se implementará con ports inyectados antes de marcar task 2.4.
- Assignment, lease, waiting pool y yield permanecen fuera de este slice.
