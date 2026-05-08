# pr-creator-skill Specification

## Purpose
TBD - created by archiving change add-pr-creator-skill. Update Purpose after archive.
## Requirements
### Requirement: Skill pr.creator con estructura estándar
El sistema SHALL incluir un skill OpenCode llamado `pr.creator` con estructura estándar tipo Anthropic, autocontenido bajo `.opencode/skills/pr.creator/`, con `SKILL.md` como punto de entrada y recursos/templates auxiliares versionados cuando sean necesarios.

#### Scenario: Skill discoverable por OpenCode
- **WHEN** una persona o agente inspecciona `.opencode/skills/pr.creator/`
- **THEN** encuentra un `SKILL.md` que describe propósito, uso, inputs esperados, límites y outputs del skill `pr.creator`

#### Scenario: Recursos del skill autocontenidos
- **WHEN** el skill necesita un template o referencia auxiliar para generar contenido de PR
- **THEN** esos recursos se ubican dentro de `.opencode/skills/pr.creator/` y no requieren archivos fuera del skill salvo políticas o artefactos existentes del repo referenciados explícitamente

### Requirement: Template de Pull Request para GitHub
El skill SHALL generar contenido Markdown para Pull Requests de GitHub con secciones consistentes, redactadas en español para contenido humano y preservando identificadores técnicos.

#### Scenario: Secciones mínimas del PR
- **WHEN** se usa `pr.creator` para preparar un Pull Request
- **THEN** el resultado incluye al menos resumen, `Why`, trazabilidad/ticket, evidencia de pruebas, `Monitors`, `Migraciones`, riesgos/follow-ups y checklist o notas de validación

#### Scenario: No ejecuta creación de PR
- **WHEN** el skill genera el contenido del Pull Request
- **THEN** no ejecuta `gh pr create`, `git push`, sync de proyectos ni otra operación remota, y deja esas acciones al rol `delivery` con autorización explícita

### Requirement: Modo manual de invocación
El skill SHALL poder invocarse manualmente por una persona o agente para generar o revisar el título y cuerpo Markdown de un Pull Request sin ejecutar operaciones de delivery.

#### Scenario: Usuario invoca pr.creator manualmente
- **WHEN** una persona o agente solicita usar `pr.creator` y provee contexto como change-id, diff, evidencia o notas del cambio
- **THEN** el skill produce un título sugerido y un cuerpo de PR en Markdown siguiendo el template estándar

#### Scenario: Contexto manual incompleto
- **WHEN** la invocación manual no incluye datos suficientes para completar secciones obligatorias
- **THEN** el skill usa valores explícitos como `Pendiente de confirmar`, `No configurado` o `No aplica`, y lista preguntas o datos faltantes sin inventar evidencia

### Requirement: Integración obligatoria desde delivery al crear PR
El subagente `delivery` SHALL usar `pr.creator` al preparar un Pull Request cuando el skill esté disponible, antes o durante la ejecución autorizada de `gh pr create`.

#### Scenario: Delivery crea PR con pr.creator instalado
- **WHEN** `delivery` tiene autorización explícita para crear un PR y existe `.opencode/skills/pr.creator/`
- **THEN** `delivery` usa `pr.creator` para generar el título/cuerpo del PR antes de ejecutar `gh pr create`

#### Scenario: Delivery conserva operaciones Git/GH
- **WHEN** `pr.creator` genera el contenido del PR para `delivery`
- **THEN** `delivery` sigue siendo responsable de branch safety, validación final, commit, push, `gh pr create`, issue/project sync y reporte de evidencia

#### Scenario: Skill no disponible durante delivery
- **WHEN** `delivery` debe crear un PR pero `.opencode/skills/pr.creator/` no existe o no puede cargarse
- **THEN** `delivery` reporta la limitación y usa la política/template vigente solo si eso no contradice la autorización ni los gates de delivery

### Requirement: Resumen funcional derivable desde OpenSpec
El skill SHALL preferir `openspec/changes/<change>/proposal.md` como fuente para el resumen de nueva funcionalidad y el `why` cuando exista una propuesta OpenSpec asociada.

