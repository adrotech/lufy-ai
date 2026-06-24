## Why

El harness ya clasifica trabajo T1/T2 y recomienda slices de revision, pero no tiene un protocolo explicito para deliberar entre alternativas de artifacts cuando hay alta incertidumbre o varios ejes de riesgo. Esto provoca que propuestas complejas dependan de una sola formulacion inicial o de discusiones ad hoc, sin criterios claros para ramificar, comparar y colapsar antes de implementar.

## What Changes

- Introducir un protocolo MVP de multi-artifact branching para generar hasta 2 candidates de `proposal` cuando `sdd-router` detecte T1 o T2 multi-risk con alta incertidumbre.
- Permitir branching opcional de `design` solo cuando, tras seleccionar o mergear una proposal, sigan existiendo decisiones tecnicas sustantivas.
- Mantener branching de `tasks` como excepcional y explicito, solo por riesgo real de estrategia de implementacion.
- Requerir un join/merge obligatorio antes de pasar a implementation-ready, con un unico artifact set canonico como fuente de verdad.
- Usar roles existentes: `sdd-router` recomienda, `orchestrator` coordina candidates y join, `implementer`/solution writer redacta alternativas, `reviewer` compara calidad/riesgo y el humano decide trade-offs no objetivos.
- Respetar `workflow_limits` y `parallel_execution`: candidates paralelos solo si son artifacts independientes, cada candidate escribe paths aislados, hay merge plan y la validacion se agrupa despues del join.
- Definir no-goals del MVP: no crear nuevos agentes/roles, no paralelizar delivery, no paralelizar decisiones de seguridad/contratos publicos no resueltas, no implementar runtime ni assets de agentes en esta propuesta.

## Capabilities

### New Capabilities
- `multi-artifact-branching-deliberation`: protocolo para ramificar, comparar, seleccionar/mergear y colapsar candidates de artifacts OpenSpec antes de implementacion.

### Modified Capabilities
- `sdd-harness-routing`: ampliar el contrato de routing para recomendar `artifact_branching` proporcionalmente y preservar `workflow_limits`/`parallel_execution`.
- `systemic-workflow`: incorporar estados sistemicos de deliberacion y join para que el comportamiento emergente del workflow sea trazable antes de implementacion.
- `harness-adapter-architecture`: aclarar que el comportamiento de branching/join es adapter-neutral y que cada adapter solo renderiza/coordina sus artifacts compatibles.

## Impact

- Afecta contratos y documentacion OpenSpec del harness; no cambia runtime, CLI, agentes raiz, skills instaladas ni delivery en esta propuesta.
- El cambio posterior de implementacion debera actualizar instrucciones de roles existentes y cualquier renderer/asset gestionado aplicable, manteniendo un unico plan canonico antes de `/opsx-apply` o flujo equivalente.
- La validacion esperada para esta propuesta es documental/OpenSpec: markers core v2, scenarios WHEN/THEN y `openspec validate "add-multi-artifact-branching-deliberation" --strict`.
