# Propose

Elige el mode proporcional y ejecuta `lufy-ai sdd new --change <name> --mode <full|lite> [--capability <capability>]`. Full completa `proposal.md`, `design.md`, `tasks.md` y specs delta; Lite completa `proposal.md` y `tasks.md`. Valida con `lufy-ai sdd validate --change <name> --strict` antes de declarar el change listo.

`change-overview.html` se crea y refresca automáticamente; no invoques otro comando o skill para renderizarlo ni lo edites como fuente.

No implementes runtime durante esta acción.
