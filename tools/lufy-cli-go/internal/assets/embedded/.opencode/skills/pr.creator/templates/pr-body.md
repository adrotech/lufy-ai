# <Título sugerido>

## Resumen

- <Cambio funcional principal o `Pendiente de confirmar`>
- <Impacto/alcance relevante o `No aplica`>

## Why

<Motivo del cambio, preferiblemente derivado de `openspec/changes/<change>/proposal.md`. Si falta contexto: `Pendiente de confirmar`.>

## Tarea asociada

- Tracking: <link/ID Jira, GitHub Issue/Project, Notion u otro; o `No configurado` / `No aplica` / `Pendiente de confirmar`>
- OpenSpec: <`openspec/changes/<change-id>` o `No aplica`>

## Datos faltantes / pendientes de confirmar

- <dato, decisión o evidencia faltante; o `No aplica`>

## Evidencia de pruebas

### Comandos y resultados

| Comando | Resultado | Notas |
| --- | --- | --- |
| `<comando>` | `<pass/fail/no disponible/no aplica>` | `<salida resumida o limitación>` |

### Evidencia funcional / adjuntos

- Capturas: <links/rutas o `No aplica`>
- JSON/curls: <resumen/link/bloque o `No aplica`>
- Validación manual/estática: <detalle o `Pendiente de confirmar`>

### Checks remotos del PR

- Comando: <`gh pr checks <PR>` / `gh pr view <PR> --json statusCheckRollup,mergeStateStatus,url` / `Pendiente de delivery`>
- Estado: <`pass` / `fail` / `pending` / `no disponible` / `Pendiente de delivery`>
- Notas/recovery: <link al PR, checks pendientes/fallidos o `No aplica`>

### Guardrail de paths ignorados/internos

- Comando: <`lufy-ai pr guard --base <base>` / fallback `git check-ignore -v --no-index --stdin` / `Pendiente de delivery`>
- Estado: <`Sin hallazgos` / `Detectados` / `Pendiente de delivery`>
- Evidencia: <paths, patrón `.gitignore`, prefijo interno, override explícito o `No aplica`>
- Nota: `.gitignore` no impide que archivos ya trackeados entren por cherry-pick, worktree o commits existentes.

## Monitors

| Sistema | Monitor/Dashboard | Link | Estado |
| --- | --- | --- | --- |
| <Grafana/New Relic/Datadog/otro> | <nombre o `No configurado`> | <link o `No aplica`> | <activo/pendiente/no aplica> |

## Migraciones

- Estado: <`Detectadas` / `No detectadas` / `Pendiente de confirmar`>
- Evidencia: <rutas, patrones de diff o `No se detectaron patrones; revisión heurística`>
- Plan/rollback: <plan de ejecución/rollback o `No aplica` / `Pendiente de confirmar`>

## Riesgos / Follow-ups

- <riesgo, deuda o `No aplica`>

## Checklist / notas de validación

- [ ] Branch/base revisados por `delivery` según `.opencode/policies/delivery.md`.
- [ ] Evidencia de validación incluida o limitación declarada.
- [ ] `lufy-ai pr guard --base <base>` ejecutado o fallback documentado; paths ignorados/internos corregidos u override explícito registrado.
- [ ] Checks remotos del PR consultados/esperados por `delivery` antes de reportar `delivered`/`closed`, o marcados como pendientes con recovery.
- [ ] Tarea asociada/tracking declarado o marcado como no configurado.
- [ ] Monitors declarados o marcados como no aplica/no configurado.
- [ ] Migraciones revisadas con heurística y estado explícito.
