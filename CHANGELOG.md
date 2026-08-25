# Changelog

Los cambios relevantes del proyecto se documentarán en este archivo.

El formato recomendado sigue categorías simples:

- `Added` para funcionalidades nuevas.
- `Changed` para cambios de comportamiento existente.
- `Fixed` para correcciones de bugs.
- `Security` para cambios de seguridad o supply chain.

Las releases públicas deben enlazar su tag y resumir validación relevante.

## Unreleased

## [v0.6.23] - 2026-08-25

### Added

- Lufy SDD Full y Lite como metodología nativa instalable, con templates más completos para objetivos del LLM, restricciones, entorno, arquitectura, datos, patrones y diagramas.
- Comandos y skills `lufy.sdd-*` para explorar, proponer, aplicar, verificar, sincronizar y archivar cambios Lufy SDD desde el harness.

### Changed

- Assets fuente y embebidos sincronizados para distribuir Lufy SDD Full/Lite desde el CLI Go.
- Specs OpenSpec de routing actualizadas para exigir mayor profundidad metodológica, criterios verificables y guía explícita de validación.

### Fixed

- `pr guard` permite archives OpenSpec versionables y sigue bloqueando changes activos internos en PRs normales.

## [v0.6.22] - 2026-08-17

### Added

- `lufy-ai memory capture/connect/index` para persistir decisiones, correcciones del usuario y backlinks Obsidian con validación determinística.
- Registry portable `.lufy/skill-registry.json` con precedencia project-over-global, raíces por adapter y paths exactos a los `SKILL.md` fuente.
- Comandos `lufy-ai skills ensure/refresh/status` para mantener y diagnosticar el índice sin modificar skills user-owned.
- Setter y pruebas aisladas para mantener alineadas la versión canónica y las referencias copiables.
- Verificación de la versión embebida en artifacts construidos y publicados.

### Changed

- Agentes y skills de memoria ahora tratan correcciones explícitas del usuario como memoria durable (`rule`/`lesson`) y usan el CLI para capturar/conectar notas.
- El routing prioriza memoria Obsidian y el grafo de contexto configurados, exige fallbacks explícitos y devuelve hints compactos en lugar de dumps amplios.
- `install`, `setup` y `sync` aseguran el skill registry en modo best-effort; OpenCode repite el ensure al crear una sesión y Codex conserva un fallback explícito de lifecycle/primer uso.
- `doctor` y `verify --deep` incorporan el estado read-only del skill registry con recuperación accionable.
- Los workflows de tag y release ahora exigen consistencia entre `RELEASE_VERSION`, tag calculado y changelog.
- `actions/setup-go` usa el `go.sum` del módulo anidado para resolver el cache sin warnings.

### Fixed

- Comparaciones de paths del registry compatibles con repositorios movidos y normalización de rutas en macOS/Windows.
- Paridad entre assets fuente y embebidos para hooks, skills, agentes y documentación operativa.
- El gate documental ya no queda verde cuando falta la versión canónica o el source tree está atrasado respecto del release planificado.

## [v0.6.11] - 2026-06-09

### Added

- Memoria Obsidian portable instalable con `lufy-ai memory init/status/validate/search`.
- Defaults `memory` y `parallel_execution` en `.lufy/config/project.yaml`, preservados por `init --rescan` y `scan`.
- Assets instalables de memoria: comandos `/lufy.mem-*`, skills `lufy.mem-*`, hooks `memory-*` y template `memory-note.md`.
- Paralelismo gobernado por `sdd-router` para `review_slices` independientes con plan de merge y validación agrupada.

### Changed

- `doctor` reporta estado de memoria sin bloquear instalación normal.
- `verify --deep` valida memoria Obsidian cuando existe.
- `sync` mantiene los assets de memoria gestionados, pero preserva `.lufy/memory` como contenido privado user-owned.
- Agentes y guías operativas usan Obsidian como memoria canónica portable.

### Fixed

- Tests de memoria compatibles con salida de `rg` en rutas Windows.
- Smoke POSIX de merge omitido correctamente en Windows.

## [v0.4.0] - 2026-05-27

### Added

- Fast path OpenSpec/docs-only para micro-slices de planificación de 1-2 artefactos, con `program_tier`, `slice_tier` y `fast_path_allowed` en Result Contract.
- Documentación de instalación y quickstart actualizada para la versión estable `v0.4.0`.

### Changed

- `orchestrator` debe sintetizar Result Contracts de subagentes en respuestas humanas en español, reservando YAML crudo para handoffs o solicitudes explícitas.
- Assets embebidos del CLI Go sincronizados con las reglas locales del harness OpenCode/OpenSpec.

### Fixed

- Dirty worktree pasa a tratarse como riesgo de delivery, no como bloqueo de validación para slices OpenSpec/docs-only sin runtime ni delivery.
