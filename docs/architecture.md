# Arquitectura

`lufy-ai` distribuye una CLI Go en `tools/lufy-cli-go` y assets gestionados para OpenCode/OpenSpec. La capa instalada es un harness SDD proporcional: clasifica trabajo en T1 Full SDD, T2 SDD Lite o T3 Express antes de elegir agentes, contexto, permisos y validación.

## Componentes

- `internal/assets`: catálogo, hashes SHA-256 y assets embebidos.
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

## Decisiones relevantes

- Ver `docs/adr/0001-managed-asset-source-of-truth.md` para fuente de verdad de assets.
- Ver `docs/adr/0002-managed-directories-extra-files.md` para archivos extra en directorios gestionados.
- Ver `docs/adr/0003-opencode-json-ownership.md` para ownership parcial de `opencode.json`.
- Ver `docs/adr/0004-sync-new-catalog-assets.md` para semántica de nuevos assets en `sync`.
- Ver `docs/adr/0005-recovery-and-rollback.md` para rollback acotado.
- `openspec/UPSTREAM.json` declara la baseline OpenSpec, versión mínima compatible y cache `.lufy-ai/openspec-cache/<version>/`; el instalador no descarga OpenSpec remoto durante `install`/`sync` por defecto.
