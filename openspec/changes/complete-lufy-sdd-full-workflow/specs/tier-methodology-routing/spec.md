# tier-methodology-routing Specification Delta

## MODIFIED Requirements

### Requirement: Lufy SDD methodology adapter provides executable lite and full modes

El sistema SHALL proveer un adapter `lufy-sdd` instalable con modos `lite` y `full`; `full` SHALL incluir lifecycle nativo de changes y specs, mientras `lite` SHALL conservar un flujo acotado sin specs persistentes obligatorias.

#### Scenario: Lufy SDD full renders executable structure

- **WHEN** el registry renderiza `lufy-sdd/full`
- **THEN** declara configuración, changes, specs, decisions, verification, archive y acciones necesarias para ejecutar el lifecycle nativo

#### Scenario: Lufy SDD lite remains bounded

- **WHEN** el registry renderiza `lufy-sdd/lite`
- **THEN** declara configuración, changes, decisions y verification, pero no exige specs activas ni sync de deltas

#### Scenario: Adapter verification detects broken workflow

- **WHEN** `VerifyWorkflow` evalúa un target seleccionado con Lufy SDD y falta configuración, un directorio requerido o un artifact obligatorio inválido
- **THEN** retorna checks accionables con nivel y path en vez de un mensaje informativo fijo

#### Scenario: Mutating CLI accepts both supported modes

- **WHEN** el usuario selecciona `T1:lufy-sdd/full` o `T2:lufy-sdd/lite`
- **THEN** install y sync persisten la selección y filtran assets según el mode sin introducir OpenSpec como fallback implícito
