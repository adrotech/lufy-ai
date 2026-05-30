## MODIFIED Requirements

### Requirement: Lufy SDD methodology adapter foundation
El sistema SHALL proveer una fundacion de adapter para `lufy-sdd` con modos `lite` y `full`, y SHALL permitir instalar sus assets metodologicos minimos cuando un tier seleccione esa metodologia.

#### Scenario: Lufy SDD adapter supports lite and full
- **WHEN** el registry resuelve la metodologia `lufy-sdd`
- **THEN** el adapter SHALL declarar soporte para `lite` y `full`
- **AND** SHALL NOT declarar soporte para `none`

#### Scenario: Lufy SDD full renders conceptual structure
- **WHEN** se renderiza `lufy-sdd/full`
- **THEN** la salida SHALL describir assets conceptuales bajo `.lufy/sdd/changes`, `.lufy/sdd/specs`, `.lufy/sdd/decisions` y `.lufy/sdd/verification`

#### Scenario: Lufy SDD lite renders bounded structure
- **WHEN** se renderiza `lufy-sdd/lite`
- **THEN** la salida SHALL describir assets conceptuales bajo `.lufy/sdd/changes`, `.lufy/sdd/decisions` y `.lufy/sdd/verification`
- **AND** SHALL NOT requerir `.lufy/sdd/specs`

#### Scenario: Mutating CLI accepts Lufy SDD selection
- **WHEN** el usuario ejecuta `lufy-ai install --methodology-tier T2:lufy-sdd/lite --target <repo> --yes --no-engram`
- **THEN** la CLI SHALL persistir `lufy-sdd/lite` en `methodologyByTier.T2`
- **AND** SHALL instalar assets minimos bajo `.lufy/sdd/`

#### Scenario: Lufy SDD full includes specs
- **WHEN** el usuario instala con `--methodology-tier T1:lufy-sdd/full`
- **THEN** el catalogo efectivo SHALL incluir `.lufy/sdd/specs`

#### Scenario: Lufy SDD lite omits specs when no full tier exists
- **WHEN** todos los tiers que usan `lufy-sdd` seleccionan `lite`
- **THEN** el catalogo efectivo SHALL NOT requerir `.lufy/sdd/specs`

#### Scenario: OpenSpec assets are methodology scoped
- **WHEN** ningun tier selecciona `openspec`
- **THEN** el catalogo efectivo SHALL NOT requerir `openspec/config.yaml`
