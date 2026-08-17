## ADDED Requirements

### Requirement: Ensure idempotente del registry

El sistema MUST ofrecer una operación `skills ensure` que escriba el registry
solamente cuando esté ausente o desactualizado y MUST conservar intacto un archivo
que ya coincida con el índice esperado.

#### Scenario: Registry ready

- **WHEN** el registry existente coincide con las skills y raíces efectivas
- **THEN** `skills ensure` termina correctamente sin reescribir el archivo

#### Scenario: Registry ausente o stale

- **WHEN** el registry no existe, el repositorio fue movido o cambió una skill
- **THEN** `skills ensure` genera atómicamente el índice esperado en el target actual

### Requirement: Auto-ensure durante el lifecycle

El sistema MUST intentar actualizar el registry después de una instalación o sync
real exitosos usando el tool efectivo y MUST tratar el fallo como warning
accionable, sin revertir las mutaciones principales ya verificadas.

#### Scenario: Lifecycle exitoso con registry desactualizado

- **WHEN** `install`, `setup` o `sync` termina sus mutaciones principales
- **THEN** el sistema asegura el registry para el tool seleccionado

#### Scenario: Fallo best-effort

- **WHEN** el ensure posterior no puede construir o escribir el registry
- **THEN** el comando informa recuperación manual y conserva exitoso el lifecycle principal

### Requirement: Inicio de sesión según capacidades reales

El sistema MUST ejecutar un ensure best-effort en `session.created` de OpenCode y
MUST evitar declarar hooks nativos de Codex que no tengan un contrato verificado.

#### Scenario: Inicio de OpenCode

- **WHEN** OpenCode emite `session.created` y `lufy-ai` está disponible
- **THEN** el plugin local ejecuta el ensure sin bloquear la sesión ante errores

#### Scenario: Inicio sin CLI disponible

- **WHEN** el hook de OpenCode no encuentra `lufy-ai`
- **THEN** finaliza silenciosamente y la sesión continúa

#### Scenario: Primer uso en Codex

- **WHEN** Codex necesita resolver skills y el registry no está ready
- **THEN** el handoff indica ejecutar `skills ensure --tool codex` antes de resolverlas

### Requirement: Diagnóstico compartido y de solo lectura

El sistema MUST exponer en `doctor` y `verify --deep` los estados
`ready|stale|not_available` con una recuperación accionable y MUST NOT mutar el
registry desde esos comandos.

#### Scenario: Doctor detecta registry stale

- **WHEN** `doctor` inspecciona un registry que no coincide con el índice esperado
- **THEN** informa un warning stale y el comando exacto de recuperación

#### Scenario: Verify deep detecta registry ausente

- **WHEN** `verify --deep` inspecciona una instalación sin registry
- **THEN** registra un fallo con recuperación y no crea el archivo
