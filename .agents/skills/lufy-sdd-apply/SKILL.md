---
name: lufy-sdd-apply
description: Implementar tasks aprobadas de un change Lufy SDD.
---

# Lufy SDD apply

Lee `.lufy/workflows/sdd/actions/apply.md` y los Markdown fuente del change. Implementa por bloques y ejecuta validate para refrescar el overview derivado; no confundas implementación con validación, sync, delivery o cierre.

## Output contract

Reporta `methodology_id: lufy-sdd`, mode efectivo, artifacts leídos o escritos, evidencia, riesgos, gate state y siguiente acción. No inventes artifacts OpenSpec.
