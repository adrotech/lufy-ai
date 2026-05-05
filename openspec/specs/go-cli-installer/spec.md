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
La CLI SHALL implementar los comandos iniciales `install`, `verify`, `backup` y `restore` antes de añadir comandos posteriores como `sync` o `update`.

#### Scenario: Install con flags mínimos
- **WHEN** el usuario ejecuta `lufy-ai install --target . --dry-run --yes --no-engram`
- **THEN** la CLI construye un plan de instalación para el target actual, omite Engram y no escribe archivos por estar en dry-run

#### Scenario: Verify de un target
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir>`
- **THEN** la CLI valida estructura instalada, archivos esperados, JSON parseable cuando aplique y estado de integración Engram según flags

#### Scenario: Backup explícito
- **WHEN** el usuario ejecuta `lufy-ai backup --target <dir>`
- **THEN** la CLI crea un backup con manifest de los archivos gestionados o relevantes para rollback dentro del target

#### Scenario: Restore desde manifest
- **WHEN** el usuario ejecuta `lufy-ai restore --target <dir> --backup <manifest-or-dir>`
- **THEN** la CLI valida el manifest y restaura los archivos registrados de forma controlada

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

### Requirement: Plan de instalación antes de escribir
El comando `install` SHALL construir un plan explícito antes de modificar el filesystem.

#### Scenario: Plan de acciones
- **WHEN** `install` analiza un target
- **THEN** el plan enumera acciones como crear directorio, copiar asset, omitir archivo idéntico, mergear JSON, crear backup o reportar conflicto

#### Scenario: Plan con conflictos
- **WHEN** existe un archivo distinto en el target y no hay merge seguro
- **THEN** el plan marca el conflicto y no lo sobrescribe silenciosamente

### Requirement: Idempotencia y preservación de trabajo del usuario
La instalación SHALL ser idempotente y MUST NOT sobrescribir trabajo del usuario sin estrategia explícita, backup y confirmación cuando corresponda.

#### Scenario: Reinstalación sin cambios
- **WHEN** el usuario ejecuta `install` dos veces sobre el mismo target sin modificaciones intermedias
- **THEN** la segunda ejecución reporta archivos idénticos como `skip` o equivalente y no produce conflictos falsos

#### Scenario: AGENTS.md existente
- **WHEN** el target ya contiene `AGENTS.md`
- **THEN** la CLI preserva el archivo, reporta la decisión y solo propone backup/merge si hay una estrategia segura

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
La CLI SHALL crear o mergear `opencode.json` mediante JSON válido, preservando claves desconocidas del usuario.

#### Scenario: Crear opencode.json faltante
- **WHEN** el target no contiene `opencode.json` y la instalación requiere configuración OpenCode
- **THEN** la CLI crea un JSON válido con las claves gestionadas por `lufy-ai`

#### Scenario: Preservar claves existentes
- **WHEN** el target contiene `opencode.json` válido con claves no gestionadas
- **THEN** la CLI preserva esas claves y modifica solo secciones gestionadas por `lufy-ai`

#### Scenario: JSON inválido
- **WHEN** el target contiene `opencode.json` inválido
- **THEN** la CLI falla sin sobrescribirlo y reporta una instrucción accionable para corregir o respaldar el archivo

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
La implementación SHALL incluir validación incremental con comandos reales disponibles después de introducir el toolchain Go.

#### Scenario: Validación Go disponible
- **WHEN** existen `tools/lufy-cli-go/go.mod` y paquetes Go
- **THEN** el implementador puede ejecutar `go test ./...` y `go build ./cmd/lufy-ai` desde `tools/lufy-cli-go/` como validación mínima

#### Scenario: Prueba dry-run en temp dir
- **WHEN** el binario Go compila
- **THEN** el implementador puede ejecutar una instalación `--dry-run` contra un directorio temporal y confirmar que no se escriben archivos de instalación

#### Scenario: Verify tras instalación temporal
- **WHEN** una instalación real se ejecuta en un directorio temporal de prueba
- **THEN** `lufy-ai verify --target <temp>` valida el resultado sin depender de modificar el repositorio fuente
