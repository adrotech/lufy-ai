# Estado del proyecto

Este documento separa capacidades reales de roadmap. El README debe enlazar solo capacidades instalables o explícitamente marcadas como dry-run/preview.

## Implementado

### Harness y arquitectura

- Harness SDD proporcional con T1 Full SDD, T2 SDD Lite y T3 Express.
- Result Contract envelope v1 para handoffs, evidencia, riesgos y siguiente acción.
- Review Workload Harness con `review_slices` para T1/T2 con varios riesgos.
- Skill resolution local-first con AutoSkills solo como bootstrap opcional y autorizado.
- Memoria Obsidian portable como fuente canónica cuando `.lufy/config/project.yaml` declara `memory.provider: obsidian`.
- Paralelismo gobernado para `review_slices` independientes con plan de merge y validación agrupada.
- Core neutral con separación inicial de tool adapters y methodology adapters.
- Registry de adapters con `opencode` como adapter escribible default.
- Adapter `codex` escribible core con assets project-locales bajo `.agents` y `.codex`.
- Adapter dry-run para `claude-code`, sin escritura real.
- Methodology adapters para `openspec`, `lufy-sdd` y `none`.
- Methodology por tier registrada en manifest y expuesta por `verify/status --json`.

### CLI Go

- CLI Go canónica en `tools/lufy-cli-go`.
- `scripts/install.sh` como wrapper estricto de `lufy-ai install`, sin fallback legacy.
- `install` idempotente con catálogo, SHA-256, manifest schema v2 y backups.
- `uninstall` gestionado con dry-run, backup, drift guard, preservación de `opencode.json` y remoción mínima de referencia en `AGENTS.md`.
- `verify` estructural con salida humana, `--json`, `--quiet`, `--verbose` y `--deep`.
- `status` humano/JSON con drift, frozen assets, `.lufy-new` pendiente y detalles por asset.
- `info` humano/JSON con catálogo efectivo, manifest, stacks, surfaces y conteos operativos.
- `doctor` humano/JSON para diagnóstico read-only de `.lufy/config/project.yaml`, manifest, drift y conflictos pendientes.
- `pin`/`unpin` para congelar assets gestionados desde el manifest sin editar el archivo target.
- `sync` conservador para assets gestionados sin drift.
- `merge` para reconciliar `.lufy-new` con ancestor seguro y cerrar estado/sidecar.
- `backup` y `restore` con manifest, targetRoot y validación de paths/hashes.
- `upgrade` autoreemplazante con versión fija y SHA-256.
- `version` con metadata de build.
- `init` y `--rescan` para `.lufy/config/project.yaml` stack-aware.
- `memory init/status/validate/search` para crear, diagnosticar, validar y buscar memoria Obsidian en repos destino.

### Assets instalables

- Agentes OpenCode: `orchestrator`, `sdd-router`, `explorer`, `implementer`, `test-writer`, `validator`, `reviewer`, `delivery`.
- Skills OpenSpec `sdd-workflow`.
- Templates `sdd-lite.md` y `result-contract.md`.
- Comandos, skills, hooks y template de memoria Obsidian: `/lufy.mem-*`, `lufy.mem-*`, `memory-*.sh` y `memory-note.md`.
- Policy de delivery.
- Comandos `/opsx-*`.
- Comandos `/lufy.*` instalables según catálogo.
- Agent Observatory TUI.
- Superficie Codex core instalable con `lufy-ai install --tool codex`: custom agents Lufy bajo `.codex/agents`, multi-agent habilitado en `.codex/config.toml`, mapeo explícito native/emulated/inline y contrato `@<rol-lufy>` como delegación en `.codex/lufy-agent-mapping.md`, skills/workflows repo-locales bajo `.agents/skills`, `.codex/hooks.json`, `.codex/rules` y bloque gestionado en `AGENTS.md`.
- OpenSpec core v2/stay-updated: config action-based, specs delta, `/opsx-sync`, `UPSTREAM.json`, `opsx-version` y resolver PATH/cache/embedded.
- Lufy SDD inicial bajo `.lufy/workflows/sdd/` cuando se selecciona.

### Release, calidad y seguridad

- Bootstrap remoto con checksum, validación de tar entries, retry y timeouts.
- Release workflow con artifacts por OS/arch, checksums, SBOM, provenance y firma cosign.
- `scripts/validate.sh` como gate local agrupado.
- Coverage objetivo `80.0%`.
- Action pinning y YAML checks.
- Release estable solo desde tags `v*` alcanzables desde `main`.

## Pendiente o futuro

- Promover `develop` a `main` y publicar la próxima release estable desde tag `v*`.
- Plugin marketplace, Observatory y reporting avanzado para Codex.
- Adapter escribible real para Claude Code.
- Lufy SDD full como alternativa completa a OpenSpec.
- Templates por stack como paquetes instalables.
- Subagentes de dominio adicionales.
- Planner 8-state completo.
- Reconciliation hook opt-in.
- Autocomplete/help avanzado mediante Cobra u otro framework.
- Verificación cosign integrada en `upgrade`.
- Two-phase apply completo si el rollback acotado resulta insuficiente.
- Migración a GoReleaser si reduce mantenimiento frente al script actual.

## Límites actuales

- Los adapters escribibles actuales son `opencode` y `codex`; `opencode` sigue siendo el default efectivo.
- Algunas sesiones Codex pueden exponer solo roles genéricos (`default`, `explorer`, `worker`); Lufy debe usar roles nativos cuando tool discovery los expone y degradar explícitamente a `emulated`/`inline` cuando no.
- `claude-code` no debe documentarse como instalable real; sigue dry-run/preview.
- `none` no es metodología universal: T1/T2 siguen protegidos por policy.
- `AGENTS.md`, `opencode.json` y `.lufy/config/project.yaml` son user-owned o user-managed.
- No existe suite Node/TS global en la raíz.
- No hacer delivery sin autorización explícita.
