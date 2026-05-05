# Tooling local de OpenCode

Este directorio contiene la configuración local de OpenCode para `lufy-ai`.

Las reglas compartidas viven en `../AGENTS.md` (guía real del repositorio). La plantilla genérica queda en `../AGENTS.md.template`.

## Agents

- `agents/orchestrator.md`: coordinador primario por defecto.
- `agents/explorer.md`: subagente read-only para exploración.
- `agents/implementer.md`: subagente de implementación.
- `agents/validator.md`: subagente read-only para validación.
- `agents/reviewer.md`: subagente read-only para revisión.
- `agents/delivery.md`: subagente de delivery para Git/GH y PRs.

Todos los agentes siguen un estándar común de frontmatter (`description`, `mode`, `temperature`, `steps`, permisos mínimos) y secciones: `Mission`, `Use When`, `Do Not Use When`, `Inputs Expected`, `Workflow`, `Boundaries`, `Validation / Evidence`, `Escalation`, `Required Output`.

Las reglas compartidas de delivery viven en `policies/delivery.md`.

### Checklist para nuevos agentes

- Mantener permisos mínimos; no conceder `edit`, `bash` o Git/GH si no son necesarios.
- Definir `steps` y un contrato de salida claro.
- Explicar cuándo usar/no usar el agente y cómo escalar.
- No prometer tests ni validación sin evidencia real.
- Usar español para contenido humano y preservar identificadores técnicos.

## Commands

Los slash commands viven en `commands/`.

- `opsx-explore`: explorar el codebase sin implementar.
- `opsx-propose`: crear artefactos de propuesta OpenSpec.
- `opsx-apply`: implementar tareas OpenSpec.
- `opsx-verify`: verificar implementación contra la spec.
- `opsx-archive`: archivar un cambio completado.

## Skills

- `skills/sdd-workflow`: OpenSpec/SDD lifecycle

Skills opcionales de delivery, project sync, memoria y release pueden agregarse en proyectos downstream. El kit base solo incluye el lifecycle OpenSpec.

## Agent Observatory TUI Plugin

El plugin de sidebar TUI se carga desde `tui.json` en la raíz:

```json
{
  "$schema": "https://opencode.ai/tui.json",
  "plugin": ["./.opencode/plugins/agent-observatory.tsx"],
  "plugin_enabled": {
    "lufy-ai.observatory": true
  }
}
```

Slash commands registrados por el plugin actual:

- `/observatory`: mostrar/ocultar el panel.
- `/observatory-agents`: contraer/expandir la lista de agentes.
- `/observatory-subagents`: contraer/expandir la sección de subagentes.
- `/observatory-cost`: mostrar/ocultar costo.

V1 es local/TUI-only. No agregar telemetría externa sin una propuesta separada.
