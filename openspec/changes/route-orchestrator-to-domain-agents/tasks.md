## 1. Agentes de dominio

- [ ] 1.1 Crear `.opencode/agents/frontend-developer.md` adaptado a reglas locales, sin dependencia de `context-manager` externo.
- [ ] 1.2 Crear `.opencode/agents/backend-developer.md` adaptado a reglas locales, con límites claros de delivery y validación.
- [ ] 1.3 Crear `.opencode/agents/mobile-developer.md` adaptado a reglas locales y sin asumir toolchains móviles disponibles salvo evidencia.
- [ ] 1.4 Crear `.opencode/agents/microservices-architect.md` como agente de diseño/handoff para arquitectura distribuida, no como sustituto directo de delivery.
- [ ] 1.5 Definir permisos de cada agente para permitir edición cuando corresponde y negar commit/push/PR/project sync.

## 2. Routing en orchestrator

- [ ] 2.1 Actualizar permisos `task` de `.opencode/agents/orchestrator.md` para permitir los nuevos agentes.
- [ ] 2.2 Añadir reglas de routing para `frontend-developer`, `backend-developer`, `mobile-developer` y `microservices-architect`.
- [ ] 2.3 Mantener `implementer` como fallback explícito para tooling, CI, scripts, OpenSpec, docs, configuración, agentes, installer local y cambios pequeños.
- [ ] 2.4 Documentar que `orchestrator` decide el routing y que `implementer` no delega internamente.
- [ ] 2.5 Mantener `validator`, `reviewer` y `delivery` como gates separados posteriores.

## 3. Handoff y boundaries

- [ ] 3.1 Definir formato estándar de handoff desde `orchestrator` a agentes de dominio: objetivo, alcance, archivos/specs relevantes, restricciones y validación esperada.
- [ ] 3.2 Definir formato de salida de agentes de dominio: cambios aplicados, evidencia, riesgos y ready state.
- [ ] 3.3 Añadir regla de `blocked` cuando un agente detecte que la tarea no pertenece a su dominio.
- [ ] 3.4 Añadir regla para usar `explorer` primero cuando el dominio o alcance sea ambiguo.

## 4. Documentación operativa

- [ ] 4.1 Actualizar `AGENTS.md` si se decide que el routing de dominio sea parte de la guía raíz.
- [ ] 4.2 Revisar `.opencode/commands/` y skills OpenSpec para referencias que asuman `implementer` como único ejecutor.
- [ ] 4.3 Documentar que `implementer` no se remueve en esta etapa y que su uso se evaluará después.

## 5. Validación

- [ ] 5.1 Ejecutar inspección estática de los archivos de agentes para confirmar permisos y nombres correctos.
- [ ] 5.2 Ejecutar `openspec status --change route-orchestrator-to-domain-agents --json`.
- [ ] 5.3 Ejecutar `git diff --check`.
- [ ] 5.4 Revisar que no se introduzcan comandos de validación inventados ni supuestos de toolchain inexistente.
