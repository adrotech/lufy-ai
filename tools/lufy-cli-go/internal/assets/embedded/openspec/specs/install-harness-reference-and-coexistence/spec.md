## Purpose
Definir el modelo de coexistencia donde las instrucciones gestionadas por Lufy viven en assets gestionados y `AGENTS.md` permanece user-owned, integrado mediante un bloque LUFY compacto sin ser un asset completo gestionado por hash.

## Requirements
### Requirement: Harness Lufy gestionado en AGENTS
La instalación SHALL mantener assets de instrucciones gestionados por Lufy y SHALL integrar el harness en el proyecto destino mediante un bloque LUFY gestionado desde `AGENTS.md` sin convertir `AGENTS.md` en asset completo gestionado por hash. La referencia legacy `@lufy-ia.harness.md` SHALL seguir siendo aceptada por compatibilidad.

#### Scenario: Target vacío instala harness y AGENTS mínimo
- **WHEN** el usuario ejecuta `lufy-ai install --target <dir>` sobre un target sin instalación previa, sin `AGENTS.md` y sin assets Lufy
- **THEN** la CLI instala `lufy-ia.harness.md` como asset gestionado, crea un `AGENTS.md` mínimo con el bloque LUFY gestionado, instala los assets restantes permitidos y escribe `.lufy/managed-state/install-state.json` con el harness registrado por SHA-256

#### Scenario: AGENTS propio recibe solo bloque gestionado
- **WHEN** el target ya contiene un `AGENTS.md` user-owned sin referencia al harness y el usuario confirma la mutación o pasa `--yes`
- **THEN** install crea backup del `AGENTS.md` existente, agrega solo el bloque LUFY gestionado sin insertar el contenido completo de Lufy y no registra `AGENTS.md` como asset completo gestionado

#### Scenario: Integración existente no se duplica
- **WHEN** el target ya contiene `AGENTS.md` con el bloque LUFY gestionado o la referencia legacy `@lufy-ia.harness.md`
- **THEN** install planifica `skip` para la integración y MUST NOT duplicarla ni reescribir el archivo solo por esa integración

### Requirement: Sync actualiza harness sin mutar AGENTS
El comando `sync` SHALL actualizar `lufy-ia.harness.md` mediante las reglas de manifest, SHA-256, backup e idempotencia de assets gestionados, y MUST NOT modificar `AGENTS.md` durante sync salvo que una acción explícita futura lo autorice.

#### Scenario: Sync actualiza harness y preserva AGENTS
- **WHEN** `.lufy/managed-state/install-state.json` registra `lufy-ia.harness.md`, el target no tiene drift local en el harness, el source hash del harness cambió y `AGENTS.md` contiene contenido propio del usuario
- **THEN** sync crea backup del harness si existe, actualiza `lufy-ia.harness.md`, actualiza su estado en el manifest y preserva `AGENTS.md` byte-for-byte

#### Scenario: Sync reporta referencia ausente sin auto-fix
- **WHEN** el target contiene `lufy-ia.harness.md` gestionado pero `AGENTS.md` no contiene el bloque LUFY gestionado ni la referencia legacy `@lufy-ia.harness.md`
- **THEN** sync reporta un warning o acción explícita requerida para agregar la integración y MUST NOT modificar `AGENTS.md` silenciosamente

### Requirement: Migración legacy de AGENTS gestionado no destructiva
La CLI SHALL detectar instalaciones legacy donde `AGENTS.md` fue registrado como asset gestionado y SHALL migrar hacia el modelo de harness sin sobrescribir ni borrar el `AGENTS.md` existente.

#### Scenario: Manifest legacy conserva AGENTS
- **WHEN** `.lufy/managed-state/install-state.json` registra `AGENTS.md` como asset completo gestionado legacy durante install, sync o migración de estado
- **THEN** la CLI preserva el `AGENTS.md` existente, instala `lufy-ia.harness.md`, retira o marca la entrada legacy de `AGENTS.md` de forma trazable cuando sea seguro y reporta cualquier bloque legado que requiera revisión manual

#### Scenario: Legacy con drift local queda bloqueado accionable
- **WHEN** `AGENTS.md` aparece en el manifest legacy pero su contenido actual difiere del último hash registrado y falta la referencia al harness
- **THEN** la CLI MUST NOT sobrescribir ni borrar `AGENTS.md`, reporta conflicto o bloque legado accionable y mantiene el harness instalado cuando pueda hacerlo sin modificar el archivo en conflicto

### Requirement: Coexistencia con OpenCode y OpenSpec propios
La instalación SHALL preservar archivos OpenCode/OpenSpec propios del target fuera del catálogo gestionado y SHALL bloquear conflictos exactos en paths gestionados antes de escribir.

#### Scenario: Archivos propios se preservan
- **WHEN** el target ya contiene configuración OpenCode, OpenSpec, comandos, skills o documentación propia fuera del catálogo gestionado por Lufy
- **THEN** install y sync MUST NOT borrar, modificar ni registrar esos archivos como gestionados por Lufy

#### Scenario: Conflicto exacto se bloquea
- **WHEN** un path del catálogo gestionado de Lufy ya existe en el target, no está registrado como gestionado por Lufy y su policy no permite merge/adopción segura
- **THEN** install y sync lo marcan como `conflict`, MUST NOT sobrescribirlo y reportan la acción manual necesaria

### Requirement: Verify diferencia harness gestionado y referencia AGENTS
`verify` SHALL validar `lufy-ia.harness.md` como asset gestionado estricto con manifest y SHA-256, y SHALL validar `AGENTS.md` solo como integración user-owned que debe referenciar el harness sin requerir hash completo del archivo.

#### Scenario: Verify valida harness y manifest
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir>` sobre un target instalado con el nuevo modelo
- **THEN** verify exige que `lufy-ia.harness.md` exista, sea archivo seguro, esté registrado en `.lufy/managed-state/install-state.json` y coincida con el SHA-256 registrado

#### Scenario: Verify valida referencia sin hash completo
- **WHEN** el target contiene `AGENTS.md` con contenido propio y el bloque LUFY gestionado o la referencia legacy `@lufy-ia.harness.md`
- **THEN** verify reporta ok para la integración de `AGENTS.md` sin exigir que `AGENTS.md` esté registrado como asset completo ni comparar su hash completo

#### Scenario: Verify reporta referencia faltante accionable
- **WHEN** el target contiene harness gestionado válido pero `AGENTS.md` falta o no contiene `@lufy-ia.harness.md`
- **THEN** verify reporta un requisito incumplido o warning accionable que explica cómo agregar la referencia, sin intentar modificar el target
