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
