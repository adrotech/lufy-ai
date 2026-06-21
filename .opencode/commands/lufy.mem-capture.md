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

1. Verifica `.lufy/config/project.yaml` y `memory.provider=obsidian`.
2. Si falta estructura, recomienda `lufy-ai memory init --target <repo>`.
3. Ejecuta `lufy-ai memory capture --target <repo> --title <title> --type <type> [--link <slug>] <texto>` para crear o actualizar la nota.
4. Usa `lufy-ai memory connect` cuando el usuario pida relacionar notas o haya notas existentes claramente relacionadas.
5. Evita duplicar memoria rutinaria o transitoria y valida con `lufy-ai memory validate --target <repo>`.
