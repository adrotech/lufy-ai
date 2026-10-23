## ADDED Requirements

### Requirement: Comparable Codex efficiency samples
Lufy SHALL comparar muestras con aceptación idéntica y procedencia observada estable de ejecutable, fuentes, instrucciones y perfil de ejecución, sin adoptar variantes automáticamente.

#### Scenario: Stable complete pair
- **WHEN** ambas muestras conservan procedencia y aceptación válidas, con métricas observadas completas
- **THEN** se comparan bytes, tokens de entrada y duración sin inventar valores ausentes

#### Scenario: Drift or incomplete sample
- **WHEN** cambia procedencia/perfil o un caso falla o queda incompleto
- **THEN** no es elegible, se detienen nuevas llamadas y se conserva el baseline

#### Scenario: Sample supports candidate
- **WHEN** dos pares válidos mejoran estrictamente bytes y tokens sin mayor duración
- **THEN** sólo se declara soporte en esa muestra y adopted permanece false

#### Scenario: Honest negative result
- **WHEN** el ensayo real no reúne evidencia suficiente
- **THEN** se publica esa limitación sin reintentos para obtener verde ni cambios globales
