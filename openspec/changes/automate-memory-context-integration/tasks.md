## 1. Config y Context Graph

- [x] Agregar `context_graph.exclude` al modelo de proyecto.
- [x] Definir defaults para omitir `.lufy/managed-state/backups/**` y `.lufy/managed-state/ancestors/**`.
- [x] Aplicar exclusiones en discovery/status hash del context graph.

## 2. Lifecycle OpenCode

- [x] Agregar plugin local OpenCode para orientación de memoria al crear sesión.
- [x] Validar memoria best-effort cuando se editen archivos bajo `.lufy/memory/`.
- [x] Mantener scripts de memoria no bloqueantes cuando la CLI/config no existe.

## 3. Diagnóstico

- [x] Extender `doctor` con estado de memoria, contexto y hooks/plugin.
- [x] Extender `verify --deep` con checks de contexto y lifecycle hooks.
- [x] Mostrar recovery `lufy-ai context build` para grafo stale/not_available.

## 4. Documentación y Validación

- [x] Documentar hooks/plugin y exclusiones.
- [x] Ejecutar tests Go focalizados.
- [x] Ejecutar validación agrupada disponible.
