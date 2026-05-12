# OpenSpec

Workflow de especificación驱动 desenvolvimento (SDD) para este proyecto.

## Estructura

```
openspec/
├── config.yaml          # Configuración del workflow
├── specs/               # Specs activas
│   └── <feature-name>/
│       ├── spec.md      # Especificación técnica
│       └── tasks.md    # Checklist de tareas
└── changes/            # Cambios en progreso
    └── archive/         # Historial de cambios completados
```

## Comandos

Usa los skills de `.opencode/skills/sdd-workflow/`:
- `/openspec-propose` - Crear nueva feature spec
- `/openspec-apply` - Implementar tareas de un cambio
- `/openspec-verify` - Verificar implementación
- `/openspec-archive` - Archivar cambio completado
- `/openspec-explore` - Explorar specs existentes

## Flujo

1. **Propose**: Crear spec en `openspec/specs/`
2. **Apply**: Implementar tareas checklist
3. **Verify**: Correr validación final agrupada, incluyendo tests y coverage cuando existan para el alcance real
4. **Archive**: Mover a `openspec/changes/archive/`

## Workflow sistémico

- Analizar al inicio los archivos existentes relevantes, dependencias, interconexiones, feedback y relación estructura-comportamiento.
- Implementar tareas sin relecturas repetidas de archivos viejos ya analizados, salvo modificación, conflicto, bloqueo, nueva evidencia, cambio de alcance o riesgo explícito.
- Revisar al final los archivos viejos modificados o afectados antes de la validación.
- Ejecutar tests, coverage y validación completa al final de todas las tareas de la propuesta cuando los comandos existan; si no existen, reportar la limitación y la evidencia real disponible.
