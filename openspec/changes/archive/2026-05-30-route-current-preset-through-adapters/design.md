# Design: Effective catalog from adapters

## Decision

Se agrega una capa pequeña entre catalogo base y casos de uso:

```text
base catalog root/embedded
  + tool adapter RenderSurface
  + methodology adapter RenderWorkflow por tier
  -> effective catalog
```

El catalogo base sigue enumerando assets disponibles. El catalogo efectivo decide que subset aplica para el harness seleccionado.

## Adapter usage

- `opencode.RenderSurface` define `.opencode/*`, `tui.json` y `opencode.json` como config especial.
- `openspec.RenderWorkflow` define `openspec/*`.
- `lufy-sdd.RenderWorkflow` define `.lufy/sdd/*`.
- `none.RenderWorkflow` no agrega assets.

El resolver debe ignorar specs que no son archivos/directorios gestionados por hash, por ejemplo `merge-json` para `opencode.json`; esa config sigue aplicandose por el servicio actual de OpenCode mientras `opencode` sea la unica tool escribible.

## Compatibility

El default `opencode + openspec full/lite + none` debe instalar el mismo preset OpenCode/OpenSpec observable.

`lufy-sdd` debe conservar el comportamiento del slice anterior: `lite` omite `specs` cuando no existe ningun tier full; `full` incluye `specs`.

## Validation

- Tests unitarios del resolver de catalogo efectivo.
- Tests existentes de install/sync/verify/CLI.
- `scripts/validate.sh`.
