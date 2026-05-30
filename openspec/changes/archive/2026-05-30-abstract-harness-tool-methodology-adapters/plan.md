# Plan end-to-end: Lufy harness hexagonal y multi-tool

Este plan consolida el refactor completo para convertir `lufy-ai` en un harness engine neutral, desacoplado de una tool concreta y de una metodologia concreta.

El objetivo no es soportar muchas tools por cantidad. El objetivo es que Lufy tenga un core estable basado en tiers, roles, contratos, validacion y delivery; y que OpenCode, Codex, Claude Code u otra tool sean adapters.

## Estado objetivo

```text
Lufy Core
  - tiers T1/T2/T3
  - roles principales y subagentes
  - contratos compactos de salida
  - skill slots neutrales
  - result contract
  - validation gates
  - delivery gates
  - managed assets, drift, backup, restore, sync, verify

Tool Adapters
  - opencode
  - codex
  - claude-code

Methodology Adapters
  - openspec
  - lufy-sdd
  - none

Instruction Renderer
  - compone rol core + binding de tool + binding de metodologia
  - genera agentes, subagentes, commands, skills, templates y policies

Skill Registry
  - index-first
  - local-first
  - pasa paths exactos de SKILL.md
```

## Principios de arquitectura

1. El core no sabe de `.opencode`, `opencode.json`, `/opsx-*`, `openspec/`, `.claude`, `CLAUDE.md`, `.codex` ni paths especificos de una tool.
2. El tier sigue siendo el diferenciador principal de Lufy.
3. La metodologia se selecciona por tier.
4. `none` significa sin artefacto metodologico formal, no sin control.
5. Los agentes intercambian contratos compactos, no prompts largos ni historia completa.
6. Los skills se pasan como slots y paths exactos, no como resumen libre.
7. OpenCode/OpenSpec siguen siendo el preset default hasta que adapters nuevos esten validados.
8. Codex y Claude Code entran primero como dry-run/render verification, no como escritura real.

## Defaults iniciales

```yaml
tool: opencode
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

## Metodologias soportadas

Solo estas:

- `openspec`: metodologia actual.
- `lufy-sdd`: metodologia propia futura, con modos `full` y `lite`.
- `none`: ejecucion directa gobernada por Result Contract y validacion proporcional.

No se agregan metodologias externas hasta que una propuesta posterior justifique el costo.

## Tools objetivo

### OpenCode

Primer adapter real. Debe preservar compatibilidad actual:

- `.opencode/agents`
- `.opencode/commands`
- `.opencode/skills`
- `.opencode/templates`
- `.opencode/policies`
- `.opencode/plugins`
- `opencode.json`
- TUI plugin
- hooks

### Codex

Adapter futuro. Objetivo inicial:

- generar instrucciones compatibles con `AGENTS.md`;
- modelar ejecucion solo-agent cuando no haya subagentes nativos equivalentes;
- mapear skills/core contracts si la superficie local lo permite;
- dry-run primero.

### Claude Code

Adapter futuro. Objetivo inicial:

- generar `CLAUDE.md` y superficies compatibles;
- mapear commands/agents/settings segun capabilities reales;
- dry-run primero.

## Roles core

Roles principales:

- `orchestrator`: coordina y preserva estado.
- `router`: clasifica tier, metodologia, permissions, review workload y skill slots.
- `delivery`: ejecuta Git/remoto solo con autorizacion explicita.

Subagentes:

- `explorer`: analiza impacto.
- `implementer`: implementa cambios acotados.
- `test-writer`: escribe/revisa tests T1/T2.
- `validator`: valida sin editar.
- `reviewer`: revisa calidad y riesgo.

Cada rol tiene:

- permisos;
- skill slots;
- fallback inline para tools sin subagentes;
- output contract compacto;
- max handoff focus.

## Skill slots

Los roles no deben nombrar skills concretos. Deben nombrar slots:

- `methodology.explore`
- `methodology.propose`
- `methodology.apply`
- `methodology.verify`
- `methodology.sync`
- `methodology.archive`
- `delivery.pr_content`
- `delivery.git`
- `stack_config.lookup`
- `validation.grouped`
- `skill_registry.lookup`

El binding actual `opencode-openspec-current` mapea esos slots a skills reales:

- `openspec-explore`
- `openspec-propose`
- `openspec-apply-change`
- `openspec-verify-change`
- `openspec-sync`
- `openspec-archive-change`
- `pr.creator`
- `lufy.onboard`
- `lufy.timereport`

## Contratos compactos entre agentes

Cada rol debe responder con payload minimo:

| Rol | Payload minimo |
| --- | --- |
| `router` | tier, metodologia, permisos, review workload, skill slots, context slice |
| `explorer` | archivos relevantes, comportamiento actual, riesgos, limite de implementacion |
| `implementer` | archivos cambiados, comportamiento cambiado, comandos corridos, gaps |
| `test-writer` | archivos de test, conducta cubierta, fases TDD, gaps |
| `validator` | matriz de comandos, pass/fail, evidencia faltante, owner probable |
| `reviewer` | findings por severidad, score, pruebas faltantes, riesgo de release |
| `delivery` | autorizacion, branch state, staged scope, PR/check status, recovery |

Esto reduce tokens porque el siguiente agente recibe:

- decision;
- evidencia;
- skill paths exactos;
- gaps;
- siguiente owner.

No recibe prompts completos ni historia innecesaria.

## Arquitectura Go objetivo

```text
tools/lufy-cli-go/internal/
  core/
    domain/
      harness.go
      role.go
      asset.go
      result_contract.go
      workflow.go
    usecase/
      install.go
      sync.go
      verify.go
      status.go
      backup.go
      restore.go

  ports/
    adapters.go
    filesystem.go
    state_store.go
    backup_store.go
    catalog.go
    renderer.go
    skill_registry.go

  adapters/
    tool/
      opencode/
      codex/
      claude-code/
    methodology/
      openspec/
      lufy-sdd/
      none/
    filesystem/
      osfs/
    state/
      jsonstate/

  instructions/
    roles/
    bindings/
    render/
    checks/
