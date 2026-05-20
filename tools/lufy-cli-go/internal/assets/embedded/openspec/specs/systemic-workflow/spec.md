# systemic-workflow Specification

## Purpose
Define the systemic working model for agents: initial context analysis, bounded rereads, final grouped validation and evidence-based reporting.

## Requirements
### Requirement: Analisis inicial sistemico
El workflow SHALL analizar el sistema al inicio de una propuesta o bloque antes de implementar, identificando archivos existentes relevantes, componentes, dependencias, interconexiones y riesgos de comportamiento.

#### Scenario: Analisis antes de implementar
- **WHEN** una propuesta OpenSpec o bloque de trabajo requiere cambios sobre codigo, configuracion, agentes o documentacion existente
- **THEN** el agente responsable realiza una inspeccion inicial dirigida que cubre el contexto necesario para planificar sin depender de relecturas repetidas durante la implementacion normal

#### Scenario: Vision holistica del cambio
- **WHEN** el analisis inicial identifica varios componentes relacionados, como agentes, skills, politica, tests, documentacion, APIs, base de datos o servicios
- **THEN** el reporte o handoff explica como interactuan las partes relevantes y que dependencias condicionan la implementacion

### Requirement: Implementacion sin relecturas repetidas innecesarias
El workflow SHALL evitar volver a leer archivos viejos ya analizados durante la implementacion normal, excepto cuando el archivo haya sido modificado, exista conflicto, aparezca nueva evidencia, el alcance cambie, se detecte un bloqueo o el riesgo requiera confirmacion.

#### Scenario: Archivo viejo no modificado
- **WHEN** un archivo existente fue revisado en el analisis inicial y no fue modificado ni afectado por nueva evidencia
- **THEN** el agente no lo relee de forma rutinaria durante cada tarea de implementacion

#### Scenario: Relectura justificada
- **WHEN** un archivo viejo fue modificado, entra en conflicto con cambios concurrentes, participa en una falla, o nueva evidencia invalida el analisis inicial
- **THEN** el agente lo relee de forma dirigida y reporta la razon si afecta el flujo o la evidencia final

### Requirement: Revisión final acotada de archivos viejos modificados
El workflow SHALL incluir una revision final acotada de archivos viejos modificados o afectados antes de cerrar la implementacion o ejecutar validacion final.

#### Scenario: Cierre de bloque con archivos existentes modificados
- **WHEN** una propuesta modifica archivos existentes revisados previamente
- **THEN** el agente revisa al final esos archivos o diffs para comprobar coherencia con el analisis inicial, dependencias y comportamiento esperado

#### Scenario: Sin archivos viejos modificados
- **WHEN** la propuesta solo crea archivos nuevos y no afecta archivos existentes
- **THEN** no se requiere una relectura final de archivos viejos, salvo que una dependencia o riesgo detectado lo justifique

### Requirement: Validacion final agrupada con tests y coverage
El workflow SHALL ejecutar tests, coverage y validacion completa al final de todas las tareas de una propuesta cuando esos comandos existan y apliquen al alcance.

#### Scenario: Propuesta con tareas completas
- **WHEN** todas las tareas de implementacion de una propuesta estan finalizadas
- **THEN** el agente ejecuta la validacion agrupada disponible, incluyendo tests y coverage cuando existan para el toolchain real del alcance, y reporta comandos exactos y resultados

#### Scenario: Validacion no disponible
- **WHEN** no existe toolchain, comando de coverage o suite aplicable al alcance
- **THEN** el agente declara la limitacion y reporta evidencia estatica, documental o manual real sin afirmar que tests o coverage pasaron

### Requirement: Excepciones para feedback temprano
El workflow SHALL permitir relectura o validacion temprana solo cuando exista bloqueo, cambio riesgoso, diagnostico de falla, incertidumbre que afecte seguridad/correctness, o necesidad de feedback para autorregular el sistema.

#### Scenario: Bloqueo o falla durante implementacion
- **WHEN** una tarea queda bloqueada o falla por una causa no entendida
- **THEN** el agente puede releer archivos relevantes o ejecutar validacion enfocada antes del final para diagnosticar y destrabar el trabajo

#### Scenario: Cambio riesgoso
- **WHEN** el cambio afecta areas de alto impacto como contratos publicos, autenticacion, persistencia, instalador, release o delivery
- **THEN** el agente puede solicitar o ejecutar validacion temprana enfocada antes de continuar, manteniendo la validacion completa para el final

### Requirement: Relacion estructura-comportamiento
El workflow SHALL evaluar que el comportamiento esperado del sistema emerge de la estructura modificada y de sus relaciones, no solo de archivos individuales.

#### Scenario: Cambio transversal
- **WHEN** una propuesta cambia reglas compartidas, agentes, skills, politica o documentacion que afecta varias fases
- **THEN** la revision final verifica coherencia entre estructura estatica, dependencias y comportamiento dinamico esperado del workflow

### Requirement: Block-scoped proportional validation
The workflow SHALL run validation/testing proportionally at the end of a task, coherent block, proposal block, or review slice, and SHALL avoid constant test loops for individual micro-checkboxes unless an exception gate applies.

#### Scenario: Validation waits for coherent block boundary
- **WHEN** an agent completes an internal micro-step that is part of a larger coherent task or block
- **THEN** the workflow SHALL NOT require full validation/testing for that micro-step and SHALL defer grouped validation to the coherent block boundary

#### Scenario: Validation runs before validated state
- **WHEN** a task, coherent block, proposal block, or review slice is ready to move from `implemented` to `validated`
- **THEN** the workflow SHALL run the real applicable validation commands or document proportional static/manual evidence and SHALL report exact evidence before using the `validated` state

#### Scenario: Exception allows early validation
- **WHEN** a blocker, risky change, feedback loop, or failure diagnosis requires earlier evidence
- **THEN** the workflow MAY run focused validation before the block boundary while preserving grouped final validation for the block when applicable
