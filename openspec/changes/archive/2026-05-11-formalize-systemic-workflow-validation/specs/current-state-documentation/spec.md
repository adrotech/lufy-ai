## ADDED Requirements

### Requirement: Documentacion del workflow sistemico
La documentacion operativa SHALL describir el workflow sistemico vigente como una practica de analisis inicial, implementacion por bloques, relectura final acotada y validacion final agrupada.

#### Scenario: Documentacion refleja fases reales
- **WHEN** una persona revisa guias operativas como `AGENTS.md`, `.opencode/policies/delivery.md` u OpenSpec docs relevantes
- **THEN** encuentra que el analisis de archivos existentes ocurre al inicio, las relecturas se evitan salvo justificacion, y tests/coverage se agrupan al final de la propuesta cuando apliquen

#### Scenario: Documentacion conserva limites de toolchain
- **WHEN** la documentacion menciona tests, coverage o validacion final
- **THEN** aclara que solo se ejecutan comandos reales disponibles para el alcance y que las limitaciones se reportan explicitamente
