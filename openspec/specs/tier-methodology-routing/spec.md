# tier-methodology-routing Specification

## Purpose
TBD - created by archiving change abstract-harness-tool-methodology-adapters. Update Purpose after archive.
## Requirements
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

#### Scenario: Project config carries compatible methodology defaults
- **WHEN** `lufy-ai init` genera o reescanea la configuración project-local
- **THEN** la configuración SHALL incluir `tool: opencode` y `methodology_by_tier` con `T1=openspec/full/required`, `T2=openspec/lite/required` y `T3=none/none/not-required` salvo overrides compatibles preservados

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
- **THEN** el Result Contract SHALL incluir `adapter_context` con `tool_id`, `methodology_id`, `methodology_mode`, `methodology_required` y `execution_mode` cuando aplique

### Requirement: Unsafe none override is blocked or justified
El sistema SHALL bloquear o exigir justificación explícita cuando se use `none` para T1 o T2.

#### Scenario: T1 none is unsafe by default
- **WHEN** una solicitud T1 intenta usar metodología `none`
- **THEN** el sistema SHALL bloquear el routing o exigir una justificación explícita registrada como riesgo

#### Scenario: T2 none requires bounded rationale
- **WHEN** una solicitud T2 intenta usar metodología `none`
- **THEN** el sistema SHALL permitirlo solo si el router documenta por qué no se requiere artefacto persistente y qué evidencia proporcional reemplaza la metodología

### Requirement: CLI methodology tier override
La CLI SHALL permitir overrides explicitos de metodologia por tier sin desplazar el tier como decision central de gobernanza.

#### Scenario: T3 none override is accepted
- **WHEN** el usuario ejecuta `lufy-ai install --methodology-tier T3:none --target <repo> --yes --no-engram`
- **THEN** el sistema SHALL seleccionar `none/none/not-required` para T3
- **AND** SHALL preservar defaults compatibles para T1 y T2

#### Scenario: T2 OpenSpec lite override is accepted
- **WHEN** el usuario ejecuta `lufy-ai install --methodology-tier T2:openspec/lite --target <repo> --yes --no-engram`
- **THEN** el sistema SHALL seleccionar `openspec/lite/required` para T2

#### Scenario: Repeated tier overrides compose
- **WHEN** el usuario pasa multiples flags `--methodology-tier`
- **THEN** el sistema SHALL aplicar todos los overrides validos sobre los defaults
- **AND** el ultimo override de un mismo tier SHALL ganar de forma deterministica

### Requirement: Unsafe methodology override is blocked
La CLI SHALL bloquear overrides que reduzcan gobernanza de T1 o T2 a `none` mientras no exista justificacion persistida y validada por otro spec.

#### Scenario: T1 none is rejected
- **WHEN** el usuario ejecuta `lufy-ai install --methodology-tier T1:none`
- **THEN** el sistema SHALL fallar con error de uso
- **AND** SHALL indicar que T1 requiere metodologia full

#### Scenario: T2 none is rejected in CLI
- **WHEN** el usuario ejecuta `lufy-ai install --methodology-tier T2:none`
- **THEN** el sistema SHALL fallar con error de uso
- **AND** SHALL indicar que T2 requiere metodologia lite o justificacion futura

#### Scenario: Unsupported methodology is rejected
- **WHEN** el usuario ejecuta `lufy-ai install --methodology-tier T3:spec-kit`
- **THEN** el sistema SHALL fallar con error de uso
- **AND** SHALL listar `openspec` y `none` como valores operativos de este slice

### Requirement: Lufy SDD methodology adapter foundation
El sistema SHALL proveer una fundacion de adapter para `lufy-sdd` con modos `lite` y `full`, y SHALL permitir instalar sus assets metodologicos minimos cuando un tier seleccione esa metodologia.

#### Scenario: Lufy SDD adapter supports lite and full
- **WHEN** el registry resuelve la metodologia `lufy-sdd`
- **THEN** el adapter SHALL declarar soporte para `lite` y `full`
- **AND** SHALL NOT declarar soporte para `none`

#### Scenario: Lufy SDD full renders conceptual structure
- **WHEN** se renderiza `lufy-sdd/full`
- **THEN** la salida SHALL describir assets conceptuales bajo `.lufy/workflows/sdd/changes`, `.lufy/workflows/sdd/specs`, `.lufy/workflows/sdd/decisions` y `.lufy/workflows/sdd/verification`

#### Scenario: Lufy SDD lite renders bounded structure
- **WHEN** se renderiza `lufy-sdd/lite`
- **THEN** la salida SHALL describir assets conceptuales bajo `.lufy/workflows/sdd/changes`, `.lufy/workflows/sdd/decisions` y `.lufy/workflows/sdd/verification`
- **AND** SHALL NOT requerir `.lufy/workflows/sdd/specs`

#### Scenario: Mutating CLI accepts Lufy SDD selection
- **WHEN** el usuario ejecuta `lufy-ai install --methodology-tier T2:lufy-sdd/lite --target <repo> --yes --no-engram`
- **THEN** la CLI SHALL persistir `lufy-sdd/lite` en `methodologyByTier.T2`
- **AND** SHALL instalar assets minimos bajo `.lufy/workflows/sdd/`

#### Scenario: Lufy SDD full includes specs
- **WHEN** el usuario instala con `--methodology-tier T1:lufy-sdd/full`
- **THEN** el catalogo efectivo SHALL incluir `.lufy/workflows/sdd/specs`

#### Scenario: Lufy SDD lite omits specs when no full tier exists
- **WHEN** todos los tiers que usan `lufy-sdd` seleccionan `lite`
- **THEN** el catalogo efectivo SHALL NOT requerir `.lufy/workflows/sdd/specs`

#### Scenario: OpenSpec assets are methodology scoped
- **WHEN** ningun tier selecciona `openspec`
- **THEN** el catalogo efectivo SHALL NOT requerir `openspec/config.yaml`
