## 1. Politica central

- [x] 1.1 Actualizar `AGENTS.md` con la regla de pensamiento sistemico, fases de analisis inicial, implementacion sin relecturas repetidas, revision final acotada y validacion final agrupada.
- [x] 1.2 Actualizar `.opencode/policies/delivery.md` para que los tiers de validacion incorporen tests/coverage al final de la propuesta y excepciones tempranas justificadas.

## 2. Agentes

- [x] 2.1 Actualizar `.opencode/agents/orchestrator.md` para coordinar fases sistemicas y evitar duplicacion entre explorer, implementer y validator.
- [x] 2.2 Actualizar `.opencode/agents/explorer.md` para exigir analisis inicial de interconexiones, dependencias, riesgos y comportamiento esperado.
- [x] 2.3 Actualizar `.opencode/agents/implementer.md` para evitar relecturas repetidas de archivos viejos y reservar la relectura final para archivos modificados, conflictos, bloqueos o nueva evidencia.
- [x] 2.4 Actualizar `.opencode/agents/validator.md` para ejecutar tests, coverage y validacion completa al final de la propuesta cuando existan, manteniendo excepciones para diagnostico/riesgo.

## 3. Skills OpenSpec

- [x] 3.1 Actualizar `.opencode/skills/sdd-workflow/openspec-apply-change/SKILL.md` para aplicar tareas con validacion agrupada final y relectura acotada.
- [x] 3.2 Actualizar `.opencode/skills/sdd-workflow/openspec-verify-change/SKILL.md` para verificar pensamiento sistemico, relectura final y evidencia de tests/coverage final cuando aplique.

## 4. Documentacion

- [x] 4.1 Actualizar documentacion OpenSpec relevante para reflejar que tests y coverage se corren al final de todas las tareas de la propuesta.
- [x] 4.2 Revisar coherencia entre politica, agentes, skills y docs sin introducir comandos de validacion inexistentes.

## 5. Validacion final

- [x] 5.1 Ejecutar validacion estatica/documental disponible para los artefactos modificados.
- [x] 5.2 Ejecutar tests/coverage aplicables al final si existe toolchain real para el alcance; si no existe, reportar la limitacion explicitamente.
