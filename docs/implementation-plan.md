# Plan de implementacion v0.2.0 -> v0.3.0

Este plan consolida los RFCs `LUFY-AI-DRIFT-RESOLUTION-RFC.md`, `LUFY-AI-OPENSPEC-V2-RFC.md` y `LUFY-AI-ROADMAP-RFC.md` como guia operativa para las proximas iteraciones de `lufy-ai`.

Documento historico de planificacion. Varias piezas ya fueron implementadas en ciclos posteriores: CLI Go canónica, drift resolution, OpenSpec core v2, harness SDD proporcional, foundation hexagonal de adapters, methodology por tier y `lufy-ai uninstall`. Para estado actual, usar [`docs/status.md`](status.md); para arquitectura vigente, usar [`docs/architecture.md`](architecture.md).

No debe leerse como contrato actual de producto ni como lista exhaustiva de pendientes. Cada bloque nuevo debe convertirse en proposal OpenSpec antes de implementar, salvo tareas puramente administrativas autorizadas.

## Objetivo

Hacer que `lufy-ai` pueda actualizarse en repositorios reales sin bloquear por drift esperado ni pisar trabajo local, y luego modernizar el workflow OpenSpec instalado hasta paridad funcional con OpenSpec v1.3.1.

## Orden recomendado

1. `v0.2.0`: Drift Resolution.
2. `v0.3.0`: OpenSpec v2 / paridad v1.3.1.
3. Patches `v0.2.x` o `v0.3.x`: bugfixes acotados y mejoras de release/UX.

La razon para hacer Drift Resolution primero es operacional: `AGENTS.md`, `.opencode/policies/*`, `openspec/` y otros assets son lugares donde el usuario naturalmente personaliza. Modernizar OpenSpec antes de resolver ese modelo aumenta conflictos en upgrades brownfield.

## Release v0.2.0 — Drift Resolution

Rama sugerida: `feature/drift-resolution` desde `develop`.

### Entregable 1: Policies declarativas por asset

Objetivo: reemplazar el ownership binario por policies que expresen como actualizar cada asset.

Cambios esperados:

- Expandir `assets.Policy` a `managed`, `no-replace`, `merge-block`, `merge-json` y `metadata`.
- Persistir `policy`, `scope` y datos de ancestor en `.lufy-ai/install-state.json` con migracion silenciosa desde schema actual.
- Ajustar `install`, `sync`, `verify` y `status` para reportar policy por asset.
- Mantener compatibilidad con assets existentes y no romper targets ya instalados.

Acceptance:

- Un target con `install-state.json` anterior se lee y migra sin perder hashes.
- Assets `managed` se actualizan con backup cuando no hay drift local.
- Assets `no-replace` con drift dejan `<path>.lufy-new` y preservan el archivo original.
- `verify --json` expone drift/policy de forma estructurada.

### Entregable 2: Scope global/proyecto

Objetivo: reducir drift aprovechando discovery global + project de OpenCode.

Cambios esperados:

- Agregar `--scope=global|project|both` a `install`, `sync`, `verify` y `status` donde aplique.
- Definir scope por entry del catalogo.
- Instalar assets OpenCode compartidos en `~/.config/opencode/` cuando el usuario elige scope global o both.
- Mantener `--scope=project` como reproduccion del comportamiento actual.
- Mantener assets intrinsicamente per-project en el target: `openspec/`, `AGENTS.md`, `tui.json` y `.lufy-ai/`.

Decision pendiente antes de implementar:

- Default de release para `v0.2.0`: mantener `--scope=project` para preservar comportamiento actual. `global` y `both` quedan disponibles como opt-in hasta validar una RC/brownfield más amplia.

Acceptance:

- `install --scope=project` preserva el comportamiento actual.
- `install --scope=global` no crea `.opencode/` de proyecto salvo assets per-project requeridos.
- `verify` explica claramente que esta verificando global, project o both.

### Entregable 3: MergeBlock para `AGENTS.md`

Objetivo: permitir convenciones locales del usuario sin bloquear upgrades del roster/reglas gestionadas.

Cambios esperados:

- Convertir `AGENTS.md.template` a bloques delimitados por `<!-- LUFY:BEGIN <id> -->` y `<!-- LUFY:END <id> -->`.
- Implementar motor `merge-block` stdlib-only.
- Hacer que `install` y `sync` actualicen solo bloques gestionados, preservando texto fuera de bloques.
- Reportar bloques faltantes, duplicados o mal cerrados como errores accionables.

Acceptance:

- Un `AGENTS.md` con texto local antes/despues de bloques conserva ese texto tras upgrade.
- Un cambio de template actualiza solo el contenido dentro del bloque gestionado.
- Marcadores corruptos bloquean escritura y explican como reparar.

### Entregable 4: `.lufy-new`, merge y restore UX

Objetivo: desbloquear upgrades con drift sin perder datos.

Cambios esperados:

- Crear `<file>.lufy-new` para nuevas versiones de assets `no-replace` con drift.
- Guardar ancestors bajo `.lufy-ai/ancestors/` para three-way merge futuro.
- Agregar o extender comando `lufy-ai merge <path>` con `LUFY_MERGE_TOOL` y default seguro documentado.
- Extender `restore` para listar backups por ID y restaurar el ultimo o uno especifico con dry-run.
- Mantener retencion de backups y no borrar datos del usuario sin confirmacion.

