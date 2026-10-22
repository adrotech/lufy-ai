## MODIFIED Requirements

### Requirement: Numeric stop rules for workload guard

El workflow SHALL conservar checkpoints explícitos para impedir continuación silenciosa de trabajo amplio, sin alcance definido o sin evidencia recuperable. Los checkpoints de amplitud SHALL evaluar cobertura por plan/slice sin convertirse en clasificación de riesgo ni anular gates independientes.

Un archivo significativo SHALL ser aquel cuyo cambio requiere una decisión sustantiva sobre comportamiento, aceptación, contrato, configuración operacional o instrucciones del harness. El formato documental no lo excluye. La repetición mecánica demostrablemente equivalente no añade por sí misma una decisión de riesgo. Los mirrors sincronizados SHALL seguir visibles en diff y review aunque reflejen la misma decisión lógica.

#### Scenario: Four-file checkpoint has adequate coverage
- **GIVEN** un plan/slice vigente y pertinente cubre alcance, aceptación y validación
- **WHEN** el bloque alcanza cuatro o más archivos significativos dentro de ese alcance
- **THEN** registra la cobertura como decisión de workload y continúa sólo si los gates independientes están satisfechos
- **AND** no repite aprobación sólo por el recuento ni escala el tier sin nueva evidencia de riesgo

#### Scenario: Four-file rule pauses for workload decision
- **WHEN** se descubren cuatro o más archivos significativos sin plan/slice suficiente o con alcance ampliado no cubierto
- **THEN** pausa en una frontera segura, reporta evidencia y requiere decisión de workload/plan antes de continuar escrituras amplias

#### Scenario: Twenty-tool-calls rule pauses long routing
- **WHEN** un bloque supera veinte tool calls sin alcanzar estado recuperable
- **THEN** pausa para resumen compacto y decisión de continuar, compactar, escalar o dividir
- **AND** no usa el número aislado para inventar riesgo o autorización

#### Scenario: Multi-file write rule requires plan
- **WHEN** un paso propone escrituras sobre múltiples archivos no triviales
- **THEN** verifica plan/slice con alcance y aceptación observables antes de mutar ampliamente

#### Scenario: Long-session rule requires resumable handoff
- **WHEN** decisiones, evidencia o siguiente acción dejan de ser fácilmente recuperables
- **THEN** crea o solicita resumen/handoff antes de continuar

## ADDED Requirements

### Requirement: Workflow constraint provenance and precedence

El workflow SHALL preservar permisos, seguridad, aprobación post-plan y delivery como invariantes/gates independientes; SHALL consumir límites configurables sólo desde `workflow_limits` disponible; y SHALL identificar separadamente checkpoints del harness y configuración leída. Una restricción activa SHALL NOT desaparecer porque un plan cubra amplitud o un budget esté satisfecho.

#### Scenario: Configuration is unavailable
- **WHEN** falta configuración canónica de límites
- **THEN** reporta `not_available` para esos paths, mantiene guardrails declarados y no inventa budgets efectivos desde defaults del producto

#### Scenario: Covered plan meets an active stop rule
- **GIVEN** un plan cubre la amplitud del bloque y una stop rule independiente está activa
- **WHEN** el workflow evalúa si continuar
- **THEN** reporta `blocked` o `escalated` con regla, evidencia y recovery en vez de continuar por cobertura de plan

#### Scenario: Constraints disagree or are invalid
- **WHEN** configuración o instrucciones presentan una ambigüedad con consecuencias sobre autorización o seguridad
- **THEN** conserva las restricciones y pide diagnóstico/decisión concreta sin ampliar permisos ni continuar optimistamente

#### Scenario: Validated workload is not delivery authority
- **WHEN** un assessment de workload permite proceder
- **THEN** validación funcional, review requerido, autorización explícita y checks remotos siguen sujetos a sus propios gates
