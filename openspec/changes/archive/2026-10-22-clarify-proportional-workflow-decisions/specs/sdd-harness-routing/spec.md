## MODIFIED Requirements

### Requirement: Numeric workload guard

El harness SHALL hacer observables las decisiones de workload usando volumen estimado y `workflow_limits` canónico disponible. El recuento de archivos o LOC SHALL NOT elevar el tier por sí solo: la clasificación SHALL conservar riesgo, incertidumbre e impacto como fundamento. Los gates de seguridad y autorización SHALL permanecer independientes.

#### Scenario: LOC budget requires workload decision
- **GIVEN** existe `workflow_limits.sizing.loc_budget`
- **WHEN** el router estima LOC superiores al presupuesto
- **THEN** emite `workload_decision_needed: true` y recomienda una decisión de workload antes de continuar
- **AND** cualquier escalado de tier identifica el riesgo o incertidumbre que lo justifica, no sólo el exceso numérico

#### Scenario: Five or more files trigger escalation or slicing
- **WHEN** el router estima cinco o más archivos
- **THEN** hace observable una decisión de workload y propone slices cuando reducen carga o riesgo
- **AND** conserva una única unidad justificada si fragmentar un cambio mecánico no aporta valor
- **AND** no eleva el tier por recuento aislado ni fragmenta T3 artificialmente

#### Scenario: Missing sizing config is not invented
- **GIVEN** falta `project.yaml` o `workflow_limits.sizing.loc_budget`
- **WHEN** el router evalúa workload
- **THEN** reporta `not_available` y no consume aliases legacy ni defaults inventados
- **AND** distingue los checkpoints declarados del harness de la configuración efectivamente leída

#### Scenario: A one-line security change retains its controls
- **WHEN** una modificación de una línea afecta auth runtime o configuración sensible
- **THEN** conserva routing previo, al menos T2 y fast path deshabilitado según las reglas de seguridad vigentes
- **AND** no usa el bajo volumen como evidencia para reducir controles

## ADDED Requirements

### Requirement: Proportional decision axes

El harness SHALL distinguir tier por riesgo/incertidumbre/impacto, workload por volumen/carga de review, tracking por recuperabilidad y paralelismo por independencia. Una decisión de workload o plan suficiente SHALL NOT conceder autorización, sustituir validación ni ampliar permisos.

#### Scenario: Mechanical multi-file change remains proportionate
- **GIVEN** un cambio entendido, mecánico y sin riesgo relevante abarca ocho archivos
- **WHEN** se seleccionan tier y controles
- **THEN** el tier no aumenta sólo por el recuento y los checkpoints de workload siguen evaluándose

#### Scenario: Existing plan does not grant implementation approval
- **GIVEN** un plan cubre la amplitud del bloque y falta la aprobación post-plan requerida
- **WHEN** el siguiente paso mutaría archivos de implementación
- **THEN** el workflow espera la aprobación y no usa cobertura, `auto-chain` o next owner como autorización

#### Scenario: Source and mirror remain visible in review
- **WHEN** una fuente gestionada y su mirror contienen la misma decisión sincronizada
- **THEN** el handoff puede referenciar una sola decisión lógica
- **AND** mantiene ambos archivos y su churn real en Git/review y verifica paridad sin excluirlos del assessment ejecutable

### Requirement: Codex proportional decision consistency

El contrato neutral, agentes Codex y sus mirrors SHALL expresar decisiones equivalentes sobre riesgo, volumen, cobertura y autorización para la superficie Codex. Las pruebas textuales y de paridad SHALL identificarse como tales y no como evaluación conductual de un modelo. Este change SHALL NOT exigir modificaciones a otros tool adapters.

#### Scenario: Managed instructions change
- **WHEN** cambia una regla proporcional distribuida
- **THEN** se actualizan consumidores afectados y mirrors exactos y se ejecutan pruebas de paridad/contrato pertinentes
- **AND** se mantienen controles de seguridad y aprobación post-plan

#### Scenario: Contract tests pass without agent evaluation
- **WHEN** pasan pruebas de presencia textual y paridad sin ejecutar un agente real
- **THEN** el resultado informa validación de instrucciones y no afirma cumplimiento conductual, menor latencia ni ahorro de tokens
