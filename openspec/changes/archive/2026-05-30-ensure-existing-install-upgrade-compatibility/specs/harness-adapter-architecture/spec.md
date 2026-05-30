## MODIFIED Requirements

### Requirement: Backward-compatible default preset
El preset inicial `tool=opencode` y `methodology=openspec` SHALL conservar el comportamiento observable actual de `lufy-ai install`, `sync`, `verify` y `status` salvo cambios documentados por la propuesta.

#### Scenario: Existing default install
- **WHEN** un usuario ejecuta `lufy-ai install --target <repo> --yes --no-engram` sin flags nuevos
- **THEN** la instalación SHALL producir el preset OpenCode/OpenSpec compatible con versiones anteriores
- **AND** el manifest SHALL registrar `tool: opencode`
- **AND** los assets efectivos SHALL provenir de los adapters registrados para OpenCode y OpenSpec

#### Scenario: Effective catalog comes from adapters
- **WHEN** install, sync o verify calculan los assets gestionados del harness
- **THEN** SHALL resolver el catalogo efectivo desde `ToolAdapter.RenderSurface` y `MethodologyAdapter.RenderWorkflow`
- **AND** SHALL fallar explicitamente si el adapter requerido no existe

#### Scenario: Default install does not opt into Lufy SDD
- **WHEN** un usuario ejecuta `lufy-ai install --target <repo> --yes --no-engram` sin flags de metodología
- **THEN** el target SHALL contener los assets OpenCode/OpenSpec actuales
- **AND** SHALL NOT contener assets `.lufy/sdd`
- **AND** el manifest SHALL registrar `methodologyByTier` default con `openspec`

#### Scenario: Existing default install syncs after adapter routing
- **GIVEN** un target instalado con el preset default OpenCode/OpenSpec
- **WHEN** el usuario ejecuta `lufy-ai sync --target <repo> --yes --no-engram` sin flags nuevos
- **THEN** sync SHALL actualizar assets gestionados cuyo source cambio sin introducir `.lufy/sdd`
- **AND** SHALL preservar `tool: opencode` y `methodologyByTier` OpenSpec en el manifest
- **AND** `lufy-ai verify --target <repo> --no-engram` SHALL reportar una instalación válida
