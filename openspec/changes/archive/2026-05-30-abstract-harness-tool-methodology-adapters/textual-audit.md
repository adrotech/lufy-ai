# Auditoria textual de acoplamientos harness/tool/metodologia

Este inventario inicia el primer slice de implementacion del refactor `abstract-harness-tool-methodology-adapters`.

Objetivo: distinguir que texto operativo pertenece al core Lufy, que texto es binding de tool, que texto es binding de metodologia y que texto queda como legacy transitorio. No intenta eliminar acoplamientos en este slice; los vuelve visibles y agrega un check para que las futuras superficies neutrales no los reintroduzcan.

## Taxonomia

| Categoria | Definicion | Ejemplos actuales |
| --- | --- | --- |
| `core` | Regla propia de Lufy, independiente de tool/metodologia. | T1/T2/T3, Result Contract, delivery gates, validacion proporcional, roles y permisos. |
| `tool-binding` | Texto necesario para expresar Lufy en una tool concreta. | `.opencode/agents`, `.opencode/skills`, `.opencode/project.yaml`, `opencode.json`, TUI plugin. |
| `methodology-binding` | Texto necesario para ejecutar una metodologia concreta. | OpenSpec, `openspec/`, `/opsx-*`, `openspec validate`, specs delta. |
| `legacy` | Texto que mezcla core/tool/metodologia y debe migrarse a renderer/bindings. | Agentes que hablan de T1/T2/T3 y a la vez de `/opsx-*` como contrato directo. |

## Superficies auditadas

| Superficie | Clasificacion actual | Observacion |
| --- | --- | --- |
| `.opencode/agents/orchestrator.md` | `legacy` | Mezcla coordinacion core, permisos, OpenCode, OpenSpec, `/opsx-*`, `.opencode/project.yaml` y delivery policy. |
| `.opencode/agents/sdd-router.md` | `legacy` | El tiering es core, pero execution modes y skill resolution estan expresados con OpenSpec y `.opencode/skills`. |
| `.opencode/agents/delivery.md` | `legacy` | Delivery gates son core; path `.opencode/policies/delivery.md` es tool-binding transitorio. |
| `.opencode/agents/explorer.md` | `legacy` | Rol core read-only con referencias OpenSpec/OpenCode para flujo instalado. |
| `.opencode/agents/implementer.md` | `legacy` | Rol core de edicion con reglas de OpenSpec/tasks y assets instalados. |
| `.opencode/agents/test-writer.md` | `legacy` | Rol core de testing stack-aware con dependencia a `.opencode/project.yaml`. |
| `.opencode/agents/validator.md` | `legacy` | Rol core read-only con validacion OpenSpec y workflow limits en `.opencode/project.yaml`. |
| `.opencode/agents/reviewer.md` | `legacy` | Rol core de review con metadata stack-aware en `.opencode/project.yaml`. |
| `.opencode/commands/opsx-*.md` | `methodology-binding` + `tool-binding` | Commands OpenSpec renderizados para OpenCode. |
| `.opencode/commands/lufy.*.md` | `tool-binding` | Commands Lufy renderizados para OpenCode. |
| `.opencode/skills/sdd-workflow/openspec-*` | `methodology-binding` + `tool-binding` | Skills OpenSpec instalados bajo superficie OpenCode. |
| `.opencode/skills/lufy.*` | `legacy` | Skills Lufy con paths OpenCode y futuras responsabilidades core/tool separables. |
| `.opencode/templates/sdd-lite.md` | `legacy` | T2 lite es core/metodologia; hoy vive como template OpenCode. |
| `.opencode/templates/result-contract.md` | `core` + `tool-binding` | El contrato es core; el path actual es OpenCode. |
| `.opencode/policies/delivery.md` | `core` + `tool-binding` | Politica core alojada en path OpenCode. |
| `AGENTS.md.template` | `legacy` | Template instalado mezcla instrucciones core con referencias `.opencode` y OpenSpec. |
| `lufy-ia.harness.md` | `legacy` | Harness compartido mezcla core, OpenSpec y `.opencode`. |
| `openspec/README.md` | `methodology-binding` | Documenta OpenSpec como metodologia instalada actual. |
| `tools/lufy-cli-go/internal/assets/catalog.go` | `tool-binding` + `methodology-binding` | Catalogo actual hardcodea `.opencode/*` y `openspec`. |
| `tools/lufy-cli-go/internal/config/config.go` | `tool-binding` | `opencode.json` y schema OpenCode. |
| `tools/lufy-cli-go/internal/platform/opencode.go` | `tool-binding` | Resolucion de config global OpenCode. |

## Reglas de migracion

- Las reglas T1/T2/T3, Result Contract, stop rules, role boundaries y delivery gates deben moverse a definiciones core neutrales.
- `.opencode/*`, `opencode.json`, TUI, hooks y commands OpenCode deben pasar a `tool=opencode`.
- `openspec/`, `/opsx-*`, `openspec validate` y delta specs deben pasar a `methodology=openspec`.
- `none` no debe mencionar OpenSpec ni `/opsx-*`.
- Adapters futuros `codex` y `claude-code` no deben depender de `.opencode` ni `opencode.json`.

## Check automatizado

Se agrega `scripts/check-harness-coupling.sh`.

El check:

- imprime inventario informativo de referencias actuales en superficies legacy;
- falla si futuras superficies neutrales contienen referencias de tool/metodologia;
- falla si `methodology/none` menciona OpenSpec, `openspec/` o `/opsx-*`;
- falla si adapters futuros `codex` o `claude-code` dependen de `.opencode` u `opencode.json`.

Este check se integra en `scripts/validate.sh` como gate temprano del programa.
