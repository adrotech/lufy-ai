# Tasks

## 1. Spec y validacion inicial

- [x] Crear deltas OpenSpec para seleccion CLI de tool/metodologia.
- [x] Validar `openspec validate add-harness-selection-flags --strict`.

## 2. Parsing CLI y dominio

- [x] Agregar parser reusable de `--tool` y `--methodology-tier`.
- [x] Validar tools no soportadas con error claro.
- [x] Validar `none` inseguro en T1/T2 con error claro.
- [x] Cubrir parsing con tests unitarios.

## 3. Install / verify / status

- [x] Propagar `HarnessConfig` a `install` y persistir manifest v2 segun seleccion efectiva.
- [x] Permitir `verify --tool opencode` para validar expectativa contra manifest.
- [x] Exponer contexto efectivo en `status --json` y `verify --json`.
- [x] Mantener `install` sin flags equivalente al comportamiento actual.

## 4. Documentacion y ayuda

- [x] Actualizar help/README/docs para `--tool` y `--methodology-tier`.
- [x] Documentar que Codex/Claude y Lufy SDD requieren specs posteriores.

## 5. Validacion final

- [x] `openspec validate add-harness-selection-flags --strict`
- [x] `git diff --check`
- [x] `scripts/validate.sh`
