---
name: lufy-sdd-propose
description: Crear y validar artifacts de un change Lufy SDD.
---

# Lufy SDD propose

Lee `.lufy/workflows/sdd/actions/propose.md`. Elige `--mode full|lite`, crea el scaffold con `lufy-ai sdd new`, completa los Markdown del mode y ejecuta `lufy-ai sdd validate --change <name> --strict` antes de recomendar apply. El overview HTML es automático y derivado; no requiere otra acción.

## Output contract

Reporta `methodology_id: lufy-sdd`, mode efectivo, artifacts leídos o escritos, evidencia, riesgos, gate state y siguiente acción. No inventes artifacts OpenSpec.
