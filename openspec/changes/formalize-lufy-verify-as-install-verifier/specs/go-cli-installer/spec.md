## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: `lufy-ai verify` canónico
La CLI Go SHALL usar `lufy-ai verify` como verificador canónico de instalaciones y MUST NOT requerir ni introducir `scripts/verify-install.sh`.

#### Scenario: Verificación estructural de categorías críticas
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir> --no-engram` sobre un target instalado
- **THEN** la CLI valida que `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/plugins` y `.opencode/policies` existen como directorios seguros no symlink

#### Scenario: Verificación de archivos críticos
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir> --no-engram` sobre un target instalado
- **THEN** la CLI valida que `.opencode/plugins/agent-observatory.tsx`, `AGENTS.md`, `tui.json`, `openspec/config.yaml` y `.lufy-ai/install-state.json` existen como archivos seguros no symlink

#### Scenario: Archivos críticos presentes en manifest
- **WHEN** un archivo crítico gestionado existe en el target pero no está registrado en `.lufy-ai/install-state.json`
- **THEN** `lufy-ai verify` falla indicando que el asset clave no está en el manifest

#### Scenario: Hashes de assets gestionados
- **WHEN** un asset listado en `.lufy-ai/install-state.json` existe pero su SHA-256 actual no coincide con `targetSHA256`
- **THEN** `lufy-ai verify` falla reportando drift con hashes abreviados expected/actual

#### Scenario: No existe script verificador paralelo
- **WHEN** se documenta o valida una instalación local/CI
- **THEN** la guía usa `lufy-ai verify` y no define `scripts/verify-install.sh` como objetivo ni dependencia
