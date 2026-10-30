> Informe histórico del modelo anterior. Sus resultados se conservan; sus comandos y roles no describen la rama inline actual. Ver [estado vigente](status.md).

# Bloque 3: observabilidad operativa Codex

## Resultado
Se corrigió el emparejado de reviews por (run_id, task_ref), preservando pares legacy sin task. Dos inicios para la misma clave vuelven ambiguo el par; no se sobrescribe el timestamp ni se inventa duración. El servicio conserva su shape JSON público. Timestamps ausentes en eventos legacy/read-side de rework/reopened degradan a partial.

## Evidencia
- RED test-writer: cruces entre tasks y sobrescritura de inicios ambiguos antes del fix. La primera assertion de timestamps usaba un fixture inválido para ese propósito: el store asigna ObservedAt si OccurredAt falta. Se corrigió a proyección legacy de lectura, sin cambiar defaults del ledger.
- GREEN focalizado: cinco métodos TestMetricsIntegrity, package PASS 0.629s.
- scripts/validate.sh: exit 0, coverage 80.4% >= 80.0%, vet/build y controles existentes correctos. Shell lint omitido por shellcheck ausente; sin cambios shell.
- Fixtures CLI en /private/tmp/lufy-block3-operational-fixtures-20261022/summary.json: dos tasks en un run suman 70000 ms sin cruces, rework 1, reopened 1; partial/unavailable sin duración; duplicate_noop y fuentes idénticas antes/después de consultar; payload libre rechazado sin persistir contenido.

## Disponibilidad por superficie
| Superficie | Soporte verificado | Disponibilidad operativa |
| --- | --- | --- |
| Run Ledger portable | review.started/completed/rework y defect.reopened aceptados como metadata content-free | Depende de productores y eventos reales |
| Lifecycle Codex | SessionStart/SubagentStart/SubagentStop/Stop/SessionEnd; bindings pseudonimizados, sin transcript | No produce por sí solo review.* ni defect.reopened |
| Context metrics | Pares y counts observados con partial/unavailable explícitos | No convierte lifecycle ni fixtures en tasas operativas |
| Runtime del repo | Consulta real `context metrics --target <repo> --json`, sin inicializar configuración/grafo/memoria | `ledger_availability=available`, `status=unavailable`, `recognized_events=0`; faltan eventos semánticos de review |

Los fixtures prueban mecanismos, no tiempos de review del usuario. Counts son eventos observados, no ceros operativos cuando falta instrumentación. La duración acumula sólo pares inequívocos; un resultado partial no garantiza cobertura total. No hay nuevo backend ni instrumentación OpenCode/Pi.

## Sync y delivery
Delta sincronizado a systemic-workflow y mirror exacto. Autorización global de delivery recibida; el merge sigue condicionado a checks del HEAD final y gates de archive. La fuente del ledger no se modifica por el cálculo de métricas.
