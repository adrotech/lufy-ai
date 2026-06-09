---
description: Captura aprendizajes durables en la memoria Obsidian portable del proyecto.
agent: orchestrator
---

Ejecuta `/lufy.mem-capture` delegando en el skill local `lufy.mem-capture`.

## Entradas

- Texto libre con decisión, regla, lesson, flow o concepto durable.
- `--type <decision|rule|flow|lesson|concept>` para clasificar la nota.
- `--dry-run` para proponer la nota sin escribirla.

## Comportamiento

1. Verifica `.lufy/project.yaml` y `memory.provider=obsidian`.
2. Si falta estructura, recomienda `lufy-ai memory init --target <repo>`.
3. Crea o propone una nota bajo `.lufy/memory/knowledge/` usando el template gestionado.
4. Evita duplicar memoria rutinaria o transitoria.
5. Si Engram MCP está disponible, solo puede aportar hints; Obsidian es la memoria canónica.
