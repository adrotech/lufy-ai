# Design: integración runtime del skill registry

## Decisiones

### Un único primitive idempotente

`skillregistry.Service.Ensure` construye el índice esperado y lo compara byte a
byte con `.lufy/skill-registry.json`. Si coincide, no toca el filesystem. Si falta
o está stale, usa la misma escritura atómica que `refresh`.

El comando público `lufy-ai skills ensure` permite reutilizar este primitive desde
hooks y automatizaciones sin duplicar lógica ni interpretar la salida de `status`.

### Lifecycle best-effort

`install` y `sync` invocan el ensure después de completar sus mutaciones y
validaciones principales, pasando el tool efectivo del plan. `setup` hereda el
comportamiento mediante su paso de instalación. Un fallo se emite como warning
accionable y no convierte una instalación ya válida en rollback o error fatal.

### Startup por harness

- OpenCode: el plugin gestionado escucha `session.created` y ejecuta un hook shell
  best-effort. El hook sale silenciosamente si `lufy-ai` no está disponible o el
  proyecto no está inicializado.
- Codex: el estado se actualiza en el lifecycle y el handoff indica ejecutar
  `skills ensure --tool codex` antes de la primera resolución cuando el índice no
  esté ready. No se activa `.codex/hooks.json` sin un contrato nativo verificable.

### Diagnóstico sin mutación

El servicio expone una inspección programática para que `status`, `doctor` y
`verify --deep` compartan la misma semántica `ready|stale|not_available`.
`doctor` informa stale/ausente como warning; `verify --deep` lo considera fallo
porque su objetivo es validar una instalación lista para ejecución. Ninguno de los
dos repara automáticamente.

## Fallos y recuperación

- Binario ausente en un hook: no bloquea la sesión; la recuperación permanece en
  CLI/documentación.
- Skill inválida: se conserva como warning del índice; las skills válidas siguen
  disponibles.
- Repositorio movido: las rutas absolutas vuelven al estado stale y `ensure`
  reescribe el índice en la ubicación actual.
- Windows: el primitive Go es portable; el hook `.sh` es específico de OpenCode
  en entornos que puedan ejecutarlo y no altera la ruta CLI soportada.

## Review slices

1. Servicio/CLI idempotente y lifecycle de instalación.
2. Startup OpenCode y fallback documentado de Codex.
3. Diagnósticos, pruebas, paridad de assets y documentación.
