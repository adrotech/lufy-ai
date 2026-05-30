# Proposal: Codex dry-run tool adapter

## Why

Lufy ya separa core, tool adapters y methodology adapters, pero solo OpenCode existe como adapter real. Para avanzar hacia compatibilidad multi-tool sin romper instalaciones actuales, necesitamos un primer adapter Codex que sea verificable en código y tests, pero que no escriba assets reales en repos destino.

## What Changes

- Agregar `codex` como tool conocida del dominio.
- Registrar un adapter Codex dry-run en el registry de adapters.
- Declarar capabilities conservadoras: single-agent/fallback inline, sin slash commands, sin hooks, sin TUI y sin config OpenCode.
- Renderizar una superficie dry-run compatible con `AGENTS.md` como preview conceptual, no como asset instalable.
- Mantener `lufy-ai install/sync/verify --tool codex` bloqueado para escritura real.
- Agregar tests de no fuga textual para evitar `.opencode`, `opencode.json` y otros paths OpenCode en el adapter Codex.

## Non-goals

- No instalar assets Codex reales.
- No crear `AGENTS.md` desde el adapter Codex en repos de usuario.
- No integrar ejecución de Codex CLI.
- No agregar soporte Claude Code en este cambio.
- No cambiar el default OpenCode/OpenSpec.
