---
description: Archivar un change Lufy SDD con gates completos.
agent: orchestrator
---

Lee `.lufy/workflows/sdd/actions/archive.md`. Ejecuta `lufy-ai sdd archive --change <name>` únicamente con tasks, evidencia, validación y delivery requeridos resueltos; Full además exige sync. Archive refresca y preserva el overview.

El handoff MUST incluir `methodology_id: lufy-sdd`, `methodology_mode`, estado de gates y siguiente acción.
