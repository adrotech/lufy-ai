# Tasks

## 1. Spec

- [x] Crear proposal/tasks y delta OpenSpec.
- [x] Validar `openspec validate encapsulate-opencode-runtime-config --strict`.

## 2. Runtime

- [x] Crear package runtime para project/global config de la tool efectiva.
- [x] Reemplazar acoples directos en install y sync.
- [x] Reemplazar acoples directos en verify y status.
- [x] Cubrir runtime con tests de opencode y rechazo explícito de tools no escribibles.

## 3. Validación

- [x] Ejecutar tests Go enfocados.
- [x] Ejecutar `scripts/validate.sh`.
- [x] Ejecutar `openspec validate --specs --strict`.
- [x] Ejecutar `git diff --check origin/develop...HEAD` y `git diff --check`.
