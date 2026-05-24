# Estado del proyecto

## Implementado

- CLI Go en `tools/lufy-cli-go`.
- Instalación y sync con catálogo, hashes SHA-256 y manifest de estado.
- Backups con manifest, retención local y restore validado.
- Rollback automático acotado cuando existe backup de recovery.
- `verify` estructural con salida humana, `--json` y `--quiet`.
- `status` con salida humana y `--json`.
- `upgrade` autoreemplazante con versión fija y verificación SHA-256.
- `verify --deep` para referencias de plugins en `tui.json` y `opencode.json`.
- Bootstrap remoto con checksum, validación de tar entries, retry y timeouts.
- Release con actions pinneadas, SBOM, provenance y firma cosign.
- Drift Resolution publicado en `v0.2.0`: policies por asset, ancestors, `.lufy-new`, `merge-block` para `AGENTS.md`, `--scope`, `merge` y restore por ID/listado.
- OpenSpec core v2/stay-updated listo para `v0.3.0`: config action-based, specs delta, scenarios testables, `/opsx-sync`, `UPSTREAM.json`, `opsx-version` y resolver PATH/cache/embedded.
- Harness SDD proporcional: `sdd-router`, T1 Full SDD, T2 SDD Lite, T3 Express, execution modes, result contracts, context slicing y review workload.
- Review Workload Harness: `review_slices` proporcionales para T1 y T2 con varios riesgos, orientados al reviewer humano y a entregables pequeños cuando conviene.
- Templates instalables de proceso: `.opencode/templates/sdd-lite.md` y `.opencode/templates/result-contract.md`.
- Skill resolution local-first con AutoSkills documentado solo como bootstrap opcional mediante `npx autoskills --dry-run` y autorización explícita antes de comandos mutantes.
- `sdd-router` limitado a clasificación read-only/no-shell y delivery con gate explícito de checks remotos de PR antes de `delivered`/`closed`.

## Pendiente o futuro

- Promover `develop` a `main` y publicar `v0.3.5` como release estable desde un tag `v*` alcanzable desde `origin/main`, con evidencia de CI, checks remotos de PR y workflow de release.
- OpenSpec expanded profile queda pendiente para un sprint posterior; no forma parte del release `v0.3.5` objetivo.
- Backlog stack-aware: Release A (`lufy-ai init`, `.opencode/project.yaml` y rescan/drift) queda cubierta por PRs previos; `/lufy.timereport` (LUFY-3) fue entregado por PR #66 y archivado post-merge. Siguen pendientes consumidores como `test-writer`, reviewer ponderado, `/lufy.onboard`, hooks dinámicos y planner 8-state; ver `docs/backlog.md`.
- Reconciliation hook opt-in para detectar cambios sin spec asociada.
- Autocomplete/help avanzado mediante Cobra u otro framework.
- Verificación cosign integrada en `upgrade`.
- Deep verify de plugins y schemas externos.
- Two-phase apply completo si el rollback acotado resulta insuficiente.
- Migración a GoReleaser si reduce mantenimiento frente al script actual.
