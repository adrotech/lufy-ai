## Why

`opencode.json` es configuración del repositorio destino y puede contener proveedores, modelos, MCPs o claves locales del usuario. Tratarlo como asset completo gestionado por hash crea riesgo de sobrescritura y conflictos falsos durante `install`, `sync` y `verify`.

## What Changes

- Definir `opencode.json` como asset especial `merge-json`, no como archivo completo `managed` con SHA-256 en `.lufy-ai/install-state.json`.
- Mantener merge conservador en `install`: crear el archivo cuando falta, preservar claves desconocidas y fallar sin overwrite si el JSON existente es inválido.
- Permitir que `sync` aplique el mismo merge seguro cuando corresponde, sin `copy`/`update-managed` por hash para `opencode.json`.
- Extender `verify` para validar JSON y estructura mínima merge-managed de `opencode.json` sin exigir entrada de hash en el manifest.
- Documentar el contrato en roadmap y documentación de CLI.

## Capabilities

### Modified Capabilities

- `go-cli-installer`: contrato de `install`, `sync` y `verify` para `opencode.json` merge-managed.
- `managed-assets-install`: catálogo/manifest e idempotencia distinguen assets completos gestionados de JSON merge-managed.

## Impact

- `tools/lufy-cli-go/internal/config/`: planificación, merge y validación mínima de `opencode.json`.
- `tools/lufy-cli-go/internal/installer/`: install plan/apply y tests de preservación/JSON inválido.
- `tools/lufy-cli-go/internal/syncer/`: sync plan/apply y tests de merge-json sin hash completo.
- `tools/lufy-cli-go/internal/verify/`: verify estructural de JSON merge-managed.
- `docs/roadmap.md`, `docs/getting-started.md`, `tools/lufy-cli-go/README.md`: documentación del contrato.
