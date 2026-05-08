## ADDED Requirements

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

## MODIFIED Requirements

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
