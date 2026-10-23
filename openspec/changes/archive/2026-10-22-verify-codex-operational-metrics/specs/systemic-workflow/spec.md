## ADDED Requirements

### Requirement: Conservative operational review metrics
El sistema SHALL derivar métricas de eventos content-free existentes sin confundir soporte estático con disponibilidad runtime.

#### Scenario: Independent task review pairs
- **WHEN** un run contiene reviews de tasks diferentes
- **THEN** los pares se emparejan por run y task sin cruzar timestamps; un par simple sin task conserva compatibilidad

#### Scenario: Ambiguous or incomplete review evidence
- **WHEN** existen inicios duplicados, huérfanos o timestamps ausentes/inválidos en review/rework/reopened
- **THEN** las métricas se degradan a partial y no se fabrica duración ni disponibilidad operativa

#### Scenario: Runtime support is distinct from event availability
- **WHEN** el adapter Codex soporta lifecycle pero no se observan eventos semánticos de review
- **THEN** se informa unavailable con recovery, sin derivar ceros operativos de evidencia sintética

#### Scenario: Metrics preserve ledger integrity
- **WHEN** se consultan métricas o se reingresa una clave idempotente equivalente
- **THEN** las fuentes permanecen inmutables y no contienen prompts, respuestas ni secretos
