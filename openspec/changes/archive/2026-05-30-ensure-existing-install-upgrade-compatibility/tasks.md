# Tasks

## 1. Spec

- [x] Crear proposal/tasks y delta OpenSpec.
- [x] Validar `openspec validate ensure-existing-install-upgrade-compatibility --strict`.

## 2. Compatibilidad install/update

- [x] Cubrir que install default conserva OpenCode/OpenSpec y no instala `.lufy/sdd`.
- [x] Cubrir que sync default sobre una instalación existente no introduce assets `lufy-sdd`.
- [x] Cubrir que sync actualiza assets gestionados existentes y verify pasa después.

## 3. Validación

- [x] Ejecutar tests Go enfocados.
- [x] Ejecutar `scripts/validate.sh`.
- [x] Ejecutar `git diff --check origin/develop...HEAD` y `git diff --check`.
