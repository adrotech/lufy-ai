# Arquitectura

`lufy-ai` distribuye una CLI Go en `tools/lufy-cli-go` y assets gestionados para el preset actual OpenCode/OpenSpec. La capa instalada es un harness SDD proporcional: clasifica trabajo en T1 Full SDD, T2 SDD Lite o T3 Express antes de elegir agentes, contexto, permisos y validación.

El core está migrando a un modelo hexagonal: Lufy conserva tiers, roles, Result Contract, policies, validación y managed assets como dominio neutral; las superficies concretas viven en adapters de tool y metodología. Hoy el único tool adapter instalable es `opencode`; `codex` y `claude-code` existen solo como adapters dry-run para capabilities/render preview, sin escritura real. Las metodologías soportadas por configuración son `openspec`, `lufy-sdd` como adapter foundation no instalable y `none`.

## Componentes

- `internal/assets`: catálogo, hashes SHA-256 y assets embebidos.
- `internal/core/domain`: modelos neutrales de harness, tiers, roles, metodología por tier y routing policy.
- `internal/adapters/tool/opencode`: adapter inicial para paths, capabilities y render de agentes OpenCode.
- `internal/adapters/tool/codex`: adapter dry-run para perfilar capabilities Codex y render preview basado en `AGENTS.md`, sin instalar assets.
- `internal/adapters/tool/claudecode`: adapter dry-run para perfilar capabilities Claude Code y render preview basado en `CLAUDE.md`, sin instalar assets.
- `internal/adapters/methodology/openspec` y `internal/adapters/methodology/none`: adapters iniciales de metodología.
- `internal/instructions/registry` y `internal/instructions/render`: contratos neutrales de roles/skills y superficie renderizable sin paths de tool hardcodeados.
- `internal/opsx`: resolución stay-updated de OpenSpec en capas `PATH`, cache local y baseline embebida.
- `internal/installer`: plan/apply de instalación inicial y actualización gestionada.
- `internal/syncer`: sincronización conservadora de assets registrados.
- `internal/backup`: backups, manifests, restore y rollback acotado.
- `internal/verify`: verificación estructural, drift y salida JSON.
- `internal/config`: merge conservador de `opencode.json`.
- `internal/platform`: path safety, locks y resolución portable de targets.
- `scripts/bootstrap.sh`: descarga verificada del binario publicado.

## Harness instalado

- `.opencode/agents/sdd-router.md`: router read-only para tiering, execution mode, context slicing, skill status y review workload.
- `.opencode/templates/sdd-lite.md`: mini-spec T2 para cambios acotados con criterios observables `WHEN`/`THEN`.
- `.opencode/templates/result-contract.md`: contrato compacto para handoffs, recuperación de contexto y reportes finales.
- `.opencode/policies/delivery.md`: invariantes compartidas de branch safety, validación, autorización y release.
- `.opencode/agents/delivery.md`: runbook operativo del subagente que aplica la policy cuando hay autorización explícita.

El Review Workload Harness se integra al routing: T1 y T2 con varios ejes de riesgo pueden producir `review_slices` para que el cambio sea revisable por partes, con evidencia y riesgos claros. La guía de PR es recomendación hasta que `delivery` reciba autorización explícita.

## Assets embebidos e installer

El binario standalone usa `go:embed` para incluir `tools/lufy-cli-go/internal/assets/embedded`. Cuando cambian agentes, templates, policies, comandos o specs base, el cambio debe reflejarse en esos assets embebidos y en el catálogo de `internal/assets/catalog.go`; si no, una instalación nueva desde binario quedaría detrás de la documentación raíz.

`.lufy-ai/install-state.json` usa schema v2 y registra ownership por asset: `tool`, `methodology`, `component`, `policy`, `scope` y hashes. La lectura de estados schema v1 se mantiene: al cargar se completan defaults compatibles para poder verificar o sincronizar instalaciones existentes.

## Decisiones relevantes

- Ver `docs/adr/0001-managed-asset-source-of-truth.md` para fuente de verdad de assets.
- Ver `docs/adr/0002-managed-directories-extra-files.md` para archivos extra en directorios gestionados.
- Ver `docs/adr/0003-opencode-json-ownership.md` para ownership parcial de `opencode.json`.
- Ver `docs/adr/0004-sync-new-catalog-assets.md` para semántica de nuevos assets en `sync`.
- Ver `docs/adr/0005-recovery-and-rollback.md` para rollback acotado.
- `openspec/UPSTREAM.json` declara la baseline OpenSpec, versión mínima compatible y cache `.lufy-ai/openspec-cache/<version>/`; el instalador no descarga OpenSpec remoto durante `install`/`sync` por defecto.
