## Purpose
Definir la CLI Go `lufy-ai` como motor portable de instalación, verificación, backup, restore y sync del kit OpenCode/OpenSpec, manteniendo `scripts/install.sh` como wrapper estricto sin fallback legacy.
## Requirements
### Requirement: CLI Go instalable
El sistema SHALL proveer una CLI Go llamada `lufy-ai` como motor de instalación progresiva del kit OpenCode/OpenSpec.

#### Scenario: Compilación del binario inicial
- **WHEN** el repositorio contiene `tools/lufy-cli-go/go.mod` y el código bajo `tools/lufy-cli-go/cmd/lufy-ai`
- **THEN** ejecutando `go build ./cmd/lufy-ai` desde `tools/lufy-cli-go/` se genera el binario sin depender de toolchains no-Go globales

#### Scenario: Punto de entrada delgado
- **WHEN** `cmd/lufy-ai/main.go` recibe un comando soportado
- **THEN** delega la lógica de negocio a paquetes internos en vez de implementar instalación completa dentro de `main.go`

### Requirement: Comandos base de instalación
La CLI SHALL implementar los comandos iniciales `install`, `verify`, `backup`, `restore` y `sync` antes de añadir comandos posteriores como `update`.

#### Scenario: Install con flags mínimos
- **WHEN** el usuario ejecuta `lufy-ai install --target . --dry-run --yes --no-engram`
- **THEN** la CLI construye un plan de instalación para el target actual, omite Engram y no escribe archivos por estar en dry-run

