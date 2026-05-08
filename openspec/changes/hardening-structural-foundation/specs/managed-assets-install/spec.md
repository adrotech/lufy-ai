## ADDED Requirements

### Requirement: Install state uses real tool metadata
`.lufy-ai/install-state.json` SHALL record runtime tool metadata from the built CLI instead of hardcoded proposal-era constants.

#### Scenario: Release install records release metadata
- **WHEN** a release-built `lufy-ai` applies install or sync and writes `.lufy-ai/install-state.json`
- **THEN** `toolVersion` reflects `version.Current().Version` and does not remain hardcoded as `dev` unless the binary is actually a development build

#### Scenario: Development install is explicit
- **WHEN** a local development binary writes `.lufy-ai/install-state.json`
- **THEN** the state may record `dev`, but it does so through the same runtime version metadata path used by release builds

### Requirement: Source fingerprint reflects effective catalog
`.lufy-ai/install-state.json` SHALL store a `sourceRootFingerprint` derived from the effective managed asset catalog.

#### Scenario: Fingerprint written after install
- **WHEN** install writes `.lufy-ai/install-state.json`
- **THEN** `sourceRootFingerprint` equals the stable fingerprint of the catalog used for that install

#### Scenario: Fingerprint updated after sync
- **WHEN** sync writes `.lufy-ai/install-state.json` after applying managed updates
- **THEN** `sourceRootFingerprint` reflects the catalog used by that sync

### Requirement: Backup manifest uses real tool metadata
Backup manifests SHALL record runtime tool metadata from the CLI version source rather than a hardcoded state constant.

#### Scenario: Backup records runtime version
- **WHEN** install, sync, manual backup or restore recovery creates `.lufy-ai/backups/<timestamp>/manifest.json`
- **THEN** `toolVersion` is populated from runtime CLI version metadata

### Requirement: Managed copies are atomic
Install, sync, backup and restore SHALL avoid direct non-atomic writes for managed file payloads.

#### Scenario: Interrupted copy cannot leave truncated managed file
- **WHEN** a managed file payload is being copied into a target or backup destination
- **THEN** the implementation writes via temp file plus rename in the destination directory instead of writing directly to the final path

## MODIFIED Requirements

### Requirement: Manifest de estado de instalación
La CLI SHALL persistir estado de instalación en `.lufy-ai/install-state.json` con schema versionado, metadata real del binario, fingerprint estable del catalogo y hashes por asset completo gestionado, excluyendo configuraciones merge-managed especiales como `opencode.json`.

#### Scenario: Estado escrito tras install exitoso
- **WHEN** install aplica acciones exitosamente
- **THEN** escribe `.lufy-ai/install-state.json` con `schemaVersion`, `toolVersion`, `sourceChangeID`, `sourceRootFingerprint`, timestamps y lista de assets completos gestionados con `sourceSHA256` y `targetSHA256`

#### Scenario: Estado usa paths relativos
- **WHEN** la CLI registra assets en install state
- **THEN** cada asset usa paths relativos al target para portabilidad y trazabilidad

#### Scenario: Estado compatible con verify
- **WHEN** `verify` lee `.lufy-ai/install-state.json`
- **THEN** puede recalcular hashes destino y comparar contra el estado registrado

#### Scenario: Merge-managed no tiene hash completo
- **WHEN** `opencode.json` se crea o actualiza por `merge-json`
- **THEN** el manifest de estado no registra `opencode.json` como asset completo ni usa su hash para detectar drift

### Requirement: Resolución segura de source y target
La CLI SHALL resolver el source root y el target project root de forma portable y segura antes de planificar o escribir, y SHALL rechazar paths relativos que escapen del root con separadores Unix, Windows o mixtos.

#### Scenario: Source root detectado desde checkout
- **WHEN** la CLI se ejecuta desde un checkout de desarrollo
- **THEN** detecta el source root usando marcadores del repo como `AGENTS.md`, `.opencode/` y `openspec/config.yaml`

#### Scenario: Target default
- **WHEN** el usuario omite `--target`
- **THEN** la CLI usa `.` como target y lo resuelve a una ruta absoluta/canonical antes de construir el plan

#### Scenario: Escritura fuera del target bloqueada
- **WHEN** una acción planificada normaliza a un path fuera de `--target`
- **THEN** la CLI MUST reject la acción y MUST NOT escribir archivos

#### Scenario: Symlink peligroso bloqueado
- **WHEN** un path source o target usa un symlink que escapa del root permitido
- **THEN** la CLI MUST reportar error y MUST NOT seguir el symlink para copiar o escribir
