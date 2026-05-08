## Why

El flujo actual ya recomienda validacion agrupada, pero la regla no explicita cuando analizar archivos existentes, cuando releerlos ni cuando ejecutar tests y coverage. Esto provoca relecturas y ejecuciones repetidas durante una propuesta, aumentando costo, ruido y riesgo de decisiones locales sin vision del sistema completo.

## What Changes

- Introducir una capacidad de workflow sistemico para propuestas OpenSpec y agentes.
- Definir que el analisis de codigo existente, archivos viejos, dependencias e interconexiones ocurre al inicio del bloque/propuesta.
- Definir que archivos viejos solo se releen al final si fueron modificados, si hay conflicto, bloqueo, nueva evidencia o riesgo explicito.
- Definir que tests, coverage y validacion completa se ejecutan al final de todas las tareas de la propuesta, salvo excepciones justificadas por bloqueo, cambio riesgoso o diagnostico.
- Incorporar criterios de pensamiento sistemico: interconexiones y dependencias, pensamiento holistico, bucles de retroalimentacion, y relacion entre estructura estatica y comportamiento dinamico.

## Capabilities

### New Capabilities
- `systemic-workflow`: Reglas operativas para analisis inicial, implementacion por bloques, relectura final acotada y validacion final agrupada en propuestas OpenSpec.

### Modified Capabilities
- `current-state-documentation`: Documentar el workflow operativo vigente para agentes, validacion y OpenSpec.

## Impact

- `.opencode/policies/delivery.md`
- `.opencode/agents/orchestrator.md`
- `.opencode/agents/explorer.md`
- `.opencode/agents/implementer.md`
- `.opencode/agents/validator.md`
- `.opencode/skills/sdd-workflow/openspec-apply-change/SKILL.md`
- `.opencode/skills/sdd-workflow/openspec-verify-change/SKILL.md`
- `AGENTS.md`
- Documentacion OpenSpec relevante
