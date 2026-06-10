# harness-adapter-architecture Specification

## Purpose
TBD - created by archiving change abstract-harness-tool-methodology-adapters. Update Purpose after archive.
## Requirements
### Requirement: Core harness neutrality
El core de `lufy-ai` SHALL modelar tiers, roles, policies, result contracts, validation gates, delivery gates y managed assets sin depender de nombres, rutas o comandos específicos de una tool o metodología.

#### Scenario: Core model avoids tool-specific paths
- **WHEN** se define o valida un modelo core de rol, tier, policy o asset
- **THEN** el modelo core SHALL NOT requerir rutas como `.opencode`, `.claude`, `.codex`, `opencode.json` u otras rutas propias de una tool

#### Scenario: Core model avoids methodology-specific commands
- **WHEN** se define o valida un modelo core de workflow
- **THEN** el modelo core SHALL NOT requerir comandos como `/opsx-*` ni paths `openspec/` como única forma de ejecución

### Requirement: Tool adapter registry
El sistema SHALL proveer un registry de tool adapters que declare identidad, detección, capabilities, paths, estrategias de configuración y renderizado de superficies operativas.

#### Scenario: OpenCode is registered as first-class adapter
- **WHEN** el usuario ejecuta `lufy-ai install` sin seleccionar tool explícita
- **THEN** el sistema SHALL resolver `opencode` como adapter default compatible con el comportamiento actual

#### Scenario: Unsupported tool is explicit
- **WHEN** el usuario selecciona una tool sin adapter implementado
- **THEN** el sistema SHALL fallar o degradar a dry-run con un mensaje explícito, sin instalar assets parciales ni asumir compatibilidad

### Requirement: Tool capability matrix
Cada tool adapter SHALL declarar capabilities como subagents, slash commands, skills, hooks, MCP, TUI, global config, project config y system prompt.

#### Scenario: Full delegation tool
- **WHEN** una tool declara `subagents: true`
- **THEN** el renderer MAY emitir roles como subagentes aislados para esa tool

#### Scenario: Solo-agent tool
- **WHEN** una tool declara `subagents: false`
- **THEN** el renderer SHALL producir un fallback inline que preserve fases, gates y Result Contract sin prometer aislamiento de subagente

### Requirement: Adapter-owned paths
Los paths concretos de configuración SHALL pertenecer al adapter de tool, no al core ni a componentes metodológicos.

#### Scenario: Component requests skills path
- **WHEN** un componente necesita instalar skills
- **THEN** SHALL consultar el adapter efectivo para obtener el directorio de skills en vez de hardcodear `.opencode/skills` u otro path

#### Scenario: Use cases request project config through tool runtime
- **WHEN** install, sync o verify necesitan planificar, aplicar o validar configuracion project-level de la tool efectiva
- **THEN** SHALL hacerlo mediante una capa runtime/adaptador de tool
- **AND** SHALL NOT invocar directamente servicios especificos de OpenCode desde el caso de uso

#### Scenario: Use cases request global config root through tool runtime
- **WHEN** install, sync, verify o status necesitan resolver config global por scope
- **THEN** SHALL hacerlo mediante una capa runtime/adaptador de tool
- **AND** SHALL preservar el path global actual para `opencode`

#### Scenario: Non writable tool runtime is explicit
- **WHEN** la capa runtime recibe `codex`, `claude-code` u otra tool sin escritura real autorizada
- **THEN** SHALL retornar un error explicito sin resolver paths OpenCode por fallback implicito

### Requirement: Backward-compatible default preset
El preset inicial `tool=opencode` y `methodology=openspec` SHALL conservar el comportamiento observable actual de `lufy-ai install`, `sync`, `verify` y `status` salvo cambios documentados por la propuesta.

#### Scenario: Existing default install
- **WHEN** un usuario ejecuta `lufy-ai install --target <repo> --yes --no-engram` sin flags nuevos
- **THEN** la instalación SHALL producir el preset OpenCode/OpenSpec compatible con versiones anteriores
- **AND** el manifest SHALL registrar `tool: opencode`
- **AND** los assets efectivos SHALL provenir de los adapters registrados para OpenCode y OpenSpec

#### Scenario: Effective catalog comes from adapters
- **WHEN** install, sync o verify calculan los assets gestionados del harness
- **THEN** SHALL resolver el catalogo efectivo desde `ToolAdapter.RenderSurface` y `MethodologyAdapter.RenderWorkflow`
- **AND** SHALL fallar explicitamente si el adapter requerido no existe

#### Scenario: Default install does not opt into Lufy SDD
- **WHEN** un usuario ejecuta `lufy-ai install --target <repo> --yes --no-engram` sin flags de metodología
- **THEN** el target SHALL contener los assets OpenCode/OpenSpec actuales
- **AND** SHALL NOT contener assets `.lufy/workflows/sdd`
- **AND** el manifest SHALL registrar `methodologyByTier` default con `openspec`

#### Scenario: Existing default install syncs after adapter routing
- **GIVEN** un target instalado con el preset default OpenCode/OpenSpec
- **WHEN** el usuario ejecuta `lufy-ai sync --target <repo> --yes --no-engram` sin flags nuevos
- **THEN** sync SHALL actualizar assets gestionados cuyo source cambio sin introducir `.lufy/workflows/sdd`
- **AND** SHALL preservar `tool: opencode` y `methodologyByTier` OpenSpec en el manifest
- **AND** `lufy-ai verify --target <repo> --no-engram` SHALL reportar una instalación válida

### Requirement: Manifest identifies adapter ownership
El manifest de instalación SHALL evolucionar para registrar tool, metodología, componente y scope de cada asset sin impedir leer manifests legacy.

