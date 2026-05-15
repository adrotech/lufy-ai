## Why

`v0.2.0` ya resolvió el riesgo operativo de upgrades con drift; el siguiente bloqueo del roadmap es que el flujo OpenSpec instalado aún no expresa contratos modernos con deltas y scenarios verificables. Este cambio inicia `v0.3.0` cerrando el gap core con OpenSpec v1.3.1 antes de sumar perfiles expandidos o stay-updated remoto.

## What Changes

- Modernizar la configuración OpenSpec instalada hacia un schema v2 action-based para que los comandos `/opsx-*` dependan de acciones explícitas y no de convenciones implícitas.
- Exigir que los specs de cambios usen delta markers `ADDED`, `MODIFIED` y `REMOVED` para separar cambios propuestos de specs principales.
- Exigir scenarios testables con formato `GIVEN`/`WHEN`/`THEN` o equivalente definido por OpenSpec v1.3.1.
- Agregar `/opsx-sync` y la skill `openspec-sync` para aplicar deltas validados a specs principales antes de archivar.
- Agregar `UPSTREAM.json` como baseline local de la versión efectiva de OpenSpec cubierta por los assets instalados.
- Agregar `opsx-version` para reportar la versión efectiva, baseline y fuente del workflow OpenSpec instalado.

## Capabilities

### New Capabilities
- `openspec-core-v2-workflow`: cubre el schema v2 action-based, validación core de deltas/scenarios, `/opsx-sync`, `openspec-sync`, `UPSTREAM.json` y reporte `opsx-version`.

### Modified Capabilities
- `managed-assets-install`: el catálogo instalado deberá incluir y verificar los nuevos assets OpenSpec core v2 como assets gestionados.
- `go-cli-installer`: la instalación, sync y verify deberán reflejar la nueva superficie OpenSpec core v2 sin romper targets existentes.

## Impact

- Afecta assets instalables en `.opencode/commands/`, `.opencode/skills/` y `openspec/`.
- Afecta assets embebidos en `tools/lufy-cli-go/internal/assets/embedded/`.
- Puede requerir ajustes en validación Go si el catálogo o verify incorporan nuevos archivos obligatorios.
- No introduce todavía el resolver stay-updated de 3 capas, perfiles expandidos, hooks de reconciliación ni comandos OpenSpec extra fuera del core.
