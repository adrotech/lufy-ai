# Tasks

## 1. Spec

- [x] Crear proposal/design/tasks y deltas OpenSpec.
- [x] Validar `openspec validate add-lufy-sdd-methodology-adapter --strict`.

## 2. Adapter

- [x] Implementar `internal/adapters/methodology/lufysdd`.
- [x] Registrar el adapter en el registry default.
- [x] Cubrir `full`/`lite`, render y verify con tests.

## 3. CLI y docs

- [x] Mantener CLI mutante bloqueando `lufy-sdd` hasta integrar catalog/renderer.
- [x] Documentar estado foundation de Lufy SDD.

## 4. Validacion

- [x] `openspec validate add-lufy-sdd-methodology-adapter --strict`
- [x] `git diff --check`
- [x] `scripts/validate.sh`
