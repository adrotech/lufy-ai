# Design: Claude Code dry-run adapter

## Modelo

`claude-code` se agrega como `ToolID` conocido y `DryRunOnly`. Igual que Codex, esto habilita tests y diseño incremental sin mezclarlo con instalación real.

## Capabilities

```text
Subagents: false
SlashCommands: false
Skills: false
Hooks: false
MCP: false
TUI: false
GlobalConfig: false
ProjectConfig: true
SystemPrompt: true
DryRunOnly: true
```

El preview se centra en `CLAUDE.md`, roles inline y gaps. No se declara compatibilidad con rutas ni config de OpenCode.

## CLI

`install`, `sync` y `verify` siguen rechazando `--tool claude-code` para escritura o verificación real de manifest. La habilitación futura requiere una propuesta separada con flag experimental y reglas de rollback.

## Checks

Los tests del adapter deben fallar si el render incluye referencias a rutas OpenCode o `opencode.json`.
