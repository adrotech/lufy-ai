# portable-skill-registry Specification

## Purpose
TBD - created by archiving change abstract-harness-tool-methodology-adapters. Update Purpose after archive.
## Requirements
### Requirement: Portable skill registry
El sistema SHALL proveer un registry portable de skills que indexe skills disponibles por nombre, descripción, scope y path exacto a `SKILL.md`.

#### Scenario: Registry stores exact paths
- **WHEN** el registry detecta un skill
- **THEN** SHALL registrar el path exacto al `SKILL.md` fuente y SHALL NOT reemplazarlo por un resumen compacto como fuente de verdad

#### Scenario: Registry is project-local
- **WHEN** se refresca el registry para un proyecto
- **THEN** SHALL escribir un índice project-local bajo una ruta Lufy-managed o user-visible que pueda ser consumida por cualquier tool adapter compatible

### Requirement: Local-first skill precedence
El registry SHALL priorizar skills del proyecto sobre skills globales y SHALL mantener precedencia explícita.

#### Scenario: Project skill wins
- **GIVEN** existe un skill con el mismo nombre en el proyecto y en un directorio global de una tool
- **WHEN** el registry se refresca
- **THEN** el skill project-local SHALL tener prioridad y el registry SHALL preservar evidencia del path elegido

### Requirement: Tool-aware skill roots
El registry SHALL escanear raíces de skills declaradas por los adapters activos sin hardcodear solo `.opencode/skills`.

#### Scenario: OpenCode skill root
- **WHEN** el adapter efectivo declara un directorio de skills OpenCode
- **THEN** el registry SHALL escanear esa raíz además de skills project-locales configurados

#### Scenario: Future tool skill root
- **WHEN** un adapter futuro declara una raíz de skills compatible
- **THEN** el registry SHALL poder indexarla sin modificar el core del registry

### Requirement: Delegation passes skill paths
Los roles que delegan trabajo SHALL pasar paths exactos de skills relevantes a subagentes o fases inline en vez de inyectar resúmenes parciales.

#### Scenario: Subagent receives skill paths
- **WHEN** el router u orchestrator decide que un skill aplica a un paso
- **THEN** el handoff SHALL listar los paths exactos que el agente/subagente debe leer antes de ejecutar

#### Scenario: Skill intent is preserved
- **WHEN** un subagente necesita usar un skill
- **THEN** SHALL leer el `SKILL.md` original seleccionado por el registry para preservar su contrato completo

### Requirement: Registry refresh is non-destructive
El refresh del registry SHALL ser no destructivo respecto de skills user-owned.

#### Scenario: Refresh scans without mutating skills
- **WHEN** se ejecuta refresh del registry
- **THEN** el sistema SHALL leer metadata y escribir/actualizar solo el índice gestionado, sin modificar archivos `SKILL.md` existentes

