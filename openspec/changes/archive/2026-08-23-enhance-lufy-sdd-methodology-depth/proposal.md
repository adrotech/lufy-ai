# Proposal: enhance-lufy-sdd-methodology-depth

## Why

Lufy SDD Full y Lite ya existen como lifecycle nativo, pero sus scaffolds actuales son demasiado delgados para guiar a un LLM en trabajo real. No fuerzan objetivos observables, restricciones de desarrollo, patrones de diseño, entorno operativo, criterios de calidad ni diagramas de arquitectura o datos.

Esto deja espacio para propuestas ambiguas, implementaciones sin límites claros y revisiones costosas. El workflow propio de Lufy debe convertirse en una metodología suficientemente prescriptiva para que el LLM sepa qué debe lograr, qué no puede tocar, cómo debe diseñar, cómo validar y qué diagramas debe mantener según el tier.

## What Changes

- Profundizar los templates nativos de Lufy SDD Full:
  - `proposal.md` con objetivos del LLM, outcomes, restricciones, entorno, alcance, aceptación y entregables.
  - `design.md` con arquitectura, patrones, componentes, datos, seguridad, operaciones, tradeoffs y diagramas Mermaid.
  - `tasks.md` con bloques coherentes, validación, sync, delivery y documentación.
  - `spec.md` con requirements verificables y criterios de testabilidad.
- Profundizar Lufy SDD Lite sin convertirlo en Full:
  - mantener solo `proposal.md` y `tasks.md`;
  - exigir objetivos, restricciones, contexto mínimo, aceptación y validación proporcional;
  - permitir diagramas compactos solo cuando reduzcan ambigüedad.
- Alinear el scaffold runtime de `lufy-ai sdd new` con los templates instalables.
- Actualizar acciones lifecycle para que `propose`, `apply` y `verify` preserven estos contratos.
- Sincronizar assets embebidos para que el instalador standalone entregue la misma metodología.

## Non-Goals

- No cambiar el lifecycle de comandos `new|status|validate|sync|archive`.
- No agregar un renderer nuevo ni cambiar el contrato automático de `change-overview.html`.
- No exigir base de datos o diagramas de datos cuando el cambio no toca persistencia.
- No convertir T2 Lite en T1 Full; Lite debe seguir siendo compacto y proporcional.

## Acceptance Criteria

- **WHEN** `lufy-ai sdd new --mode full` crea un change
- **THEN** `proposal.md`, `design.md`, `tasks.md` y `specs/<capability>/spec.md` contienen secciones prescriptivas para objetivos del LLM, restricciones, entorno, patrones, arquitectura, datos, diagramas, aceptación y validación.

- **WHEN** `lufy-ai sdd new --mode lite` crea un change
- **THEN** crea solo `proposal.md` y `tasks.md`, con profundidad suficiente para objetivos, restricciones, contexto, aceptación y validación proporcional.

- **WHEN** cambian templates o assets Lufy SDD gestionados
- **THEN** las copias bajo `tools/lufy-cli-go/internal/assets/embedded` quedan sincronizadas.

- **WHEN** se ejecuta validación local
- **THEN** pasan `openspec validate enhance-lufy-sdd-methodology-depth --strict`, `go test ./internal/lufysdd` y `go test ./internal/assets`.
