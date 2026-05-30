## 1. Análisis y propuesta

- [x] Crear rama de trabajo para la propuesta.
- [x] Revisar estado actual de arquitectura, specs y acoplamientos principales.
- [x] Crear proposal/design/tasks para el refactor.
- [x] Crear specs delta para arquitectura de adapters, routing por tier, render de instrucciones y registry de skills.

## 2. Auditoría textual de assets operativos

- [x] Inventariar referencias a OpenCode, `.opencode`, `opencode.json`, OpenSpec, `openspec/` y `/opsx-*` en agentes, subagentes, skills, commands, templates, policies y docs.
- [x] Clasificar cada referencia como `core`, `tool-binding`, `methodology-binding` o `legacy`.
- [x] Definir checks automáticos de fugas para roles neutrales y adapters futuros.

## 3. Modelo neutral de roles

- [x] Definir contrato neutral para `orchestrator`, `router`, `delivery`, `explorer`, `implementer`, `test-writer`, `validator` y `reviewer`.
- [x] Separar responsabilidades principales/subagentes de los detalles de OpenCode y OpenSpec.
- [x] Definir fallback inline para tools sin subagentes.

## 4. Instruction renderer

- [ ] Diseñar estructura de templates/bindings para role core, tool binding y methodology binding.
- [ ] Renderizar assets OpenCode/OpenSpec equivalentes a los actuales.
- [ ] Agregar golden tests de salida renderizada.
- [ ] Validar que el output inicial no cambie comportamiento de `lufy-ai install`.

## 5. Adapter registry

- [ ] Introducir `ToolAdapter` y `ToolCapabilities`.
- [ ] Mover OpenCode a adapter real inicial.
- [ ] Introducir `MethodologyAdapter`.
- [ ] Mover OpenSpec a methodology adapter real inicial.
- [ ] Implementar methodology `none`.

## 6. Routing por tier y configuración

- [ ] Agregar configuración `methodology_by_tier` con defaults compatibles.
- [ ] Propagar tool/metodología/mode al Result Contract.
- [ ] Bloquear o justificar overrides inseguros de `none` en T1/T2.

## 7. Manifest, sync y verify

- [ ] Diseñar manifest v2 compatible con v1.
- [ ] Registrar `tool`, `methodology`, `component` y `scope` por asset.
- [ ] Actualizar `verify`, `status` y `sync` para detectar assets por adapter.
- [ ] Mantener lectura de instalaciones v1 sin romper.

## 8. Validación y documentación

- [ ] Actualizar README, architecture, installation, getting-started y backlog.
- [ ] Actualizar assets embebidos si cambia cualquier asset instalable.
- [ ] Ejecutar `scripts/validate.sh`.
- [ ] Ejecutar validación OpenSpec estricta del cambio cuando el CLI esté disponible.
- [ ] Reportar evidencia real, riesgos y estado del programa.
