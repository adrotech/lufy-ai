## Why

Lufy AI necesita pasar de un flujo agéntico correcto a un harness SDD proporcional: un entorno que decida cuánto contexto, proceso, permisos y agentes hacen falta antes de ejecutar trabajo real. Hoy el orquestador puede terminar cargando demasiado contexto o activando un flujo más pesado de lo necesario para tareas simples.

Este cambio introduce un router liviano que clasifica pedidos por tier y elige el camino mínimo seguro, aplicando el principio: no matar una mosca con una bazuca.

## What Changes

- Agregar una capability de routing SDD basada en harness engineering.
- Definir T1, T2 y T3 como clasificación de propuestas, funcionalidades y tareas, no de áreas técnicas del repositorio.
- Introducir el subagente liviano `sdd-router` para clasificar el pedido antes de activar subagentes con más contexto o permisos.
- Definir contratos estructurados de entrada y salida entre `orchestrator`, `sdd-router` y subagentes, incluyendo execution mode, result contract y contexto mínimo.
- Establecer SDD Lite para T2 como flujo compacto, profesional y proporcional.
- Establecer reglas de permisos mínimos y contexto mínimo por tier.
- Registrar resolución de skills y permitir bootstrap opcional con AutoSkills cuando no existan skills locales suficientes.
- Definir aislamiento de subagentes, artifact store mínimo y tamaño proporcional de revisión según tier.
- Incorporar Review Workload Harness: diseñar propuestas/features pensando en la carga del reviewer humano, separando funcionalidades grandes en slices revisables y entregables pequeños cuando reduzcan riesgo.
- Mantener delivery separado y sujeto a autorización explícita.
- Actualizar la documentación pública y operativa para reflejar que la herramienta evolucionó hacia un harness SDD proporcional.
- Actualizar los assets embebidos de la CLI Go para que nuevas instalaciones reciban `sdd-router`, templates T2/result y la documentación/políticas vigentes.
- Generar un nuevo binario local del installer con los assets embebidos actualizados para validación, sin publicar release ni cambiar tags.
- Revisar la separación entre `.opencode/policies/delivery.md` y `.opencode/agents/delivery.md`, manteniendo la política compartida como fuente canónica y el agente como runbook operativo.

## Capabilities

### New Capabilities

- `sdd-harness-routing`: clasificación T1/T2/T3, routing proporcional, contratos entre agentes, execution modes, result contracts, permisos mínimos, contexto mínimo, SDD Lite, review workload, delivery slices y skill bootstrap opcional.

### Modified Capabilities

- Ninguna.

## Impact

- Afecta la metodología de trabajo de `.opencode/agents/orchestrator.md` y potencialmente agrega `.opencode/agents/sdd-router.md`.
- Puede requerir ajustes en `.opencode/agents/explorer.md` para aceptar handoffs más estructurados.
- Puede requerir documentación nueva o actualización de `AGENTS.md`, `.opencode/README.md` y `.opencode/policies/delivery.md` para describir el flujo proporcional.
- Puede requerir documentación de templates o convenciones para T2 Lite, result envelopes y resolución de skills.
- Afecta docs públicas (`README.md`, `docs/*`, `tools/lufy-cli-go/README.md`, `openspec/README.md`) para evitar drift entre estado real y roadmap.
- Afecta assets embebidos bajo `tools/lufy-cli-go/internal/assets/embedded/` y el catálogo instalable.
- No cambia la CLI Go, el instalador ni el modelo de releases.
- No introduce AutoSkills como dependencia obligatoria ni ejecuta instalación automática de skills externas.
- No publica un release estable; cualquier versión pública nueva sigue requiriendo promoción a `main`, tag `v*` y workflow de release autorizado.