#### Scenario: Verify de un target
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir>`
- **THEN** la CLI valida estructura instalada, archivos esperados, JSON parseable cuando aplique, `.lufy-ai/install-state.json`, manifest de assets gestionados, hashes SHA-256 y estado de integración Engram según flags

#### Scenario: Backup explícito
- **WHEN** el usuario ejecuta `lufy-ai backup --target <dir>`
- **THEN** la CLI crea un backup con manifest de los archivos gestionados o relevantes para rollback dentro del target

#### Scenario: Restore desde manifest
- **WHEN** el usuario ejecuta `lufy-ai restore --target <dir> --backup <manifest-or-dir>`
- **THEN** la CLI valida el manifest y restaura los archivos registrados de forma controlada

#### Scenario: Sync de assets gestionados
- **WHEN** el usuario ejecuta `lufy-ai sync --target <dir> --dry-run --yes --no-engram`
- **THEN** la CLI construye un plan de sincronización de assets gestionados para el target actual, omite Engram y no escribe archivos por estar en dry-run

### Requirement: Flags y defaults seguros
La CLI SHALL soportar `--target`, `--dry-run`, `--yes`, `--no-engram` y `--backup` con defaults seguros que minimicen escrituras inesperadas y prompts ambiguos.

#### Scenario: Target por defecto
- **WHEN** el usuario omite `--target`
- **THEN** la CLI usa `.` como target y lo resuelve a una ruta segura antes de planificar o escribir

#### Scenario: Dry-run sin mutaciones
- **WHEN** el usuario pasa `--dry-run`
- **THEN** la CLI muestra el plan y MUST NOT crear, modificar, borrar, clonar ni respaldar archivos reales

#### Scenario: Confirmación requerida
- **WHEN** una acción puede sobrescribir o restaurar archivos y el usuario no pasa `--yes`
- **THEN** la CLI solicita confirmación interactiva o falla de forma accionable si no hay TTY

#### Scenario: Opt-out de Engram
- **WHEN** el usuario pasa `--no-engram`
- **THEN** la CLI omite detección, configuración y verificación obligatoria de Engram

#### Scenario: Flag inválido
- **WHEN** el usuario pasa un flag desconocido a cualquier comando
- **THEN** la CLI falla con exit code distinto de cero y muestra ayuda breve del comando

#### Scenario: Sync comparte flags seguros
- **WHEN** el usuario ejecuta `lufy-ai sync` con `--target`, `--dry-run`, `--yes` o `--no-engram`
- **THEN** la CLI aplica los mismos defaults seguros y semántica de flags definidos para comandos de instalación gestionada

### Requirement: Plan de instalación antes de escribir
El comando `install` SHALL construir un plan explícito antes de modificar el filesystem.

#### Scenario: Plan de acciones
- **WHEN** `install` analiza un target
- **THEN** el plan enumera acciones como crear directorio, copiar asset, omitir archivo idéntico, mergear JSON, crear backup o reportar conflicto

#### Scenario: Plan con conflictos
- **WHEN** existe un archivo distinto en el target y no hay merge seguro
- **THEN** el plan marca el conflicto y no lo sobrescribe silenciosamente

### Requirement: Idempotencia y preservación de trabajo del usuario
La instalación SHALL ser idempotente y MUST NOT sobrescribir trabajo del usuario sin estrategia explícita, backup y confirmación cuando corresponda; `AGENTS.md` SHALL be treated as user-owned and integrated only through the minimal `@lufy-ia.harness.md` reference.

#### Scenario: Reinstalación sin cambios
- **WHEN** el usuario ejecuta `install` dos veces sobre el mismo target sin modificaciones intermedias
- **THEN** la segunda ejecución reporta archivos idénticos como `skip` o equivalente y no produce conflictos falsos

#### Scenario: AGENTS.md existente
- **WHEN** el target ya contiene `AGENTS.md` sin referencia al harness
- **THEN** la CLI preserva el archivo, planifica únicamente la inserción de la referencia `@lufy-ia.harness.md` con backup y confirmación/`--yes`, y no inserta el contenido completo de Lufy

#### Scenario: AGENTS.md ausente
- **WHEN** el target no contiene `AGENTS.md`
- **THEN** install crea un archivo mínimo user-owned que referencia `@lufy-ia.harness.md` y no lo registra como asset completo gestionado por hash

#### Scenario: AGENTS.md con referencia existente
- **WHEN** el target contiene `AGENTS.md` con la referencia `@lufy-ia.harness.md`
- **THEN** install no duplica la referencia y no reescribe `AGENTS.md` solo por esa integración

#### Scenario: Archivo desconocido del usuario
- **WHEN** el target contiene archivos no gestionados por `lufy-ai`
- **THEN** la CLI no los borra ni modifica durante `install`

### Requirement: Backup y rollback con manifest
La CLI SHALL registrar backups en un manifest portable antes de cambios riesgosos y SHALL permitir restore usando ese manifest.

#### Scenario: Manifest creado antes de sobrescribir
- **WHEN** `install` necesita modificar un archivo existente y backup está habilitado o requerido por seguridad
- **THEN** la CLI copia el estado previo a `.lufy-ai/backups/<timestamp>/` y registra path relativo, operación y hash en `manifest.json`

#### Scenario: Error durante instalación
- **WHEN** una acción de instalación falla después de crear un backup
- **THEN** la CLI reporta el manifest disponible y, si es seguro, intenta revertir las acciones aplicadas o indica el comando `restore` necesario

#### Scenario: Restore dry-run
- **WHEN** el usuario ejecuta `restore --dry-run --backup <manifest-or-dir>`
- **THEN** la CLI muestra qué archivos restauraría sin escribir cambios

### Requirement: Merge conservador de opencode.json
La CLI SHALL crear o mergear `opencode.json` mediante JSON válido, preservando claves desconocidas del usuario, y SHALL tratarlo como configuración `merge-json` especial en vez de asset completo gestionado por hash.

#### Scenario: Crear opencode.json faltante
- **WHEN** el target no contiene `opencode.json` y la instalación requiere configuración OpenCode
- **THEN** la CLI crea un JSON válido con las claves gestionadas mínimas por `lufy-ai`

#### Scenario: Preservar claves existentes
- **WHEN** el target contiene `opencode.json` válido con claves no gestionadas
- **THEN** la CLI preserva esas claves y modifica solo secciones gestionadas por `lufy-ai`

#### Scenario: JSON inválido
- **WHEN** el target contiene `opencode.json` inválido
- **THEN** la CLI falla sin sobrescribirlo y reporta una instrucción accionable para corregir o respaldar el archivo

#### Scenario: Opencode no se registra con hash completo
- **WHEN** `install` o `sync` escriben `opencode.json` mediante `merge-json`
- **THEN** `.lufy-ai/install-state.json` no contiene una entrada de asset completo para `opencode.json` ni requiere comparar su SHA-256 como asset gestionado

### Requirement: Engram portable
La CLI SHALL resolver Engram de forma portable con `exec.LookPath("engram")` o abstracción equivalente y MUST NOT hardcodear `/opt/homebrew/bin/engram`.

#### Scenario: Engram encontrado en PATH
- **WHEN** Engram existe en `PATH` y `--no-engram` no está activo
- **THEN** la CLI usa la ruta resuelta o una invocación portable compatible con la configuración OpenCode

#### Scenario: Engram ausente
- **WHEN** Engram no existe en `PATH` y `--no-engram` no está activo
- **THEN** la instalación base continúa sin fallar, dejando la integración deshabilitada o no configurada y reportando una nota accionable

#### Scenario: Ruta hardcodeada prohibida
- **WHEN** la CLI genera configuración relacionada con Engram
- **THEN** el contenido generado no contiene `/opt/homebrew/bin/engram`

### Requirement: Wrapper Bash estricto
`scripts/install.sh` SHALL permanecer como wrapper de compatibilidad que delega exclusivamente en `lufy-ai install` y MUST NOT conservar lógica legacy de instalación.

#### Scenario: Uso histórico con argumento posicional
- **WHEN** el usuario ejecuta `scripts/install.sh <target-project-dir>`
- **THEN** el wrapper conserva compatibilidad y delega o mapea a `lufy-ai install --target <target-project-dir>` cuando el binario Go está disponible

#### Scenario: Delegación al binario Go
- **WHEN** el wrapper detecta `tools/lufy-cli-go/bin/lufy-ai` o `lufy-ai` en `PATH`
- **THEN** delega la instalación a la CLI Go preservando flags compatibles

#### Scenario: CLI Go ausente
- **WHEN** el wrapper no encuentra `lufy-ai` en `PATH` ni `tools/lufy-cli-go/bin/lufy-ai`
- **THEN** falla sin ejecutar fallback legacy y muestra la instrucción `cd tools/lufy-cli-go && mkdir -p bin && go build -o bin/lufy-ai ./cmd/lufy-ai`

#### Scenario: Sin lógica legacy
- **WHEN** se inspecciona `scripts/install.sh`
- **THEN** no contiene lógica legacy de copia, detección de stack, Engram, `copy_files`, generación de `opencode.json` ni prompts de instalación

#### Scenario: Sin descarga remota insegura
- **WHEN** el wrapper no encuentra binario Go
- **THEN** no descarga ni ejecuta binarios remotos sin mecanismo explícito de integridad y autorización

### Requirement: Validación por fases
La implementación SHALL incluir validación incremental con comandos reales disponibles después de introducir el toolchain Go, y SHALL ser ejecutable tanto localmente como desde CI mínima, incluyendo checks de paridad de assets, path safety portable, metadata de state y escrituras atomicas cuando aplique.

#### Scenario: Validación Go disponible
- **WHEN** existen `tools/lufy-cli-go/go.mod` y paquetes Go
- **THEN** el implementador puede ejecutar `go test ./...` y `go build ./cmd/lufy-ai` desde `tools/lufy-cli-go/` como validación mínima

#### Scenario: Prueba dry-run en temp dir
- **WHEN** el binario Go compila
- **THEN** el implementador puede ejecutar una instalación `--dry-run` contra un directorio temporal y confirmar que no se escriben archivos de instalación

#### Scenario: Verify tras instalación temporal
- **WHEN** una instalación real se ejecuta en un directorio temporal de prueba
- **THEN** `lufy-ai verify --target <temp>` valida el resultado sin depender de modificar el repositorio fuente

#### Scenario: Validación automática en CI
- **WHEN** se ejecuta el workflow de CI mínima del instalador Go
- **THEN** la validación incluye tests, build y smoke temporal de install/verify/idempotencia/backup/restore con `--no-engram`

### Requirement: Comando sync de CLI Go
La CLI Go SHALL exponer `lufy-ai sync` como comando para sincronizar assets gestionados de forma segura en un target existente y aplicar merges seguros para assets `merge-json` explícitos, actualizando `lufy-ia.harness.md` como asset gestionado sin mutar `AGENTS.md` automáticamente.

#### Scenario: Help incluye sync
- **WHEN** el usuario solicita ayuda de la CLI o del comando `sync`
- **THEN** la salida describe `sync`, sus flags soportados y que opera sobre assets gestionados con manifest/hash/backup

#### Scenario: Sync delega fuera de main
- **WHEN** `cmd/lufy-ai/main.go` recibe el comando `sync`
- **THEN** delega la lógica de negocio a paquetes internos en vez de implementar planificación o copia completa dentro de `main.go`

#### Scenario: Wrapper Bash no cambia para sync
- **WHEN** se inspecciona `scripts/install.sh` después de añadir `sync`
- **THEN** permanece como wrapper estricto de `lufy-ai install` y no contiene lógica propia ni fallback legacy para sincronizar assets

#### Scenario: Sync aplica merge-json de opencode
- **WHEN** un target instalado tiene `opencode.json` válido que necesita claves merge-managed mínimas
- **THEN** `sync` planifica/aplica `merge-json` para `opencode.json`, preserva claves desconocidas y no usa `copy` ni `update-managed` por hash para ese archivo

#### Scenario: Sync aplica harness gestionado
- **WHEN** un target instalado tiene `lufy-ia.harness.md` registrado sin drift local y el source del harness cambió
- **THEN** `sync` planifica/aplica backup y `update-managed` para `lufy-ia.harness.md` y actualiza su entrada de manifest

#### Scenario: Sync no auto-repara AGENTS
- **WHEN** un target instalado tiene `AGENTS.md` sin la referencia `@lufy-ia.harness.md`
- **THEN** `sync` reporta warning o acción explícita requerida y MUST NOT modificar `AGENTS.md` silenciosamente

### Requirement: Validación de sync en CLI Go
La implementación SHALL incluir validación real del comando `sync` usando comandos disponibles del toolchain Go y pruebas de filesystem confinadas a directorios temporales.

#### Scenario: Validación Go de sync disponible
- **WHEN** existen `tools/lufy-cli-go/go.mod` y paquetes Go con el comando `sync`
- **THEN** el implementador puede ejecutar `go test ./...` y `go build ./cmd/lufy-ai` desde `tools/lufy-cli-go/` como validación mínima

#### Scenario: Sync dry-run en temp dir
- **WHEN** el binario Go compila
- **THEN** el implementador puede ejecutar `lufy-ai sync --target <temp> --dry-run` y confirmar que no se escriben archivos de sincronización

#### Scenario: Verify tras sync temporal
- **WHEN** una instalación temporal y un sync real se ejecutan en un directorio temporal de prueba
- **THEN** `lufy-ai verify --target <temp>` valida el resultado sin depender de modificar el repositorio fuente

### Requirement: `lufy-ai verify` canónico
La CLI Go SHALL usar `lufy-ai verify` como verificador canónico de instalaciones y MUST NOT requerir ni introducir `scripts/verify-install.sh`; verify SHALL validate `lufy-ia.harness.md` as the managed agent-instructions asset and validate `AGENTS.md` as a user-owned reference integration.

#### Scenario: Verificación estructural de categorías críticas
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir> --no-engram` sobre un target instalado
- **THEN** la CLI valida que `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/plugins` y `.opencode/policies` existen como directorios seguros no symlink

