# Lufy SDD

Workflow SDD nativo gestionado por `lufy-ai`.

## Layout

- `changes/<change>/`: artifacts fuente y `change-overview.html` derivado.
- `specs/<capability>/spec.md`: requirements vigentes del mode full.
- `decisions/`: decisiones transversales durables.
- `verification/<change>/`: evidencia local requerida antes de archive.
- `archive/YYYY-MM-DD-<change>/`: changes cerrados.
- `actions/`: contrato tool-neutral del lifecycle.
- `templates/`: scaffolding humano de referencia.

## CLI

```bash
lufy-ai sdd new --change <name> --mode full --capability <name>
lufy-ai sdd new --change <name> --mode lite
lufy-ai sdd status --change <name>
lufy-ai sdd validate --change <name> --strict
```

Full crea proposal, design, tasks y specs delta. Lite crea proposal y tasks para cambios acotados; `sync` es `not_applicable` porque no mantiene specs activas. Ninguna acción Lufy SDD requiere el binario OpenSpec.

`change-overview.html` se crea con el scaffold y se refresca automáticamente durante validate, sync y archive. Es offline, determinístico y derivado: los Markdown y `change.yaml` continúan siendo la fuente de verdad. No requiere un comando ni skill separado.

## Spec delta

Usa secciones `ADDED`, `MODIFIED` o `REMOVED`. Cada requirement agregado o modificado requiere al menos un scenario con `WHEN` y `THEN`; `GIVEN` es opcional.
