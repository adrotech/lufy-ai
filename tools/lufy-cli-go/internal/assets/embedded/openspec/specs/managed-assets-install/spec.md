## Purpose
Definir la instalación gestionada de assets de `lufy-ai` mediante la CLI Go, incluyendo catálogo permitido, idempotencia por SHA-256, manifest de estado, backups/restores y verificación estructural confinada al target.
## Requirements
### Requirement: Catálogo completo de assets gestionados
La CLI Go SHALL instalar el conjunto completo de assets gestionados de `lufy-ai` desde el repo fuente o assets embebidos hacia un proyecto destino y SHALL distinguir assets completos gestionados de configuraciones merge-managed o integraciones user-owned especiales.

#### Scenario: Catálogo incluye assets requeridos
- **WHEN** la CLI construye el catálogo de instalación
- **THEN** el catálogo incluye `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/policies`, `.opencode/plugins`, `.opencode/agent-observatory`, `lufy-ia.harness.md`, `tui.json`, `openspec/` y metadatos requeridos bajo `.lufy/managed-state/`

#### Scenario: AGENTS es integración user-owned
- **WHEN** la CLI construye el catálogo de instalación
- **THEN** `AGENTS.md` no se trata como asset completo gestionado por SHA-256 y se maneja únicamente mediante la integración especial de referencia `@lufy-ia.harness.md`

#### Scenario: Catálogo excluye archivos fuera de alcance
- **WHEN** el repo fuente contiene archivos no listados dentro del set raíz permitido
- **THEN** la CLI MUST NOT copiarlos al target como parte de `install`

#### Scenario: Opencode es merge-managed especial
- **WHEN** la CLI planifica `opencode.json`
- **THEN** lo trata como `merge-json` especial fuera del catálogo de archivos completos con SHA-256 y preserva claves desconocidas del usuario

#### Scenario: Assets embebidos conservan paridad del harness
- **WHEN** la CLI se compila como binario standalone
- **THEN** el catálogo embebido contiene `lufy-ia.harness.md` con el mismo target, policy y SHA-256 que el catálogo raíz efectivo

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

### Requirement: Plan de instalación fiel
El comando `install` SHALL construir un plan explícito antes de cualquier mutación y `--dry-run` SHALL mostrar ese plan sin escribir.

#### Scenario: Plan para target vacío
- **WHEN** el target no contiene assets de `lufy-ai`
- **THEN** el plan incluye `create-dir` para directorios requeridos y `copy` para archivos gestionados ausentes

#### Scenario: Dry-run sin mutaciones
- **WHEN** el usuario ejecuta `lufy-ai install --target <dir> --dry-run`
- **THEN** la CLI muestra acciones planificadas y MUST NOT crear directorios, copiar archivos, crear backups ni escribir `.lufy/managed-state/install-state.json`

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
- **WHEN** `.lufy/managed-state/install-state.json` indica que un archivo fue gestionado previamente, el target no tiene drift local y el source hash actual difiere del source hash instalado
- **THEN** el plan incluye `backup` y `update-managed` para actualizarlo de forma segura

### Requirement: Conflictos no se sobrescriben silenciosamente
La CLI MUST detectar conflictos y MUST NOT sobrescribir archivos no gestionados o modificados localmente sin una decisión explícita soportada por la policy del asset.

#### Scenario: Archivo existente no gestionado
- **WHEN** un asset destino existe pero no aparece como gestionado en `.lufy/managed-state/install-state.json` y su policy no permite adopción o merge seguro
- **THEN** el plan lo marca como `conflict` y apply MUST NOT sobrescribirlo

#### Scenario: Archivo gestionado con drift local y policy managed
- **WHEN** un asset `managed` existe pero su hash actual no coincide con el último hash target registrado
- **THEN** el plan lo marca como `conflict` y apply MUST NOT sobrescribirlo como `update-managed`

