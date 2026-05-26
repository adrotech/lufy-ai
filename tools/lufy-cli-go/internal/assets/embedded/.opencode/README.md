# Tooling local de OpenCode

Este directorio contiene la configuración local de OpenCode para `lufy-ai`.

Las reglas compartidas viven en `../AGENTS.md` (guía real del repositorio). La plantilla genérica queda en `../AGENTS.md.template`.

## Agents

- `agents/orchestrator.md`: coordinador primario por defecto.
- `agents/sdd-router.md`: subagente read-only para clasificación T1/T2/T3, contexto mínimo y routing proporcional.
- `agents/explorer.md`: subagente read-only para exploración.
- `agents/implementer.md`: subagente de implementación.
- `agents/test-writer.md`: subagente TDD stack-aware para pruebas sustantivas T1/T2.
- `agents/validator.md`: subagente read-only para validación.
- `agents/reviewer.md`: subagente read-only para revisión stack-aware con scoring L1-L5.
- `agents/delivery.md`: subagente de delivery para Git/GH y PRs.

Todos los agentes siguen un estándar común de frontmatter (`description`, `mode`, `temperature`, `steps`, permisos mínimos) y secciones: `Mission`, `Use When`, `Do Not Use When`, `Inputs Expected`, `Workflow`, `Boundaries`, `Validation / Evidence`, `Escalation`, `Required Output`.

Las reglas compartidas de delivery viven en `policies/delivery.md`.

## Harness Routing

`lufy-ai` usa routing proporcional para elegir el flujo más pequeño que resuelva el pedido con seguridad.

- **T1 Full SDD**: nuevas capabilities, impacto transversal, decisiones de arquitectura, contratos públicos, seguridad, política de delivery o alta incertidumbre. Usa OpenSpec completo.
- **T2 SDD Lite**: cambio funcional acotado, bug relevante, ajuste de agente/skill o refactor controlado. Usa mini-spec o handoff estructurado con criterios WHEN/THEN.
- **T3 Express**: cambio trivial, mecánico, documental o local sin riesgo relevante. Puede ir directo a implementación acotada y validación proporcional.

El `orchestrator` puede invocar `sdd-router` antes de activar agentes más pesados. El router devuelve `tier`, `confidence`, `execution_mode`, `recommended_flow`, `context_slice`, `skill_status`, `review_workload`, `stop_reason` y `next_agent`.

Execution modes soportados:

- `full_sdd`: OpenSpec completo.
- `sdd_lite`: mini-spec T2 o handoff estructurado.
- `express`: implementación directa acotada.
- `clarify`: pregunta bloqueante breve.
- `explore_only`: investigación read-only.
- `verify_only`: evidencia o diagnóstico.
- `delivery_pending`: Git/GH bloqueado hasta autorización explícita.

Templates locales:

- `templates/sdd-lite.md`: artefacto compacto para T2.
- `templates/result-contract.md`: contrato de salida para handoffs y recuperación de contexto.

Hooks locales:

- `hooks/format-dispatch.sh`: dispatcher silencioso para PostToolUse que lee `.opencode/project.yaml`, matchea extensiones y ejecuta el formatter/autofix configurado del stack cuando aplica.

Skill resolution es local-first: `.opencode/skills` y `AGENTS.md` tienen prioridad. Si falta cobertura local, el router puede sugerir AutoSkills solo como bootstrap opcional, empezando por `npx autoskills --dry-run` y requiriendo autorización explícita antes de cualquier comando mutante.

Subagent isolation: cada subagente recibe solo el `context_slice`, rutas y constraints necesarios para su rol. La revisión se dimensiona como `none`, `focused` o `full` según tier, superficie modificada y riesgo.

Review Workload Harness: para T1 y T2 con varios ejes de riesgo, el router puede recomendar `review_slices`. Cada slice debe tener objetivo, archivos esperados, criterios `WHEN`/`THEN`, validación, riesgo principal y guía de PR. Esto piensa en el reviewer humano y favorece entregables pequeños, pero no obliga a micro-PRs: T3 no se fragmenta y toda guía de PR sigue dependiendo de autorización explícita de delivery.

Contexto operativo del repo:

- La CLI Go del producto vive en `../tools/lufy-cli-go`.
- `../scripts/install.sh` es un wrapper estricto del CLI Go, sin fallback legacy.
- Preferir validación agrupada al final de un bloque/proposal; no correr tests constantemente salvo bloqueo, cambio riesgoso o diagnóstico.
- Foco OpenSpec actual: `install-managed-assets-with-hash-idempotency` (assets gestionados, SHA-256, manifest, idempotencia, backup/restore, verify estructural).
- `migrate-installer-to-go-cli` no debe archivarse mientras tenga tasks incompletas.

### Checklist para nuevos agentes

- Mantener permisos mínimos; no conceder `edit`, `bash` o Git/GH si no son necesarios.
- Definir `steps` y un contrato de salida claro.
- Para agentes de routing, incluir execution mode, permisos mínimos, context slicing y result contract.
- Explicar cuándo usar/no usar el agente y cómo escalar.
- No prometer tests ni validación sin evidencia real.
- Usar español para contenido humano y preservar identificadores técnicos.

## Commands

Los slash commands viven en `commands/`.

Regla de namespace:

- `/opsx-*`: comandos canónicos del workflow OpenSpec. Se preservan y no se renombran desde el kit Lufy.
- `/lufy.*`: extras propios del kit Lufy, como reportes o utilidades operativas que complementan OpenSpec.

Comandos OpenSpec:

- `/opsx-explore`: explorar el codebase sin implementar.
- `/opsx-propose`: crear artefactos de propuesta OpenSpec.
- `/opsx-apply`: implementar tareas OpenSpec.
- `/opsx-verify`: verificar implementación contra la spec.
- `/opsx-sync`: aplicar deltas validados a specs principales sin archivar.
- `/opsx-archive`: archivar un cambio completado; tasks incompletas implican `blocked`, no archive.
- `/opsx-version`: reportar la fuente OpenSpec efectiva y diagnósticos de fallback.

Comandos Lufy:

- `/lufy.timereport`: generar un reporte local de tiempo/ROI como extra del kit.

## Skills

- `skills/sdd-workflow/openspec-explore`: explorar cambios.
- `skills/sdd-workflow/openspec-propose`: proponer cambios.
- `skills/sdd-workflow/openspec-apply-change`: implementar tasks.
- `skills/sdd-workflow/openspec-verify-change`: verificar implementación contra artefactos.
- `skills/sdd-workflow/openspec-archive-change`: archivar solo cambios completos.

Skills opcionales de delivery, project sync, memoria y release pueden agregarse en proyectos downstream. El kit base solo incluye el lifecycle OpenSpec.

## Agent Observatory TUI Plugin

El plugin de sidebar TUI se carga desde `tui.json` en la raíz:

```json
{
  "$schema": "https://opencode.ai/tui.json",
  "plugin": ["./.opencode/plugins/agent-observatory.tsx"],
  "plugin_enabled": {
    "lufy-ai.observatory": true
  }
}
```

Slash commands registrados por el plugin actual:

- `/observatory`: mostrar/ocultar el panel.
- `/observatory-agents`: contraer/expandir la lista de agentes.
- `/observatory-subagents`: contraer/expandir la sección de subagentes.
- `/observatory-cost`: mostrar/ocultar costo.

V1 es local/TUI-only. No agregar telemetría externa sin una propuesta separada.
