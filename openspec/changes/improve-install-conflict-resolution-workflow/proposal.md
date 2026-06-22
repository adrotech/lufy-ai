## Why

Los proyectos que ya tienen assets LUFY/OpenCode/OpenSpec pero no tienen manifest gestionado quedan bloqueados durante `install` con una lista plana de conflictos no gestionados. Esto obliga al usuario o agente a razonar manualmente sobre muchos archivos y dificulta revisar categorías en paralelo.

El setup interactivo debe ser la experiencia principal para humanos, pero necesita consumir el mismo contrato scriptable que usan agentes y automatización para no duplicar lógica.

## What Changes

- Agregar un contrato de plan de conflictos de instalación con salida humana y JSON.
- Agrupar conflictos por categorías revisables: agentes, comandos, skills, templates, OpenSpec specs y root/config.
- Recomendar una acción segura por archivo: `keep-local`, `accept-managed`, `merge`, `backup-and-replace` o `block`.
- Exponer un comando read-only `lufy-ai conflicts plan --target <dir> [--json]`.
- Hacer que `setup`, `status` y `verify` orienten al usuario hacia el plan cuando falta manifest o hay una migración bloqueada probable.
- Mantener la UI Bubble Tea como presentador sobre el mismo contrato; no resolver decisiones destructivas solo con flags nuevos.
- Documentar que `lufy-ia.harness.md` es un asset gestionado activo por compatibilidad y no debe borrarse como legacy sin una migración específica.
- Extender el harness para que los grupos del plan puedan convertirse en slices paralelizables cuando los archivos sean independientes.

## Non-Goals

- No sobrescribir conflictos no gestionados automáticamente.
- No borrar `lufy-ia.harness.md` ni renombrarlo a `lufy-ai.harness.md` en este cambio.
- No agregar flags masivos como `--accept-managed-all` o `--delete-deprecated`.
- No reemplazar comandos scriptables por UI interactiva.

## Review Slices

### Slice 1: Contrato de plan de conflictos

- Objetivo: modelar reporte por archivo/categoría a partir del plan de instalación existente.
- Archivos esperados: `tools/lufy-cli-go/internal/conflictplan/*`, tests unitarios.
- Criterios:
  - WHEN hay conflictos no gestionados, THEN cada item incluye path, categoría, riesgo, recomendación, razón y acciones disponibles.
  - WHEN existen rutas legacy `.lufy-ai/*`, THEN el reporte las marca como legacy/deprecated y sugiere `migrate-layout` sin borrar.

### Slice 2: CLI scriptable

- Objetivo: exponer `lufy-ai conflicts plan` con salida humana y JSON.
- Archivos esperados: `internal/cli/app.go`, tests CLI, command palette.
- Criterios:
  - WHEN el usuario ejecuta `lufy-ai conflicts plan --json`, THEN recibe JSON parseable sin mutaciones.
  - WHEN el usuario ejecuta salida humana, THEN ve grupos y próximos pasos accionables.

### Slice 3: Setup/status/verify guidance

- Objetivo: aclarar estados de migración bloqueada cuando falta manifest.
- Archivos esperados: `internal/setup`, `internal/status`, `internal/verify`.
- Criterios:
  - WHEN falta manifest pero hay conflictos probables, THEN el output sugiere `lufy-ai conflicts plan --target <dir>`.
  - WHEN setup detecta conflictos, THEN no aplica mutaciones y apunta al plan.

### Slice 4: Harness paralelizable

- Objetivo: usar categorías del plan como grupos revisables en paralelo cuando sean independientes.
- Archivos esperados: `.opencode/agents/orchestrator.md`, `.opencode/agents/sdd-router.md`, specs.
- Criterios:
  - WHEN un plan produce grupos independientes, THEN el harness puede delegar revisión por categoría en paralelo con validación agrupada posterior.

## Validation

- `openspec validate "improve-install-conflict-resolution-workflow" --strict`
- `go test ./...` desde `tools/lufy-cli-go`
- `scripts/validate.sh`