#### Scenario: Archivo gestionado con drift local y policy no-replace
- **WHEN** un asset `no-replace` existe, tiene drift local y existe una nueva versión source
- **THEN** el plan usa una acción no destructiva que preserva el archivo original y escribe la nueva versión como `.lufy-new`

#### Scenario: Archivo gestionado con drift local y policy merge-block
- **WHEN** un asset `merge-block` existe, tiene drift local fuera de bloques lufy y los marcadores lufy son válidos
- **THEN** el plan puede actualizar solo los bloques lufy gestionados sin tratar el texto del usuario como conflicto

#### Scenario: Estado corrupto
- **WHEN** `.lufy/managed-state/install-state.json` existe pero no es JSON válido o usa schema no soportado
- **THEN** install MUST fail de forma accionable o marcar conflictos bloqueantes y MUST NOT asumir que archivos son seguros de sobrescribir

### Requirement: Manifest de estado de instalación
La CLI SHALL persistir estado de instalación en `.lufy/managed-state/install-state.json` con schema versionado, metadata real del binario, fingerprint estable del catalogo y hashes por asset completo gestionado, excluyendo configuraciones merge-managed especiales como `opencode.json` e integraciones user-owned especiales como `AGENTS.md`.

#### Scenario: Estado escrito tras install exitoso
- **WHEN** install aplica acciones exitosamente
- **THEN** escribe `.lufy/managed-state/install-state.json` con `schemaVersion`, `toolVersion`, `sourceChangeID`, `sourceRootFingerprint`, timestamps y lista de assets completos gestionados con `sourceSHA256` y `targetSHA256`, incluyendo `lufy-ia.harness.md` y excluyendo `AGENTS.md`

#### Scenario: Estado usa paths relativos
- **WHEN** la CLI registra assets en install state
- **THEN** cada asset usa paths relativos al target para portabilidad y trazabilidad

#### Scenario: Estado compatible con verify
- **WHEN** `verify` lee `.lufy/managed-state/install-state.json`
- **THEN** puede recalcular hashes destino y comparar contra el estado registrado

#### Scenario: Merge-managed no tiene hash completo
- **WHEN** `opencode.json` se crea o actualiza por `merge-json`
- **THEN** el manifest de estado no registra `opencode.json` como asset completo ni usa su hash para detectar drift

#### Scenario: AGENTS user-owned no tiene hash completo
- **WHEN** `AGENTS.md` se crea o recibe la referencia `@lufy-ia.harness.md` como integración user-owned
- **THEN** el manifest de estado no registra `AGENTS.md` como asset completo ni usa su hash para detectar drift

### Requirement: Backup multiasset antes de mutar
La CLI SHALL crear backups de todos los assets afectados antes de actualizaciones gestionadas o restore que sobrescriba archivos.

#### Scenario: Backup antes de update-managed
- **WHEN** install va a aplicar `update-managed` sobre uno o más archivos
- **THEN** crea un backup bajo `.lufy/managed-state/backups/<timestamp>/` antes de sobrescribir cualquier archivo existente

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
El comando `verify` SHALL validar estructura, estado y hashes de una instalación gestionada, además de validar configuraciones merge-managed e integraciones user-owned especiales sin exigir hash completo para esas integraciones.

#### Scenario: Instalación completa válida
- **WHEN** el target contiene todos los assets gestionados, `lufy-ia.harness.md` gestionado por manifest/hash, `AGENTS.md` con referencia al harness, `opencode.json` merge-managed válido y el install state coincide por hash para assets completos
- **THEN** `verify --target <dir>` reporta checks `ok` y exit code cero

#### Scenario: Asset crítico faltante
- **WHEN** falta un asset requerido del catálogo, incluyendo `lufy-ia.harness.md`
- **THEN** `verify` reporta `fail` y retorna exit code distinto de cero

