## MODIFIED Requirements

### Requirement: Catálogo completo de assets gestionados
La CLI Go SHALL instalar el conjunto completo de assets gestionados de `lufy-ai` desde el repo fuente o assets embebidos hacia un proyecto destino y SHALL distinguir assets completos gestionados de configuraciones merge-managed o integraciones user-owned especiales.

#### Scenario: Catálogo incluye assets requeridos
- **WHEN** la CLI construye el catálogo de instalación
- **THEN** el catálogo incluye `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/policies`, `.opencode/plugins`, `.opencode/agent-observatory`, `lufy-ia.harness.md`, `tui.json`, `openspec/` y metadatos requeridos bajo `.lufy-ai/`

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
- **WHEN** `.lufy-ai/install-state.json` falta o no puede parsearse
- **THEN** `verify` reporta `fail` o `warn` según severidad definida y explica la recuperación recomendada

#### Scenario: Opencode merge-managed inválido
- **WHEN** `opencode.json` falta, no parsea como JSON o carece de estructura merge-managed mínima
- **THEN** `verify` reporta `fail` para `opencode.json` sin buscarlo en `.lufy-ai/install-state.json` como asset completo

#### Scenario: Referencia AGENTS ausente o incompleta
- **WHEN** `AGENTS.md` falta o no contiene la referencia `@lufy-ia.harness.md`
- **THEN** `verify` reporta un requisito incumplido o warning accionable para la integración de `AGENTS.md` sin requerir hash completo ni entrada de manifest para ese archivo

### Requirement: Sync seguro de assets gestionados
El comando `sync` SHALL reaplicar assets gestionados desde el source hacia un target existente usando el catálogo permitido, estado de instalación, SHA-256 y políticas de conflicto existentes, y MUST treat `AGENTS.md` as user-owned reference integration rather than a managed file payload.

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

#### Scenario: Sync actualiza harness sin mutar AGENTS
- **WHEN** `lufy-ia.harness.md` tiene upstream nuevo sin drift local y `AGENTS.md` contiene contenido user-owned
- **THEN** `sync` planifica backup y `update-managed` para `lufy-ia.harness.md`, actualiza su manifest al aplicar y MUST NOT modificar `AGENTS.md`

#### Scenario: Sync advierte referencia faltante
- **WHEN** `AGENTS.md` no contiene `@lufy-ia.harness.md` durante sync
- **THEN** `sync` reporta warning o acción explícita requerida y MUST NOT insertar la referencia silenciosamente

### Requirement: Manifest de estado de instalación
La CLI SHALL persistir estado de instalación en `.lufy-ai/install-state.json` con schema versionado, metadata real del binario, fingerprint estable del catalogo y hashes por asset completo gestionado, excluyendo configuraciones merge-managed especiales como `opencode.json` e integraciones user-owned especiales como `AGENTS.md`.

#### Scenario: Estado escrito tras install exitoso
- **WHEN** install aplica acciones exitosamente
- **THEN** escribe `.lufy-ai/install-state.json` con `schemaVersion`, `toolVersion`, `sourceChangeID`, `sourceRootFingerprint`, timestamps y lista de assets completos gestionados con `sourceSHA256` y `targetSHA256`, incluyendo `lufy-ia.harness.md` y excluyendo `AGENTS.md`

#### Scenario: Estado usa paths relativos
- **WHEN** la CLI registra assets en install state
- **THEN** cada asset usa paths relativos al target para portabilidad y trazabilidad

#### Scenario: Estado compatible con verify
- **WHEN** `verify` lee `.lufy-ai/install-state.json`
- **THEN** puede recalcular hashes destino y comparar contra el estado registrado

#### Scenario: Merge-managed no tiene hash completo
- **WHEN** `opencode.json` se crea o actualiza por `merge-json`
- **THEN** el manifest de estado no registra `opencode.json` como asset completo ni usa su hash para detectar drift

#### Scenario: AGENTS user-owned no tiene hash completo
- **WHEN** `AGENTS.md` se crea o recibe la referencia `@lufy-ia.harness.md` como integración user-owned
- **THEN** el manifest de estado no registra `AGENTS.md` como asset completo ni usa su hash para detectar drift