#### Scenario: Verificación de archivos críticos
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir> --no-engram` sobre un target instalado
- **THEN** la CLI valida que `.opencode/plugins/agent-observatory.tsx`, `lufy-ia.harness.md`, `tui.json`, `openspec/config.yaml` y `.lufy-ai/install-state.json` existen como archivos seguros no symlink, y valida `AGENTS.md` como archivo user-owned que referencia el harness cuando esté presente

#### Scenario: Archivos críticos presentes en manifest
- **WHEN** un archivo crítico gestionado existe en el target pero no está registrado en `.lufy-ai/install-state.json`
- **THEN** `lufy-ai verify` falla indicando que el asset clave no está en el manifest; esta regla aplica a `lufy-ia.harness.md` y no exige entrada de manifest para `AGENTS.md`

#### Scenario: Hashes de assets gestionados
- **WHEN** un asset listado en `.lufy-ai/install-state.json` existe pero su SHA-256 actual no coincide con `targetSHA256`
- **THEN** `lufy-ai verify` falla reportando drift con hashes abreviados expected/actual

#### Scenario: Verificación de opencode merge-managed
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir> --no-engram` sobre un target instalado
- **THEN** la CLI valida que `opencode.json` sea JSON parseable y contenga la estructura mínima merge-managed sin requerir entrada de hash completo en el manifest

