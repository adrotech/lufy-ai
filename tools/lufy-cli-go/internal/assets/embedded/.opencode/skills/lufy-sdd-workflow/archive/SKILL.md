---
name: lufy-sdd-archive
description: Archivar un change Lufy SDD con gates completos.
---

# Lufy SDD archive

Lee `.lufy/workflows/sdd/actions/archive.md`. Ejecuta `lufy-ai sdd archive --change <name>` únicamente con tasks, evidencia, validación y delivery requeridos resueltos; Full además exige sync. Archive refresca y preserva el overview.

## Output contract

Reporta `methodology_id: lufy-sdd`, mode efectivo, artifacts leídos o escritos, evidencia, riesgos, gate state y siguiente acción. No inventes artifacts OpenSpec.
