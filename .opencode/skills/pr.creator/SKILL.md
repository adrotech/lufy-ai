---
name: pr.creator
description: Genera título sugerido y cuerpo Markdown de Pull Request para GitHub con trazabilidad, evidencia, monitores y migraciones, sin ejecutar delivery.
license: MIT
compatibility: OpenCode skill autocontenido; no requiere toolchain externo.
metadata:
  author: lufy-ai
  version: "1.0"
---

# Skill: pr.creator

Genera contenido de Pull Request para GitHub en español: título sugerido y cuerpo Markdown. Puede usarse manualmente o desde el subagente `delivery` antes de ejecutar `gh pr create`.

## Fuente de verdad y límites

- `.opencode/policies/delivery.md` sigue siendo la fuente de verdad para branch safety, validación, PRs, sync y gates de delivery.
- `pr.creator` estructura y redacta contenido; `delivery` conserva commit, push, `gh pr create`, sync remoto, branch safety, validación final y reporte.
- No ejecutes `git`, `gh`, sync de GitHub Projects, Jira, Notion, herramientas remotas ni mutaciones de delivery.
- Cuando `delivery` provea evidencia de `lufy-ai pr guard --base <base>`, inclúyela en el PR body. Si falta, marca `Pendiente de delivery`; no ejecutes el guardrail desde `pr.creator`.
- No inventes evidencia, links, tickets, monitores, resultados de pruebas ni migraciones.
- Si falta información, usa explícitamente `Pendiente de confirmar`, `No configurado` o `No aplica` y lista los datos faltantes.
- El contenido humano debe estar en español; preserva identificadores técnicos, rutas, flags, IDs y nombres de comandos.

## Inputs esperados

Usa cualquier dato disponible, en este orden de preferencia:

1. `change-id` OpenSpec, si existe.
2. Contenido de `openspec/changes/<change-id>/proposal.md`, especialmente secciones `Why` y `What Changes`.
3. `tasks.md`, specs, design, resumen funcional, commits, diff o notas proporcionadas por la persona/agente.
4. Evidencia de validación: comandos exactos, resultados, capturas, JSON, curls, salidas resumidas y limitaciones.
5. Evidencia del PR guard: salida de `lufy-ai pr guard --base <base>` o fallback `git check-ignore -v --no-index --stdin`.
6. Tarea asociada: links o IDs de Jira, GitHub Issues/Projects, Notion u otro sistema de tracking configurado/proporcionado.
7. Monitores o dashboards: Grafana, New Relic, Datadog u otros sistemas configurados/proporcionados.
8. Archivos modificados o diff disponible para detectar migraciones y cambios de schema.

## Outputs

Devuelve solamente contenido listo para PR, sin ejecutar delivery:

```markdown
Título sugerido: <título>

<cuerpo Markdown del PR>
```

El cuerpo debe seguir el template en `templates/pr-body.md` y mantener secciones obligatorias: resumen, `Why`, tarea asociada/trazabilidad, evidencia de pruebas, `Monitors`, `Migraciones`, riesgos/follow-ups y checklist/notas de validación.

## Flujo manual

Cuando una persona o agente invoque `pr.creator` directamente:

1. Recolecta los inputs proporcionados: `change-id`, diff, notas, evidencia, tickets, monitores o contexto del cambio.
2. Si hay `change-id`, intenta usar `openspec/changes/<change-id>/proposal.md` como fuente principal de resumen y `Why`.
3. Si no hay `proposal.md`, usa diff/notas/contexto y marca supuestos como `Pendiente de confirmar`.
4. Aplica las heurísticas de `Migraciones` documentadas en `references/detection.md`.
5. Produce título sugerido y cuerpo Markdown. No ejecutes `git`, `gh` ni operaciones remotas.

## Flujo integrado desde delivery

Cuando `delivery` tenga autorización explícita para crear un PR y exista `.opencode/skills/pr.creator/`:

1. `delivery` reúne branch/workspace state, diff, commits, evidencia de validación, tracking y contexto OpenSpec según `.opencode/policies/delivery.md`.
2. `delivery` invoca/usa `pr.creator` para generar título y cuerpo del PR.
3. `pr.creator` devuelve solo contenido del PR.
4. `delivery` ejecuta las operaciones autorizadas restantes: staging/commit/push/`gh pr create`, sync y reporte.

Si `pr.creator` no está disponible, `delivery` debe reportar esa limitación y usar la política/template vigente solo si no contradice autorización ni gates de delivery.

## Construcción del PR

### Resumen y Why

- Si existe `openspec/changes/<change-id>/proposal.md`, deriva:
  - `Resumen` desde `## What Changes` y capacidades relevantes.
  - `Why` desde `## Why`.
- Resume en bullets claros; no copies bloques largos si pueden sintetizarse.
- Si no existe proposal, usa contexto provisto y marca lo incierto como `Pendiente de confirmar`.

### Tarea asociada / tracking

Incluye links o IDs solo cuando haya evidencia/configuración disponible:

- Jira: claves tipo `ABC-123` o URLs de Jira proporcionadas.
- GitHub Issues/Projects: URLs `github.com/.../issues/<n>`, referencias `#<n>` con repo conocido o Project links proporcionados.
- Notion: URLs o IDs de página proporcionados.
- Otros sistemas: referencias explícitas en contexto, propuesta, tareas o notas.

Si no hay herramienta configurada ni ticket provisto, usa `No configurado` o `Pendiente de confirmar`.

### Evidencia de pruebas

- Lista comandos exactos y resultado observado (`pass`, `fail`, `no disponible`, `no aplica`) solo si fueron proporcionados o ejecutados por el agente llamador.
- Incluye capturas, JSON o curls con contexto cuando existan.
- Si no se ejecutó validación, declara la limitación; no afirmes éxito.

### Guardrail de paths ignorados/internos

- Si `delivery` ejecutó `lufy-ai pr guard --base <base>`, incluye comando, resultado y notas.
- Si el guard reportó paths ignorados o internos, marca el PR como `Pendiente de corregir` o documenta el override explícito del usuario.
- El texto debe explicar que `.gitignore` no evita que archivos ya trackeados entren por cherry-pick, worktree o commits existentes.
- Si el CLI no está disponible y hay fallback, registra `git diff --name-only <base>...HEAD -- | git check-ignore -v --no-index --stdin` y la revisión manual de `openspec/`, `.lufy/`, `.lufy-ai/`, `pr_review/`.

### Monitors

Captura monitores o dashboards solo si están configurados/proporcionados:

- Grafana: nombre/link de dashboard o panel.
- New Relic: entidad, dashboard, alerta o link.
- Datadog: monitor, dashboard, SLO o link.
- Otros: sistema, nombre y link.

Si no aplica o no hay datos, usa `No aplica`, `No configurado` o `Pendiente de confirmar`.

### Migraciones

Analiza rutas y diff disponible con las heurísticas de `references/detection.md`. La sección debe indicar uno de estos estados:

- `Detectadas`: listar archivos/patrones y pedir plan/evidencia de ejecución si aplica.
- `No detectadas`: indicar que no se encontraron patrones y que la revisión fue heurística.
- `Pendiente de confirmar`: cuando hay señales ambiguas de persistencia o schema.

## Template

Usa `templates/pr-body.md` como base. Sustituye placeholders con datos reales o fallbacks explícitos.
