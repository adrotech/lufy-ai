---
name: lufy.mem-connect
description: Conecta notas Obsidian existentes con backlinks seguros y evita enlaces rotos.
license: MIT
compatibility: OpenCode skill autocontenido; usa lufy-ai memory search/validate.
metadata:
  author: lufy-ai
  version: "1.0"
---

# Skill: lufy.mem-connect

Usar para enlazar conocimiento existente. El objetivo es mejorar navegación, no crear una red de enlaces decorativa.

## Flujo

1. Confirmar que `.lufy/config/project.yaml` declara `memory.provider: obsidian` y que `.lufy/memory` existe con `lufy-ai memory status --target <repo> --json`.
2. Buscar notas candidatas con `lufy-ai memory search`.
3. Elegir enlaces que representen dependencia real, decisión relacionada, regla base o lesson aplicable.
4. Ejecutar `lufy-ai memory connect --target <repo> [--bidirectional] <from-slug> <to-slug>` para editar solo las notas necesarias.
5. Reconstruir índice con `lufy-ai memory index --target <repo>` cuando se editen relaciones manualmente.
6. Validar con `lufy-ai memory validate --target <repo>`.

## Criterios

- No crear backlinks a slugs inexistentes.
- No enlazar notas `deprecated` salvo que el objetivo sea explicar historia o migración.
- Si una nota está `superseded`, preferir su reemplazo activo.
