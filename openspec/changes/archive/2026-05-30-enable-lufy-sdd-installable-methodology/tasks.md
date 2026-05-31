# Tasks

## 1. Spec

- [x] Crear proposal/design/tasks y deltas OpenSpec.
- [x] Validar `openspec validate enable-lufy-sdd-installable-methodology --strict`.

## 2. Catalogo efectivo

- [x] Agregar assets `.lufy/sdd` al catalogo root y embebido.
- [x] Agregar filtro de catalogo por `HarnessConfig`.
- [x] Aplicar el filtro en install, sync y verify.
- [x] Cubrir ownership root/embedded y filtro full/lite con tests.

## 3. CLI

- [x] Permitir `lufy-sdd/full` y `lufy-sdd/lite` en `--methodology-tier`.
- [x] Mantener bloqueos de `none` inseguros y tools no escribibles.
- [x] Cubrir manifest resultante con tests CLI.

## 4. Docs

- [x] Documentar `lufy-sdd` como metodologia instalable inicial.
- [x] Mantener Codex/Claude como dry-run.

## 5. Validacion

- [x] `openspec validate enable-lufy-sdd-installable-methodology --strict`
- [x] `go test ./internal/assets ./internal/installer ./internal/syncer ./internal/verify ./internal/cli`
- [x] `git diff --check`
- [x] `scripts/validate.sh`
