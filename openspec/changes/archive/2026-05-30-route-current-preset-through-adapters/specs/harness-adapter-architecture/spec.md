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
