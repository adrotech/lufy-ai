## MODIFIED Requirements

### Requirement: Comandos base de instalación
La CLI SHALL implementar los comandos iniciales `install`, `verify`, `backup`, `restore` y `sync` antes de añadir comandos posteriores como `update`.

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

## ADDED Requirements

### Requirement: Comando sync de CLI Go
La CLI Go SHALL exponer `lufy-ai sync` como comando para sincronizar assets gestionados de forma segura en un target existente.

#### Scenario: Help incluye sync
- **WHEN** el usuario solicita ayuda de la CLI o del comando `sync`
- **THEN** la salida describe `sync`, sus flags soportados y que opera sobre assets gestionados con manifest/hash/backup

#### Scenario: Sync delega fuera de main
- **WHEN** `cmd/lufy-ai/main.go` recibe el comando `sync`
- **THEN** delega la lógica de negocio a paquetes internos en vez de implementar planificación o copia completa dentro de `main.go`

#### Scenario: Wrapper Bash no cambia para sync
- **WHEN** se inspecciona `scripts/install.sh` después de añadir `sync`
- **THEN** permanece como wrapper estricto de `lufy-ai install` y no contiene lógica propia ni fallback legacy para sincronizar assets

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