```

## Backlog por etapas

### Etapa 0: Propuesta y contrato de programa

Estado: en curso.

Entregables:

- `proposal.md`
- `design.md`
- `tasks.md`
- specs delta
- este `plan.md`

Acceptance:

- `openspec validate abstract-harness-tool-methodology-adapters --strict`
- plan end-to-end revisable por humanos

### Etapa 1: Auditoria textual y gates de acoplamiento

Estado: implementado parcialmente.

Entregables:

- `textual-audit.md`
- `scripts/check-harness-coupling.sh`
- integracion en `scripts/validate.sh`

Acceptance:

- inventario de acoplamientos actuales;
- roles neutrales no pueden mencionar tool/metodologia concreta;
- `methodology/none` no puede mencionar OpenSpec ni `/opsx-*`;
- adapters futuros Codex/Claude no pueden depender de `.opencode`.

### Etapa 2: Contratos neutrales de roles y skills

Estado: en curso.

Entregables:

- `tools/lufy-cli-go/internal/instructions/roles/*.yaml`
- `tools/lufy-cli-go/internal/instructions/bindings/opencode-openspec-current/role-skills.yaml`
- `role-contracts.md`
- `internal/core/domain/role.go`

Acceptance:

- cada rol declara skill slots;
- cada rol declara contrato de salida compacto;
- binding actual mapea slots a skills reales;
- tests de dominio pasan.

### Etapa 3: Registry/loader de roles y skills

Objetivo:

Leer roles neutrales y bindings de skills para que el renderer no dependa de prompts sueltos.

Entregables:

- parser/loader de role contracts;
- parser/loader de role-skill bindings;
- validacion de slots sin resolver;
- tests con fixtures.

Acceptance:

- dado `role=delivery` y preset actual, el loader devuelve `delivery.pr_content -> pr.creator path`;
- dado `role=router`, no carga skills metodologicos directos;
- paths inexistentes opcionales reportan fallback, no error fatal.

### Etapa 4: Renderer de instruction surface

Objetivo:

Renderizar assets OpenCode actuales desde roles neutrales + binding tool + binding metodologia.

Entregables:

- `internal/instructions/render`;
- templates o ensamblado de agentes/subagentes;
- golden tests.

Acceptance:

- preset OpenCode/OpenSpec genera salida equivalente a assets actuales o diferencias justificadas;
- no cambia comportamiento publico de `lufy-ai install`;
- golden tests detectan cambios accidentales.

### Etapa 5: Tool capability registry

Objetivo:

Declarar capabilities por tool para no asumir subagentes, commands, skills o MCP en todas.

Entregables:

- `ToolCapabilities`;
- registry inicial con OpenCode;
- placeholders no-write para Codex y Claude Code.

Acceptance:

- OpenCode declara subagents, slash commands, skills, hooks, MCP, TUI, project/global config segun soporte actual;
- Codex/Claude Code quedan como unsupported/dry-run hasta implementacion posterior;
- tools no soportadas fallan explicitamente.

### Etapa 6: Methodology adapters

Objetivo:

Separar OpenSpec y `none` como metodologias.

Entregables:

- `adapters/methodology/openspec`;
- `adapters/methodology/none`;
- ID reservado `lufy-sdd`.

Acceptance:

- `openspec` renderiza assets metodologicos actuales;
- `none` no instala ni exige `openspec/` ni `/opsx-*`;
- T3 default usa `none` sin perder Result Contract ni validacion proporcional.

### Etapa 7: Tool adapter OpenCode

Objetivo:

Mover `.opencode/*`, `opencode.json`, TUI, hooks y config global a adapter OpenCode.

Entregables:

- `adapters/tool/opencode`;
- paths y config merge propios del adapter;
- verification checks propios de OpenCode.

Acceptance:

- `lufy-ai install` default conserva output actual;
- `lufy-ai install --tool opencode` equivale al default;
- tests de catalogo dejan de hardcodear OpenCode desde el core.

### Etapa 8: Manifest v2 y assets por componente

Objetivo:

Registrar ownership por tool, metodologia y componente.

Entregables:

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
      "tool": "opencode",
      "methodology": "openspec",
      "component": "instruction-surface"
    }
  ]
}
```

Acceptance:

- manifests v1 siguen verificando;
- manifests v2 registran component ownership;
- sync no mezcla assets de tool/metodologia incorrecta.

### Etapa 9: CLI flags de seleccion

Objetivo:

Permitir seleccionar tool/metodologia sin romper defaults.

Flags iniciales posibles:

```bash
lufy-ai install --tool opencode
lufy-ai install --methodology-tier T3:none
lufy-ai verify --tool opencode
lufy-ai status --json
```

Acceptance:

- `lufy-ai install` sigue funcionando igual;
- flags invalidos fallan con mensaje claro;
- `none` en T1/T2 requiere justificacion o se bloquea.

### Etapa 10: Codex dry-run adapter

Objetivo:

Preparar compatibilidad Codex sin mutar repos reales.

Entregables:

- capability profile Codex;
- render dry-run de instrucciones compatibles;
- mapping de solo-agent fallback;
- checks de fuga anti-OpenCode.

Acceptance:

- no escribe assets reales sin flag experimental;
- no menciona `.opencode`;
- documenta gaps de capabilities.

### Etapa 11: Claude Code dry-run adapter

Objetivo:

Preparar compatibilidad Claude Code sin mutar repos reales.

Entregables:

- capability profile Claude Code;
- render dry-run de instrucciones compatibles;
- mapping de commands/agents/settings cuando corresponda;
- checks de fuga anti-OpenCode.

Acceptance:

- no escribe assets reales sin flag experimental;
- no depende de `opencode.json`;
- documenta gaps de capabilities.

### Etapa 12: Lufy SDD lite/full

Objetivo:

Crear metodologia propia, primero lite y luego full.

Entregables esperados:

```text
.lufy/sdd/
  changes/
  specs/
  decisions/
  verification/
```

Acceptance:

- T2 puede usar `lufy-sdd/lite`;
- T1 puede usar `lufy-sdd/full`;
- migracion desde OpenSpec documentada;
- `none` sigue disponible para T3.

### Etapa 13: Documentacion publica y migracion

Objetivo:

Comunicar el nuevo modelo sin prometer adapters incompletos.

Entregables:

- README actualizado;
- architecture actualizado;
- getting-started actualizado;
- installation actualizado;
- backlog actualizado;
- ADR si aplica.

Acceptance:

- OpenCode/OpenSpec descritos como preset actual;
- Lufy descrito como harness engine;
- Codex/Claude descritos solo cuando existan al menos como dry-run validado.

### Etapa 14: Release readiness

Objetivo:

Preparar una version estable cuando el preset actual siga funcionando y el core hexagonal este listo.

Acceptance:

- `scripts/validate.sh` pasa;
- tests Go pasan;
- smokes de install/sync/verify pasan;
- golden tests del renderer pasan;
- PR checks remotos pasan;
- docs y assets embebidos sincronizados;
- manifest v1/v2 verificados.

## Orden recomendado de PRs

1. Proposal + plan + specs.
2. Auditoria textual + check de acoplamiento.
3. Roles neutrales + skill slots + output contracts.
4. Loader/registry de roles y skills.
5. Renderer inicial con golden tests.
6. Methodology adapters `openspec` y `none`.
7. Tool adapter OpenCode.
8. Manifest v2 y sync/verify/status.
9. CLI flags de seleccion.
10. Codex dry-run.
11. Claude Code dry-run.
12. Lufy SDD lite/full.

## Riesgos

- Crear abstraccion demasiado generica antes de tener uso real.
- Romper comportamiento actual de OpenCode/OpenSpec.
- Aumentar tokens si los contratos no son compactos.
- Mezclar tool adapter y methodology adapter.
- Permitir `none` en T1/T2 sin evidencia proporcional.
- Mantener docs y embedded assets fuera de sync.

## Mitigaciones

- OpenCode/OpenSpec queda default hasta el final.
- Adapters nuevos empiezan en dry-run.
- Roles neutrales no pueden mencionar tool/metodologia concreta.
- Skill slots pasan paths exactos, no resúmenes.
- Manifest v1 se lee durante la transicion.
- Validacion agrupada y golden tests antes de reportar readiness.
