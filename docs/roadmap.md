# Roadmap de hardening y evolución de lufy-ai

Este roadmap captura trabajo futuro taggeable derivado del análisis comparativo con `Gentleman-Programming/gentle-ai`. El alcance real de `lufy-ai` sigue siendo un kit OpenCode/OpenSpec instalable en repositorios destino. La CLI Go actual existe para instalar, verificar y sincronizar assets gestionados del kit; no convierte el repo en un framework de aplicación ni en un producto multiagente separado.

## Visión

Convertir `lufy-ai` en una capa operativa instalable, segura e idempotente para proyectos existentes, con validación básica automatizada y capacidad de reaplicar assets gestionados sin destruir configuración local.

La evolución debe priorizar primero la seguridad del instalador y la confianza del usuario antes de expandir templates, automatizaciones o empaquetado como producto.

## Principios de diseño

- **Instalación segura por defecto**: ningún archivo crítico debe sobrescribirse sin backup, estrategia explícita o confirmación.
- **Idempotencia verificable**: ejecutar el instalador o sync más de una vez debe producir un estado estable y auditable.
- **Compatibilidad con repositorios existentes**: respetar `.opencode/`, `AGENTS.md`, `tui.json`, `opencode.json` y `openspec/` cuando ya existan.
- **Flags predecibles para automatización**: `--dry-run`, `--yes`, `--no-engram` y `--target` deben funcionar sin prompts inesperados.
- **Detección portable**: usar `command -v engram` para resolver Engram, evitando paths hardcodeados como `/opt/homebrew/bin/engram`.
- **Validación simple antes de crecer**: ejecutar checks básicos de shell cuando estén disponibles, JSON e instalación en temp dir con `lufy-ai verify` antes de introducir features mayores.
- **Documentación honesta**: documentar solo templates y assets que realmente se instalan.
- **Evolución incremental**: productizar como CLI queda para una fase futura, no como prioridad inmediata.
- **Migración compatible de Bash a Go**: Bash queda como wrapper estricto de compatibilidad que delega en `lufy-ai install`, sin fallback legacy.
- **Release estable desde producción**: integrar trabajo en `develop`, promover a `main` y publicar releases estables solo desde tags `v*` sobre commits alcanzables desde `main`.

## Estado instalable vs roadmap

Esta página contiene ideas futuras y decisiones estratégicas. Salvo que una sección indique explícitamente que algo ya existe en la rama actual, no debe leerse como contrato instalable.

Para el backlog priorizado de mejoras detectadas por análisis externo, ver [`docs/backlog.md`](backlog.md). Ese backlog agrupa oportunidades por olas y sirve como entrada para futuras proposals OpenSpec.

Para política de firma, provenance, SBOM y labels de release, ver [`docs/release-security.md`](release-security.md).

Para gates de coverage, lint, shellcheck, matriz multi-OS y E2E post-release, ver [`docs/ci-quality-gates.md`](ci-quality-gates.md).

Estado actual documentable:

- CLI Go en `tools/lufy-cli-go/` con `install`, `verify`, `backup`, `restore` y `sync`.
- Assets gestionados con estado `.lufy-ai/install-state.json`, hashes SHA-256, idempotencia y backups antes de actualizaciones gestionadas.
- Drift Resolution en rama: policies declarativas, `AGENTS.md` como `merge-block`, `tui.json` como `no-replace`, `.lufy-new`, ancestors, `--scope=project|global|both`, `lufy-ai merge <path>` y restore con `--list`/ID.
- `opencode.json` se maneja como configuración `merge-json`: se crea/mergea de forma conservadora, preserva claves desconocidas, falla ante JSON inválido y no se registra como asset completo por hash.
- Wrapper `scripts/install.sh` estricto, sin fallback legacy ni detección de stack en Bash.
- Harness SDD proporcional instalable: `sdd-router`, T1/T2/T3, SDD Lite, result contracts, context slicing, review workload y skill resolution local-first.
- Review Workload Harness instalable: slices revisables para features/propuestas grandes, sin forzar micro-entregables en T3.
- Templates de proceso instalables: `.opencode/templates/sdd-lite.md` y `.opencode/templates/result-contract.md`.
- Workflow mínimo `.github/workflows/go-cli-install.yml` presente en esta rama para tests/build/smokes de la CLI Go y `git diff --check`; su existencia no implica archive automático de proposals OpenSpec.
- Distribución versionada implementada en la rama: `lufy-ai version`, artifacts release por OS/arch, checksums SHA-256, bootstrap `scripts/bootstrap.sh` y assets embebidos para instalar sin checkout fuente. Las releases públicas instalables dependen de publicar tags `v*` desde commits alcanzables desde `main` y sus artifacts en GitHub Releases; antes de que exista un tag publicado, el bootstrap solo funciona contra fixtures/local mirrors o fallará al intentar descargar la release inexistente.
- README, `docs/getting-started.md` y README de la CLI ya describen el flujo sin clone con pinning/inspección, sujeto a que exista la release pública taggeada correspondiente.

