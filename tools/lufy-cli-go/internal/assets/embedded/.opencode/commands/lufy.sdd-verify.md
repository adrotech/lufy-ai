---
description: Verificar implementación y gates de un change Lufy SDD.
agent: orchestrator
---

Lee `.lufy/workflows/sdd/actions/verify.md`. Ejecuta validación strict —que refresca el overview—, contrasta scenarios con código/tests y registra evidencia regular bajo `.lufy/workflows/sdd/verification/<change>/`.

El handoff MUST incluir `methodology_id: lufy-sdd`, `methodology_mode`, estado de gates y siguiente acción.
