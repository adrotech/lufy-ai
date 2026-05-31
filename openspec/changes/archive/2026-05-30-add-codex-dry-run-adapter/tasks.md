# Tasks

## 1. Spec

- [x] Crear proposal/design/tasks y deltas OpenSpec.
- [x] Validar `openspec validate add-codex-dry-run-adapter --strict`.

## 2. Dominio y registry

- [x] Agregar `codex` como `ToolID` conocido.
- [x] Agregar `DryRunOnly` a capabilities de tool.
- [x] Registrar adapter Codex dry-run sin hacerlo escribible por CLI.
- [x] Cubrir registry/capabilities con tests.

## 3. Render dry-run y leak checks

- [x] Implementar `internal/adapters/tool/codex`.
- [x] Renderizar preview conceptual para `AGENTS.md` con fallback inline.
- [x] Agregar tests que bloqueen referencias `.opencode` y `opencode.json`.

## 4. Documentacion

- [x] Actualizar docs para explicar que Codex existe solo como dry-run adapter.
- [x] Mantener OpenCode/OpenSpec como preset default.

## 5. Validacion

- [x] `openspec validate add-codex-dry-run-adapter --strict`
- [x] `git diff --check`
- [x] Tests Go focalizados