Flujo operativo de ramas/release:

- `develop`: base normal de PRs y rama de integración diaria.
- `main`: rama productiva/estable, actualizada por promoción `develop` → `main` o hotfix explícito.
- `v*`: tags de release estable creados después de la promoción, sobre commits alcanzables desde `origin/main`.
- Settings remotos esperados: default branch `develop` y protecciones para `develop`/`main`; ver [`docs/github-branch-settings.md`](github-branch-settings.md).

No son capacidades instalables actuales:

- templates por stack como `frontend-react`, `frontend-nextjs`, `frontend-astro`, `mobile-expo` o `backend-spring`;
- detección automática de stack integrada en la CLI;
- subagentes especializados adicionales como `infra-cloud-sre`, `react-ui`, `nextjs-app-router` o `astro-islands-content`.
- instalación automática de skills externas; AutoSkills solo queda como bootstrap opcional con dry-run y autorización explícita.

Esos elementos se conservan abajo como roadmap para futuras iteraciones y solo deberían moverse al README cuando existan como assets reales, estén instalados por la CLI y tengan validación local/CI coherente.

## Prioridad estrategica 2026-05

Los RFCs externos revisados el 2026-05-11 cambian el foco inmediato del roadmap: antes de sumar mas superficie OpenSpec, el instalador debe resolver upgrades con drift sin bloquear ni destruir trabajo local. Despues de eso, el proyecto puede modernizar su flujo OpenSpec hacia paridad v1.3.1 y una arquitectura stay-updated.

RFCs incorporados como input de roadmap:

| RFC | Objetivo | Target sugerido | Orden |
| --- | --- | --- | --- |
| `LUFY-AI-DRIFT-RESOLUTION-RFC.md` | Resolver drift/upgrades con policies declarativas, `.lufy-new`, merge/restore y scope global/proyecto | `v0.2.0` | Primero |
| `LUFY-AI-OPENSPEC-V2-RFC.md` | Modernizar OpenSpec a paridad v1.3.1 con deltas, scenarios, sync, profiles y fallback 3 capas | `v0.3.0` | Segundo |
| `LUFY-AI-ROADMAP-RFC.md` | Coordinar releases, dependencias, testing sandbox y comunicacion | `v0.2.0` -> `v0.3.0` | Guia operativa |

Decisiones de roadmap:

- Implementar Drift Resolution antes de OpenSpec v2; no hay dependencia tecnica dura, pero si una dependencia operativa de UX y menor riesgo.
- Mantener PRs normales contra `develop` y releases estables desde `main` con tags `v*` alcanzables desde `origin/main`.
- Tratar cada tier/sprint como entregable independiente y mergeable; no mezclar Drift Resolution y OpenSpec v2 en una sola rama.
- No documentar `v0.2.0` o `v0.3.0` como disponibles hasta que existan tags y artifacts publicados.
- Mantener Go stdlib-only para la CLI salvo decision explicita posterior.

Plan de implementacion detallado: ver [`docs/implementation-plan.md`](implementation-plan.md).

