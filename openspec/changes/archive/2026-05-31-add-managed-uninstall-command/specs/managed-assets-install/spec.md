## ADDED Requirements

### Requirement: Uninstall gestionado de assets
La CLI SHALL proveer un comando `uninstall` que remueva de forma segura los assets gestionados por Lufy sin borrar archivos user-owned ni merge-managed.

#### Scenario: Uninstall planifica antes de mutar
- **WHEN** el usuario ejecuta `lufy-ai uninstall --target <dir> --dry-run`
- **THEN** la CLI SHALL mostrar los archivos gestionados que removería
- **AND** SHALL NOT borrar archivos, crear backups ni modificar estado

#### Scenario: Uninstall requiere confirmación
- **WHEN** el usuario ejecuta `lufy-ai uninstall --target <dir>` y existen mutaciones reales
- **THEN** la CLI SHALL fallar de forma accionable indicando que requiere `--yes`
- **AND** SHALL NOT borrar archivos ni escribir estado

#### Scenario: Uninstall real crea backup previo
- **WHEN** el usuario ejecuta `lufy-ai uninstall --target <dir> --yes`
- **THEN** la CLI SHALL crear un backup bajo `.lufy-ai/backups/<timestamp>/` antes de borrar cualquier archivo
- **AND** el backup SHALL incluir assets gestionados existentes, ancestors gestionados, `AGENTS.md` si contiene la referencia Lufy e `.lufy-ai/install-state.json`

#### Scenario: Uninstall remueve solo assets sin drift
- **WHEN** un asset registrado en `.lufy-ai/install-state.json` existe y su hash actual coincide con el hash registrado
- **THEN** uninstall SHALL remover ese archivo gestionado y su ancestor registrado cuando corresponda
- **AND** SHALL remover directorios gestionados que queden vacíos

#### Scenario: Drift local bloquea uninstall
- **WHEN** un asset gestionado existe pero su hash actual no coincide con el hash registrado
- **THEN** uninstall SHALL reportar conflicto
- **AND** SHALL NOT borrar archivos ni modificar estado

#### Scenario: Integraciones user-owned se preservan
- **WHEN** `AGENTS.md` contiene `@lufy-ia.harness.md`
- **THEN** uninstall SHALL remover solo esa referencia y preservar el resto del archivo
- **AND** SHALL NOT borrar `opencode.json` porque es merge-managed/user-owned

#### Scenario: Reinstall posterior funciona
- **WHEN** un target fue desinstalado exitosamente
- **AND** el usuario ejecuta `lufy-ai install --target <dir> --yes --no-engram`
- **THEN** la instalación SHALL reconstruir los assets gestionados y `verify` SHALL pasar
