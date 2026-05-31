# Tasks

## 1. Spec

- [x] Crear proposal/design/tasks y delta OpenSpec.
- [x] Validar `openspec validate route-current-preset-through-adapters --strict`.

## 2. Adapter catalog

- [x] Implementar resolver de catalogo efectivo basado en `ToolAdapter` y `MethodologyAdapter`.
- [x] Usar el resolver en install, sync y verify.
- [x] Ajustar `lufy-sdd` para declarar assets instalables.
- [x] Mantener bloqueo de Codex/Claude para escritura.

## 3. Compatibilidad

- [x] Cubrir que el default conserva assets OpenCode/OpenSpec actuales.
- [x] Cubrir que `lufy-sdd/lite` y `lufy-sdd/full` filtran assets correctamente.
- [x] Cubrir que un adapter no registrado falla explicitamente.

## 4. Validacion

- [x] `openspec validate route-current-preset-through-adapters --strict`
- [x] `go test ./internal/harnesscatalog ./internal/assets ./internal/installer ./internal/syncer ./internal/verify ./internal/cli`
- [x] `git diff --check`
- [x] `scripts/validate.sh`
