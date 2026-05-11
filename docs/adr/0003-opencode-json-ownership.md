# ADR 0003: Ownership parcial de `opencode.json`

## Estado

Aceptada.

## Contexto

`opencode.json` es configuración compartida entre el usuario, OpenCode y `lufy-ai`. Reescribirlo completo produce drift silencioso y puede borrar preferencias del usuario.

## Decisión

`lufy-ai` gestiona solo un subconjunto de claves:

- `$schema`
- `plugin`
- `mcp.engram.type`
- `mcp.engram.command`

Las opciones existentes de `mcp.engram` como `enabled`, `timeout`, `env` y otras claves no gestionadas se preservan.

Se registra metadata bajo `x-lufy-ai` con versión y claves gestionadas.

## Consecuencias

- `opencode.json` no se registra como asset gestionado por hash completo en `install-state.json`.
- Los merges deben ser conservadores y validar tipos antes de escribir.
- `mcp: null` se trata como ausente y puede ser reemplazado por un objeto gestionado.
