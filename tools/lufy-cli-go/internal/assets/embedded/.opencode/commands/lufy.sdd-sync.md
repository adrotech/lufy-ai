---
description: Sincronizar deltas Lufy SDD con specs activas.
agent: orchestrator
---

Lee `.lufy/workflows/sdd/actions/sync.md`. En Full, ejecuta `lufy-ai sdd sync --change <name>` tras validación y reporta cualquier ambigüedad como bloqueo. En Lite, acepta `not_applicable` sin mutaciones. Sync refresca el overview y no archiva.

El handoff MUST incluir `methodology_id: lufy-sdd`, `methodology_mode`, estado de gates y siguiente acción.
