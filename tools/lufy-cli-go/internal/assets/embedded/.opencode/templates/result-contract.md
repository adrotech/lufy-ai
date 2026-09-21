# OpenCode Result Contract Overlay

The canonical Result Contract envelope lives at `.lufy/contracts/result-contract.md` and MUST be used for substantive handoffs and workflow results.

This file is a compatibility pointer for older OpenCode surfaces. It MUST NOT define a second envelope.

Antes de aceptar un handoff sustantivo, valida el documento canonical con `lufy-ai result validate --stdin`; la fuente normativa permanece en `.lufy/contracts/result-contract.md`.
