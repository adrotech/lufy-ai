## ADDED Requirements

### Requirement: Explicit semantic review producer
Lufy SHALL registrar eventos semánticos únicamente por operaciones explícitas de review, con IDs hasheados, timestamps observados e idempotencia; lifecycle genérico SHALL NOT inferir reviews.

#### Scenario: Explicit review pair
- **WHEN** el host inicia y completa un intento con evidencia de reviewer válida
- **THEN** el ledger permite derivar un par review.started/review.completed sin duplicar retries

#### Scenario: Invalid or conflicting completion
- **WHEN** falta inicio/evidencia, el outcome contradice el contrato o una operación reutiliza identidad con contenido diferente
- **THEN** se rechaza sin registrar una conclusión semántica falsa

#### Scenario: Telemetry does not advance gates
- **WHEN** se registra una operación semántica
- **THEN** se conserva metadata content-free sin Checkpoint/KindFinish, sin cambiar permisos ni autorizar delivery

#### Scenario: Generic lifecycle remains generic
- **WHEN** se recibe SubagentStart, SubagentStop o SessionEnd sin operación explícita de review
- **THEN** no se fabrican review.started/completed/rework ni defect.reopened