## Prioridades por fases historicas

Las fases siguientes se conservan como trazabilidad del roadmap original. Varias ya estan implementadas o archivadas en OpenSpec; las nuevas prioridades ejecutables son las de la seccion anterior y el plan de implementacion.

### Fase 1 — Hardening del installer y seguridad de instalación

Objetivo: hacer que `scripts/install.sh` sea seguro, auditable e idempotente para uso local y CI.

- `RM-001`: Añadir flags `--dry-run`, `--yes`, `--no-engram` y `--target` robusto.
- `RM-002`: Implementar backup/rollback mínimo antes de tocar `.opencode/`, `AGENTS.md`, `tui.json`, `opencode.json` y `openspec/`.
- `RM-003`: Resolver Engram con `command -v engram` y escribir ese path en `opencode.json` solo cuando aplique.
- `RM-004`: Formalizar `lufy-ai verify` como verificador canónico para validar estructura instalada, manifest, hashes SHA-256 y JSON parseable.
- `RM-005`: Mejorar merge/idempotencia para no sobrescribir configs existentes; detectar bloques gestionados o pedir estrategia.

### Fase 2 — Validación continua mínima

Objetivo: asegurar que el kit instalable no se rompa por cambios documentales, scripts o JSON.

- `RM-006`: Añadir GitHub Actions básica para `shellcheck scripts/install.sh`, validación de JSON y test de instalación en temp dir.
- `RM-007`: Integrar `lufy-ai verify` en el workflow y documentar cómo ejecutarlo localmente.

### Fase 3 — Sync seguro de assets gestionados

Objetivo: permitir reaplicar assets de `lufy-ai` sin convertir la instalación en una sobrescritura destructiva.

- `RM-008`: Añadir sync simple para reaplicar assets gestionados de forma segura.
- `RM-009`: Definir manifest o metadata de assets gestionados para distinguir archivos propios de personalizaciones locales.

### Fase 4 — Templates reales y documentación sincronizada

Objetivo: introducir templates por stack solo cuando existan como archivos instalables y estén validados.

- `RM-010`: Añadir templates reales por stack únicamente si se instalan realmente.
- `RM-011`: Separar roadmap del README moviendo futuro/plantillas a `docs/roadmap.md` y dejando README como estado real + quickstart.

### Fase 5 — Productización futura

Objetivo: evaluar empaquetado y DX avanzada cuando el instalador, sync y CI ya sean confiables.

- `RM-012`: Evaluar productización como CLI como fase futura, no prioridad inmediata.
- `RM-013`: Migrar installer a CLI Go y mantener Bash solo como wrapper estricto de compatibilidad, sin lógica legacy.

### Fase 6 — Distribución versionada sin clone

Objetivo: permitir instalación sin obligar a clonar este repositorio, usando releases binarios versionados, checksums y un bootstrap remoto seguro.

- `RM-014`: Publicar binarios `lufy-ai` versionados por OS/arch en GitHub Releases, con checksums SHA-256 y comando `lufy-ai version`.
- `RM-015`: Añadir bootstrap remoto inspeccionable que detecta OS/arch, permite pinning de versión, descarga la release correcta, verifica checksum antes de instalar y coloca el binario en un destino de `PATH` elegido por el usuario.
- `RM-016`: Resolver instalación standalone real llevando assets gestionados embebidos en el binario o mediante un release bundle versionado con assets; estrategia incremental recomendada: primero `go:embed`, bundle solo si el tamaño/frecuencia de assets lo justifica.
- `RM-017`: Actualizar README, `docs/getting-started.md` y README de la CLI solo al final de la implementación, sin presentar `curl | bash`, Homebrew, Scoop, `go install` ni releases como estado actual antes de existir.

## Propuesta/design por iniciativa

### `RM-001` — Flags robustos del installer

**Propuesta**: reemplazar el manejo posicional actual por parsing explícito de argumentos.

**Design**:

