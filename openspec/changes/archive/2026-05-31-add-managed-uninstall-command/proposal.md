# Proposal: agregar uninstall gestionado

## Why

Necesitamos probar el ciclo completo de usuarios reales: instalar Lufy, desinstalarlo de forma segura y volver a instalarlo. Hoy la CLI tiene backup/restore, pero no ofrece una operación explícita para remover los assets gestionados por Lufy sin tocar archivos propios del usuario.

## What Changes

- Agregar `lufy-ai uninstall` con `--target`, `--dry-run`, `--yes` y `--keep-state`.
- El uninstall remueve solo archivos registrados como gestionados y sin drift local.
- Antes de mutar, crea backup de assets gestionados existentes, ancestors, `AGENTS.md` si contiene la referencia Lufy y `install-state.json`.
- Remueve la referencia `@lufy-ia.harness.md` de `AGENTS.md` sin borrar el archivo.
- Preserva `opencode.json` porque es merge-managed/user-owned.

## Non Goals

- No borrar archivos no gestionados.
- No forzar eliminación de assets con drift local.
- No implementar uninstall global/both todavía.
