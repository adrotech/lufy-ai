## 1. OpenSpec

- [x] Crear propuesta, tareas y spec delta para setup/version preflight.
- [x] Validar propuesta con `openspec validate "add-setup-wizard-and-version-preflight" --strict`.

## 2. Version preflight

- [x] Agregar servicio `versioncheck` para consultar ultima release estable y comparar semver.
- [x] Cubrir update disponible, al dia, build dev y errores de red.
- [x] Integrar version preflight como primer paso de setup.

## 3. Setup plan/aplicacion

- [x] Crear servicio `setup` con plan de features instalables/configurables.
- [x] Detectar estado de install, project config, memory, context graph y verify.
- [x] Detectar layout, stack/superficies y metodologia SDD como features explicitas.
- [x] Exponer registry versionado de features con metadata `since`.
- [x] Aplicar acciones con `--yes` reutilizando servicios existentes.
- [x] Mantener `--dry-run` sin mutaciones.

## 4. CLI

- [x] Agregar comando `lufy-ai setup` y help.
- [x] Soportar `--target`, `--dry-run`, `--yes`, `--json`, `--skip-version-check`, `--require-latest`, `--check-new-features`.
- [x] Agregar selector interactivo TUI cuando hay TTY y no se usan flags no interactivos.
- [x] Sugerir setup post-upgrade.

## 5. Validacion final

- [x] Ejecutar tests Go aplicables.
- [x] Ejecutar `scripts/validate.sh`.
- [x] Ejecutar `openspec validate "add-setup-wizard-and-version-preflight" --strict`.
- [x] Revisar diff final y evidenciar limitaciones.
