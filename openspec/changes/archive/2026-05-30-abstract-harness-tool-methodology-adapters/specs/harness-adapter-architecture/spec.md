## ADDED Requirements

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

### Requirement: Backward-compatible default preset
El preset inicial `tool=opencode` y `methodology=openspec` SHALL conservar el comportamiento observable actual de `lufy-ai install`, `sync`, `verify` y `status` salvo cambios documentados por la propuesta.

#### Scenario: Existing default install
- **WHEN** un usuario ejecuta `lufy-ai install --target <repo> --yes --no-engram` sin flags nuevos
- **THEN** la instalación SHALL producir el preset OpenCode/OpenSpec compatible con versiones anteriores

### Requirement: Manifest identifies adapter ownership
El manifest de instalación SHALL evolucionar para registrar tool, metodología, componente y scope de cada asset sin impedir leer manifests legacy.

#### Scenario: Legacy manifest remains readable
- **GIVEN** un target contiene `.lufy-ai/install-state.json` de schema anterior
- **WHEN** `lufy-ai verify` o `lufy-ai sync` se ejecuta
- **THEN** el sistema SHALL leer el manifest legacy y reportar cualquier limitación de migración sin romper por ausencia de campos nuevos

#### Scenario: New asset records origin
- **WHEN** un asset se registra en manifest v2
- **THEN** la entrada SHALL identificar al menos `tool`, `methodology` cuando aplique, `component`, `policy`, `scope` y hashes gestionados
