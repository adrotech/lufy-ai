# Tasks

## 1. Spec

- [x] Crear proposal/tasks y delta OpenSpec.
- [x] Validar `openspec validate add-managed-uninstall-command --strict`.

## 2. Implementación

- [x] Crear servicio `internal/uninstaller` con plan, dry-run, backup y mutaciones.
- [x] Integrar comando CLI `lufy-ai uninstall`.
- [x] Remover referencia Lufy de `AGENTS.md` sin borrar contenido user-owned.
- [x] Limpiar directorios vacíos gestionados y preservar `opencode.json`.

## 3. Tests y validación

- [x] Cubrir dry-run sin mutaciones.
- [x] Cubrir uninstall real y reinstall posterior.
- [x] Cubrir bloqueo por drift local.
- [x] Ejecutar tests Go enfocados.
- [x] Ejecutar `scripts/validate.sh`.
- [x] Ejecutar `openspec validate --specs --strict`.
- [x] Ejecutar `git diff --check origin/develop...HEAD` y `git diff --check`.
