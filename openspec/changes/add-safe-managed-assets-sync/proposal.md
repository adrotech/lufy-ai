## Why

Los assets gestionados necesitan un mecanismo seguro para reaplicarse cuando el source evoluciona sin sobrescribir personalizaciones locales ni archivos fuera del catálogo. El roadmap RM-008/RM-009 requiere un sync idempotente basado en manifest/hash/backup que reutilice la instalación gestionada existente y deje una ruta clara para actualizar proyectos ya instalados.

## What Changes

- Añadir un comando `lufy-ai sync` en la CLI Go para sincronizar assets gestionados desde el source hacia un target existente.
- Reusar catálogo, resolución segura de paths, estado `.lufy-ai/install-state.json`, SHA-256, backup/restore y verify estructural existentes.
- Construir un plan antes de mutar, con `--dry-run` sin escrituras y acciones explicables por asset.
- Actualizar solo assets previamente gestionados y sin drift local; crear backups antes de cualquier actualización gestionada.
- Reportar conflictos para archivos ausentes/no gestionados, drift local, estado inválido o escapes de target, sin sobrescribirlos silenciosamente.
- Mantener `scripts/install.sh` como wrapper estricto de `install`; no ampliar el wrapper para lógica de sync en este cambio.
- No introducir templates nuevos, productización, descarga remota ni fallback legacy.

## Capabilities

### New Capabilities
- Ninguna. El cambio extiende capacidades existentes de instalación gestionada y CLI Go.

### Modified Capabilities
- `managed-assets-install`: define sync seguro de assets gestionados con hash, manifest, backup, idempotencia y límites de alcance.
- `go-cli-installer`: añade el comando base `sync`, flags/defaults seguros y validación aplicable al flujo Go.

## Impact

- Código Go bajo `tools/lufy-cli-go/`, especialmente parsing de comandos, planificación/aplicación de assets gestionados, estado, backup y verificación.
- Contrato CLI: nuevo comando `lufy-ai sync --target <dir> [--dry-run] [--yes] [--no-engram]` o comportamiento equivalente documentado por la CLI.
- Specs OpenSpec existentes `managed-assets-install` y `go-cli-installer`.
- Sin cambios esperados en `scripts/install.sh` salvo validación de que permanece estricto y sin fallback legacy.
