---
description: Busca contexto durable en la memoria Obsidian portable del proyecto.
agent: orchestrator
---

Ejecuta `/lufy.mem-search` delegando en el skill local `lufy.mem-search`.

## Entradas

- Query textual.

## Comportamiento

1. Ejecuta `lufy-ai memory status --target <repo>`.
2. Si la memoria está inicializada, busca con `lufy-ai memory search --target <repo> <query>`.
3. Devuelve solo hints compactos: nota, estado, línea y relevancia.
4. No trata memoria como evidencia más fuerte que archivos, comandos o instrucciones explícitas.
