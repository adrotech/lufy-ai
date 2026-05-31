## ADDED Requirements

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
