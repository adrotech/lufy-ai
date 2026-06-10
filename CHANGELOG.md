# Changelog

Los cambios relevantes del proyecto se documentarán en este archivo.

El formato recomendado sigue categorías simples:

- `Added` para funcionalidades nuevas.
- `Changed` para cambios de comportamiento existente.
- `Fixed` para correcciones de bugs.
- `Security` para cambios de seguridad o supply chain.

Las releases públicas deben enlazar su tag y resumir validación relevante.

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
