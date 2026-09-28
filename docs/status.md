# Estado del producto

Este documento describe el estado instalable de la release estable `v0.6.24`. El código en una rama, una proposal o un change activo no cuenta como capacidad publicada hasta completar validación, merge a `main`, tag y release.

## Capacidades vigentes

| Área | Estado | Superficie principal |
| --- | --- | --- |
| Tool adapters | Disponible | OpenCode es el default; Codex instala una superficie project-local autocontenida. |
| Methodology adapters | Disponible | OpenSpec Full/Lite, Lufy SDD Full/Lite y `none` sólo donde la policy lo permite. |
| Routing proporcional | Disponible | T1 Full SDD, T2 SDD Lite y T3 Express, con Review Workload Harness y slices revisables. |
| Surface Execution Plan | Disponible | `lufy-ai plan` resuelve superficies, contratos conectados, capabilities y validaciones sin ejecutarlas. |
| Result Contract | Disponible | `lufy-ai result validate|normalize|transition`; las transiciones usan policy, evidencia y CAS. |
| Run Ledger | Disponible | `lufy-ai run record|checkpoint|status|summary|verify|prune`, con eventos append-only y payload content-free. |
| Adaptive routing | Disponible y opt-in | `disabled` por default; `shadow` observa y `advisory` permite assign/yield explícitos con leases y budget. |
| Context Graph | Disponible | Build/query/path/explain/diff más coverage, trace, review y metrics como índices secundarios. |
| Memoria | Disponible | Obsidian portable bajo `.lufy/memory`, con capture, connect, validate, search e index. |
| Skill registry | Disponible | `.lufy/skill-registry.json` derivado, local-first y con precedencia project-over-global. |
| Managed assets | Disponible | Install/sync/uninstall, manifest schema v2, SHA-256, backup/restore, pin/unpin y merge seguro. |
| Observabilidad humana | Disponible | Agent Observatory para OpenCode y reportes HTML offline de proposal, PR y tiempo. |
| Release | Disponible | Binarios multi-OS/arch, checksums, SBOM, provenance, firma keyless y E2E post-release. |

## Roles del harness

El núcleo instala ocho roles estables: `orchestrator`, `sdd-router`, `explorer`, `implementer`, `test-writer`, `validator`, `reviewer` y `delivery`. Adaptive routing puede sugerir un `role_hint`, pero no crea roles, no cambia identidad, no transfiere ownership y no amplía permisos.

## Fuentes de estado

En orden de autoridad:

1. código, configuración y contratos actuales;
2. tests, validaciones y resultados de CI;
3. release/tag alcanzable desde `main`;
4. specs activas y ADRs;
5. Context Graph y memoria como orientación secundaria;
6. roadmap como hipótesis futura, nunca como evidencia de disponibilidad.

## Límites actuales

- `claude-code` permanece como preview dry-run; no es un adapter escribible.
- Codex no incluye todavía marketplace, Observatory TUI ni reporting avanzado equivalente a OpenCode.
- `setup` usa defaults `opencode`/`project`; para otro tool, scope o metodología por tier se usan los comandos individuales.
- No hay templates instalables por stack ni subagentes de dominio adicionales.
- Context Graph es lexical y determinístico; sus inferencias no reemplazan archivos, tests, logs ni comandos.
- Run Ledger no guarda prompts, respuestas, transcripts, diffs, secretos ni contenido libre.
- Adaptive routing no avanza gates y preempta ante delivery, seguridad, contratos públicos, schema de base de datos o migraciones destructivas.
- Delivery, commit, push, PR, merge y promoción siguen requiriendo autorización explícita.

## Cómo verificar un target

```bash
lufy-ai version
lufy-ai status --target <repo> --verbose
lufy-ai doctor --target <repo>
lufy-ai verify --target <repo> --deep
```

Para el flujo completo y la relación entre componentes, ver [`harness-workflow.md`](harness-workflow.md).
