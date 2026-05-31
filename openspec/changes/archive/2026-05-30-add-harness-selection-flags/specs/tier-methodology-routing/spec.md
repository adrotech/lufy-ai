## ADDED Requirements

### Requirement: CLI methodology tier override
La CLI SHALL permitir overrides explicitos de metodologia por tier sin desplazar el tier como decision central de gobernanza.

#### Scenario: T3 none override is accepted
- **WHEN** el usuario ejecuta `lufy-ai install --methodology-tier T3:none --target <repo> --yes --no-engram`
- **THEN** el sistema SHALL seleccionar `none/none/not-required` para T3
- **AND** SHALL preservar defaults compatibles para T1 y T2

#### Scenario: T2 OpenSpec lite override is accepted
- **WHEN** el usuario ejecuta `lufy-ai install --methodology-tier T2:openspec/lite --target <repo> --yes --no-engram`
- **THEN** el sistema SHALL seleccionar `openspec/lite/required` para T2

#### Scenario: Repeated tier overrides compose
- **WHEN** el usuario pasa multiples flags `--methodology-tier`
- **THEN** el sistema SHALL aplicar todos los overrides validos sobre los defaults
- **AND** el ultimo override de un mismo tier SHALL ganar de forma deterministica

### Requirement: Unsafe methodology override is blocked
La CLI SHALL bloquear overrides que reduzcan gobernanza de T1 o T2 a `none` mientras no exista justificacion persistida y validada por otro spec.

#### Scenario: T1 none is rejected
- **WHEN** el usuario ejecuta `lufy-ai install --methodology-tier T1:none`
- **THEN** el sistema SHALL fallar con error de uso
- **AND** SHALL indicar que T1 requiere metodologia full

#### Scenario: T2 none is rejected in CLI
- **WHEN** el usuario ejecuta `lufy-ai install --methodology-tier T2:none`
- **THEN** el sistema SHALL fallar con error de uso
- **AND** SHALL indicar que T2 requiere metodologia lite o justificacion futura

#### Scenario: Unsupported methodology is rejected
- **WHEN** el usuario ejecuta `lufy-ai install --methodology-tier T3:spec-kit`
- **THEN** el sistema SHALL fallar con error de uso
- **AND** SHALL listar `openspec` y `none` como valores operativos de este slice
