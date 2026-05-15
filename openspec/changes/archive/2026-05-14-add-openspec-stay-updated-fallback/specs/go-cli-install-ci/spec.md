## ADDED Requirements

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
