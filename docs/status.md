# Estado del proyecto

## Implementado

- CLI Go en `tools/lufy-cli-go`.
- Instalación y sync con catálogo, hashes SHA-256 y manifest de estado.
- Backups con manifest, retención local y restore validado.
- Rollback automático acotado cuando existe backup de recovery.
- `verify` estructural con salida humana, `--json` y `--quiet`.
- `status` con salida humana y `--json`.
- `upgrade` autoreemplazante con versión fija y verificación SHA-256.
- `verify --deep` para referencias de plugins en `tui.json` y `opencode.json`.
- Bootstrap remoto con checksum, validación de tar entries, retry y timeouts.
- Release con actions pinneadas, SBOM, provenance y firma cosign.

## Pendiente o futuro

- Autocomplete/help avanzado mediante Cobra u otro framework.
- Verificación cosign integrada en `upgrade`.
- Deep verify de plugins y schemas externos.
- Two-phase apply completo si el rollback acotado resulta insuficiente.
- Migración a GoReleaser si reduce mantenimiento frente al script actual.
