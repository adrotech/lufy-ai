# Change: integrar el skill registry con el runtime

## Why

El registry portable ya puede generarse y diagnosticarse, pero depende de que una
persona recuerde ejecutar `lufy-ai skills refresh`. Esa separación deja sesiones
con un índice ausente o desactualizado después de instalar, sincronizar, mover el
repositorio o editar skills.

## What Changes

- Agregar `lufy-ai skills ensure`, idempotente: conserva un registry `ready` y
  reescribe solamente estados `stale` o `not_available`.
- Ejecutar el ensure en modo best-effort al finalizar `install`, `setup` y `sync`
  con el tool efectivo, sin revertir una instalación válida si el índice falla.
- Integrar OpenCode con su evento local `session.created` para asegurar el índice
  al inicio de cada sesión.
- Mantener una integración honesta para Codex: ensure durante el lifecycle y una
  instrucción de primer uso, sin declarar un hook nativo no verificado.
- Incorporar el estado del registry a `doctor` y a `verify --deep`, ambos de solo
  lectura y con recuperación accionable.
- Cubrir repositorios movidos, Windows, ausencia del binario y skills inválidas.

## Impact

- Affected specs: `portable-skill-registry`.
- Affected code: CLI Go, installer, syncer, diagnostics, assets OpenCode/Codex y
  documentación operativa.
- Compatibilidad: `refresh` y `status` conservan su contrato; no se agregan
  integraciones con Claude Code.

## Non-goals

- Instalar skills externas o modificar el contenido de `SKILL.md`.
- Bloquear una sesión, instalación o sync por un fallo best-effort del registry.
- Inventar o documentar como soportado un hook de inicio de Codex sin contrato
  oficial disponible.
