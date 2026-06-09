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
2. Propone backlinks explícitos y evita enlaces rotos.
3. No crea duplicados si ya existe una conexión suficiente.
4. Actualiza la nota mínima necesaria y valida con `lufy-ai memory validate`.
