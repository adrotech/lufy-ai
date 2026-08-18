# lufy-sdd-workflow Specification Delta

## ADDED Requirements

### Requirement: Lufy SDD Full provides a native change lifecycle

El sistema SHALL proveer un lifecycle Lufy SDD Full ejecutable mediante `lufy-ai sdd` sin requerir el binario, cache ni schemas internos de OpenSpec.

#### Scenario: Create a full change

- **WHEN** el usuario ejecuta `lufy-ai sdd new --change add-audit-log --capability audit-log` sobre una instalación `lufy-sdd/full`
- **THEN** el sistema crea metadata, proposal, design, tasks y spec delta editables bajo `.lufy/workflows/sdd/changes/add-audit-log/`

#### Scenario: Existing change is preserved

- **WHEN** el destino del change ya existe
- **THEN** `sdd new` falla sin sobrescribir ningún artifact existente

#### Scenario: Create a lite change

- **WHEN** el usuario ejecuta `lufy-ai sdd new --change quick-fix --mode lite`
- **THEN** el sistema crea metadata, proposal, tasks y overview HTML sin exigir design ni specs delta

### Requirement: Lufy SDD validates artifacts semantically

El sistema SHALL validar artifacts requeridos, markers delta, requirements, scenarios, tareas y seguridad de paths con diagnósticos determinísticos.

#### Scenario: Strict validation accepts a testable delta

- **WHEN** un change contiene artifacts no vacíos y cada requirement agregado o modificado tiene un scenario con `WHEN` y `THEN`
- **THEN** `lufy-ai sdd validate --strict` retorna éxito y un reporte `lufy-sdd-report/v1` sin errores

#### Scenario: Invalid scenario blocks validation

- **WHEN** un requirement agregado carece de scenario o de cláusula `THEN`
- **THEN** la validación falla, identifica path y línea, no modifica artifacts fuente ni specs activas y solo puede refrescar el overview derivado

### Requirement: Lufy SDD reports effective workflow status

El sistema SHALL derivar estado y progreso desde metadata, artifacts, tareas, digest de sync y evidencia sin confundir checkboxes completas con cierre operativo.

#### Scenario: Completed tasks still require sync

- **WHEN** un change Full tiene todas las tareas marcadas pero el digest actual de deltas no fue sincronizado
- **THEN** `lufy-ai sdd status` reporta `sync_pending` y no `completed` ni `archived`

#### Scenario: JSON output is deterministic

- **WHEN** se ejecuta dos veces `sdd status --json` sin cambios de filesystem
- **THEN** los campos y diagnósticos observables mantienen contenido y orden estables

### Requirement: Lufy SDD syncs delta specs safely

El sistema SHALL planificar y aplicar `ADDED`, `MODIFIED` y `REMOVED` sobre specs activas únicamente después de un preflight global exitoso.

#### Scenario: Valid deltas update main specs

- **WHEN** todos los requirements tienen destinos no ambiguos y el usuario ejecuta `lufy-ai sdd sync --change <name>`
- **THEN** el sistema crea, reemplaza o elimina requirements según el marker, preserva contenido no relacionado y registra el digest sincronizado después del éxito completo

#### Scenario: Ambiguous target aborts all writes

- **WHEN** un requirement `MODIFIED` o `REMOVED` no tiene exactamente una coincidencia en su spec activa
- **THEN** sync falla durante preflight y no modifica ninguna spec ni metadata

#### Scenario: Write failure rolls back touched specs

- **WHEN** una escritura falla después de modificar un destino previo
- **THEN** el sistema restaura los destinos tocados y no registra el digest como sincronizado

### Requirement: Lufy SDD archive enforces local closure gates

El sistema SHALL archivar solo changes estrictamente válidos, con tareas completas, deltas sincronizados sin cambios posteriores y evidencia de verificación regular.

#### Scenario: Incomplete gate blocks archive

- **WHEN** falta una tarea, falta evidencia de verificación o, en Full, el digest está desactualizado
- **THEN** `lufy-ai sdd archive` falla con recovery accionable y mantiene el change activo

#### Scenario: Ready change archives exclusively

- **WHEN** todos los gates locales pasan y el destino fechado no existe
- **THEN** el change se mueve a `.lufy/workflows/sdd/archive/YYYY-MM-DD-<name>/` sin sobrescribir otro archive

### Requirement: Lufy SDD rejects unsafe filesystem surfaces

El sistema SHALL restringir change y capability IDs a kebab-case seguro y SHALL rechazar symlinks o tipos no regulares en paths relevantes de lectura o mutación.

#### Scenario: Unsafe identifier is rejected

- **WHEN** un comando recibe `../escape`, un path absoluto o un ID fuera del formato permitido
- **THEN** falla antes de resolver o escribir fuera de `.lufy/workflows/sdd/`

#### Scenario: Symlink is not followed

