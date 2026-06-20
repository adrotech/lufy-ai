---
description: Consulta el grafo de contexto local como índice secundario de hints compactos.
agent: orchestrator
---

Ejecuta `/lufy.context` delegando en el skill local `lufy.context-search`.

## Entradas

- Query textual, ruta, símbolo o intención de impacto.
- Opcional: `--base <ref>` cuando se busca impacto de diff.

## Comportamiento

1. Verifica disponibilidad con `lufy-ai context status --target <repo> --json`.
2. Si está `ready`, consulta con `lufy-ai context query --target <repo> --json <query>`.
3. Si se entrega `--base`, puede usar `lufy-ai context diff --target <repo> --json --base <ref>`.
4. Si el grafo falta, falla o está stale, reporta `context_graph_hints.status: not_available` o `stale` y `recovery: lufy-ai context build`.
5. Devuelve solo hints compactos y recuerda que el grafo no reemplaza archivos actuales, diff, tests ni comandos de validación.
