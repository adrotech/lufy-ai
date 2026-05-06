# go-cli-install-ci Specification

## Purpose
Definir el gate mínimo de GitHub Actions y validación local reproducible para la CLI Go `lufy-ai` y el wrapper estricto `scripts/install.sh`.

## Requirements
### Requirement: CI mínima del instalador Go
El sistema SHALL ejecutar una validación continua mínima para la CLI Go y el wrapper de instalación en GitHub Actions.

#### Scenario: Tests y build Go en CI
- **WHEN** se abre o actualiza un pull request que afecta el repositorio
- **THEN** el workflow ejecuta `go test ./...` y `go build ./cmd/lufy-ai` desde `tools/lufy-cli-go/` sin depender de toolchains Node/TS en la raíz

#### Scenario: Smoke de instalación en target temporal
- **WHEN** el workflow compila el binario `lufy-ai`
- **THEN** ejecuta un smoke en un directorio temporal que cubre dry-run sin mutaciones, install real, `verify`, idempotencia básica, `backup` y `restore`

#### Scenario: Wrapper Bash validado por CI
- **WHEN** existe `tools/lufy-cli-go/bin/lufy-ai` construido durante el job
- **THEN** el workflow ejecuta `scripts/install.sh` contra un target temporal y confirma que delega en la CLI Go sin usar fallback legacy

#### Scenario: CI portable sin Engram obligatorio
- **WHEN** el workflow ejecuta smokes del instalador
- **THEN** usa `--no-engram` para no depender de que `engram` exista en el runner

### Requirement: Validación local reproducible
El sistema SHALL documentar cómo ejecutar localmente la validación mínima equivalente a la CI.

#### Scenario: Documentación de comandos locales
- **WHEN** un mantenedor quiere reproducir el gate antes de abrir un pull request
- **THEN** la documentación lista los comandos locales para tests, build y smoke del instalador Go

#### Scenario: Sin comandos inventados de raíz
- **WHEN** la documentación describe validación local
- **THEN** no presenta `npm test`, `tsc` u otros comandos globales de raíz como requeridos para este gate
