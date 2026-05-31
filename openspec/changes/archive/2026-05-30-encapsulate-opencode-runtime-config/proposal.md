# Proposal: encapsular runtime config de OpenCode

## Why

El catálogo de assets ya se resuelve por adapters, pero los casos de uso de install, sync, verify y status todavía llaman directamente a helpers específicos de OpenCode para `opencode.json` y config global. Ese acople complica convertir a Lufy en un harness neutral porque mezcla reglas de aplicación con detalles concretos de la tool actual.

## What Changes

- Agregar una capa interna de runtime de tool para resolver configuración project/global de la tool efectiva.
- Mantener `opencode` como única tool escribible actual y preservar el comportamiento existente.
- Reemplazar llamadas directas desde install, sync, verify y status por la capa runtime.

## Non Goals

- No habilitar escritura real para Codex o Claude Code.
- No cambiar paths, defaults ni contenido de `opencode.json`.
- No mover la implementación concreta de `internal/config`; solo ocultarla detrás del runtime.
