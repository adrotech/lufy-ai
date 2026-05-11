# ADR 0004: Semántica de `sync` para assets nuevos del catálogo

## Estado

Aceptada.

## Contexto

Cuando una versión nueva de `lufy-ai` agrega assets al catálogo, un target existente puede no tener esos archivos registrados en `install-state.json`. Instalar automáticamente esos archivos durante `sync` puede ser útil, pero también puede sorprender a usuarios con targets personalizados.

## Decisión

`sync` no instala automáticamente assets nuevos del catálogo que están ausentes del target y no estaban registrados previamente.

`verify` deriva requirements del catálogo de forma compatible: exige sentinels mínimos históricos y assets catalogados que ya están registrados en el manifest.

## Consecuencias

- La incorporación de assets nuevos requiere una estrategia explícita futura si se quiere auto-adopción.
- `verify` no falla por assets nuevos del catálogo que `sync` decidió no instalar.
- Si un asset nuevo del catálogo colisiona con un archivo local no gestionado, la operación debe bloquearse como conflicto antes de sobrescribir.
