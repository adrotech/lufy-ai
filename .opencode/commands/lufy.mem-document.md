---
description: Documenta decisiones, reglas y flujos durables en Obsidian con frontmatter validable.
agent: orchestrator
---

Ejecuta `/lufy.mem-document` delegando en el skill local `lufy.mem-document`.

## Entradas

- Tema o archivo/ruta a documentar.
- `--decision`, `--rule`, `--flow`, `--lesson` o `--concept`.
- `--dry-run` para revisar estructura antes de escribir.

## Comportamiento

1. Usa `.lufy/memory/knowledge/` como destino privado por defecto.
2. Exige frontmatter con `name`, `description`, `type` y `status`.
3. Para decisiones, incluye sección `**Why:**`.
4. Prefiere `lufy-ai memory capture` para crear o actualizar notas validables en vez de editar Markdown manualmente.
5. Conecta la nota con `lufy-ai memory connect` solo cuando la nota destino exista o será creada en el mismo bloque.
6. Ejecuta o recomienda `lufy-ai memory validate --target <repo>` al final.
