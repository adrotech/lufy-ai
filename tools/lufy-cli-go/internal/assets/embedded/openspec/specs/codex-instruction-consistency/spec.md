# Codex instruction consistency

## Purpose
Mantener reglas compartidas y distribución Codex coherentes sin cambiar autoridad ni permisos.

## Requirements

### Requirement: Consistent Codex instruction distribution
Lufy SHALL mantener una fuente identificada de wording proporcional común y distribuirla inline sin divergencia, preservando autoridad de delivery.md y permisos particulares de roles.

#### Scenario: Drift detected
- **WHEN** una fuente, segmento compartido o mirror Codex diverge
- **THEN** el check offline rechaza y exige render/sync explícito

#### Scenario: Bounded rendering
- **WHEN** se renderiza el wording común
- **THEN** sólo se actualizan segmentos marcados y mirrors esperados, conservando instrucciones particulares y sandbox_mode

#### Scenario: Honest runtime guidance
- **WHEN** se consultan instrucciones de hooks o delegación
- **THEN** diferencian configuración y evidencia observada, conservan recovery de handoffs vacíos y no amplían permisos ni gates

#### Scenario: Instruction-only CI coverage
- **WHEN** un PR o push cambia sólo .codex o el renderer de instrucciones
- **THEN** se dispara la validación offline de distribución y drift