#### Scenario: Drift local detectado
- **WHEN** un asset gestionado fue modificado en el target después de instalarse
- **THEN** `verify` reporta drift con hash esperado y hash actual

#### Scenario: Manifest ausente o inválido
- **WHEN** `.lufy/managed-state/install-state.json` falta o no puede parsearse
- **THEN** `verify` reporta `fail` o `warn` según severidad definida y explica la recuperación recomendada

#### Scenario: Opencode merge-managed inválido
- **WHEN** `opencode.json` falta, no parsea como JSON o carece de estructura merge-managed mínima
- **THEN** `verify` reporta `fail` para `opencode.json` sin buscarlo en `.lufy/managed-state/install-state.json` como asset completo

#### Scenario: Referencia AGENTS ausente o incompleta
- **WHEN** `AGENTS.md` falta o no contiene la referencia `@lufy-ia.harness.md`
- **THEN** `verify` reporta un requisito incumplido o warning accionable para la integración de `AGENTS.md` sin requerir hash completo ni entrada de manifest para ese archivo

### Requirement: Wrapper Bash permanece estricto
`scripts/install.sh` SHALL seguir delegando en la CLI Go y MUST NOT reintroducir lógica de instalación de assets.

#### Scenario: Install mediante wrapper
- **WHEN** el usuario ejecuta `scripts/install.sh <target>`
- **THEN** el wrapper delega en `lufy-ai install --target <target>` preservando flags compatibles

#### Scenario: Sin fallback legacy
- **WHEN** se inspecciona `scripts/install.sh`
- **THEN** no contiene lógica propia para copiar `.opencode`, `openspec`, `AGENTS.md`, `tui.json`, backups o hash/idempotencia

#### Scenario: Mutaciones confinadas
- **WHEN** install, backup, restore o verify operan con `--target <dir>`
- **THEN** cualquier escritura se limita al target y a `.lufy/managed-state/` dentro del target

### Requirement: Sync seguro de assets gestionados
El comando `sync` SHALL reaplicar assets gestionados desde el source hacia un target existente usando el catálogo permitido, estado de instalación, SHA-256 y políticas de conflicto existentes, y MUST treat `AGENTS.md` as user-owned reference integration rather than a managed file payload.

#### Scenario: Sync planifica antes de escribir
- **WHEN** el usuario ejecuta `lufy-ai sync --target <dir>`
- **THEN** la CLI construye un plan explícito antes de crear backups, copiar archivos, actualizar estado o modificar cualquier contenido

#### Scenario: Sync dry-run sin mutaciones
- **WHEN** el usuario ejecuta `lufy-ai sync --target <dir> --dry-run`
- **THEN** la CLI muestra acciones planificadas y MUST NOT crear directorios, copiar archivos, crear backups, reparar estado ni escribir `.lufy/managed-state/install-state.json`

#### Scenario: Sync usa catálogo permitido
- **WHEN** el source contiene archivos fuera del catálogo de assets gestionados
- **THEN** `sync` MUST NOT copiarlos ni registrarlos como gestionados

#### Scenario: Escritura de sync confinada al target
- **WHEN** una acción de sync normaliza a un path fuera de `--target`
- **THEN** la CLI MUST reject la acción y MUST NOT escribir archivos

#### Scenario: Sync actualiza harness sin mutar AGENTS
- **WHEN** `lufy-ia.harness.md` tiene upstream nuevo sin drift local y `AGENTS.md` contiene contenido user-owned
- **THEN** `sync` planifica backup y `update-managed` para `lufy-ia.harness.md`, actualiza su manifest al aplicar y MUST NOT modificar `AGENTS.md`

#### Scenario: Sync advierte referencia faltante
- **WHEN** `AGENTS.md` no contiene `@lufy-ia.harness.md` durante sync
- **THEN** `sync` reporta warning o acción explícita requerida y MUST NOT insertar la referencia silenciosamente

