---
description: Conecta notas Obsidian existentes mediante backlinks y actualiza el índice local.
agent: orchestrator
---

Ejecuta `/lufy.mem-connect` delegando en el skill local `lufy.mem-connect`.

## Entradas

- Dos o más notas, conceptos o slugs.
- `--dry-run` para mostrar conexiones propuestas sin mutar.

## Comportamiento

1. Busca notas existentes con `lufy-ai memory search`.
2. Ejecuta `lufy-ai memory connect --target <repo> [--bidirectional] <from-slug> <to-slug>` para backlinks explícitos y sin enlaces rotos.
3. No crea duplicados si ya existe una conexión suficiente.
4. Reconstruye el índice con `lufy-ai memory index --target <repo>` cuando haga falta y valida con `lufy-ai memory validate`.
