## Contexto

El sistema actual ya tiene piezas que pueden convertirse en una arquitectura neutral: la CLI Go separa `installer`, `syncer`, `verify`, `assets`, `state`, `backup`, `projectconfig` y `platform`; el harness ya separa roles y usa Result Contract; y `.opencode/project.yaml` ya captura metadata stack-aware y `workflow_limits`.

El problema es que el boundary de producto sigue mezclado con OpenCode/OpenSpec:

- El catálogo instala `.opencode/*`, `opencode.json`, `openspec/*` y assets relacionados como si fueran el dominio.
- Los agentes y skills dicen explícitamente cómo operar con OpenCode y OpenSpec.
- T1/T2/T3 están definidos por valor operativo, pero muchas instrucciones los expresan como comandos `/opsx-*`.

La migración debe empezar por separar lenguaje operativo y contratos antes de mover demasiada lógica Go.

## Arquitectura objetivo

```text
Lufy Core
  - tiers T1/T2/T3
  - roles y permisos
  - result contract
  - validation/delivery gates
  - policies
  - managed assets, drift, backup, restore, sync, verify

Tool Adapter
  - opencode inicial
  - codex futuro
  - claude-code futuro

Methodology Adapter
  - openspec inicial
  - none inicial
  - lufy-sdd futuro

Instruction Renderer
  - compone roles neutrales + tool binding + methodology binding
  - produce agentes, subagentes, commands, skills, templates y policies

Skill Registry
  - indexa SKILL.md por path exacto
  - local-first
  - evita resumir o compactar intención del skill
```

## Modelo neutral de roles

Los roles Lufy deben existir antes que sus renderizados por tool:

```yaml
role: explorer
kind: subagent
can_edit: false
can_shell: true
delivery_allowed: false
delegation:
  preferred: delegated_when_supported
  fallback: inline_phase
methodology_bindings:
  openspec:
    phase: explore
  lufy-sdd:
    phase: explore
  none:
    phase: direct_analysis
```

Roles principales:

- `orchestrator`: coordina, enruta y conserva estado; no implementa ni hace delivery.
- `router`: clasifica T1/T2/T3, selecciona metodología por tier y revisa capabilities de la tool.
- `delivery`: opera Git/GH solo con autorización explícita.

Subagentes:

- `explorer`: investigación read-only/focused.
- `implementer`: edición acotada.
- `test-writer`: pruebas T1/T2 cuando aplica.
- `validator`: validación read-only.
- `reviewer`: review stack-aware y riesgos.

## Tool adapter

`ToolAdapter` declara capabilities y paths. El core no debe consultar rutas concretas de OpenCode, Codex o Claude Code.

```go
type ToolAdapter interface {
    ID() ToolID
    Capabilities() ToolCapabilities
    Detect(ctx context.Context, env Env) DetectionResult
    RenderSurface(input HarnessModel) ([]AssetSpec, error)
    Verify(target Target) ([]Check, error)
}
```

Capabilities iniciales:

- `subagents`
- `slash_commands`
- `skills`
- `hooks`
- `mcp`
- `tui`
- `global_config`
- `project_config`
- `system_prompt`

OpenCode será el primer adapter real. Codex y Claude Code quedan como adapters futuros, primero en `dry-run`/diseño.

## Methodology adapter

La metodología se elige por tier. El core decide el rigor; el adapter genera o valida los artefactos.

```go
type MethodologyAdapter interface {
    ID() MethodologyID
    SupportedModes() []MethodologyMode
    RenderWorkflow(input WorkflowModel) ([]AssetSpec, error)
    VerifyWorkflow(target Target, tier Tier) ([]Check, error)
}
```

Metodologías permitidas inicialmente:

- `openspec`
- `lufy-sdd`
- `none`

`lufy-sdd` puede existir como ID reservado sin implementación runtime completa en este cambio.

## Routing por tier

Defaults iniciales compatibles:

```yaml
methodology_by_tier:
  T1:
    id: openspec
    mode: full
    required: true
  T2:
    id: openspec
    mode: lite
    required: true
  T3:
    id: none
    mode: none
    required: false
```

`none` no significa sin control. Significa:

- no proposal/spec/tasks persistente obligatorio;
- sí Result Contract;
- sí evidencia proporcional;
- sí riesgos y siguiente acción;
- sí stop rules y autorización de delivery.

## Instruction renderer

Los assets de texto no deben mantenerse como prompts irreductiblemente OpenCode/OpenSpec. Deben dividirse en:

- role core: reglas neutrales del rol;
- tool binding: cómo se expresa ese rol en una tool;
- methodology binding: qué operaciones metodológicas existen para ese tier;
- output binding: formato de reportes/result contracts.

El renderer debe producir o validar:

- agentes/subagentes;
- commands;
- skills;
- templates;
- policies;
- referencias en `AGENTS.md`/prompt raíz;
- assets embebidos.

## Skill registry portable

Tomar el enfoque index-first: el registry almacena descripción, scope y path exacto al `SKILL.md`. El agente o subagente lee el skill real solo cuando corresponde.

Esto evita romper la intención del skill con resúmenes parciales y permite compartir skills entre tools con distinta estructura.

## Manifest y compatibilidad

El manifest actual debe poder leerse como v1. La evolución v2 debe identificar origen de cada asset:

```json
{
  "schemaVersion": 2,
  "tool": "opencode",
  "methodologyByTier": {
    "T1": {"id": "openspec", "mode": "full"},
    "T2": {"id": "openspec", "mode": "lite"},
    "T3": {"id": "none", "mode": "none"}
  },
  "assets": [
    {
      "id": ".opencode/agents/orchestrator.md",
      "tool": "opencode",
      "methodology": "openspec",
      "component": "instruction-surface",
      "policy": "managed"
    }
  ]
}
```

## Migración por slices

1. Auditoría textual y specs.
2. Modelo neutral de roles y renderer con golden output equivalente.
3. Adapter OpenCode y Methodology OpenSpec/None sin cambio de UX.
4. Manifest v2 compatible y CLI flags.
5. Codex/Claude Code dry-run futuro.
6. `lufy-sdd` lite/full futuro.

## Riesgos

- Arquitectura Go limpia pero prompts todavía acoplados: mitigado con auditoría textual y leak checks.
- `none` mal usado en T1/T2: mitigado con default T3-only y justificación explícita para override.
- Manifest rompe instalaciones existentes: mitigado con lectura v1 y escritura v2 gradual.
- Demasiados adapters antes de estabilizar core: mitigado dejando solo OpenCode real al inicio.
