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

### Requirement: Release artifact CI
The CI system SHALL build and validate versioned release artifacts for the `lufy-ai` CLI.

#### Scenario: Matrix build for release artifacts
- **WHEN** the release workflow runs for an authorized tag or release event
- **THEN** it builds `lufy-ai` artifacts from `tools/lufy-cli-go/` for the supported OS/arch matrix without depending on root Node/TS tooling

#### Scenario: Version metadata injected
- **WHEN** the release workflow builds binaries
- **THEN** it injects version, commit and build date metadata consumed by `lufy-ai version`

### Requirement: Release smoke validation
The CI system SHALL validate release artifacts before they are published as installable artifacts.

#### Scenario: Artifact version smoke
- **WHEN** a release artifact is built for the runner platform or can be executed in CI
- **THEN** CI runs `lufy-ai version` from the packaged artifact and confirms the expected version metadata is present

#### Scenario: Install smoke from artifact
- **WHEN** a release artifact is executable in CI
- **THEN** CI runs at least `install --dry-run`, a temporary install and `verify --target <temp> --no-engram` using the packaged binary or unpacked release artifact

#### Scenario: Checksum smoke
- **WHEN** release artifacts and checksum files are generated
- **THEN** CI recalculates SHA-256 hashes and confirms they match the published checksum entries

### Requirement: Bootstrap CI validation
The CI system SHALL validate the bootstrap installer without requiring live mutation of user machines.

#### Scenario: Bootstrap dry-run
- **WHEN** bootstrap installer code changes
- **THEN** CI or local validation can run a dry-run or fixture-backed mode that resolves OS/arch, version and artifact URL without writing outside a temporary directory

#### Scenario: Bootstrap checksum failure test
- **WHEN** bootstrap validation runs against a fixture with an incorrect checksum
- **THEN** the test confirms installation is blocked and the binary is not executed

### Requirement: Structural hardening checks in CI
The Go installer CI SHALL include checks that protect structural hardening guarantees for assets, paths, state metadata and atomic writes.

#### Scenario: Catalog parity covered by tests
- **WHEN** CI runs Go tests for `tools/lufy-cli-go/`
- **THEN** tests fail if the root managed asset catalog and embedded catalog drift for target paths, policies or source hashes

#### Scenario: Windows traversal semantics covered
- **WHEN** CI runs path safety tests
- **THEN** tests include traversal inputs with Windows separators or mixed separators and verify they are rejected

#### Scenario: State metadata covered
- **WHEN** CI runs install/sync tests that write `.lufy-ai/install-state.json`
- **THEN** tests verify tool metadata and source fingerprint are populated from runtime/catalog sources rather than hardcoded proposal-era constants

#### Scenario: Atomic copy behavior covered
- **WHEN** CI runs install/sync/backup tests
- **THEN** tests cover that managed payload writes use the atomic write path or verify equivalent behavior through the copy helper