- Soportar `--target <dir>` y mantener compatibilidad con el argumento posicional actual cuando no haya flag.
- `--dry-run` debe listar acciones planeadas sin escribir archivos ni clonar/cambiar estado persistente salvo temporales limpiables.
- `--yes` debe aceptar prompts seguros, pero no debe ocultar conflictos destructivos sin backup.
- `--no-engram` debe saltar detección, prompts y cambios MCP de Engram.
- Flags desconocidos deben fallar con ayuda breve y código distinto de cero.

### `RM-002` — Backup/rollback mínimo

**Propuesta**: crear un backup por ejecución antes de cualquier escritura sobre assets sensibles.

**Design**:

- Directorio sugerido: `.lufy-ai/backups/<timestamp>/` dentro del target o temp externo configurable.
- Respaldar solo si existe el path destino y será modificado.
- Registrar un manifest con paths respaldados, acciones realizadas y timestamp.
- Si una acción falla, intentar rollback de paths tocados en esa ejecución y reportar el estado final.
- No borrar backups automáticamente en la primera iteración.

### `RM-003` — Engram portable

**Propuesta**: resolver Engram con `command -v engram` al momento de instalar.

**Design**:

- Si `--no-engram`, no consultar ni escribir integración Engram.
- Si Engram existe y el usuario acepta integración, escribir el path resuelto en `opencode.json`.
- Si no existe, dejar MCP deshabilitado sin path hardcodeado o con una recomendación explícita de instalación.
- Evitar depender de `/opt/homebrew/bin/engram`, porque rompe Linux, macOS no-Homebrew y entornos CI.

### `RM-004` — `lufy-ai verify` canónico

**Propuesta**: usar `lufy-ai verify` como verificador local canónico para confirmar que un target quedó instalable, sin crear un script Bash paralelo.

**Design**:

- Aceptar `--target <dir>` y validar desde fuera del proyecto destino.
- Confirmar presencia de `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/plugins`, `.opencode/policies`, `.opencode/plugins/agent-observatory.tsx`, `AGENTS.md`, `tui.json`, `openspec/config.yaml` y `.lufy-ai/install-state.json`.
- Validar JSON parseable desde la CLI Go cuando aplique.
- Validar que los archivos críticos gestionados estén registrados en manifest y que los hashes SHA-256 coincidan.
- Fallar con mensajes accionables y código distinto de cero.

### `RM-005` — Merge/idempotencia de configs

**Propuesta**: distinguir archivos gestionados de archivos del usuario y evitar sobrescrituras silenciosas.

**Design**:

- Para archivos completos gestionados, comparar checksums o manifest antes de reemplazar.
- Para `AGENTS.md`, usar bloque gestionado con marcadores si se decide mezclar contenido.
- Para `opencode.json`, usar estrategia especial `merge-json`: merge conservador, backup antes de escribir cuando exista, preservación de claves desconocidas y fallo accionable ante JSON inválido.
- Para otros JSON como `tui.json`, preferir estrategia explícita futura (`skip`, `backup-and-replace`, `merge`) antes de permitir sobrescrituras amplias.
- El instalador repetido debe reportar `unchanged`, `updated`, `skipped` o `conflict` por asset.

**Siguiente evolucion aprobada por RFC Drift Resolution:** reemplazar el bloqueo generico ante drift por policies declarativas por asset: `managed`, `no-replace`, `merge-block`, `merge-json` y `metadata`. `AGENTS.md` pasa a `merge-block`; configuraciones user-owned reciben nueva version como `.lufy-new`; JSON estructurado usa merge conservador.

### `RM-006` — GitHub Actions básica

**Propuesta**: crear un workflow CI mínimo enfocado en assets reales del repo.

**Design**:

- `shellcheck scripts/install.sh` cuando esté disponible y smoke con `lufy-ai verify`.
- Validar JSON de `tui.json`, `.opencode/package.json`, `.opencode/package-lock.json` y archivos JSON relevantes.
- Ejecutar instalación en un temp dir y luego `lufy-ai verify --target <temp> --no-engram`.
- No agregar pipelines de build/test de producto inexistente.