- **WHEN** un ancestor, change, spec destino o archive relevante es un symlink
- **THEN** la operación falla sin seguir el enlace ni mutar el destino enlazado

### Requirement: Lufy SDD exposes methodology-specific agent actions

El sistema SHALL instalar acciones y skills equivalentes para explore, propose, apply, verify, sync y archive en cada tool adapter escribible seleccionado junto con Lufy SDD.

#### Scenario: Lufy-only selection excludes OpenSpec workflow assets

- **WHEN** una instalación selecciona `lufy-sdd/full` para todos los tiers metodológicos requeridos y no selecciona OpenSpec
- **THEN** recibe acciones Lufy SDD y no recibe comandos `/opsx-*`, skills OpenSpec ni directorio `openspec/` como assets requeridos

#### Scenario: Agent handoff identifies Lufy SDD

- **WHEN** una acción Lufy SDD produce un handoff
- **THEN** reporta `methodology_id: lufy-sdd`, mode efectivo, estado de gates y siguiente acción sin inventar artifacts OpenSpec

### Requirement: Lufy SDD generates an integrated HTML overview

El sistema SHALL materializar automáticamente un `change-overview.html` autocontenido y determinístico a partir de los artifacts Markdown, sin requerir un comando o skill público adicional.

#### Scenario: Full overview contains all source artifacts

- **WHEN** se crea o valida un change Full
- **THEN** el overview reúne proposal, design, tasks y specs delta en orden estable, sin scripts ni recursos remotos

#### Scenario: Lite overview contains the bounded artifact set

- **WHEN** se crea o valida un change Lite
- **THEN** el overview reúne proposal y tasks y no reporta design ni specs como artifacts faltantes

#### Scenario: Validation refreshes a stale overview

- **WHEN** un usuario modifica un Markdown fuente y ejecuta `sdd validate`
- **THEN** el overview se reemplaza atómicamente y refleja el contenido nuevo sin modificar los Markdown ni las specs activas

### Requirement: Lufy SDD Lite uses mode-aware lifecycle gates

El sistema SHALL reconocer `mode: lite` y adaptar validación, estado, sync y archive al conjunto reducido de artifacts.

#### Scenario: Lite sync is not applicable

- **WHEN** el usuario ejecuta `sdd sync` sobre un change Lite válido
- **THEN** el sistema responde `not_applicable` sin crear ni modificar specs activas

#### Scenario: Lite archive does not require a delta digest

- **WHEN** un change Lite es válido, tiene todas sus tareas completas y evidencia regular
- **THEN** archive permite el cierre sin exigir un digest de sync inexistente

### Requirement: Harness routing respects the selected SDD methodology

El sistema SHALL separar tier, execution mode y methodology adapter, y SHALL ejecutar el lifecycle Lufy SDD cuando `lufy-sdd/full` o `lufy-sdd/lite` sea la selección efectiva del tier.

#### Scenario: T1 routes to native Lufy Full

- **WHEN** el router clasifica T1 y el contexto efectivo selecciona `lufy-sdd/full`
- **THEN** el handoff usa skills y comandos Lufy SDD Full y no sustituye el flujo por OpenSpec

#### Scenario: T2 routes to native Lufy Lite

- **WHEN** el router clasifica T2 y el contexto efectivo selecciona `lufy-sdd/lite`
- **THEN** el handoff crea o mantiene el change Lite nativo en vez de limitarse a un template o handoff informal

#### Scenario: Lufy overview is reported as automatic

- **WHEN** un agente completa new, validate, sync o archive para un change Lufy SDD
- **THEN** el Result Contract reporta policy `automatic`, status, trigger y path del overview sin preguntar al usuario si desea generarlo

### Requirement: Delivery distinguishes user-owned SDD artifacts from internal metadata

El sistema SHALL permitir que los artifacts user-owned bajo el workflow Lufy SDD participen del delivery sin desactivar el guard general de metadata `.lufy/`.

#### Scenario: User-owned change artifacts pass internal-prefix classification

- **WHEN** un PR contiene archivos bajo `.lufy/workflows/sdd/changes/`, `specs/`, `decisions/`, `verification/` o `archive/` y no están ignorados explícitamente
- **THEN** `lufy-ai pr guard` no los clasifica como metadata interna solo por el prefijo `.lufy/`

#### Scenario: Other Lufy metadata remains protected

- **WHEN** un PR contiene `.lufy/managed-state/`, `.lufy/context/` u otra ruta interna no exceptuada
- **THEN** el guard mantiene el bloqueo y reporta la ruta interna

#### Scenario: Explicit ignore still protects a user-owned artifact

- **WHEN** un artifact de una raíz Lufy SDD user-owned coincide explícitamente con `.gitignore`
- **THEN** el guard lo reporta como ignored y bloquea aunque no lo duplique como metadata interna
