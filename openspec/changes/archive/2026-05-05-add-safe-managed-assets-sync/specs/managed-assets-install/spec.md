## ADDED Requirements

### Requirement: Sync seguro de assets gestionados
El comando `sync` SHALL reaplicar assets gestionados desde el source hacia un target existente usando el catálogo permitido, estado de instalación, SHA-256 y políticas de conflicto existentes.

#### Scenario: Sync planifica antes de escribir
- **WHEN** el usuario ejecuta `lufy-ai sync --target <dir>`
- **THEN** la CLI construye un plan explícito antes de crear backups, copiar archivos, actualizar estado o modificar cualquier contenido

#### Scenario: Sync dry-run sin mutaciones
- **WHEN** el usuario ejecuta `lufy-ai sync --target <dir> --dry-run`
- **THEN** la CLI muestra acciones planificadas y MUST NOT crear directorios, copiar archivos, crear backups, reparar estado ni escribir `.lufy-ai/install-state.json`

#### Scenario: Sync usa catálogo permitido
- **WHEN** el source contiene archivos fuera del catálogo de assets gestionados
- **THEN** `sync` MUST NOT copiarlos ni registrarlos como gestionados

#### Scenario: Escritura de sync confinada al target
- **WHEN** una acción de sync normaliza a un path fuera de `--target`
- **THEN** la CLI MUST reject la acción y MUST NOT escribir archivos

### Requirement: Sync idempotente por manifest y hash
El comando `sync` SHALL decidir acciones por hash de source actual, target actual y último estado registrado para preservar idempotencia y detectar cambios seguros.

#### Scenario: Asset gestionado sin cambios se omite
- **WHEN** un asset gestionado existe en el target, el hash target actual coincide con el hash target registrado y el hash source actual coincide con el hash source registrado
- **THEN** el plan de sync lo marca como `skip` y apply MUST NOT reescribirlo

#### Scenario: Upstream cambiado sin drift local se actualiza
- **WHEN** `.lufy-ai/install-state.json` indica que un asset fue gestionado previamente, el hash target actual coincide con el último hash target registrado y el hash source actual difiere del hash source registrado
- **THEN** el plan de sync incluye `backup` y `update-managed` para ese asset

#### Scenario: Segunda sync sin cambios
- **WHEN** el usuario ejecuta sync dos veces contra el mismo target sin cambios intermedios después de una actualización exitosa
- **THEN** la segunda ejecución produce `skip` para los assets ya sincronizados y no modifica contenido ni timestamps de archivos gestionados

#### Scenario: Asset retirado del catálogo se preserva rastreado
- **WHEN** un asset registrado previamente en `.lufy-ai/install-state.json` ya no existe en el catálogo source actual y el archivo target conserva el hash registrado
- **THEN** `sync` lo reporta como `retired`, no lo borra y mantiene su entrada en `.lufy-ai/install-state.json` para que siga siendo verificable o limpiable manualmente

#### Scenario: Estado actualizado tras sync exitoso
- **WHEN** sync aplica `update-managed` exitosamente sobre uno o más assets
- **THEN** escribe `.lufy-ai/install-state.json` con hashes source y target actualizados solo después de completar las mutaciones requeridas

### Requirement: Sync preserva personalizaciones fuera de scope
El comando `sync` MUST NOT sobrescribir archivos no gestionados, assets con drift local ni personalizaciones fuera del catálogo permitido.

#### Scenario: Asset gestionado con drift local bloqueado
- **WHEN** un asset gestionado existe pero su hash target actual no coincide con el último hash target registrado
- **THEN** el plan de sync lo marca como `conflict` y apply MUST NOT sobrescribirlo como `update-managed`

#### Scenario: Archivo existente no gestionado bloqueado
- **WHEN** un path del catálogo existe en el target pero no aparece como gestionado en `.lufy-ai/install-state.json`
- **THEN** el plan de sync lo marca como `conflict` y apply MUST NOT sobrescribirlo

#### Scenario: Archivo desconocido del usuario preservado
- **WHEN** el target contiene archivos fuera del catálogo de assets gestionados
- **THEN** `sync` MUST NOT borrarlos, modificarlos ni registrarlos como assets gestionados

#### Scenario: Asset retirado con drift local bloqueado
- **WHEN** un asset registrado previamente ya no existe en el catálogo source actual pero su archivo target difiere del hash registrado
- **THEN** el plan de sync lo marca como `conflict` y apply MUST NOT borrarlo, reemplazarlo ni eliminarlo del estado gestionado

#### Scenario: Estado ausente o corrupto bloquea sobrescrituras
- **WHEN** `.lufy-ai/install-state.json` falta, no es JSON válido o usa schema no soportado
- **THEN** `sync` MUST fail de forma accionable o marcar conflictos bloqueantes y MUST NOT asumir que archivos existentes son seguros de sobrescribir

### Requirement: Sync crea backup antes de actualizaciones gestionadas
El comando `sync` SHALL crear un backup multiasset antes de sobrescribir cualquier asset gestionado existente.

#### Scenario: Backup previo a update-managed por sync
- **WHEN** sync va a aplicar `update-managed` sobre uno o más archivos existentes
- **THEN** crea un backup bajo `.lufy-ai/backups/<timestamp>/` antes de sobrescribir cualquier archivo

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
