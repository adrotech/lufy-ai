## ADDED Requirements

### Requirement: Catálogo completo de assets gestionados
La CLI Go SHALL instalar el conjunto completo de assets gestionados de `lufy-ai` desde el repo fuente hacia un proyecto destino.

#### Scenario: Catálogo incluye assets requeridos
- **WHEN** la CLI construye el catálogo de instalación
- **THEN** el catálogo incluye `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/policies`, `.opencode/plugins`, `.opencode/agent-observatory`, `AGENTS.md`, `tui.json`, `openspec/` y metadatos requeridos bajo `.lufy-ai/`

#### Scenario: Catálogo excluye archivos fuera de alcance
- **WHEN** el repo fuente contiene archivos no listados dentro del set raíz permitido
- **THEN** la CLI MUST NOT copiarlos al target como parte de `install`

### Requirement: Resolución segura de source y target
La CLI SHALL resolver el source root y el target project root de forma portable y segura antes de planificar o escribir.

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

### Requirement: Plan de instalación fiel
El comando `install` SHALL construir un plan explícito antes de cualquier mutación y `--dry-run` SHALL mostrar ese plan sin escribir.

#### Scenario: Plan para target vacío
- **WHEN** el target no contiene assets de `lufy-ai`
- **THEN** el plan incluye `create-dir` para directorios requeridos y `copy` para archivos gestionados ausentes

#### Scenario: Dry-run sin mutaciones
- **WHEN** el usuario ejecuta `lufy-ai install --target <dir> --dry-run`
- **THEN** la CLI muestra acciones planificadas y MUST NOT crear directorios, copiar archivos, crear backups ni escribir `.lufy-ai/install-state.json`

#### Scenario: Acciones explicables
- **WHEN** el plan contiene acciones
- **THEN** cada acción identifica tipo (`create-dir`, `copy`, `skip`, `update-managed`, `conflict` o `backup`), path relativo, razón y hashes relevantes cuando existan

### Requirement: Idempotencia por contenido/hash
La instalación SHALL decidir acciones por contenido/hash y MUST ser idempotente en ejecuciones repetidas sin cambios.

#### Scenario: Archivo ausente se copia
- **WHEN** un asset gestionado no existe en el target
- **THEN** el plan lo marca como `copy` y apply lo copia desde source preservando path relativo

#### Scenario: Archivo igual se omite
- **WHEN** un asset gestionado existe en el target y su SHA-256 coincide con el source actual
- **THEN** el plan lo marca como `skip` y apply MUST NOT reescribirlo

#### Scenario: Segunda instalación sin cambios
- **WHEN** el usuario ejecuta install dos veces contra el mismo target sin cambios intermedios
- **THEN** la segunda ejecución produce `skip` para assets ya instalados y no modifica contenido ni timestamps de archivos gestionados salvo el estado si necesita reparación no destructiva

#### Scenario: Upstream cambiado en archivo gestionado
- **WHEN** `.lufy-ai/install-state.json` indica que un archivo fue gestionado previamente, el target no tiene drift local y el source hash actual difiere del source hash instalado
- **THEN** el plan incluye `backup` y `update-managed` para actualizarlo de forma segura

### Requirement: Conflictos no se sobrescriben silenciosamente
La CLI MUST detectar conflictos y MUST NOT sobrescribir archivos no gestionados o modificados localmente sin una decisión explícita soportada.

#### Scenario: Archivo existente no gestionado
- **WHEN** un asset destino existe pero no aparece como gestionado en `.lufy-ai/install-state.json`
- **THEN** el plan lo marca como `conflict` y apply MUST NOT sobrescribirlo

#### Scenario: Archivo gestionado con drift local
- **WHEN** un asset gestionado existe pero su hash actual no coincide con el último hash target registrado
- **THEN** el plan lo marca como `conflict` y apply MUST NOT sobrescribirlo como `update-managed`

#### Scenario: Estado corrupto
- **WHEN** `.lufy-ai/install-state.json` existe pero no es JSON válido o usa schema no soportado
- **THEN** install MUST fail de forma accionable o marcar conflictos bloqueantes y MUST NOT asumir que archivos son seguros de sobrescribir

### Requirement: Manifest de estado de instalación
La CLI SHALL persistir estado de instalación en `.lufy-ai/install-state.json` con schema versionado y hashes por asset.

#### Scenario: Estado escrito tras install exitoso
- **WHEN** install aplica acciones exitosamente
- **THEN** escribe `.lufy-ai/install-state.json` con `schemaVersion`, `toolVersion`, `sourceChangeID`, timestamps y lista de assets con `sourceSHA256` y `targetSHA256`

