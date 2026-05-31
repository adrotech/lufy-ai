## ADDED Requirements

### Requirement: Lufy SDD methodology adapter foundation
El sistema SHALL proveer una fundacion de adapter para `lufy-sdd` con modos `lite` y `full`, sin habilitarlo todavia como metodologia instalable por la CLI.

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

#### Scenario: Mutating CLI still blocks Lufy SDD selection
- **WHEN** el usuario ejecuta `lufy-ai install --methodology-tier T2:lufy-sdd/lite`
- **THEN** la CLI SHALL fallar con error de uso explicito
- **AND** SHALL NOT persistir `lufy-sdd` en manifest hasta que catalog y renderer lo soporten como instalable
