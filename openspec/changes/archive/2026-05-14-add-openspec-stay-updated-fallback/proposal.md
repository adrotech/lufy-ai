## Why

El core v2 ya instala un baseline local de OpenSpec, pero ese baseline sigue acoplado al release manual de `lufy-ai`. Para avanzar hacia `v0.3.0`, el workflow necesita poder mantenerse actualizado con OpenSpec upstream de forma controlada, offline-safe y auditable sin romper targets existentes.

## What Changes

- Agregar una capacidad `openspec-stay-updated` para resolver la fuente efectiva de OpenSpec en tres capas: CLI `openspec` en `PATH`, cache local versionada y baseline embebida.
- Introducir un paquete interno Go para resolver baseline/manifiestos OpenSpec, validar versiones mínimas y operar sin red cuando corresponda.
- Definir cache local `.lufy-ai/openspec-cache/<version>/` con manifiesto y escrituras atómicas.
- Agregar un workflow `sync-openspec.yml` que detecte bumps de baseline y abra PR automático, sin merge automático.
- Mantener `UPSTREAM.json` como baseline declarativo y extenderlo solo lo necesario para que el resolver pueda comparar versión/fuente/capacidades.
- No agregar todavía perfil expanded, comandos `/opsx-new`, `/opsx-continue`, hooks de reconciliación ni release `v0.3.0`.

## Capabilities

### New Capabilities

- `openspec-stay-updated`: cubre resolver 3 capas, cache versionada, manifiestos, validación de baseline y workflow de actualización por PR.

### Modified Capabilities

- `openspec-core-v2-workflow`: el baseline local deja de ser solo metadata estática y pasa a participar en resolución stay-updated offline-safe.
- `go-cli-installer`: la CLI Go incorpora resolución/caching OpenSpec sin romper instalación standalone ni dependencia stdlib-only.
- `go-cli-install-ci`: CI valida el resolver/cache y el workflow de sync de baseline sin requerir red en tests normales.

## Impact

- Afecta `tools/lufy-cli-go/internal/` con un nuevo paquete interno para resolver OpenSpec.
- Afecta `openspec/UPSTREAM.json`, assets embebidos y pruebas de paridad.
- Afecta `.github/workflows/` con un nuevo workflow programado/manual para proponer bumps.
- Puede agregar scripts o pruebas de sandbox para cache local, pero no debe introducir dependencias externas en la CLI Go sin decisión explícita.
