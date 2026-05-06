## Context

El catálogo actual instala assets completos con hash SHA-256 y los registra en `.lufy-ai/install-state.json`. Ese modelo funciona para archivos propiedad de `lufy-ai`, pero `opencode.json` es una configuración compartida con contenido del usuario. El comportamiento deseado es una estrategia parcial y conservadora.

## Decision

`opencode.json` se maneja como asset especial `merge-json`:

- No se incluye como asset completo del catálogo ni se registra con `targetSHA256` en install state.
- `install` y `sync` usan el servicio de configuración para cargar JSON existente, fallar si es inválido y escribir solo el resultado mergeado.
- El merge agrega claves gestionadas mínimas (`$schema`, `plugin`) y la integración Engram cuando aplica, preservando claves desconocidas existentes.
- Si el archivo existente será modificado, se planifica backup antes de escribir.
- `verify` parsea `opencode.json` y valida la estructura mínima merge-managed, pero no espera entrada de manifest ni hash completo.

## Alternatives Considered

- **Asset completo por hash**: rechazado porque sobrescribe o bloquea configuración local legítima y no preserva claves desconocidas.
- **Ignorar `opencode.json` en sync/verify**: rechazado porque deja sin validar un archivo crítico del runtime OpenCode.
- **Merge profundo generalizado para todos los JSON**: postergado; el alcance actual solo requiere `opencode.json` para no cambiar contratos de otros assets.

## Risks

- La estructura mínima (`$schema` string y `plugin` array) es deliberadamente reducida; futuras claves gestionadas deben agregarse explícitamente para no convertir el merge en una sobrescritura amplia.
- Si el usuario tiene `plugin` con un tipo incompatible, el merge lo preserva y `verify` falla con mensaje accionable en vez de normalizar silenciosamente.

## Validation

- Tests Go cubren install, sync, JSON inválido, preservación de claves desconocidas y exclusión de `opencode.json` del manifest de hashes.
- Smokes de CLI validan install/verify/sync desde binario compilado.
- Comandos OpenSpec validan que este cambio tenga artefactos aplicables.
