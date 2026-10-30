# Changelog

Los cambios relevantes del proyecto se documentarán en este archivo.

El formato recomendado sigue categorías simples:

- `Added` para funcionalidades nuevas.
- `Changed` para cambios de comportamiento existente.
- `Fixed` para correcciones de bugs.
- `Security` para cambios de seguridad o supply chain.

Las releases públicas deben enlazar su tag y resumir validación relevante.

## Unreleased

### Changed

- Manual HTML de uso con ocho recorridos ilustrativos, conversaciones, diagramas, skills y visor de delegación, usando los recursos visuales originales de Lufy.

- Flujo Codex inline integrado: el principal adapta trabajo y evidencia al resultado; delegación configurable `off`, `on_request` o `auto`.
- Catálogo de 16 assets y cinco skills. Configuración schema 2, manifest 4 y backups 2 ligados a estado 4; formatos incompatibles sin migrador genérico.
- Contexto lexical v3, memoria portable y registros explícitos opcionales; el hook sólo orienta `SessionStart`.
- Documentación reorganizada por uso, configuración, comandos, recuperación y publicación, con navegación histórica y defaults contrastados con `develop`.

### Removed

- Selección de metodología por tiers, motores SDD/OpenSpec, roles instalables obligatorios y Result Contract.
- Adapters OpenCode/Claude Code, parsers de artifacts y lectores de formatos anteriores.
- Guías duplicadas o vigentes que todavía dependían de roles, envelopes o integración pendiente.

### Fixed

- Identidad de paths nativos y con slash en manifest/verificación para Windows.
- Clasificador de ensayos `command-effects/v2`: argumentos literales, operadores acotados, diagnósticos y procedencia de helpers. La reevaluación offline conserva veredictos históricos y comandos desconocidos como inconclusos.

## [v0.6.25] - 2026-10-23

### Added

- Diagnóstico observable del runtime Codex, eventos semánticos explícitos de review y evidencia ligada al candidato Git.
- Regresiones offline de Codex obligatorias en CI, protocolo nativo opt-in y perfil collaboration con facetas separadas de transporte, rol y cierre.
- Procedencia reproducible de experimentos de eficiencia; se conserva el baseline inline porque la variante por referencia no completó una comparación válida.

### Changed

- Renovado el banner principal para representar el ciclo gobernado y distribuido del harness actual.
- Consolidada la documentación pública, con una guía sistémica del workflow y estado/roadmap vivos.
- Retirados los snapshots históricos `docs/backlog.md` y `docs/implementation-plan.md`; la historia entregada permanece en CHANGELOG, ADRs y archives SDD/OpenSpec.
- Unificadas las instrucciones proporcionales de Codex y sus mirrors con controles ejecutables de paridad.

### Fixed

- Corregidos cruces de reviews y duraciones ambiguas en las métricas operativas.
- El binding de evidencia evita efectos laterales de filtros, gitlinks y lazy fetch durante la inspección del candidato.
- Un timeout explícito de wait_agent sólo permite acreditar transporte si una inspección posterior del mismo agente confirma el contrato esperado; errores y respuestas desconocidas conservan aceptación cerrada.

Validación: 131 pruebas offline, cobertura Go 80,5%, vet/build, 36 specs estrictas y checks multiplataforma de promoción. La integración nativa completa sigue inconclusa por cierre no disponible; los replays offline no se presentan como nuevas corridas live ni como mejoras de rendimiento.

## [v0.6.24] - 2026-09-28

### Added

- Surface Execution Plan para detectar superficies frontend, backend y fullstack, componer contratos conectados y seleccionar reglas de validación desde el proyecto.
- Run Ledger append-only con idempotencia, control de versión, reparación auditable y Result Contract ejecutable para registrar transiciones del Loop Engine sin avanzar gates por texto libre.
- Context Graph local y Review Workload Harness para recuperar contexto estructural, dimensionar slices y mantener la carga de revisión dentro de límites explícitos.
- Asignación adaptativa `disabled`/`shadow`/`advisory` con scoring determinista, capacidad y budget acotados, leases con fencing y protocolo de yield durable.

### Changed

- Lufy SDD Full y Lite cubren el ciclo completo de explore, propose, apply, verify, sync y archive, con routing proporcional T1/T2/T3 y validación estricta por change.
- Harnesses OpenCode y Codex mantienen paridad de agentes, skills, contratos, assets gestionados y Result Contract v1.
- La orquestación integra Surface Execution, Context Graph, workload de review, Run Ledger y asignación adaptativa sin transferir autoridad de delivery ni avanzar gates automáticamente.

### Fixed

- Run Ledger reintenta de forma acotada la contención transitoria de reemplazos de archivos en Windows y conserva la semántica CAS/idempotente.
- La confirmación adaptativa revalida recommendation vigente, capacidad global y budget del actor dentro del mismo CAS.
- La recuperación de leases vencidas exige un checkpoint durable y fenced; no libera recursos implícitamente por reloj.
- Gates de coupling, paridad root/embedded, privacidad, race, stress, artifacts multi-OS y CI remota quedaron verdes para la promoción.

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