### `RM-007` — Verificación local documentada e integrada

**Propuesta**: convertir `lufy-ai verify` en la comprobación estándar de instalación local y CI.

**Design**:

- Documentar en README o `docs/getting-started.md` el comando local recomendado.
- Mantener la verificación en la CLI Go para evitar dependencias Bash/Python adicionales y divergencia de reglas.
- Reutilizar los mismos checks de `lufy-ai verify` en CI para evitar divergencia entre validación local y remota.
- Reportar claramente qué asset falta o qué JSON no parsea.

### `RM-008` — Sync seguro de assets gestionados

**Propuesta**: agregar un modo simple para reaplicar assets de `lufy-ai` luego de una instalación inicial.

**Design**:

- Puede iniciar como `scripts/install.sh --sync --target <dir>` o script separado si eso reduce complejidad.
- Debe reutilizar backup, manifest e idempotencia de Fase 1.
- Debe limitarse a assets gestionados: `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/policies`, `.opencode/plugins` y templates reales.
- No debe reescribir convenciones locales sin estrategia explícita.

### `RM-009` — Manifest de assets gestionados

**Propuesta**: registrar qué archivos pertenecen a `lufy-ai` para habilitar sync e idempotencia sin adivinar.

**Design**:

- Crear un manifest versionado con path, tipo de asset, estrategia de actualización y checksum opcional.
- Distinguir assets reemplazables completos de archivos que requieren merge o confirmación.
- Usar el manifest tanto en instalación como en sync y verificación.
- Evitar incluir archivos generados o personalizaciones del proyecto destino.

### `RM-010` — Templates reales por stack

**Propuesta**: no prometer templates de stack hasta que existan y se instalen.

**Design**:

- Cada template debe tener archivos concretos, criterio de instalación y verificación.
- Actualizar README solo cuando el template exista como asset real.
- Mantener templates recomendados como roadmap mientras no existan: `frontend-react`, `frontend-nextjs`, `frontend-astro`, `mobile-expo`, `backend-spring`.
- Evitar `backend-node` genérico salvo que se defina un caso de uso backend concreto.

### `RM-011` — README como estado real y roadmap separado

**Propuesta**: mantener README enfocado en quickstart, estado actual y enlaces; mover contenido futuro a `docs/roadmap.md` cuando se haga la reestructuración documental.

**Design**:

- Por ahora, enlazar este roadmap desde README sin reescritura amplia.
- En una iteración futura, migrar secciones especulativas de templates/subagentes a este documento o a docs específicos.
- Evitar que README venda templates o features no instalables.
- Mantener una navegación corta hacia `docs/getting-started.md`, `docs/roadmap.md` y `openspec/README.md`.

### `RM-012` — CLI futura

**Propuesta**: evaluar CLI después de estabilizar instalador, CI, sync y templates reales.

**Design**:

- No bloquear Fase 1 por selección de lenguaje o empaquetado.
- Considerar CLI solo si reduce complejidad del shell o mejora instalación multiplataforma.
- Requisitos mínimos futuros: comandos `install`, `verify`, `sync`, `doctor`, versionado y distribución clara.

### `RM-013` — Migrar installer a CLI Go

**Propuesta**: productizar el installer como CLI Go (`lufy-ai`) cuando el hardening inicial esté validado, moviendo progresivamente la lógica crítica fuera de Bash.

**Rationale Go vs Rust**:

- Go es suficiente para CLI, filesystem, JSON, OS detection y distribución como single-binary.
- Rust no es necesario para esta etapa porque el problema principal es idempotencia, backup/rollback, portabilidad y UX, no performance ni memory-safety.
- Mantener Bash solo como wrapper estricto de compatibilidad reduce fricción para usuarios actuales sin seguir aumentando la complejidad del shell ni duplicar rutas legacy.

**Design por fases**:

1. Mantener Bash como wrapper estricto que delega en la CLI Go; si no hay binario local o en `PATH`, falla con instrucciones de build local.
2. Crear una CLI Go mínima con comandos base: `lufy-ai install --target . --dry-run --yes --no-engram`, `lufy-ai verify --target .`, `lufy-ai backup` y `lufy-ai restore`.
3. Mover al binario la detección de entorno, Engram portable, backup/rollback, merge/idempotencia y `verify`.
4. Agregar `sync`/`update` después de estabilizar instalación, verificación y rollback.
5. Evaluar una TUI Go como opción futura; no es prioridad inicial.

**Estado actual (2026-05-05):**

- ✅ Scaffolding Go en carpeta dedicada `tools/lufy-cli-go/` con `go.mod` propio.
- ✅ Comandos base cableados (`install`, `verify`, `backup`, `restore`) en slice mínimo funcional.
- ✅ Wrapper `scripts/install.sh` delega exclusivamente a `lufy-ai install`, usando `tools/lufy-cli-go/bin/lufy-ai` o `lufy-ai` en `PATH`, sin fallback legacy.
- ✅ Resolución Engram portable por `PATH` sin hardcode nuevo en la ruta de migración.
- ✅ Smoke E2E reproducible validado en temp dir para install real + verify + idempotencia básica (2da ejecución con `skip`) + backup/restore (dry-run y real) sobre conflicto controlado de `AGENTS.md`.
- ✅ Install real copia assets gestionados del catálogo, escribe `.lufy-ai/install-state.json` con SHA-256 y evita sobrescribir drift local.
- ✅ `backup`/`restore` usan `manifest.json`, hashes, `targetRoot` y backup de recovery antes de restauraciones reales.
- ✅ `sync` reaplica assets gestionados con hash/idempotencia y backup previo, bloqueando drift local, estado ausente/corrupto y escapes por symlink/path inseguro.
- ✅ `opencode.json` usa contrato `merge-json` especial en `install`, `sync` y `verify`: no se copia como asset completo ni se registra con SHA-256 en `.lufy-ai/install-state.json`.

### `RM-014` — Releases binarios versionados

**Propuesta**: publicar artifacts versionados de `lufy-ai` desde GitHub Actions/GitHub Releases para que el usuario no necesite clonar y compilar el repositorio como paso de instalación.

**Design:**

- Construir artifacts por OS/arch soportado desde `tools/lufy-cli-go/` sin depender de tooling Node/TS de raíz.
- Generar nombres determinísticos que incluyan versión, OS y arquitectura.
- Publicar checksums SHA-256 junto a cada release y validarlos en CI antes de considerar el release instalable.
- Añadir `lufy-ai version` con versión, commit, fecha de build, GOOS y GOARCH; los builds locales sin metadata deben declararse como development/unknown.
- Tratar GitHub Releases con checksums como fuente de verdad para canales posteriores como Homebrew, Scoop o manifests similares.

**Estado actual:** implementado en la rama por proposal OpenSpec `add-versioned-binary-release-installer`: build de artifacts por OS/arch, checksums SHA-256 y `lufy-ai version` existen. Para uso público falta publicar un tag `v*` y los artifacts de GitHub Releases correspondientes; sin ese tag, no hay URL pública instalable aunque el código esté listo.

### `RM-015` — Bootstrap remoto seguro

**Propuesta**: ofrecer un instalador remoto que descargue una release versionada, verifique su checksum y coloque el binario en `PATH` sin ejecutar acciones destructivas por defecto.

**Design:**

- Detectar OS/arch y fallar de forma accionable para plataformas no soportadas.
- Exigir verificación SHA-256 antes de instalar o ejecutar cualquier binario descargado.
- Soportar version pinning (`vX.Y.Z`) como ruta recomendada para automatización; `latest` debe ser conveniencia explícita con trade-off documentado.
- Documentar `curl | bash` solo junto con alternativa inspeccionable: descargar script, revisarlo y ejecutarlo con versión explícita.
- No ejecutar `lufy-ai install` contra un target de proyecto salvo flag explícito del usuario; el bootstrap instala el binario, no modifica repositorios destino por sorpresa.

