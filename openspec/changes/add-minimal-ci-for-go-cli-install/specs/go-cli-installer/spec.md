## MODIFIED Requirements

### Requirement: Validación por fases
La implementación SHALL incluir validación incremental con comandos reales disponibles después de introducir el toolchain Go, y SHALL ser ejecutable tanto localmente como desde CI mínima.

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