### Requirement: Sync idempotente por manifest y hash
El comando `sync` SHALL decidir acciones por hash de source actual, target actual y último estado registrado para preservar idempotencia y detectar cambios seguros.

#### Scenario: Asset gestionado sin cambios se omite
- **WHEN** un asset gestionado existe en el target, el hash target actual coincide con el hash target registrado y el hash source actual coincide con el hash source registrado
- **THEN** el plan de sync lo marca como `skip` y apply MUST NOT reescribirlo

#### Scenario: Upstream cambiado sin drift local se actualiza
- **WHEN** `.lufy/managed-state/install-state.json` indica que un asset fue gestionado previamente, el hash target actual coincide con el último hash target registrado y el hash source actual difiere del hash source registrado
- **THEN** el plan de sync incluye `backup` y `update-managed` para ese asset

#### Scenario: Segunda sync sin cambios
- **WHEN** el usuario ejecuta sync dos veces contra el mismo target sin cambios intermedios después de una actualización exitosa
- **THEN** la segunda ejecución produce `skip` para los assets ya sincronizados y no modifica contenido ni timestamps de archivos gestionados

#### Scenario: Asset retirado del catálogo se preserva rastreado
- **WHEN** un asset registrado previamente en `.lufy/managed-state/install-state.json` ya no existe en el catálogo source actual y el archivo target conserva el hash registrado
- **THEN** `sync` lo reporta como `retired`, no lo borra y mantiene su entrada en `.lufy/managed-state/install-state.json` para que siga siendo verificable o limpiable manualmente

#### Scenario: Estado actualizado tras sync exitoso
- **WHEN** sync aplica `update-managed` exitosamente sobre uno o más assets
- **THEN** escribe `.lufy/managed-state/install-state.json` con hashes source y target actualizados solo después de completar las mutaciones requeridas

### Requirement: Sync preserva personalizaciones fuera de scope
El comando `sync` MUST NOT sobrescribir archivos no gestionados, assets con drift local ni personalizaciones fuera del catálogo permitido; solo puede modificar drift local cuando la policy del asset define una estrategia segura y no destructiva.

#### Scenario: Asset managed con drift local bloqueado
- **WHEN** un asset `managed` existe pero su hash target actual no coincide con el último hash target registrado
- **THEN** el plan de sync lo marca como `conflict` y apply MUST NOT sobrescribirlo como `update-managed`

#### Scenario: Asset no-replace con drift local genera nueva versión
- **WHEN** un asset `no-replace` existe con drift local y el source actual difiere del source registrado
- **THEN** el plan de sync escribe la nueva versión en `.lufy-new` y conserva el target original

#### Scenario: Asset merge-block con drift local fuera de bloques preservado
- **WHEN** un asset `merge-block` contiene personalizaciones fuera de bloques lufy y bloques lufy válidos
- **THEN** sync preserva esas personalizaciones y actualiza solo los bloques gestionados que cambiaron

#### Scenario: Archivo existente no gestionado bloqueado
- **WHEN** un path del catálogo existe en el target pero no aparece como gestionado en `.lufy/managed-state/install-state.json` y su policy no permite merge/adopción segura
- **THEN** el plan de sync lo marca como `conflict` y apply MUST NOT sobrescribirlo

#### Scenario: Archivo desconocido del usuario preservado
- **WHEN** el target contiene archivos fuera del catálogo de assets gestionados
- **THEN** `sync` MUST NOT borrarlos, modificarlos ni registrarlos como assets gestionados

### Requirement: Catálogo incluye assets OpenSpec core v2
La CLI Go SHALL instalar los assets OpenSpec core v2 como parte del catálogo gestionado cuando estén presentes en la fuente de `lufy-ai`.

#### Scenario: Nuevos comandos y skills se instalan
- **WHEN** la CLI construye el catálogo desde una fuente con OpenSpec core v2
- **THEN** incluye `/opsx-sync`, `openspec-sync`, `opsx-version`, `openspec/config.yaml` v2 y `openspec/UPSTREAM.json` como assets gestionados