**Estado actual:** implementado en la rama como `scripts/bootstrap.sh`. Detecta OS/arch, soporta versión explícita o `latest`, descarga artifact/checksums y verifica SHA-256 antes de instalar el binario. El bootstrap está disponible para inspección y validación local; la descarga desde GitHub solo funciona cuando exista la release taggeada publicada. `scripts/install.sh` sigue siendo wrapper local estricto y no crece fallback remoto.

### `RM-016` — Assets standalone para instalación real sin clone

**Propuesta**: eliminar la dependencia del checkout fuente para instalar assets gestionados, llevando esos assets dentro del binario o en un bundle versionado verificable.

**Design:**

- Estrategia incremental recomendada: usar `go:embed` primero para que el binario release sea autocontenido y el checksum cubra código + assets.
- Considerar bundle zip/tar versionado si el tamaño o frecuencia de assets hace incómodo recompilar el binario por cambios de contenido.
- Si se usa bundle, verificar checksum del bundle completo y manifest interno antes de usar assets.
- Mantener los contratos existentes de idempotencia, manifest `.lufy-ai/install-state.json`, hashes SHA-256, backup/restore y `verify`.
- No documentar instalación standalone real hasta que `lufy-ai install --target <dir>` funcione desde un binario distribuido sin leer el repo fuente.

**Estado actual:** implementado en la rama mediante assets gestionados embebidos en el binario Go. Un binario release puede ejecutar `lufy-ai install --target <dir>` sin leer el checkout fuente, preservando idempotencia, manifest, hashes, backup/restore, sync y verify; el quickstart público requiere que exista una release `v*` publicada para descargar ese binario.

### `RM-017` — Documentación final del flujo sin clone

**Propuesta**: actualizar la documentación pública al cierre de la implementación para mover el camino principal a instalación sin clone y retirar instrucciones obsoletas.

**Design:**

- Actualizar `README.md`, `docs/getting-started.md` y `tools/lufy-cli-go/README.md` solo cuando releases, checksums, bootstrap y assets standalone estén implementados y validados.
- Mantener instrucciones clone/build únicamente como flujo de contribuidor/desarrollo si siguen siendo útiles.
- Borrar o demotar docs obsoletos que presenten clone/build como camino primario de usuario final; no conservar retrocompatibilidad documental hacia un flujo que ya no aplique.
- Incluir verificación post-instalación con `lufy-ai verify` y advertencias sobre pinning/checksums.

**Estado actual:** implementado en la rama. README, `docs/getting-started.md` y `tools/lufy-cli-go/README.md` ya describen instalación sin clone, pinning, checksum/bootstrap e instalación con binario; deben leerse como flujo disponible para tags `v*` publicados, no como garantía de que una release pública exista antes del primer tag.

## Criterios de aceptación

### Criterios generales

- Cada iniciativa conserva un ID estable `RM-###` para issues, ramas o tags.
- Los cambios siguen siendo compatibles con el rol real del repo: kit OpenCode/OpenSpec instalable.
- La documentación humana permanece en español y preserva identificadores técnicos.
- No se documenta como existente algo que no esté implementado o instalado realmente.

### Criterios por iniciativa

