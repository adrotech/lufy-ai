## ADDED Requirements

### Requirement: Explicit collaboration diagnostic profile
El runner SHALL ofrecer un perfil opt-in collaboration que conserve el aislamiento de usuario, registre sólo el reviewer del fixture por overrides y separe evidencia de transporte, rol, lifecycle y cierre. El perfil legacy SHALL conservar su aceptación.

#### Scenario: Structured recovery is observed
- **WHEN** callbacks registran spawn aceptado con reviewer explícito, inspección del resultado vacío, una recuperación al mismo target e inspección del envelope esperado
- **THEN** la faceta de transporte pasa sólo con identidades, orden y fingerprint coherentes, sin inferirlos del texto del modelo

#### Scenario: Cleanup is unavailable
- **WHEN** el perfil no expone una operación de cierre
- **THEN** cleanup queda not_available y el caso integral permanece inconcluso, aun con transporte válido

#### Scenario: Registration stays isolated
- **WHEN** se invoca el perfil collaboration
- **THEN** registra la copia del reviewer mediante overrides de esa invocación conservando ignore-user-config, read-only y never, sin copiar credenciales ni cambiar configuración global

#### Scenario: Missing or contradictory evidence
- **WHEN** faltan callbacks, hay otro target, payload/orden incorrecto, followups adicionales o una violación observada
- **THEN** no se acredita recuperación; las violaciones conservan precedencia failed y los gaps permanecen inconclusos
