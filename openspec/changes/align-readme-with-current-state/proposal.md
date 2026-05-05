## Why

El README mezcla estado actual del kit instalable con contenido futuro sobre templates y subagentes, lo que puede hacer que usuarios esperen features que no existen como assets instalables. RM-011 requiere que la entrada principal del repo sea honesta, centrada en estado real y quickstart, dejando el roadmap y diseño futuro en documentación específica.

## What Changes

- Reestructurar `README.md` para enfocarlo en qué entrega hoy el repo, quickstart de instalación/verificación y navegación corta a docs relevantes.
- Preservar el banner, enlaces relevantes y referencias a `docs/getting-started.md`, `docs/roadmap.md`, `openspec/README.md` y `tools/lufy-cli-go/README.md`.
- Mover o consolidar contenido especulativo de templates, stacks y subagentes futuros fuera del README hacia `docs/roadmap.md` o documentación específica.
- Actualizar `docs/getting-started.md` para mantener español humano, quickstart coherente con CLI Go actual y comandos reales disponibles.
- Actualizar `tools/lufy-cli-go/README.md` para que su estado documental refleje la CLI actual: `install`, `verify`, `backup`, `restore`, `sync`, assets gestionados, SHA-256, idempotencia, backup/restore y validación real.
- Documentar como en curso, no como completado, el trabajo de CI proposal/implementation y sync implementation cuando corresponda.
- Evitar prometer templates, detección de stack o features no instalables como capacidad disponible.

## Capabilities

### New Capabilities
- `current-state-documentation`: Cubre la documentación pública del estado real del repositorio, quickstart, separación de roadmap/futuro y consistencia entre README, getting started y README de la CLI Go.

### Modified Capabilities
- Ninguna.

## Impact

- Afecta documentación humana: `README.md`, `docs/getting-started.md`, `docs/roadmap.md` y `tools/lufy-cli-go/README.md`.
- No cambia código, comandos de CLI, assets instalables, APIs, configuración de CI ni contratos públicos.
- La validación esperada es estática/documental: revisión de enlaces/rutas, ausencia de promesas de features no implementadas y coherencia con specs/roadmap existentes.
