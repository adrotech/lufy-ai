# Proposal: Claude Code dry-run tool adapter

## Why

Después del adapter Codex dry-run, el siguiente paso del plan multi-tool es modelar Claude Code sin habilitar escritura real. Esto permite validar capabilities, gaps y superficie de instrucciones antes de decidir una instalación experimental.

## What Changes

- Agregar `claude-code` como tool conocida del dominio.
- Registrar un adapter Claude Code dry-run en el registry.
- Declarar capabilities conservadoras y `DryRunOnly=true`.
- Renderizar preview conceptual basado en `CLAUDE.md`.
- Mantener comandos mutantes bloqueados para `--tool claude-code`.
- Agregar tests de no fuga OpenCode.

## Non-goals

- No escribir `CLAUDE.md` en repos destino.
- No integrar la CLI de Claude Code.
- No mapear commands/settings reales fuera de un preview.
- No cambiar el default OpenCode/OpenSpec.
