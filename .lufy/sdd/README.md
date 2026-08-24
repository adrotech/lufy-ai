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

## Methodology Contract

Lufy SDD guía el trabajo del LLM con contratos explícitos:

- objetivos medibles y outcomes esperados;
- restricciones de desarrollo y non-goals;
- entorno real, comandos disponibles y dependencias;
- patrones de diseño y límites arquitectónicos;
- diagramas Mermaid en Markdown fuente cuando reducen ambigüedad;
- criterios WHEN/THEN verificables;
- gates separados para implementación, validación, sync, delivery y cierre.

### Full

Usa Full para T1: arquitectura, contratos públicos, seguridad, datos, impacto transversal o incertidumbre alta. Full requiere `proposal.md`, `design.md`, `tasks.md` y al menos un spec delta. `design.md` debe cubrir arquitectura, componentes, datos/persistencia cuando aplique, seguridad, operaciones, tradeoffs y diagramas.

### Lite

Usa Lite para T2: cambios acotados con riesgo o comportamiento verificable. Lite conserva solo `proposal.md` y `tasks.md`, pero debe capturar objetivo del LLM, restricciones, contexto mínimo, aceptación, validación proporcional y riesgos. Puede incluir diagramas compactos si aclaran el flujo; no debe convertirse en Full.

## Spec delta

Usa secciones `ADDED`, `MODIFIED` o `REMOVED`. Cada requirement agregado o modificado requiere al menos un scenario con `WHEN` y `THEN`; `GIVEN` es opcional.
