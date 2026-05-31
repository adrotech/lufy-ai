## ADDED Requirements

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
