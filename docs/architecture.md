# Arquitectura

`lufy-ai` distribuye una CLI Go en `tools/lufy-cli-go` y assets gestionados para OpenCode/OpenSpec.

## Componentes

- `internal/assets`: catálogo, hashes SHA-256 y assets embebidos.
- `internal/opsx`: resolución stay-updated de OpenSpec en capas `PATH`, cache local y baseline embebida.
- `internal/installer`: plan/apply de instalación inicial y actualización gestionada.
- `internal/syncer`: sincronización conservadora de assets registrados.
- `internal/backup`: backups, manifests, restore y rollback acotado.
- `internal/verify`: verificación estructural, drift y salida JSON.
- `internal/config`: merge conservador de `opencode.json`.
- `internal/platform`: path safety, locks y resolución portable de targets.
- `scripts/bootstrap.sh`: descarga verificada del binario publicado.

## Decisiones relevantes

- Ver `docs/adr/0001-managed-asset-source-of-truth.md` para fuente de verdad de assets.
- Ver `docs/adr/0002-managed-directories-extra-files.md` para archivos extra en directorios gestionados.
- Ver `docs/adr/0003-opencode-json-ownership.md` para ownership parcial de `opencode.json`.
- Ver `docs/adr/0004-sync-new-catalog-assets.md` para semántica de nuevos assets en `sync`.
- Ver `docs/adr/0005-recovery-and-rollback.md` para rollback acotado.
- `openspec/UPSTREAM.json` declara la baseline OpenSpec, versión mínima compatible y cache `.lufy-ai/openspec-cache/<version>/`; el instalador no descarga OpenSpec remoto durante `install`/`sync` por defecto.
