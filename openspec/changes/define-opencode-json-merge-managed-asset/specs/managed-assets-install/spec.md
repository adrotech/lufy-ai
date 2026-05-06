## MODIFIED Requirements

### Requirement: Catálogo completo de assets gestionados
La CLI Go SHALL instalar el conjunto completo de assets gestionados de `lufy-ai` desde el repo fuente hacia un proyecto destino y SHALL distinguir assets completos gestionados de configuraciones merge-managed especiales.

#### Scenario: Catálogo incluye assets requeridos
- **WHEN** la CLI construye el catálogo de instalación
- **THEN** el catálogo incluye `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/policies`, `.opencode/plugins`, `.opencode/agent-observatory`, `AGENTS.md`, `tui.json`, `openspec/` y metadatos requeridos bajo `.lufy-ai/`

#### Scenario: Catálogo excluye archivos fuera de alcance
- **WHEN** el repo fuente contiene archivos no listados dentro del set raíz permitido
- **THEN** la CLI MUST NOT copiarlos al target como parte de `install`

#### Scenario: Opencode es merge-managed especial
- **WHEN** la CLI planifica `opencode.json`
- **THEN** lo trata como `merge-json` especial fuera del catálogo de archivos completos con SHA-256 y preserva claves desconocidas del usuario

### Requirement: Manifest de estado de instalación
La CLI SHALL persistir estado de instalación en `.lufy-ai/install-state.json` con schema versionado y hashes por asset completo gestionado, excluyendo configuraciones merge-managed especiales como `opencode.json`.

#### Scenario: Estado escrito tras install exitoso
- **WHEN** install aplica acciones exitosamente
- **THEN** escribe `.lufy-ai/install-state.json` con `schemaVersion`, `toolVersion`, `sourceChangeID`, timestamps y lista de assets completos gestionados con `sourceSHA256` y `targetSHA256`

#### Scenario: Estado usa paths relativos
- **WHEN** la CLI registra assets en install state
- **THEN** cada asset usa paths relativos al target para portabilidad y trazabilidad

#### Scenario: Estado compatible con verify
- **WHEN** `verify` lee `.lufy-ai/install-state.json`
- **THEN** puede recalcular hashes destino y comparar contra el estado registrado

#### Scenario: Merge-managed no tiene hash completo
- **WHEN** `opencode.json` se crea o actualiza por `merge-json`
- **THEN** el manifest de estado no registra `opencode.json` como asset completo ni usa su hash para detectar drift

### Requirement: Sync preserva personalizaciones fuera de scope
El comando `sync` MUST NOT sobrescribir archivos no gestionados, assets con drift local ni personalizaciones fuera del catálogo permitido; para `opencode.json` solo puede aplicar merge JSON conservador explícito.

#### Scenario: Asset gestionado con drift local bloqueado
- **WHEN** un asset gestionado existe pero su hash target actual no coincide con el último hash target registrado
- **THEN** el plan de sync lo marca como `conflict` y apply MUST NOT sobrescribirlo como `update-managed`

#### Scenario: Archivo existente no gestionado bloqueado
- **WHEN** un path del catálogo existe en el target pero no aparece como gestionado en `.lufy-ai/install-state.json`
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

### Requirement: Verify estructural completo
El comando `verify` SHALL validar estructura, estado y hashes de una instalación gestionada, además de validar configuraciones merge-managed especiales sin exigir hash completo.

#### Scenario: Instalación completa válida
- **WHEN** el target contiene todos los assets gestionados, `opencode.json` merge-managed válido y el install state coincide por hash para assets completos
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

#### Scenario: Opencode merge-managed inválido
- **WHEN** `opencode.json` falta, no parsea como JSON o carece de estructura merge-managed mínima
- **THEN** `verify` reporta `fail` para `opencode.json` sin buscarlo en `.lufy-ai/install-state.json` como asset completo
