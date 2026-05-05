## Why

El `orchestrator` ya es el punto natural de routing, pero hoy solo conoce `explorer`, `implementer`, `validator`, `reviewer` y `delivery`. Queremos capturar un modelo donde el `orchestrator` pueda enrutar implementación a agentes especializados por dominio sin perder un fallback seguro para tooling, docs, CI, OpenSpec y cambios pequeños.

## What Changes

- Añadir agentes de dominio adaptados al repositorio, inspirados en los subagents externos:
  - `frontend-developer`
  - `backend-developer`
  - `mobile-developer`
  - `microservices-architect`
- Actualizar el `orchestrator` para poder invocar esos agentes según señales de la tarea.
- Mantener `implementer` como fallback inicial para tareas de tooling, docs, CI, OpenSpec, configuración, installer local y cambios pequeños no cubiertos por un dominio.
- Definir reglas explícitas de routing por dominio y un formato de handoff entre `orchestrator` y agentes especializados.
- Adaptar los nuevos agentes a las reglas del repo: español para documentación humana, sin delivery, sin validación inventada, respeto de `AGENTS.md`, OpenSpec y `.opencode/policies/delivery.md`.
- No remover `implementer` en esta etapa; evaluar su remoción o renombre después de observar el uso del nuevo routing.

## Capabilities

### New Capabilities
- `domain-agent-routing`: routing de implementación desde `orchestrator` hacia agentes especializados por dominio con fallback controlado a `implementer`.

### Modified Capabilities

- Ninguna spec activa existente cambia su comportamiento de producto; esta proposal afecta el workflow interno de agentes.

## Impact

- `.opencode/agents/orchestrator.md`: permisos y reglas de routing hacia nuevos agentes.
- `.opencode/agents/`: nuevas definiciones para agentes de dominio.
- `AGENTS.md`: posible actualización de guía operativa si se decide documentar el nuevo flujo globalmente.
- `.opencode/commands/` o skills OpenSpec: posible ajuste documental si algún comando menciona directamente `implementer` como único ejecutor.
- Riesgo principal: routing demasiado agresivo o especialistas con supuestos externos no adaptados al repo.
