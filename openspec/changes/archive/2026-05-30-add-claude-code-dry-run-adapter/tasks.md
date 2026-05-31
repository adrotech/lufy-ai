# Tasks

## 1. Spec

- [x] Crear proposal/design/tasks y deltas OpenSpec.
- [x] Validar `openspec validate add-claude-code-dry-run-adapter --strict`.

## 2. Dominio y registry

- [x] Agregar `claude-code` como `ToolID` conocido.
- [x] Registrar adapter Claude Code dry-run sin hacerlo escribible por CLI.
- [x] Cubrir registry/capabilities con tests.

## 3. Render dry-run y leak checks

- [x] Implementar `internal/adapters/tool/claudecode`.
- [x] Renderizar preview conceptual para `CLAUDE.md`.
- [x] Agregar tests que bloqueen referencias OpenCode.

## 4. Documentacion

- [x] Actualizar docs para explicar que Claude Code existe solo como dry-run adapter.
- [x] Mantener OpenCode/OpenSpec como preset default.

## 5. Validacion

- [x] `openspec validate add-claude-code-dry-run-adapter --strict`
- [x] `git diff --check`
- [x] `scripts/validate.sh`
