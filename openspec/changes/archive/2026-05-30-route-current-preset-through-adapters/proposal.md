# Proposal: Route current preset through adapters

## Why

El preset actual OpenCode/OpenSpec ya funciona y debe seguir siendo el default. Sin embargo, `install`, `sync` y `verify` todavía derivan parte de sus decisiones desde paths hardcodeados en catalogo/config en vez de consultar el registry de adapters.

Para terminar la migracion sin agregar nuevas tools, el comportamiento observable actual debe quedar gobernado por `ToolAdapter` y `MethodologyAdapter`: OpenCode declara su surface, OpenSpec declara su workflow, `lufy-sdd` declara su workflow instalable y `none` no agrega assets formales.

## What Changes

- Agregar un resolver de catalogo efectivo que use el registry de adapters.
- Reemplazar el filtrado manual por metodologia por un filtro basado en `RenderSurface` y `RenderWorkflow`.
- Mantener `opencode` como unica tool escribible y default.
- Mantener compatibilidad de install/sync/verify/status para OpenCode/OpenSpec.
- Ajustar `lufy-sdd` para que su adapter declare assets instalables, no previews dry-run.

## Non-goals

- No habilitar Codex ni Claude Code para escritura.
- No cambiar defaults.
- No renombrar `/opsx-*`.
- No reescribir agentes/skills existentes fuera del surface actual.