#### Scenario: Assets embebidos conservan paridad
- **WHEN** la CLI se compila como binario standalone
- **THEN** los assets OpenSpec core v2 embebidos coinciden con los assets raíz usados en desarrollo

### Requirement: Verify cubre baseline OpenSpec core v2
`lufy-ai verify` SHALL reportar el estado de los assets OpenSpec core v2 requeridos por la instalación gestionada.

#### Scenario: Baseline faltante falla verify
- **WHEN** un target instalado con catálogo core v2 no contiene `openspec/UPSTREAM.json`
- **THEN** `lufy-ai verify` reporta fallo accionable para el asset faltante

#### Scenario: Sync repara assets core v2 sin pisar drift
- **WHEN** un target existente carece de un nuevo asset OpenSpec core v2 y no hay conflicto local
- **THEN** `lufy-ai sync` planifica copiarlo y preserva las policies de drift existentes para cualquier archivo modificado por el usuario

#### Scenario: Opencode merge-json preserva personalizaciones
- **WHEN** `sync` aplica `merge-json` sobre un `opencode.json` válido con claves desconocidas
- **THEN** preserva esas claves, agrega solo estructura merge-managed mínima y no lo copia/reemplaza como asset completo por hash

#### Scenario: Opencode inválido bloquea sync
- **WHEN** `opencode.json` existente no es JSON válido
- **THEN** `sync` falla sin sobrescribirlo ni escribir estado nuevo

#### Scenario: Asset retirado con drift local bloqueado
- **WHEN** un asset registrado previamente ya no existe en el catálogo source actual pero su archivo target difiere del hash registrado
- **THEN** el plan de sync lo marca como `conflict` y apply MUST NOT borrarlo, reemplazarlo ni eliminarlo del estado gestionado

#### Scenario: Estado ausente o corrupto bloquea sobrescrituras
- **WHEN** `.lufy/managed-state/install-state.json` falta, no es JSON válido o usa schema no soportado
- **THEN** `sync` MUST fail de forma accionable o marcar conflictos bloqueantes y MUST NOT asumir que archivos existentes son seguros de sobrescribir

### Requirement: Sync crea backup antes de actualizaciones gestionadas
El comando `sync` SHALL crear un backup multiasset antes de sobrescribir cualquier asset gestionado existente.

#### Scenario: Backup previo a update-managed por sync
- **WHEN** sync va a aplicar `update-managed` sobre uno o más archivos existentes
- **THEN** crea un backup bajo `.lufy/managed-state/backups/<timestamp>/` antes de sobrescribir cualquier archivo

#### Scenario: Manifest de backup identifica sync
- **WHEN** sync crea un backup
- **THEN** el backup incluye `manifest.json` con paths relativos, hashes previos, acción causante `sync`, status de captura y ubicación de cada copia respaldada

#### Scenario: Error de sync después de backup
- **WHEN** una mutación de sync falla después de crear backup
- **THEN** la CLI reporta el path del manifest y la instrucción de restore necesaria

### Requirement: Verify posterior a sync
El resultado de un sync exitoso SHALL ser verificable mediante `verify` usando el mismo manifest y catálogo de assets gestionados.

#### Scenario: Verify después de sync exitoso
- **WHEN** sync actualiza assets gestionados y escribe estado exitosamente
- **THEN** `lufy-ai verify --target <dir>` puede recalcular hashes destino y reportar checks `ok` para los assets sincronizados

#### Scenario: Sync no degrada restore
- **WHEN** sync crea un backup antes de actualizar assets
- **THEN** `restore` puede usar ese manifest para restaurar los paths registrados de forma controlada y confinada al target

### Requirement: Install state uses real tool metadata
`.lufy/managed-state/install-state.json` SHALL record runtime tool metadata from the built CLI instead of hardcoded proposal-era constants.