#### Scenario: Proposal disponible
- **WHEN** el usuario provee un `change-id` o el contexto permite ubicar `openspec/changes/<change>/proposal.md`
- **THEN** el resumen y el `Why` del PR se derivan de las secciones relevantes de `proposal.md`, ajustados al formato del Pull Request

#### Scenario: Proposal no disponible
- **WHEN** no existe `proposal.md` aplicable
- **THEN** el skill usa el contexto provisto, diff o notas del usuario para redactar el resumen y marca cualquier supuesto que requiera confirmación

### Requirement: Trazabilidad a ticket configurado
El skill SHALL incluir una sección de tarea asociada que muestre links a Jira, GitHub Issues/Projects, Notion u otra herramienta de tracking solo cuando exista configuración, contexto o evidencia disponible.

#### Scenario: Ticket detectado o proporcionado
- **WHEN** existe un enlace o identificador de tarea asociado al cambio
- **THEN** el template incluye el link o referencia en la sección de trazabilidad/ticket

#### Scenario: Ticket no configurado
- **WHEN** no hay herramienta de tracking configurada ni ticket provisto
- **THEN** el template declara `No configurado`, `No aplica` o `Pendiente de confirmar` sin inventar enlaces

### Requirement: Evidencia de pruebas y validación
El skill SHALL incluir una sección de evidencia que capture comandos ejecutados y resultados reales, y permita adjuntar capturas, JSON o curls cuando aplique.

#### Scenario: Comandos de validación disponibles
- **WHEN** el cambio tiene evidencia de validación con comandos y resultados
- **THEN** el PR body incluye los comandos exactos, resultado observado y cualquier salida relevante resumida

#### Scenario: Evidencia funcional adicional
- **WHEN** existen capturas, respuestas JSON, curls o evidencia manual relevante
- **THEN** el template incluye un lugar explícito para enlazarlas o pegarlas con contexto

#### Scenario: Validación no ejecutada o no disponible
- **WHEN** no se ejecutó validación o no existe toolchain aplicable
- **THEN** el template declara la limitación y no afirma éxito sin evidencia

### Requirement: Monitors operativos
El skill SHALL incluir una sección `Monitors` para alertas, dashboards o monitores asociados en sistemas configurados como Grafana, New Relic, Datadog u otros.

#### Scenario: Monitor configurado o proporcionado
- **WHEN** existe un monitor, dashboard o alerta asociada al cambio
- **THEN** el template incluye nombre, link y sistema de observabilidad correspondiente

#### Scenario: Monitor no aplica
- **WHEN** no hay monitor configurado o el cambio no requiere monitoreo operacional
- **THEN** la sección `Monitors` queda explícitamente marcada como `No aplica`, `No configurado` o `Pendiente de confirmar`

### Requirement: Migraciones detectadas automáticamente
El skill SHALL analizar automáticamente archivos modificados y diff disponible para detectar migrations o creación/cambio de tablas/schemas de DB, reflejando el resultado en la sección `Migraciones`.

#### Scenario: Migración detectada por ruta o patrón
- **WHEN** el diff incluye rutas o patrones como `migrations/`, `db/migrate/`, `schema.sql`, `prisma/schema.prisma`, `*.migration.*`, `CREATE TABLE` o `ALTER TABLE`
- **THEN** la sección `Migraciones` lista los archivos o indicios detectados y solicita evidencia/plan de ejecución si aplica

#### Scenario: Sin migraciones detectadas
- **WHEN** el análisis no encuentra rutas ni patrones de migración o schema en los cambios
- **THEN** la sección `Migraciones` indica que no se detectaron migraciones y menciona que la detección fue heurística

#### Scenario: Señal ambigua de persistencia
- **WHEN** el diff sugiere cambios de persistencia pero no coincide con patrones conocidos de migración
- **THEN** la sección `Migraciones` marca el estado como `Pendiente de confirmar` e incluye la evidencia ambigua encontrada
