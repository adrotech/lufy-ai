## ADDED Requirements

### Requirement: Methodology selection by tier
El sistema SHALL seleccionar metodología por tier, manteniendo T1/T2/T3 como la decisión central de gobernanza y riesgo.

#### Scenario: T1 requires full methodology
- **WHEN** una solicitud se clasifica como T1
- **THEN** el workflow SHALL seleccionar una metodología full requerida, inicialmente `openspec/full` por default

#### Scenario: T2 uses lite methodology when needed
- **WHEN** una solicitud se clasifica como T2
- **THEN** el workflow SHALL seleccionar una metodología lite o handoff estructurado requerido cuando exista comportamiento, riesgo o trazabilidad que observar

#### Scenario: T3 may use none
- **WHEN** una solicitud se clasifica como T3
- **THEN** el workflow MAY seleccionar `none` como metodología y SHALL permitir implementación directa acotada con validación proporcional

### Requirement: Supported methodology set
El sistema SHALL limitar metodologías soportadas por configuración a `openspec`, `lufy-sdd` y `none` hasta que una propuesta posterior agregue otra explícitamente.

#### Scenario: Unsupported methodology is rejected
- **WHEN** la configuración solicita una metodología distinta de `openspec`, `lufy-sdd` o `none`
- **THEN** el sistema SHALL reportar configuración inválida y no SHALL inferir un adapter alternativo

### Requirement: None methodology remains governed
La metodología `none` SHALL significar ausencia de artefacto metodológico formal, no ausencia de control operativo.

#### Scenario: T3 none still reports evidence
- **WHEN** un workflow T3 usa metodología `none`
- **THEN** el resultado SHALL incluir evidencia proporcional, riesgos si existen, estado y siguiente acción en el Result Contract o salida final equivalente

#### Scenario: None does not install formal methodology assets
- **WHEN** una instalación selecciona `none` para un tier
- **THEN** el renderer SHALL NOT instalar assets metodológicos formales para ese tier, como comandos `/opsx-*` requeridos o directorios `openspec/` requeridos por ese tier

### Requirement: Methodology decisions are explicit in handoffs
El router y orchestrator SHALL reportar tool, metodología, mode y required status en decisiones sustantivas de workflow.

#### Scenario: Router emits methodology decision
- **WHEN** el router clasifica una solicitud no trivial
- **THEN** su salida SHALL incluir `tier`, `tool.id`, `methodology.id`, `methodology.mode` y `methodology.required`

#### Scenario: Result Contract carries adapter context
- **WHEN** un agente local completa un paso sustantivo
- **THEN** el Result Contract SHALL incluir el contexto efectivo de tool y metodología cuando aplique

### Requirement: Unsafe none override is blocked or justified
El sistema SHALL bloquear o exigir justificación explícita cuando se use `none` para T1 o T2.

#### Scenario: T1 none is unsafe by default
- **WHEN** una solicitud T1 intenta usar metodología `none`
- **THEN** el sistema SHALL bloquear el routing o exigir una justificación explícita registrada como riesgo

#### Scenario: T2 none requires bounded rationale
- **WHEN** una solicitud T2 intenta usar metodología `none`
- **THEN** el sistema SHALL permitirlo solo si el router documenta por qué no se requiere artefacto persistente y qué evidencia proporcional reemplaza la metodología
