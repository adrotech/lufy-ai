# ADR 0002: Archivos extra en directorios gestionados

## Estado

Aceptada.

## Contexto

El catálogo gestiona directorios completos como `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/plugins`, `.opencode/policies` y `openspec`. Los usuarios pueden necesitar personalizar esos directorios con archivos locales.

## Decisión

Los archivos extra creados por el usuario dentro de directorios gestionados se permiten y se preservan.

`verify` los reporta como `info` para visibilidad, pero no los trata como error.

## Consecuencias

- El instalador no elimina personalizaciones locales por defecto.
- Si una versión futura del catálogo agrega una ruta que ya existe localmente y no está gestionada, la operación debe bloquearse como conflicto para evitar sobrescritura silenciosa.
- Se recomienda que los usuarios nombren personalizaciones con prefijos como `local-*` o usen subdirectorios explícitos cuando sea posible.
