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
3. **Verify**: Correr tests y coverage
4. **Archive**: Mover a `openspec/changes/archive/`