Acceptance:

- El upgrade nunca pisa un archivo user-owned con drift.
- `.lufy-new` queda registrado y visible en `status`/`verify`.
- `merge <path>` valida ancestor/user/new antes de invocar herramienta externa.
- `restore --dry-run` lista exactamente que restauraria.

Estado de implementación en rama: `restore --list` muestra IDs/timestamps/manifests y `restore --backup <id>` resuelve backups bajo `.lufy-ai/backups/`; `restore --backup <manifest-or-dir>` se mantiene compatible.

### Validacion v0.2.0

Comandos minimos antes de PR:

- `scripts/validate.sh`
- `git diff --check origin/develop`
- `tools/lufy-cli-go/scripts/smoke-install.sh`
- Smokes sandbox manuales o automatizados para greenfield, brownfield con `AGENTS.md` customizado y repo con `.opencode/` existente.

## Release v0.3.0 — OpenSpec v2 / Paridad v1.3.1

Rama sugerida: `feature/openspec-v2` desde `develop` despues de mergear `v0.2.0`.

### Sprint 1: Core gap closure

Objetivo: alinear el flujo instalado con OpenSpec v1.3.1 core.

Cambios esperados:

- Reescribir `openspec/config.yaml` a schema v2 action-based.
- Exigir delta markers `ADDED`, `MODIFIED`, `REMOVED` en specs de cambios.
- Exigir scenarios `GIVEN/WHEN/THEN` o formato equivalente testable definido por OpenSpec v1.3.1.
- Agregar `/opsx-sync` y skill `openspec-sync` para aplicar deltas a specs principales.
- Agregar `UPSTREAM.json` con baseline OpenSpec efectiva.
- Agregar `opsx-version` para reportar version efectiva y fuente.

### Sprint 2: Stay-updated 3 capas

Objetivo: desacoplar upgrades de OpenSpec upstream del release manual de `lufy-ai`.

Resolucion runtime:

1. Delegar a `openspec` CLI en PATH si existe y cumple version minima.
2. Usar cache `.lufy-ai/openspec-cache/<version>/` si existe.
3. Usar baseline embebida en el binario como fallback offline.

Cambios esperados:

- Paquete interno `opsx` con manifest, resolver, fetcher/delegator y validator.
- Workflow `sync-openspec.yml` con PR automatico para bumps de baseline, no merge automatico.
- Escrituras atomicas para cache y manifests.

### Sprint 3: Expanded profile

Objetivo: completar la superficie OpenSpec moderna para OpenCode.

Comandos/skills esperados:

- `/opsx-new` / `openspec-new`
- `/opsx-continue` / `openspec-continue`
- `/opsx-ff` / `openspec-ff`
- `/opsx-bulk-archive` / `openspec-bulk-archive`
- `/opsx-onboard` / `openspec-onboard`
- `/opsx-doctor` o `openspec-validate` para diagnostico y enforcement

Acceptance:

- Profile `core` instala superficie minima.
- Profile `expanded` instala todos los comandos nuevos.
- El default queda documentado y testeado.

### Sprint 4: Reconciliation hook

Objetivo: detectar cambios de codigo/configuracion sin spec asociada antes de que aparezca drift documental.

Cambios esperados:

- Script opt-in `scripts/hooks/pre-commit-reconcile.sh`.
- No sobrescribir hooks existentes del usuario.
- Reportar archivos cambiados sin change OpenSpec asociado y sugerir siguiente accion.

### Sprint 5: Docs, tests y release

Objetivo: cerrar comunicacion, migracion y evidencia.

Cambios esperados:

- `docs/openspec-v2.md`.
- `docs/migration-v0.2-to-v0.3.md`.
- CHANGELOG para `v0.3.0`.
- Smokes de upgrade desde target `v0.2.0` con `AGENTS.md` customizado.

### Validacion v0.3.0

Comandos minimos antes de PR:

- `scripts/validate.sh`
- `openspec validate --all`
- `git diff --check origin/develop`
- Smokes sandbox: greenfield Go, brownfield TypeScript y monorepo mixto.
- Verificar que un proposal sin deltas/scenarios falla con mensaje accionable.

## Backlog posterior

Temas fuera de `v0.2.0` y `v0.3.0`:

- Soporte multi-tool fuera de OpenCode.
- Plugin system para skills custom.
- Homebrew/Scoop antes de estabilizar releases GitHub.
- Convertir agentes a skills como estrategia general.

## Preparacion de proposals

Proposals sugeridas:

1. `resolve-install-drift-policies`: Entregable 1 de Drift Resolution.
2. `support-global-project-install-scope`: Entregable 2 de Drift Resolution.
3. `merge-block-agents-template`: Entregable 3 de Drift Resolution.
4. `add-lufy-new-merge-workflow`: Entregable 4 de Drift Resolution.
5. `modernize-openspec-core-v2`: Sprint 1 de OpenSpec v2.
6. `add-openspec-stay-updated-fallback`: Sprint 2 de OpenSpec v2.
7. `expand-openspec-opencode-profile`: Sprint 3 de OpenSpec v2.
8. `add-openspec-reconciliation-hook`: Sprint 4 de OpenSpec v2.
9. `document-openspec-v2-release`: Sprint 5 de OpenSpec v2.

Cada proposal debe incluir specs delta, tasks verificables, plan de migracion si cambia estado persistido, y evidencia real de validacion antes de delivery.
