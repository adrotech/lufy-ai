## ADDED Requirements

### Requirement: Hexagonal application boundaries
La CLI Go SHALL separar dominio, casos de uso y detalles externos mediante boundaries hexagonales aplicables a instalacion, sincronizacion, verificacion, estado, backup y runtime de tool.

#### Scenario: Domain remains free of external effects
- **WHEN** se revisa codigo bajo `internal/core/domain`
- **THEN** no SHALL importar filesystem, red, backup, state store, tool runtime, adapters concretos ni packages de plataforma
- **AND** sus funciones SHALL poder probarse sin temp dirs, variables de entorno ni procesos externos

#### Scenario: Use case coordinates through application ports
- **WHEN** `install`, `sync`, `verify`, `status`, `backup`, `restore` o `upgrade` necesitan leer/escribir archivos, estado, backups, runtime de tool o tiempo actual
- **THEN** el caso de uso SHALL depender de una abstraccion inyectable o facade de aplicacion
- **AND** los detalles concretos SHALL vivir en adapters externos o ensambladores

#### Scenario: Default preset remains compatible
- **WHEN** los boundaries se refactorizan para el preset `opencode` + `openspec`
- **THEN** `lufy-ai install`, `sync`, `verify` y `status` SHALL conservar comportamiento observable salvo cambios explicitamente documentados
- **AND** `scripts/validate.sh` SHALL pasar antes de reportar el slice como validado

### Requirement: Application service responsibility split
Los servicios de aplicacion SHALL separar planificacion, ejecucion, reporting, persistencia y validacion cuando un servicio concentre multiples razones para cambiar.

#### Scenario: Plan building has no real mutations
- **WHEN** un planner construye acciones o checks para install, sync o verify
- **THEN** SHALL NOT escribir archivos, crear backups, adquirir locks mutantes ni imprimir salida humana como parte de la decision de negocio
- **AND** SHALL devolver un modelo verificable por tests

#### Scenario: Execution is isolated
- **WHEN** un executor aplica acciones de install, sync, backup, restore o uninstall
- **THEN** las mutaciones SHALL estar confinadas al executor o adapter correspondiente
- **AND** errores durante la ejecucion SHALL preservar evidencia de recovery o rollback disponible

#### Scenario: Reporting is presentation-only
- **WHEN** se emite salida humana o JSON
- **THEN** el reporter SHALL renderizar un modelo de resultado ya construido
- **AND** SHALL NOT recalcular reglas de negocio ni mutar estado

### Requirement: Typed action semantics
Las acciones de planificacion que afecten comportamiento SHALL usar tipos o constantes declaradas, y SHALL evitar strings magicos dispersos.

#### Scenario: New action kind is declared
- **WHEN** se agrega una nueva accion de install, sync o uninstall
- **THEN** su kind SHALL estar declarado en un tipo o constante compartida del paquete correspondiente
- **AND** SHALL tener tests para planificacion, confirmacion requerida y ejecucion o no-mutacion segun aplique

#### Scenario: Unknown action fails explicitly
- **WHEN** un executor recibe una accion desconocida o no soportada
- **THEN** SHALL fallar explicitamente con error accionable
- **AND** SHALL NOT ignorarla silenciosamente

### Requirement: SOLID and clean-code review gates
Los refactors arquitectonicos SHALL reportar evidencia de SOLID y clean code proporcional al riesgo del slice.

#### Scenario: Reviewer evaluates architectural slice
- **WHEN** un slice modifica boundaries, puertos, adapters o casos de uso principales
- **THEN** el resultado SHALL explicar impacto sobre SRP, OCP, LSP, ISP y DIP
- **AND** SHALL identificar cualquier deuda residual que quede deliberadamente fuera del slice

#### Scenario: Clean-code gate checks large service responsibilities
- **WHEN** un archivo o funcion de servicio sigue mezclando decision de negocio, IO y reporting despues del slice
- **THEN** el resultado SHALL justificar por que permanece asi o registrar follow-up concreto
- **AND** SHALL NOT declarar cumplimiento hexagonal estricto para ese paquete
