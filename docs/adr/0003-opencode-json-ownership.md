# ADR 0003: Ownership parcial de `opencode.json`

## Estado

Aceptada.

## Contexto

`opencode.json` es configuración compartida entre el usuario, OpenCode y `lufy-ai`. Reescribirlo completo produce drift silencioso y puede borrar preferencias del usuario.

## Decisión

`lufy-ai` gestiona solo un subconjunto de claves:

- `$schema`
- `plugin`

Durante el merge también limpia metadata legacy `x-lufy-ai` y remueve integraciones MCP de memoria externa discontinuadas, preservando otros MCPs del usuario.

## Consecuencias

- `opencode.json` no se registra como asset gestionado por hash completo en `install-state.json`.
- Los merges deben ser conservadores y validar tipos antes de escribir.
- `mcp` se preserva como configuración user-owned salvo claves legacy que Lufy retire explícitamente.
