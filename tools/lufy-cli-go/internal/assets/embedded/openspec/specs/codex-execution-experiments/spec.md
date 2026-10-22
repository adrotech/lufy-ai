# codex-execution-experiments

## Purpose

Definir experimentos locales auditables para Lufy en Codex sin adopción automática de variantes.

## Requirements

### Requirement: Evidence-gated execution experiments
El experimento SHALL comparar variantes conservando seguridad, aceptación y procedencia y SHALL NOT cambiar automáticamente el routing o los permisos del harness.

#### Scenario: Paired context delivery
- **WHEN** se compara fuente inline y por referencia
- **THEN** se conservan instrucciones y fuentes, se informa lo medido y se separa el experimento de una garantía general

#### Scenario: Evidence is insufficient or worse
- **WHEN** faltan muestras válidas, fidelidad o mejora reproducible
- **THEN** se mantiene el baseline, se registra no adopción y no se convierte el resultado inconcluso en aprobación

#### Scenario: Recovery preserves authorization boundaries
- **WHEN** se reutiliza un checkpoint o autorización previa
- **THEN** se exige idempotencia y alcance explícito, sin inferir permisos desde un next owner o hint

#### Scenario: Experiment attempts are exhausted
- **WHEN** termina el ensayo acotado con fallos o evidencia incompleta
- **THEN** se conserva el resultado, se detienen los reintentos y no se modifica el workflow productivo
