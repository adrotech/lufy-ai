# Design: Codex dry-run adapter

## Modelo

`codex` se incorpora como `ToolID` conocido pero no escribible. La diferencia entre "conocido" y "writable" queda en el adapter:

- `domain.ToolCodex` permite modelar y testear compatibilidad.
- `ToolCapabilities.DryRunOnly` marca que no se puede usar para mutaciones reales.
- La CLI mantiene el bloqueo actual para comandos mutantes.

## Capabilities

Codex se modela inicialmente así:

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

La decisión importante es `Subagents=false`: el renderer debe representar roles Lufy como fases inline y contratos compactos, no prometer aislamiento nativo.

## Render dry-run

`RenderSurface` para Codex devuelve specs de preview con policy `dry-run` y target conceptual `AGENTS.md`. Estos specs no deben ser consumidos por `install` ni `sync` como catálogo real en este cambio.

El contenido real de `AGENTS.md` queda fuera de este slice. Solo se valida que el adapter:

- no dependa de `.opencode`;
- no mencione `opencode.json`;
- documente fallback inline;
- exponga gaps de capabilities.

## CLI

Los comandos mutantes siguen rechazando `--tool codex`. El mensaje de error debe seguir dejando claro que el único adapter escribible es `opencode`.

## Riesgos

- Confundir "adapter conocido" con "instalable". Mitigación: `DryRunOnly` y tests de CLI que mantienen el bloqueo.
- Sobreprometer capabilities de Codex. Mitigación: capabilities conservadoras y docs de gaps.
