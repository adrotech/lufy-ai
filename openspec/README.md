# OpenSpec

Workflow de especificación dirigida por specs (SDD) para este proyecto. OpenSpec es la metodología principal instalada por el adapter `opencode`, pero Lufy ya permite seleccionar metodología por tier. En la práctica: T1 usa metodología full, T2 usa metodología lite o handoff estructurado, y T3 puede usar `none` cuando la policy lo permite. Para cambios grandes, el Review Workload Harness ayuda a dividir features en slices revisables por humanos.

## Estructura

```
openspec/
├── config.yaml          # Configuración core v2 action-based
├── UPSTREAM.json        # Baseline, versión mínima y metadata de resolución stay-updated
├── specs/               # Specs activas
│   └── <feature-name>/
│       ├── spec.md      # Especificación técnica
│       └── tasks.md    # Checklist de tareas
└── changes/            # Cambios en progreso
    └── archive/         # Historial de cambios completados
```

## Comandos

Usa los comandos instalados en `.opencode/commands/` y skills de `.opencode/skills/sdd-workflow/`:
- `/opsx-explore` - Explorar specs, impacto o ideas en modo read-only
- `/opsx-propose` - Crear artefactos de cambio con specs delta
- `/opsx-apply` - Implementar tareas de un cambio
- `/opsx-verify` - Verificar implementación, deltas y scenarios
- `/opsx-sync` - Aplicar deltas validados a specs principales sin archivar
- `/opsx-archive` - Archivar cambio completado tras sync y gates
- `opsx-version` - Reportar fuente efectiva OpenSpec: PATH, cache local o baseline embebida

## Flujo

1. **Route**: Clasificar con `sdd-router` cuando el pedido sea no trivial, ambiguo o riesgoso.
2. **T1 Propose**: Crear artefactos en `openspec/changes/<change>/`, incluyendo specs delta.
3. **T2 Lite**: Usar mini-spec o handoff estructurado con criterios `WHEN`/`THEN` cuando Full SDD sea excesivo.
4. **Review slices**: Para T1 o T2 con varios riesgos, definir slices revisables con objetivo, archivos esperados, criterios, validación y riesgo.
5. **Apply**: Implementar tareas checklist.
6. **Verify**: Correr validación final agrupada, incluyendo tests y coverage cuando existan para el alcance real.
7. **Sync**: Aplicar deltas validados a `openspec/specs/` sin mover el cambio cuando exista delta OpenSpec.
8. **Archive**: Mover a `openspec/changes/archive/`.

Si el tier usa `lufy-sdd` o `none`, no inventar artefactos OpenSpec. Reportar `methodology_id`, `methodology_mode` y `execution_mode` en el handoff cuando sea relevante.

## Specs delta core v2

Los specs bajo `openspec/changes/<change>/specs/` deben usar secciones explícitas:

- `## ADDED Requirements`
- `## MODIFIED Requirements`
- `## REMOVED Requirements`

Cada requisito añadido o modificado debe incluir al menos un `#### Scenario:` con cláusulas `WHEN` y `THEN`. Usa `GIVEN` solo cuando el contexto inicial sea necesario. Los requisitos removidos deben incluir razón o guía de migración.

`/opsx-sync` es la acción explícita para llevar esos deltas a `openspec/specs/` antes de `/opsx-archive`; archive debe bloquear si detecta deltas sin sincronizar.

## Workflow sistémico

- Analizar al inicio los archivos existentes relevantes, dependencias, interconexiones, feedback y relación estructura-comportamiento.
- Implementar tareas sin relecturas repetidas de archivos viejos ya analizados, salvo modificación, conflicto, bloqueo, nueva evidencia, cambio de alcance o riesgo explícito.
- Revisar al final los archivos viejos modificados o afectados antes de la validación.
- Ejecutar tests, coverage y validación completa al final de todas las tareas de la propuesta cuando los comandos existan; si no existen, reportar la limitación y la evidencia real disponible.
