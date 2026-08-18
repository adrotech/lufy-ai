---
name: lufy-sdd-verify
description: Verificar implementación y gates de un change Lufy SDD.
---

# Lufy SDD verify

Lee `.lufy/workflows/sdd/actions/verify.md`. Ejecuta validación strict —que refresca el overview—, contrasta scenarios con código/tests y registra evidencia regular bajo `.lufy/workflows/sdd/verification/<change>/`.

## Output contract

Reporta `methodology_id: lufy-sdd`, mode efectivo, artifacts leídos o escritos, evidencia, riesgos, gate state y siguiente acción. No inventes artifacts OpenSpec.
