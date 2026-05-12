## Context

La CLI Go ya instala, sincroniza, verifica, respalda y restaura assets gestionados usando catálogo allowlist, SHA-256 e install state. El problema pendiente es que el modelo actual solo distingue entre asset gestionado sin drift y conflicto bloqueante. Ese modelo protege datos, pero bloquea upgrades cuando el usuario modifica archivos que naturalmente debe poder personalizar, como `AGENTS.md`, policies o configuraciones OpenCode.

El cambio introduce un contrato declarativo por asset para decidir cómo actuar ante drift. La implementación debe mantenerse stdlib-only, preservar el wrapper Bash estricto y sincronizar assets raíz con el mirror embebido cuando cambien templates o specs gestionadas.

## Goals / Non-Goals

**Goals:**

- Resolver drift esperado sin sobrescribir trabajo local.
- Hacer que install/sync/verify/status reporten policy, scope y estado de drift de forma accionable.
- Migrar `.lufy-ai/install-state.json` de forma silenciosa y compatible hacia metadata de policy/scope/ancestor.
- Preservar el comportamiento actual con `--scope=project`.
- Permitir `AGENTS.md` mixto mediante bloques gestionados por lufy.
- Crear `.lufy-new` y ancestors para upgrades con drift en assets `no-replace`.

**Non-Goals:**

- No implementar OpenSpec v2 ni nuevos comandos `/opsx-*` en este cambio.
- No agregar dependencias externas ni framework CLI nuevo.
- No agregar soporte multi-tool fuera de OpenCode.
- No borrar automáticamente archivos retirados o `.lufy-new` generados.
- No hacer merge semántico arbitrario de Markdown fuera de bloques marcados.

## Decisions

1. **Policies como strings persistidas.**
   Usar valores JSON estables (`managed`, `no-replace`, `merge-block`, `merge-json`, `metadata`) evita acoplar install-state a `iota` y facilita compatibilidad futura.

2. **Migración silenciosa de state.**
   El lector de estado acepta el schema actual y completa defaults: `managed` para assets completos existentes, `project` como scope inicial y ancestor vacío hasta la siguiente escritura exitosa. La escritura nueva debe usar el schema actualizado.

3. **Planificación policy-driven antes de escribir.**
   Install y sync deben resolver la action por asset antes de aplicar. El apply no debe recalcular decisiones destructivas; solo ejecutar el plan validado.

4. **`no-replace` genera `.lufy-new`.**
   Cuando hay nueva versión y drift local, se escribe la versión nueva en un path sibling `<target>.lufy-new` con backup si corresponde. El archivo original no se toca.

5. **`merge-block` limitado a marcadores explícitos.**
   El motor solo reemplaza contenido entre `<!-- LUFY:BEGIN <id> -->` y `<!-- LUFY:END <id> -->`. Texto fuera de bloques permanece intacto. Marcadores duplicados, anidados o sin cierre bloquean escritura.

6. **Ancestors dentro de `.lufy-ai/ancestors/`.**
   Guardar la última versión limpia instalada por lufy permite merge futuro y auditoría. Los paths deben mapearse de forma segura, relativa y portable, sin permitir escapes.

7. **Scope explícito y reversible.**
   Se introduce `--scope=project|global|both`. `project` conserva el comportamiento actual. El default final puede decidirse durante RC, pero la implementación debe soportar los tres valores y reportar qué target efectivo usa.

8. **Restore mejora sin romper contratos existentes.**
   `restore` puede listar backups y aceptar IDs, pero debe seguir soportando `--backup <manifest-or-dir>`.

## Risks / Trade-offs

- **Cambio de default scope sorprende usuarios** -> mantener `--scope=project`, documentar migración y validar en sandbox brownfield antes de cambiar default.
- **State migration incorrecta puede romper installs existentes** -> cubrir con fixtures v1 y tests de lectura/escritura round-trip.
- **MergeBlock corrupto puede destruir contenido** -> bloquear ante marcadores ambiguos y no intentar reparaciones automáticas.
- **`.lufy-new` puede acumular basura** -> reportarlo en `status`/`verify`, pero dejar limpieza manual para no borrar datos.
- **Global scope toca home del usuario** -> dry-run debe ser fiel, `--yes`/confirmación debe ser explícita y tests deben usar home temporal.

## Migration Plan

1. Implementar parsing/escritura de state compatible con schema actual y nuevo.
2. Introducir policy/scope en catálogo sin cambiar todavía el default efectivo.
3. Ajustar plan/apply de install y sync para policies nuevas.
4. Migrar `AGENTS.md.template` a bloques y sincronizar mirror embebido.
5. Añadir reporting en verify/status y tests/smokes.
6. Documentar RC/default scope y validar sandboxes antes de release.

Rollback: si el cambio debe revertirse antes de release, los targets nuevos seguirán teniendo backups y `.lufy-new`; la CLI anterior no entenderá campos nuevos, por lo que la migración debe preservar campos existentes y evitar depender de datos irreversibles para restores básicos.

## Open Questions

- Confirmar antes del release si el default de `install` será `project`, `global` o `both`.
- Definir lista final de policies por entry del catálogo tras revisar assets reales.
- Definir si `merge` invoca herramienta externa en este cambio o si solo deja ancestors y `.lufy-new` listos para una UX posterior dentro del mismo release.