| ID | Criterios de aceptación |
| --- | --- |
| `RM-001` | `scripts/install.sh --dry-run --target <dir>` no escribe en el target; `--yes` no requiere prompts seguros; `--no-engram` omite Engram; flags inválidos fallan con ayuda. |
| `RM-002` | Antes de modificar paths sensibles se crea backup; ante fallo se intenta rollback; queda manifest de la ejecución. |
| `RM-003` | `opencode.json` no contiene `/opt/homebrew/bin/engram` hardcodeado cuando la integración se genera; usa el resultado de `command -v engram` o queda deshabilitada. |
| `RM-004` | `lufy-ai verify --target <dir> --no-engram` valida estructura, JSON parseable, manifest, hashes y presencia de commands/skills/plugin. |
| `RM-005` | Reinstalar sobre un target existente no sobrescribe configs sin estrategia; `opencode.json` se mergea de forma conservadora y los conflictos/JSON inválido se reportan de forma accionable. |
| `RM-006` | Existe workflow en `.github/workflows/` con tests/build Go, smoke de instalación en temp dir, wrapper smoke y checks estáticos disponibles; `shellcheck` queda como mejora opcional si se incorpora al runner. |
| `RM-007` | La verificación local está documentada y CI ejecuta `lufy-ai verify` o checks equivalentes. |
| `RM-008` | Sync reaplica assets gestionados con backup y reporte por archivo, sin modificar personalizaciones fuera de scope. |
| `RM-009` | Existe manifest de assets gestionados y lo consumen instalación/sync/verificación cuando aplique. |
| `RM-010` | Cualquier template documentado como disponible existe como asset instalable y es verificado por CI o script local. |
| `RM-011` | README queda centrado en estado real + quickstart; contenido futuro vive en docs de roadmap o diseño. |
| `RM-012` | Hay decisión documentada de CLI con alcance, tradeoffs y prerequisitos cumplidos antes de implementarla. |
| `RM-013` | Existe plan de migración a CLI Go; Bash queda limitado a wrapper estricto sin fallback legacy; la primera CLI cubre `install`, `verify`, `backup` y `restore` antes de `sync`/`update`. |
| `RM-014` | GitHub Releases publica binarios versionados por OS/arch con checksums SHA-256 y `lufy-ai version` reporta metadata de release. |
| `RM-015` | El bootstrap remoto detecta OS/arch, soporta pinning de versión, verifica checksum antes de instalar y ofrece alternativa inspeccionable a `curl \| bash`. |
| `RM-016` | Un binario release o bundle verificado puede instalar assets gestionados sin leer el checkout fuente, preservando idempotencia, manifest, hashes, backup/restore y verify. |
| `RM-017` | README/getting-started/CLI docs describen instalación sin clone solo cuando está implementada y retiran clone/build como camino primario de usuario final. |

## Riesgos/dependencias

- `shellcheck` puede no estar disponible localmente; CI debe instalarlo o usar una acción mantenida.
- Merge JSON sin tooling dedicado puede volverse frágil en shell; si crece, conviene evaluar helper en Python o CLI futura.
- Rollback parcial puede ser complejo si se mezclan copias recursivas y merges; la primera versión debe limitar acciones y registrar manifest.
- `--yes` puede ser peligroso si se interpreta como permiso destructivo; debe combinarse con backup y políticas conservadoras.
- Sync requiere una fuente de verdad de assets gestionados; sin manifest puede pisar personalizaciones.
- Templates por stack pueden reintroducir drift documental si README se actualiza antes de tener archivos reales.
- Migrar lógica desde Bash a Go puede dejar brechas mientras la CLI completa assets gestionados; el wrapper evita duplicación al no mantener fallback legacy.
- Publicar binarios sin resolver assets standalone puede crear una falsa promesa de instalación sin clone; la documentación final debe esperar a `go:embed` o bundle verificado.
- `curl | bash` requiere comunicación cuidadosa: pinning e inspección deben estar documentados junto al comando directo.
- Matrices OS/arch, Homebrew y Scoop aumentan mantenimiento; conviene estabilizar GitHub Releases + checksums antes de sumar canales.
- Cambiar el default de instalacion a scope global puede sorprender a usuarios actuales; debe incluir flag `--scope=project`, migracion documentada y smoke brownfield antes de release.
- OpenSpec v2 cambia reglas de authoring al exigir delta markers y scenarios; debe introducirse con validator accionable y migracion clara, no solo con templates nuevos.

## Etiquetas sugeridas

- `roadmap`
- `installer-hardening`
- `install-safety`
- `idempotency`
- `backup-rollback`
- `engram`
- `ci`
- `verify-install`
- `sync`
- `templates`
- `documentation`
- `future-cli`
- `go-cli`
- `release-distribution`
- `checksums`
- `bootstrap-installer`
- `version-pinning`
