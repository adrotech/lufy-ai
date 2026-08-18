---
name: lufy-sdd-sync
description: Sincronizar deltas Lufy SDD con specs activas.
---

# Lufy SDD sync

Lee `.lufy/workflows/sdd/actions/sync.md`. En Full, ejecuta `lufy-ai sdd sync --change <name>` tras validación y reporta cualquier ambigüedad como bloqueo. En Lite, acepta `not_applicable` sin mutaciones. Sync refresca el overview y no archiva.

## Output contract

Reporta `methodology_id: lufy-sdd`, mode efectivo, artifacts leídos o escritos, evidencia, riesgos, gate state y siguiente acción. No inventes artifacts OpenSpec.