#### Scenario: Estado usa paths relativos
- **WHEN** la CLI registra assets en install state
- **THEN** cada asset usa paths relativos al target para portabilidad y trazabilidad

#### Scenario: Estado compatible con verify
- **WHEN** `verify` lee `.lufy-ai/install-state.json`
- **THEN** puede recalcular hashes destino y comparar contra el estado registrado

### Requirement: Backup multiasset antes de mutar
La CLI SHALL crear backups de todos los assets afectados antes de actualizaciones gestionadas o restore que sobrescriba archivos.

#### Scenario: Backup antes de update-managed
- **WHEN** install va a aplicar `update-managed` sobre uno o más archivos
- **THEN** crea un backup bajo `.lufy-ai/backups/<timestamp>/` antes de sobrescribir cualquier archivo existente

#### Scenario: Manifest de backup completo
- **WHEN** se crea un backup
- **THEN** el backup incluye `manifest.json` con paths relativos, hashes previos, acción causante, status de captura y ubicación de cada copia respaldada

#### Scenario: Error después de backup
- **WHEN** una mutación falla después de crear backup
- **THEN** la CLI reporta el path del manifest y la instrucción de restore necesaria

### Requirement: Restore controlado
El comando `restore` SHALL restaurar backups multiasset de forma controlada, verificable y confinada al target.

#### Scenario: Restore dry-run
- **WHEN** el usuario ejecuta `lufy-ai restore --target <dir> --backup <manifest-or-dir> --dry-run`
- **THEN** la CLI muestra qué archivos restauraría y MUST NOT escribir cambios

#### Scenario: Restore real
- **WHEN** el usuario ejecuta restore real con un manifest válido
- **THEN** la CLI restaura solo los paths registrados dentro del target y reporta el resultado por asset

#### Scenario: Manifest con path escape
- **WHEN** un manifest de backup contiene un path absoluto o que escapa del target
- **THEN** restore MUST reject el manifest y MUST NOT escribir archivos

#### Scenario: Backup antes de restore destructivo
- **WHEN** restore va a sobrescribir archivos existentes
- **THEN** la CLI crea un backup del estado actual antes de aplicar la restauración

### Requirement: Verify estructural completo
El comando `verify` SHALL validar estructura, estado y hashes de una instalación gestionada.

#### Scenario: Instalación completa válida
- **WHEN** el target contiene todos los assets gestionados y el install state coincide por hash
- **THEN** `verify --target <dir>` reporta checks `ok` y exit code cero

#### Scenario: Asset crítico faltante
- **WHEN** falta un asset requerido del catálogo
- **THEN** `verify` reporta `fail` y retorna exit code distinto de cero

#### Scenario: Drift local detectado
- **WHEN** un asset gestionado fue modificado en el target después de instalarse
- **THEN** `verify` reporta drift con hash esperado y hash actual

#### Scenario: Manifest ausente o inválido
- **WHEN** `.lufy-ai/install-state.json` falta o no puede parsearse
- **THEN** `verify` reporta `fail` o `warn` según severidad definida y explica la recuperación recomendada

### Requirement: Wrapper Bash permanece estricto
`scripts/install.sh` SHALL seguir delegando en la CLI Go y MUST NOT reintroducir lógica de instalación de assets.

#### Scenario: Install mediante wrapper
- **WHEN** el usuario ejecuta `scripts/install.sh <target>`
- **THEN** el wrapper delega en `lufy-ai install --target <target>` preservando flags compatibles

#### Scenario: Sin fallback legacy
- **WHEN** se inspecciona `scripts/install.sh`
- **THEN** no contiene lógica propia para copiar `.opencode`, `openspec`, `AGENTS.md`, `tui.json`, backups o hash/idempotencia

### Requirement: Seguridad y Engram portable
La CLI MUST mantener la instalación confinada y MUST NOT hardcodear rutas locales de Engram.

#### Scenario: Ausencia de Engram
- **WHEN** Engram no existe en `PATH`
- **THEN** la instalación base de assets gestionados continúa o reporta nota no bloqueante según flags, pero MUST NOT fallar por una ruta hardcodeada

#### Scenario: Ruta hardcodeada prohibida
- **WHEN** la CLI genera configuración o verifica assets
- **THEN** el contenido generado por este slice MUST NOT contener `/opt/homebrew/bin/engram`

#### Scenario: Mutaciones confinadas
- **WHEN** install, backup, restore o verify operan con `--target <dir>`
- **THEN** cualquier escritura se limita al target y a `.lufy-ai/` dentro del target
