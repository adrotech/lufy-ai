## Why

El slice anterior `migrate-installer-to-go-cli` dejó una CLI Go inicial en `tools/lufy-cli-go` y convirtió `scripts/install.sh` en un wrapper estricto, pero la instalación real todavía necesita cubrir el kit completo de `lufy-ai` con reglas robustas de idempotencia. Este cambio propone el siguiente slice: instalar assets gestionados reales en un proyecto destino y decidir `copy`, `skip`, `update-managed` o `conflict` usando contenido/hash, con estado persistente para ejecuciones repetibles y auditables.

## What Changes

- Extender la CLI Go `lufy-ai` bajo `tools/lufy-cli-go` para instalar el kit completo de assets gestionados en un `--target`.
- Incorporar un catálogo/manifest de assets versionado que cubra:
  - `.opencode/agents`
  - `.opencode/commands`
  - `.opencode/skills`
  - `.opencode/policies`
  - `.opencode/plugins`
  - `.opencode/agent-observatory`
  - `AGENTS.md`
  - `tui.json`
  - `openspec/`
  - metadatos necesarios de instalación bajo `.lufy-ai/`
- Construir un plan fiel antes de mutar el filesystem, incluyendo acciones `create-dir`, `copy`, `skip`, `update-managed`, `conflict` y `backup`.
- Persistir `.lufy-ai/install-state.json` con versión de schema, hashes source/target, timestamps, `sourceChangeID` y lista de assets instalados.
- Implementar idempotencia por hash:
  - archivo ausente => `copy`.
  - archivo existente con hash igual al source => `skip`.
  - archivo gestionado previamente y source upstream cambió => `update-managed` con backup previo.
  - archivo existente no gestionado, o con hash distinto al estado esperado => `conflict` y no sobrescritura sin decisión explícita.
- Ampliar backup/restore para capturar todos los assets tocados por instalación, actualización o restauración.
- Ampliar `verify` para validar estructura completa, manifest de estado, hashes y ausencia de mutaciones fuera de `--target`.
- Mantener `scripts/install.sh` como wrapper estricto que delega en la CLI Go; no reintroducir lógica Bash legacy.
- No hay cambios **BREAKING** previstos para el wrapper ni para comandos ya existentes; el comportamiento debe endurecer seguridad e idempotencia.

## Capabilities

### New Capabilities

- `managed-assets-install`: Instalación completa de assets gestionados de `lufy-ai` con catálogo/hash, dry-run fiel, backups multiasset, restore y verify estructural.

### Modified Capabilities

- Ninguna. No existen specs activas bajo `openspec/specs/` que deban modificarse; el contrato previo de `go-cli-installer` fue un slice inicial y este cambio agrega una capacidad más específica.

## Impact

- Código afectado: `tools/lufy-cli-go/internal/installer`, `tools/lufy-cli-go/internal/assets`, `tools/lufy-cli-go/internal/backup`, `tools/lufy-cli-go/internal/verify`, `tools/lufy-cli-go/internal/platform` y entrypoints/dispatch asociados si requieren nuevas opciones.
- Tests afectados: pruebas Go unitarias e integración con temp dirs desde `tools/lufy-cli-go/`, especialmente planificación, hash/idempotencia, conflictos, backup/restore y verify.
- Assets fuente: directorios `.opencode/**`, `AGENTS.md`, `tui.json` y `openspec/` deberán resolverse desde el checkout fuente sin hardcodear rutas absolutas.
- Estado destino: se añadirá o actualizará `.lufy-ai/install-state.json` y backups bajo `.lufy-ai/backups/` dentro del `--target`.
- Seguridad: la CLI no debe tocar fuera de `--target`, no debe seguir symlinks peligrosos, debe respaldar antes de mutar y no debe hardcodear Engram.
- Fuera de alcance: distribución binaria/release, TUI, multiagente externo y sync cloud.
