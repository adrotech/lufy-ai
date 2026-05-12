## MODIFIED Requirements

### Requirement: Conflictos no se sobrescriben silenciosamente
La CLI MUST detectar conflictos y MUST NOT sobrescribir archivos no gestionados o modificados localmente sin una decisión explícita soportada por la policy del asset.

#### Scenario: Archivo existente no gestionado
- **WHEN** un asset destino existe pero no aparece como gestionado en `.lufy-ai/install-state.json` y su policy no permite adopción o merge seguro
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
- **WHEN** `.lufy-ai/install-state.json` existe pero no es JSON válido o usa schema no soportado
- **THEN** install MUST fail de forma accionable o marcar conflictos bloqueantes y MUST NOT asumir que archivos son seguros de sobrescribir

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
- **WHEN** un path del catálogo existe en el target pero no aparece como gestionado en `.lufy-ai/install-state.json` y su policy no permite merge/adopción segura
- **THEN** el plan de sync lo marca como `conflict` y apply MUST NOT sobrescribirlo

#### Scenario: Archivo desconocido del usuario preservado
- **WHEN** el target contiene archivos fuera del catálogo de assets gestionados
- **THEN** `sync` MUST NOT borrarlos, modificarlos ni registrarlos como assets gestionados

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
- **WHEN** `.lufy-ai/install-state.json` falta, no es JSON válido o usa schema no soportado
- **THEN** `sync` MUST fail de forma accionable o marcar conflictos bloqueantes y MUST NOT asumir que archivos existentes son seguros de sobrescribir

## ADDED Requirements

### Requirement: Manifest registra policy scope y ancestor
La CLI SHALL persistir metadata de policy, scope y ancestor por asset gestionado para que install, sync, verify y status puedan tomar decisiones consistentes.

#### Scenario: Estado nuevo incluye metadata de drift
- **WHEN** install o sync escribe `.lufy-ai/install-state.json`
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
