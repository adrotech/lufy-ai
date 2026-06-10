# go-cli-install-ci Specification

## Purpose
Definir el gate mínimo de GitHub Actions y validación local reproducible para la CLI Go `lufy-ai` y el wrapper estricto `scripts/install.sh`.
## Requirements
### Requirement: CI mínima del instalador Go
El sistema SHALL ejecutar una validación continua mínima para la CLI Go y el wrapper de instalación en GitHub Actions sobre PRs y pushes dirigidos a `develop` y `main`.

#### Scenario: Tests y build Go en CI
- **WHEN** se abre o actualiza un pull request hacia `develop` o `main` que afecta el repositorio
- **THEN** el workflow ejecuta `go test ./...` y `go build ./cmd/lufy-ai` desde `tools/lufy-cli-go/` sin depender de toolchains Node/TS en la raíz

#### Scenario: Pushes protegidos en develop y main
- **WHEN** hay un push en `develop` o `main` que afecta rutas cubiertas por el workflow
- **THEN** el workflow ejecuta el gate mínimo de tests, build, smokes, sanity OpenSpec condicional y `git diff --check`

#### Scenario: Ramas legacy no son base del gate normal
- **WHEN** se configura el trigger del workflow de instalación Go
- **THEN** no usa `development` ni `master` como ramas normales de PR/push para este flujo

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
- **WHEN** CI runs install/sync tests that write `.lufy/managed-state/install-state.json`
- **THEN** tests verify tool metadata and source fingerprint are populated from runtime/catalog sources rather than hardcoded proposal-era constants

#### Scenario: Atomic copy behavior covered
- **WHEN** CI runs install/sync/backup tests
- **THEN** tests cover that managed payload writes use the atomic write path or verify equivalent behavior through the copy helper

### Requirement: CI validates supply-chain release artifacts
The CI/release validation system SHALL verify that supply-chain metadata for release artifacts is produced and internally consistent before publishing or accepting release workflow changes.

#### Scenario: Release smoke includes verification metadata
- **WHEN** release artifact smoke validation runs
- **THEN** it confirms expected signature/provenance/SBOM files exist for the generated release artifact set or reports the missing files as a release blocker

#### Scenario: PR validation covers workflow syntax and local gates
- **WHEN** a pull request changes release or auto-tag workflows
- **THEN** local/CI validation includes syntax checks and the repository grouped validation command where applicable

### Requirement: CI fails on unpinned release-sensitive actions
The validation system SHALL detect release-sensitive workflow changes that use floating third-party action refs.

#### Scenario: Floating actions detected
- **WHEN** release-sensitive workflows contain `uses:` references to third-party actions without a commit SHA
- **THEN** validation fails with the workflow path and offending action reference

#### Scenario: Local validation can reproduce action pinning check
- **WHEN** a maintainer runs the grouped validation locally for a workflow change
- **THEN** the same action pinning check can run without requiring GitHub-hosted secrets or release publication

### Requirement: CI quality gates extend installer validation
The Go installer CI SHALL include quality gates beyond basic test/build while preserving the existing install and wrapper smoke coverage.

#### Scenario: Quality gates run before smokes
- **WHEN** CI validates a pull request affecting CLI, scripts, workflows or managed assets
- **THEN** coverage, lint/static checks and shell script lint run before expensive installer/wrapper smokes where practical

#### Scenario: Existing smoke coverage preserved
- **WHEN** quality gates are added to the workflow
- **THEN** dry-run install, real install, verify, idempotence, backup/restore and wrapper delegation smokes remain covered

### Requirement: Local validation remains reproducible
The repository SHALL expose local validation commands matching the CI quality gates as closely as practical.

#### Scenario: Grouped local validation includes quality gates
- **WHEN** a maintainer runs `scripts/validate.sh`
- **THEN** it runs PR-aware whitespace, action pinning, Go tests/build and available quality gates for the CLI scope

#### Scenario: Unavailable optional tools are reported
- **WHEN** a local machine lacks optional tools such as ShellCheck or golangci-lint
- **THEN** the validation output reports the missing tool instead of inventing success

### Requirement: CI validates OpenSpec resolver without network dependency
La validación local y CI SHALL probar el resolver OpenSpec usando fixtures locales en vez de depender de red externa.

#### Scenario: Resolver tests use local fixtures
- **WHEN** se ejecuta `go test ./...` desde `tools/lufy-cli-go/`
- **THEN** las pruebas del resolver cubren PATH simulado, cache válida, cache corrupta y fallback embebido sin acceso a red

#### Scenario: Grouped validation covers cache parity
- **WHEN** se ejecuta `scripts/validate.sh`
- **THEN** la validación incluye tests relevantes del resolver y mantiene la paridad de assets raíz/embebidos

### Requirement: Baseline sync workflow is safe for protected branches
El workflow `sync-openspec.yml` SHALL crear PRs de actualización sin pushear cambios directos a `develop` o `main`.

#### Scenario: Workflow opens pull request
- **WHEN** el workflow detecta una baseline upstream compatible más nueva
- **THEN** crea una rama dedicada y abre PR contra `develop` con resumen de cambios y evidencia de validación

#### Scenario: Workflow skips when no bump exists
- **WHEN** la versión efectiva ya coincide con upstream o no se puede determinar una actualización segura
- **THEN** el workflow termina sin cambios y reporta el motivo como salida visible