#### Scenario: Legacy manifest remains readable
- **GIVEN** un target contiene `.lufy/managed-state/install-state.json` de schema anterior
- **WHEN** `lufy-ai verify` o `lufy-ai sync` se ejecuta
- **THEN** el sistema SHALL leer el manifest legacy y reportar cualquier limitación de migración sin romper por ausencia de campos nuevos

#### Scenario: New asset records origin
- **WHEN** un asset se registra en manifest v2
- **THEN** la entrada SHALL identificar al menos `tool`, `methodology` cuando aplique, `component`, `policy`, `scope` y hashes gestionados

### Requirement: CLI tool selection
La CLI SHALL permitir seleccionar explicitamente el tool adapter efectivo para comandos que instalan, sincronizan, verifican o reportan assets gestionados.

#### Scenario: Explicit OpenCode tool matches default
- **WHEN** el usuario ejecuta `lufy-ai install --tool opencode --target <repo> --yes --no-engram`
- **THEN** el sistema SHALL producir el mismo preset compatible que `lufy-ai install --target <repo> --yes --no-engram`
- **AND** el manifest SHALL registrar `tool: opencode`

#### Scenario: Unsupported write tool is rejected
- **WHEN** el usuario ejecuta un comando mutante con `--tool codex`, `--tool claude-code` u otra tool sin adapter escribible
- **THEN** el sistema SHALL fallar con error de uso explicito
- **AND** SHALL NOT instalar assets parciales ni asumir compatibilidad OpenCode

#### Scenario: Verify checks expected tool
- **GIVEN** un repo contiene manifest con `tool: opencode`
- **WHEN** el usuario ejecuta `lufy-ai verify --tool opencode --target <repo>`
- **THEN** la verificacion SHALL pasar el chequeo de tool esperado si el resto del estado es valido

#### Scenario: Verify rejects mismatched tool
- **GIVEN** un repo contiene manifest con una tool distinta a la esperada
- **WHEN** el usuario ejecuta `lufy-ai verify --tool opencode --target <repo>`
- **THEN** la verificacion SHALL reportar fallo por mismatch de tool

### Requirement: JSON reports expose harness context
Los reportes JSON de estado y verificacion SHALL exponer el contexto de harness efectivo para que otras tools puedan inspeccionarlo sin parsear texto humano.

#### Scenario: Status JSON includes adapter context
- **WHEN** el usuario ejecuta `lufy-ai status --json --target <repo>`
- **THEN** la salida JSON SHALL incluir `tool`, `methodologyByTier` y `schemaVersion` cuando exista manifest

#### Scenario: Verify JSON includes adapter context
- **WHEN** el usuario ejecuta `lufy-ai verify --json --target <repo>`
- **THEN** la salida JSON SHALL incluir `tool`, `methodologyByTier` y `schemaVersion` cuando exista manifest

### Requirement: Codex dry-run adapter
El sistema SHALL modelar `codex` como tool adapter conocido pero dry-run-only hasta que una propuesta posterior autorice escritura real.

#### Scenario: Codex adapter exposes conservative capabilities
- **WHEN** el registry resuelve el adapter `codex`
- **THEN** sus capabilities SHALL declarar `DryRunOnly=true`
- **AND** SHALL NOT declarar soporte nativo de subagents, slash commands, hooks, TUI o configuracion OpenCode

#### Scenario: Codex render is preview only
- **WHEN** se renderiza la superficie del adapter `codex`
- **THEN** la salida SHALL describir un preview compatible con `AGENTS.md`
- **AND** SHALL usar fallback inline para roles Lufy que no tengan subagentes nativos equivalentes
- **AND** SHALL NOT producir assets instalables reales para repos destino

#### Scenario: Codex adapter does not leak OpenCode paths
- **WHEN** se valida la salida del adapter `codex`
- **THEN** el check SHALL fallar si aparecen referencias a `.opencode`, `opencode.json` o paths propios de OpenCode

#### Scenario: Mutating CLI still blocks Codex
- **WHEN** el usuario ejecuta `lufy-ai install --tool codex`
- **THEN** la CLI SHALL fallar con error de uso explicito
- **AND** SHALL NOT escribir ni planificar assets Codex reales

### Requirement: Claude Code dry-run adapter
El sistema SHALL modelar `claude-code` como tool adapter conocido pero dry-run-only hasta que una propuesta posterior autorice escritura real.

#### Scenario: Claude Code adapter exposes conservative capabilities
- **WHEN** el registry resuelve el adapter `claude-code`
- **THEN** sus capabilities SHALL declarar `DryRunOnly=true`
- **AND** SHALL NOT declarar soporte nativo de subagents, slash commands, hooks, TUI o configuracion OpenCode

#### Scenario: Claude Code render is preview only
- **WHEN** se renderiza la superficie del adapter `claude-code`
- **THEN** la salida SHALL describir un preview compatible con `CLAUDE.md`
- **AND** SHALL usar fallback inline para roles Lufy que no tengan subagentes nativos equivalentes
- **AND** SHALL NOT producir assets instalables reales para repos destino

#### Scenario: Claude Code adapter does not leak OpenCode paths
- **WHEN** se valida la salida del adapter `claude-code`
- **THEN** el check SHALL fallar si aparecen referencias a `.opencode`, `opencode.json` o paths propios de OpenCode

#### Scenario: Mutating CLI still blocks Claude Code
- **WHEN** el usuario ejecuta `lufy-ai install --tool claude-code`
- **THEN** la CLI SHALL fallar con error de uso explicito
- **AND** SHALL NOT escribir ni planificar assets Claude Code reales
