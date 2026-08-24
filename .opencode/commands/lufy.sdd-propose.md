---
description: Crear y validar artifacts de un change Lufy SDD.
agent: orchestrator
---

Lee `.lufy/workflows/sdd/actions/propose.md`. Elige `--mode full|lite`, crea el scaffold con `lufy-ai sdd new`, completa los Markdown del mode y ejecuta `lufy-ai sdd validate --change <name> --strict` antes de recomendar apply. El overview HTML se crea y refresca automáticamente.

El handoff MUST incluir `methodology_id: lufy-sdd`, `methodology_mode`, estado de gates y siguiente acción.
