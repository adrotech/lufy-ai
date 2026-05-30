# Contratos compactos de roles y skills

Este slice agrega la parte que evita consumo innecesario de tokens entre agentes: cada rol declara sus `skill_slots`, su contrato de salida y el payload minimo que debe transferir al siguiente rol.

## Regla principal

Los roles neutrales no nombran skills concretos de una tool o metodologia. Declaran slots:

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

El binding efectivo traduce esos slots a paths concretos.

Para el preset actual, el binding vive en:

```text
tools/lufy-cli-go/internal/instructions/bindings/opencode-openspec-current/role-skills.yaml
```

## Skill bindings actuales

| Slot neutral | Skill actual | Uso |
| --- | --- | --- |
| `methodology.explore` | `openspec-explore` | Exploracion T1 cuando la metodologia requiere artefacto. |
| `methodology.propose` | `openspec-propose` | Propuesta full para T1. |
| `methodology.apply` | `openspec-apply-change` | Implementacion desde tareas metodologicas. |
| `methodology.verify` | `openspec-verify-change` | Verificacion contra artefactos metodologicos. |
| `methodology.sync` | `openspec-sync` | Sync de deltas validados. |
| `methodology.archive` | `openspec-archive-change` | Archivo de cambio completado. |
| `delivery.pr_content` | `pr.creator` | Redaccion estructurada de PR, sin delivery. |
| `delivery.git` | `git-delivery` | Optional; si falta, delivery usa policy local. |
| `lufy.onboard` | `lufy.onboard` | Skill invocado por comando de onboarding. |
| `lufy.timereport` | `lufy.timereport` | Skill invocado por comando de reporte. |

## Contrato de salida por rol

| Rol | Payload minimo | Proximo consumidor |
| --- | --- | --- |
| `orchestrator` | `workflow_decision`, `context_slice`, `next_recommended`, skill paths requeridos. | Cualquier rol especializado. |
| `router` | tier, metodologia, permisos, review workload, skill slots, context slice. | `orchestrator`, `explorer`, `implementer`, `validator`, `delivery`. |
| `explorer` | archivos relevantes, comportamiento actual, riesgos, limite de implementacion. | `implementer` o `user`. |
| `implementer` | archivos cambiados, comportamiento cambiado, comandos corridos, gaps de validacion/delivery. | `validator`, `reviewer`, `delivery`. |
| `test-writer` | archivos de test, conducta cubierta, fases TDD, gaps. | `implementer` o `validator`. |
| `validator` | matriz de comandos, pass/fail, evidencia faltante, owner probable. | `reviewer`, `delivery`, `implementer`. |
| `reviewer` | findings por severidad, score, pruebas faltantes, riesgo de release. | `implementer`, `validator`, `delivery`. |
| `delivery` | autorizacion, branch state, scope staged, PR/check status, recovery command. | `orchestrator` o `user`. |

## Por que reduce tokens

- El router ya no debe copiar instrucciones completas de skills: pasa slots y paths exactos.
- El siguiente rol lee solo los `SKILL.md` necesarios para su fase.
- El payload compacto evita reenviar historia, prompts completos o matrices irrelevantes.
- Los adapters futuros pueden cambiar paths sin cambiar el contrato del rol.
