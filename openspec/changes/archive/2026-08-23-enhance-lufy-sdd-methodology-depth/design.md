# Design: enhance-lufy-sdd-methodology-depth

## Context

La metodología nativa Lufy SDD vive en dos superficies sincronizadas:

- assets instalables bajo `.lufy/sdd/`;
- assets embebidos bajo `tools/lufy-cli-go/internal/assets/embedded/.lufy/sdd/`;
- scaffold runtime hardcodeado en `tools/lufy-cli-go/internal/lufysdd/service.go`.

El `change-overview.html` se deriva automáticamente de los Markdown del change. Por eso los Markdown deben ser la fuente de verdad rica; el HTML no se edita ni se usa como source.

## Decisions

### Decision: Full SDD es prescriptivo

Full debe guiar al LLM como un contrato de arquitectura y ejecución. El scaffold incluirá secciones para:

- objetivos medibles del LLM;
- restricciones y límites de cambio;
- entorno esperado y comandos reales;
- arquitectura y patrones de diseño;
- componentes, flujos y datos;
- seguridad, privacidad y operaciones;
- diagramas Mermaid obligatorios cuando apliquen;
- criterios de aceptación WHEN/THEN;
- tareas por bloques coherentes y gates.

### Decision: Lite es compacto, no informal

Lite seguirá sin `design.md` ni specs delta, pero dejará de ser un checklist mínimo. Su `proposal.md` debe capturar el contrato suficiente para cambios T2: objetivo, contexto, restricciones, alcance, aceptación, impacto, entorno y riesgos. Sus tareas deben separar implementación, validación y evidencia.

### Decision: Diagramas viven en Markdown fuente

Los diagramas se expresan como bloques Mermaid dentro de `design.md` para Full y de forma opcional en `proposal.md` para Lite. Se incluyen plantillas para:

- flujo de trabajo;
- componentes;
- arquitectura;
- datos/persistencia cuando aplique.

### Decision: Runtime y templates deben permanecer alineados

`lufy-ai sdd new` debe generar el mismo contrato que los templates de referencia. Los tests verificarán contenido clave en scaffolds Full y Lite. Después de tocar `.lufy/sdd`, se copiará el mismo contenido a `tools/lufy-cli-go/internal/assets/embedded/.lufy/sdd`.

## Risks

- Templates demasiado largos pueden generar ruido para cambios pequeños. Mitigación: Lite conserva menos artifacts y marca diagramas como opcionales salvo que haya ambigüedad.
- Mermaid en scaffolds puede ser interpretado como requerido aunque no aplique. Mitigación: incluir instrucciones `Si aplica` y placeholders borrables.
- Drift entre assets raíz y embebidos. Mitigación: ejecutar `go test ./internal/assets`.

## Validation Plan

- `openspec validate enhance-lufy-sdd-methodology-depth --strict`
- `go test ./internal/lufysdd`
- `go test ./internal/assets`
- `scripts/validate.sh` si el bloque queda listo para delivery.
