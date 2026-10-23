# codex-native-regression

## Purpose

Definir regresiones nativas observables y acotadas de Lufy en Codex.

## Requirements

### Requirement: Observed native regression
El protocolo SHALL comprobar delegación y recuperación mediante callbacks/herramientas observados, con identidad coherente del reviewer y sin inferir capacidades de flags o afirmaciones del modelo.

#### Scenario: Native recovery completes
- **WHEN** un reviewer nativo devuelve un handoff vacío y una recuperación del mismo agente produce el envelope sintético validado
- **THEN** el evaluador exige hooks, waits ordenados, fingerprint y cierre observado antes de aceptar el caso

#### Scenario: Runtime lacks traceable capabilities
- **WHEN** faltan callbacks, herramientas, identidad coherente, payload válido o fin de sesión
- **THEN** el caso queda inconcluso y no se convierte en emulación ni éxito

#### Scenario: Observed violation takes precedence
- **WHEN** se observa shell, edición o cambio de fuentes del fixture
- **THEN** el caso falla aun si también existen gaps de trazabilidad

### Requirement: Isolated vetted fixture
El ensayo live SHALL ser opt-in, acotado y usar sólo hooks/instrucciones del fixture revisados con procedencia explícita, sin modificar confianza, modelos ni permisos globales.

#### Scenario: Vetted invocation-only hooks
- **WHEN** la automatización ejecuta los hooks revisados del fixture
- **THEN** registra hashes y condición de confianza temporal, conserva sandbox read-only y no afirma que hooks no confiables funcionen por defecto

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

#### Scenario: Bounded wait expires before recovery completes
- **WHEN** wait_agent informa timed_out booleano true sin error explícito y una inspección posterior del mismo agente confirma el envelope esperado
- **THEN** el timeout no invalida la faceta de transporte ni acredita recuperación por sí solo; errores, respuestas desconocidas o falta de inspección posterior mantienen el caso inconcluso y la aceptación legacy no cambia
