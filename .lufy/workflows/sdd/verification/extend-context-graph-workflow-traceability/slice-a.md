# Slice A — dominio y extracción semántica

## Estado

- Implementación: completa.
- Validación agrupada: aprobada.
- Gate del change: `implemented`; Slices B–D, sync y delivery continúan pendientes.

## TDD

### RED

`go test ./internal/contextgraph/extractors -run '^TestWorkflowExtractor' -count=1`

Falló por ausencia de los nodos semánticos, seis relaciones explícitas y diagnostics. Los tests preexistentes del extractor permanecieron verdes.

### GREEN

Se implementaron:

- nodos deterministas para change, spec, requirement, scenario, task, decision, Go test e identificadores GitHub locales;
- markers allow-listed `implements`, `verifies`, `depends_on`, `caused_by`, `reviewed_by` y `supersedes`;
- `Diagnostic` bounded y sanitizado;
- `ResolveWorkflowReferences` puro sobre el agregado, sin escaneo global desde `Extract`;
- extractor/cache v2 con persistencia RAW por source y grafo solo con edges resueltos.

### TRIANGULATE

- corpus multi-archivo con edge válido y referencia rota;
- no fuzzy evidence para nombres similares;
- equivalencia LF/CRLF y orden repetible;
- segunda build desde cache conserva resultado;
- un marker cacheado se resuelve cuando aparece solamente su target;
- diagnostics del grafo limitados a 128;
- canaries y `.lufy/runtime/**` ausentes de metadata observable.

### REFACTOR

La resolución quedó separada de extracción. El cache preserva hechos candidatos por fuente y el build resuelve una sola vez después del join, evitando que una referencia rota desaparezca permanentemente.

## Evidencia

| Comando | Resultado |
| --- | --- |
| `go test ./internal/contextgraph/extractors -run 'TestWorkflowExtractor' -count=1` | passed |
| `go test ./internal/contextgraph/extractors -count=1` | passed |
| `go test ./internal/contextgraph/application -count=1` | passed |
| `go test ./internal/contextgraph/... -count=1` | passed |
| `git diff --check` | passed |

## Riesgos siguientes

- Slice B debe consumir `Graph.Diagnostics` y freshness para coverage/trace queries.
- Run/evidence/review nodes content-free permanecen en Slice D.