#### Scenario: Verificación de referencia AGENTS
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir> --no-engram` sobre un target instalado
- **THEN** la CLI valida que `AGENTS.md` contenga `@lufy-ia.harness.md` como requisito o warning accionable sin comparar hash completo de `AGENTS.md`

#### Scenario: No existe script verificador paralelo
- **WHEN** se documenta o valida una instalación local/CI
- **THEN** la guía usa `lufy-ai verify` y no define `scripts/verify-install.sh` como objetivo ni dependencia

### Requirement: Standalone asset source
The CLI Go SHALL support installation from a distributed binary without requiring access to the source repository checkout.

#### Scenario: Embedded assets install without clone
- **WHEN** a release binary includes managed assets embedded in the binary
- **THEN** `lufy-ai install --target <dir>` can install the managed OpenCode/OpenSpec assets without reading from the repository source tree

#### Scenario: Bundle assets install without clone
- **WHEN** a release uses a versioned asset bundle instead of embedded assets
- **THEN** the CLI or bootstrap verifies the bundle integrity before using it as the asset source for installation

#### Scenario: Source checkout remains development path only
- **WHEN** the CLI runs from a developer checkout
- **THEN** it may use local assets for development workflows, but public installation documentation does not require cloning once standalone assets are implemented

### Requirement: Release binary preserves installer safety
The release-distributed `lufy-ai` binary SHALL preserve existing install, verify, backup, restore and sync safety semantics.

#### Scenario: Distributed install remains idempotent
- **WHEN** the user runs a release binary installation twice against the same target
- **THEN** the second run reports unchanged managed assets without overwriting local drift or unmanaged user files

#### Scenario: Distributed verify uses same structural checks
- **WHEN** the user runs `lufy-ai verify --target <dir> --no-engram` from a release binary
- **THEN** it validates structure, JSON, manifest and SHA-256 managed asset hashes with the same contract as the local build

#### Scenario: Wrapper remains strict
- **WHEN** `scripts/install.sh` is retained after release distribution exists
- **THEN** it continues to delegate to `lufy-ai install` and does not implement its own remote download fallback

### Requirement: CLI state metadata comes from version package
The CLI SHALL use the same runtime version metadata source for `lufy-ai version`, install state and backup manifests.

#### Scenario: Version command and state agree
- **WHEN** a release binary writes install state after install or sync
- **THEN** the state metadata is derived from the same version package used by `lufy-ai version`

### Requirement: CLI filesystem writes are atomic for managed payloads
The CLI SHALL use atomic writes for managed file payloads copied by install, sync, backup and restore.

#### Scenario: Managed file copy uses atomic helper
- **WHEN** internal copy helpers write a managed file payload
- **THEN** they use a shared or equivalent temp-file-plus-rename pattern rather than direct final-path writes

### Requirement: CLI validates portable path traversal
The CLI SHALL reject unsafe relative paths before direct catalog, state or backup operations use them.

#### Scenario: Direct `EnsureRelativeSafe` use is safe
- **WHEN** `verify`, `backup` or catalog building validates a relative path without immediately calling `SafeJoin`
- **THEN** traversal forms using `../`, `..\` or mixed separators are rejected

### Requirement: CLI scope flag for managed operations
The CLI SHALL expose a `--scope` flag for managed install/sync/verify/status operations where scope affects target resolution.

#### Scenario: Invalid scope rejected
- **WHEN** the user passes an unsupported `--scope` value
- **THEN** the CLI exits non-zero with allowed values `project`, `global` and `both`

#### Scenario: Scope shown in dry-run
- **WHEN** the user runs install or sync with `--dry-run` and a scope value
- **THEN** the output identifies the effective scope and root paths that would be written

### Requirement: CLI merge command
The CLI SHALL expose `lufy-ai merge <path>` for policy-driven drift resolution where ancestor and `.lufy-new` data exist.

#### Scenario: Help includes merge
- **WHEN** the user requests CLI help
- **THEN** the output lists `merge` as the command for reconciling `.lufy-new` files with local edits

#### Scenario: Merge does not write before tool succeeds
- **WHEN** the merge tool cannot be started or exits unsuccessfully
- **THEN** the CLI preserves the original target, ancestor and `.lufy-new` files

### Requirement: CLI restore discovery
The CLI SHALL support backup discovery in addition to restoring from explicit manifest paths.

#### Scenario: Restore list mode is non-mutating
- **WHEN** the user asks restore to list backups
- **THEN** the CLI reads backup manifests and prints available backups without writing target files

#### Scenario: Explicit manifest remains supported
- **WHEN** the user runs `restore --backup <manifest-or-dir>`
- **THEN** the CLI preserves the existing restore behavior for that explicit backup reference

### Requirement: CLI reports drift actions consistently
The CLI SHALL use consistent action names and JSON fields for policy-driven drift handling.

#### Scenario: Plan includes lufy-new action
- **WHEN** install or sync plans a no-replace drift resolution
- **THEN** human and JSON plan output identify an action for writing `.lufy-new` rather than a destructive update

#### Scenario: Verify and status share policy fields
- **WHEN** verify or status emits JSON
- **THEN** each relevant asset result includes policy, scope, target path and recommended action when drift is detected

### Requirement: CLI instala workflow OpenSpec core v2 standalone
La CLI Go SHALL instalar el workflow OpenSpec core v2 desde checkout de desarrollo o desde assets embebidos de release sin requerir clone adicional.

#### Scenario: Instalación desde release incluye core v2
- **WHEN** el usuario ejecuta `lufy-ai install --target <dir>` con un binario release que contiene assets core v2 embebidos
- **THEN** el target recibe la configuración, comandos, skills y baseline OpenSpec core v2 gestionados

#### Scenario: Sync desde release actualiza core v2
- **WHEN** el usuario ejecuta `lufy-ai sync --target <dir>` con un binario release que contiene assets core v2 embebidos
- **THEN** la CLI compara el catálogo embebido y planifica updates seguros para assets OpenSpec core v2 gestionados

### Requirement: CLI mantiene validación de paridad de assets OpenSpec
La implementación SHALL mantener validación automática para evitar que specs o comandos OpenSpec raíz diverjan de sus copias embebidas.

#### Scenario: Tests detectan drift de assets embebidos
- **WHEN** un asset OpenSpec core v2 raíz cambia sin actualizar su copia embebida
- **THEN** los tests Go de assets fallan indicando drift entre catálogo raíz y embebido

#### Scenario: Validación agrupada ejecuta paridad relevante
- **WHEN** se ejecuta `scripts/validate.sh`
- **THEN** la validación Go incluye la comprobación de paridad entre assets raíz y embebidos

### Requirement: CLI provides OpenSpec resolver package
La CLI Go SHALL incluir un paquete interno para resolver la fuente efectiva de OpenSpec sin acoplarla al catálogo de assets gestionados.

#### Scenario: Resolver package is isolated
- **WHEN** se inspecciona la implementación de stay-updated
- **THEN** la lógica de resolución, manifiestos y cache vive en un paquete interno dedicado y no dentro de `cmd/lufy-ai/main.go`

#### Scenario: Resolver remains stdlib-only
- **WHEN** se compila la CLI Go después de agregar el resolver
- **THEN** no requiere dependencias externas nuevas salvo decisión explícita documentada

### Requirement: CLI uses embedded baseline as fallback
La CLI Go SHALL conservar instalación standalone usando baseline embebida cuando no haya fuente OpenSpec externa válida.

#### Scenario: Release binary resolves offline baseline
- **WHEN** un binario release se ejecuta sin checkout fuente, sin red y sin `openspec` en `PATH`
- **THEN** el resolver selecciona la baseline embebida y los comandos locales siguen funcionando

#### Scenario: Resolver does not mutate during install by default
- **WHEN** el usuario ejecuta `lufy-ai install` o `lufy-ai sync`
- **THEN** la CLI no descarga OpenSpec remoto ni modifica cache salvo que una acción explícita de update/cache lo solicite

### Requirement: CLI validates OpenSpec cache safely
La CLI Go SHALL validar cache OpenSpec por manifiesto y paths seguros antes de usarla.

#### Scenario: Corrupt cache falls back safely
- **WHEN** la cache local existe pero su manifiesto es inválido
- **THEN** la CLI ignora esa cache, reporta warning accionable y usa la siguiente capa válida

#### Scenario: Unsafe cache paths are blocked
- **WHEN** el manifiesto de cache contiene paths absolutos, traversal o symlinks inseguros
- **THEN** la CLI rechaza la cache y no lee ni escribe fuera del target

### Requirement: CLI init command
The CLI Go SHALL expose `lufy-ai init` as the command for generating stack-aware project configuration.

#### Scenario: Help includes init
- **WHEN** the user requests CLI help
- **THEN** the output lists `init` as the command for generating `.opencode/project.yaml`

#### Scenario: Init delegates outside main
- **WHEN** `cmd/lufy-ai/main.go` receives the `init` command
- **THEN** it delegates scanning, merging and writing logic to internal packages instead of implementing that logic in `main.go`

#### Scenario: Init supports target flag
- **WHEN** the user runs `lufy-ai init --target <dir>`
- **THEN** the CLI resolves `<dir>` with the same safe target handling used by managed commands before reading or writing `.opencode/project.yaml`

### Requirement: CLI init write safety
The `init` command SHALL use safe write semantics consistent with the existing CLI safety model.

#### Scenario: Init creates parent directory safely
- **WHEN** `.opencode/` does not exist and the user runs `lufy-ai init --target <dir>`
- **THEN** the CLI creates only the required `.opencode/` directory and `.opencode/project.yaml` inside the resolved target

#### Scenario: Init dry-run is not required
- **WHEN** the user runs `lufy-ai init --target <dir>`
- **THEN** the command may write `.opencode/project.yaml` because initialization is its explicit purpose, but it MUST NOT modify managed assets, install state, backups or unrelated files

#### Scenario: Init reports generated path
- **WHEN** `lufy-ai init` completes successfully
- **THEN** the CLI prints the generated config path and summary of detected stacks

### Requirement: CLI init validation
The implementation of `lufy-ai init` SHALL be validated with Go tests and fixture repositories for supported and unsupported stacks.

#### Scenario: Fixture tests cover supported stacks
- **WHEN** Go tests run for the CLI packages
- **THEN** fixtures verify Go, TypeScript/Next, JavaScript, Python, Java/Kotlin and multistack detection

#### Scenario: Fixture tests cover unsupported stacks
- **WHEN** Go tests run for the CLI packages
- **THEN** fixtures verify unsupported stacks such as Rust are emitted with `supported: false` rather than causing init failure

#### Scenario: Validation command covers init
- **WHEN** `scripts/validate.sh` runs after this change is implemented
- **THEN** the Go validation includes tests for `lufy-ai init` and still validates existing install/sync/verify behavior

### Requirement: CLI rescan drift reporting
The CLI Go SHALL expose `lufy-ai init --rescan` as the stack-aware project rescan mode that reports drift between `.opencode/project.yaml` and current repository evidence.

#### Scenario: Help describes rescan drift behavior
- **WHEN** the user requests help for `lufy-ai init`
- **THEN** the output describes `--rescan` as refreshing stack evidence, preserving user overrides and reporting drift without destructive cleanup

#### Scenario: Rescan delegates outside main
- **WHEN** `cmd/lufy-ai/main.go` receives `init --rescan`
- **THEN** it delegates scanning, drift comparison, merge planning, reporting and writing logic to internal packages instead of implementing that logic in `main.go`

#### Scenario: Rescan reports clean idempotent state
- **WHEN** the user runs `lufy-ai init --target <dir> --rescan` twice without target or config changes between runs
- **THEN** the second run exits successfully, reports no drift and does not create backups or modify unrelated install state

### Requirement: CLI rescan validation coverage
The implementation of `lufy-ai init --rescan` SHALL be validated with Go tests and confined filesystem fixtures for drift, stale detection and idempotency.

#### Scenario: Fixture tests cover rescan drift categories
- **WHEN** Go tests run for the CLI packages
- **THEN** fixtures verify at least no-drift, new stack drift, tooling drift, CI drift, stale stack detection, invalid existing config and unknown field preservation

#### Scenario: Validation command covers rescan
- **WHEN** `scripts/validate.sh` runs after this change is implemented
- **THEN** the Go validation includes tests for `lufy-ai init --rescan` and still validates existing install, sync, verify and init behavior
