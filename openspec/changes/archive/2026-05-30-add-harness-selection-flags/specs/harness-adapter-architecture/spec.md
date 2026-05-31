## ADDED Requirements

### Requirement: CLI tool selection
La CLI SHALL permitir seleccionar explicitamente el tool adapter efectivo para comandos que instalan, sincronizan, verifican o reportan assets gestionados.

#### Scenario: Explicit OpenCode tool matches default
- **WHEN** el usuario ejecuta `lufy-ai install --tool opencode --target <repo> --yes --no-engram`
- **THEN** el sistema SHALL producir el mismo preset compatible que `lufy-ai install --target <repo> --yes --no-engram`
- **AND** el manifest SHALL registrar `tool: opencode`

#### Scenario: Unsupported write tool is rejected
- **WHEN** el usuario ejecuta un comando mutante con `--tool codex`, `--tool claude-code` u otra tool sin adapter escribible
- **THEN** el sistema SHALL fallar con error de uso explicito
- **AND** SHALL NOT instalar assets parciales ni asumir compatibilidad OpenCode

#### Scenario: Verify checks expected tool
- **GIVEN** un repo contiene manifest con `tool: opencode`
- **WHEN** el usuario ejecuta `lufy-ai verify --tool opencode --target <repo>`
- **THEN** la verificacion SHALL pasar el chequeo de tool esperado si el resto del estado es valido

#### Scenario: Verify rejects mismatched tool
- **GIVEN** un repo contiene manifest con una tool distinta a la esperada
- **WHEN** el usuario ejecuta `lufy-ai verify --tool opencode --target <repo>`
- **THEN** la verificacion SHALL reportar fallo por mismatch de tool

### Requirement: JSON reports expose harness context
Los reportes JSON de estado y verificacion SHALL exponer el contexto de harness efectivo para que otras tools puedan inspeccionarlo sin parsear texto humano.

#### Scenario: Status JSON includes adapter context
- **WHEN** el usuario ejecuta `lufy-ai status --json --target <repo>`
- **THEN** la salida JSON SHALL incluir `tool`, `methodologyByTier` y `schemaVersion` cuando exista manifest

#### Scenario: Verify JSON includes adapter context
- **WHEN** el usuario ejecuta `lufy-ai verify --json --target <repo>`
- **THEN** la salida JSON SHALL incluir `tool`, `methodologyByTier` y `schemaVersion` cuando exista manifest