#### Scenario: Release install records release metadata
- **WHEN** a release-built `lufy-ai` applies install or sync and writes `.lufy/managed-state/install-state.json`
- **THEN** `toolVersion` reflects `version.Current().Version` and does not remain hardcoded as `dev` unless the binary is actually a development build

#### Scenario: Development install is explicit
- **WHEN** a local development binary writes `.lufy/managed-state/install-state.json`
- **THEN** the state may record `dev`, but it does so through the same runtime version metadata path used by release builds

### Requirement: Source fingerprint reflects effective catalog
`.lufy/managed-state/install-state.json` SHALL store a `sourceRootFingerprint` derived from the effective managed asset catalog.

#### Scenario: Fingerprint written after install
- **WHEN** install writes `.lufy/managed-state/install-state.json`
- **THEN** `sourceRootFingerprint` equals the stable fingerprint of the catalog used for that install

#### Scenario: Fingerprint updated after sync
- **WHEN** sync writes `.lufy/managed-state/install-state.json` after applying managed updates
- **THEN** `sourceRootFingerprint` reflects the catalog used by that sync

### Requirement: Backup manifest uses real tool metadata
Backup manifests SHALL record runtime tool metadata from the CLI version source rather than a hardcoded state constant.

#### Scenario: Backup records runtime version
- **WHEN** install, sync, manual backup or restore recovery creates `.lufy/managed-state/backups/<timestamp>/manifest.json`
- **THEN** `toolVersion` is populated from runtime CLI version metadata

### Requirement: Managed copies are atomic
Install, sync, backup and restore SHALL avoid direct non-atomic writes for managed file payloads.

#### Scenario: Interrupted copy cannot leave truncated managed file
- **WHEN** a managed file payload is being copied into a target or backup destination
- **THEN** the implementation writes via temp file plus rename in the destination directory instead of writing directly to the final path

### Requirement: Manifest registra policy scope y ancestor
La CLI SHALL persistir metadata de policy, scope y ancestor por asset gestionado para que install, sync, verify y status puedan tomar decisiones consistentes.

#### Scenario: Estado nuevo incluye metadata de drift
- **WHEN** install o sync escribe `.lufy/managed-state/install-state.json`
- **THEN** cada asset gestionado incluye `policy`, `scope` y referencia de ancestor cuando aplique

#### Scenario: Estado anterior migra con defaults seguros
- **WHEN** la CLI lee un install state de schema anterior sin `policy` ni `scope`
- **THEN** completa defaults compatibles sin perder hashes, paths ni timestamps existentes

### Requirement: Catálogo declara scope efectivo
El catálogo de assets SHALL declarar si cada entry se instala en scope project, global o both.

#### Scenario: Entry project-only permanece local
- **WHEN** el catálogo marca un asset como project-only
- **THEN** install y sync lo planifican bajo el project target incluso cuando el usuario solicita scope global para assets compartidos

#### Scenario: Entry global shared usa config global
- **WHEN** el catálogo marca un asset como global-shared y el usuario solicita `--scope=global` o `--scope=both`
- **THEN** install y sync lo planifican bajo el directorio global OpenCode resuelto

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
- **THEN** la CLI SHALL crear un backup bajo `.lufy/managed-state/backups/<timestamp>/` antes de borrar cualquier archivo
- **AND** el backup SHALL incluir assets gestionados existentes, ancestors gestionados, `AGENTS.md` si contiene la referencia Lufy e `.lufy/managed-state/install-state.json`

#### Scenario: Uninstall remueve solo assets sin drift
- **WHEN** un asset registrado en `.lufy/managed-state/install-state.json` existe y su hash actual coincide con el hash registrado
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
- **AND** el usuario ejecuta `lufy-ai install --target <dir> --yes`
- **THEN** la instalación SHALL reconstruir los assets gestionados y `verify` SHALL pasar
