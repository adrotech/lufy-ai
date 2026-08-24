# Propose

Elige el mode proporcional y ejecuta `lufy-ai sdd new --change <name> --mode <full|lite> [--capability <capability>]`.

Antes de declarar el change listo:

- completa el objetivo del LLM, outcome, restricciones, entorno, scope y aceptación;
- para Full, completa `design.md` con arquitectura, componentes, patrones, datos si aplica, seguridad, operaciones, tradeoffs y diagramas Mermaid;
- para Lite, mantén `proposal.md` compacto, pero no omitas objetivo, límites, criterios WHEN/THEN, validación y riesgos;
- completa `tasks.md` por bloques coherentes, no solo micro-checklists;
- valida con `lufy-ai sdd validate --change <name> --strict`.

Full completa `proposal.md`, `design.md`, `tasks.md` y specs delta; Lite completa `proposal.md` y `tasks.md`.

`change-overview.html` se crea y refresca automáticamente; no invoques otro comando o skill para renderizarlo ni lo edites como fuente.

No implementes runtime durante esta acción.
