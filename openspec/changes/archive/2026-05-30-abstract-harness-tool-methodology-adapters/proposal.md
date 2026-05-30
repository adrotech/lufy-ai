## Why

`lufy-ai` debe evolucionar de un instalador OpenCode/OpenSpec a un harness engine neutral. El valor diferencial del proyecto no es una tool concreta ni una metodología concreta, sino el routing proporcional por tiers T1/T2/T3, los roles operativos, la validación con evidencia, los gates de delivery y la gestión segura de assets.

Hoy esa intención está parcialmente acoplada a OpenCode y OpenSpec en tres lugares:

- CLI Go: catálogo, configuración, verificación y scopes usan rutas y contratos OpenCode/OpenSpec.
- Assets operativos: agentes, subagentes, commands, skills, templates y policies mencionan `.opencode`, `/opsx-*`, OpenCode y OpenSpec directamente.
- Documentación y specs: describen T1/T2/T3 como flujo OpenSpec/OpenCode en vez de separar core Lufy, tool adapter y methodology adapter.

El refactor debe preparar compatibilidad futura con tools como Codex o Claude Code y con metodologías `openspec`, `lufy-sdd` o `none`, sin romper el comportamiento actual.

## What Changes

- Introducir un modelo conceptual neutral de harness Lufy: tiers, roles, policies, result contract, validation gates, delivery gates y managed assets.
- Definir `tool adapters` como la capa que traduce el modelo Lufy a superficies concretas de una tool: agentes, subagentes, skills, slash commands, hooks, MCP, prompts, config y paths.
- Definir `methodology adapters` como la capa que traduce cada tier a artefactos metodológicos: `openspec`, `lufy-sdd` futuro o `none`.
- Separar metodología por tier:
  - T1: metodología full requerida.
  - T2: metodología lite requerida cuando hay comportamiento/riesgo.
  - T3: `none` permitido por defecto, sin artefacto metodológico formal.
- Crear un renderer de instrucciones que genere o valide agentes, subagentes, commands, skills, templates y policies desde roles neutrales más bindings de tool/metodología.
- Agregar un registry portable de skills basado en índice y paths exactos, para que los agentes carguen `SKILL.md` reales cuando los necesiten.
- Mantener `opencode` + `openspec` como preset/default inicial y conservar compatibilidad de `lufy-ai install`.

## Capabilities

### New Capabilities

- `harness-adapter-architecture`: arquitectura core + adapters + capabilities.
- `tier-methodology-routing`: selección de metodología por tier con `openspec`, `lufy-sdd` o `none`.
- `instruction-surface-rendering`: renderizado y validación de agentes/subagentes/skills/commands sin fugas de tool o metodología.
- `portable-skill-registry`: índice portable de skills con paths exactos y precedencia local-first.

### Modified Capabilities

- `sdd-harness-routing`: el tier sigue siendo la decisión central, pero deja de estar conceptualmente atado a OpenSpec.
- `project-stack-config`: deberá hospedar configuración neutral futura sin romper `.opencode/project.yaml` en el primer slice.
- `go-cli-installer`: deberá migrar progresivamente de catálogo OpenCode fijo a assets renderizados por adapters.
- `managed-assets-install`: deberá extender manifest/verify/sync para distinguir tool, metodología y componente.

## Impact

- Agentes principales: `orchestrator`, `sdd-router` y `delivery` deberán expresarse primero como roles Lufy neutrales.
- Subagentes: `explorer`, `implementer`, `test-writer`, `validator` y `reviewer` deberán depender de operaciones abstractas de workflow, no de `/opsx-*` ni `.opencode/*`.
- Skills: deberán clasificarse como core, methodology-specific o tool-specific.
- Commands: `/opsx-*` quedan como binding OpenSpec para OpenCode, no como contrato core del harness.
- Installer Go: requerirá adapter registry, capability registry, manifest v2 compatible y golden tests de assets renderizados.
- Docs: deberán comunicar que OpenCode/OpenSpec son el preset actual, no la identidad del producto.

## Non-Goals

- No implementar soporte real de Codex o Claude Code en este cambio inicial.
- No implementar todavía `lufy-sdd` full/lite.
- No reemplazar OpenSpec como default inmediato.
- No mover `.opencode/project.yaml` a `.lufy/config.yaml` en el primer slice.
- No reescribir todo el instalador en una única PR.
