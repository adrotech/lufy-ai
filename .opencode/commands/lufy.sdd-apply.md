---
description: Implementar tasks aprobadas de un change Lufy SDD.
agent: orchestrator
---

Lee `.lufy/workflows/sdd/actions/apply.md` y los Markdown fuente del change. Implementa por bloques y ejecuta validate para refrescar el overview derivado; no confundas implementación con validación, sync, delivery o cierre.

El handoff MUST incluir `methodology_id: lufy-sdd`, `methodology_mode`, estado de gates y siguiente acción.
