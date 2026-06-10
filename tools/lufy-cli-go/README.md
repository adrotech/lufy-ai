# lufy-cli-go

CLI Go canónica de `lufy-ai`. Vive en `tools/lufy-cli-go` y reemplaza la lógica histórica del instalador Bash. `scripts/install.sh` solo delega en `lufy-ai install`; no debe reintroducir fallback legacy.

## Propósito

- Instalar y mantener el harness en repositorios existentes.
- Gestionar assets con manifest, hashes SHA-256, backups y restore.
- Separar core de harness, tool adapters y methodology adapters.
- Permitir upgrades, sync y uninstall sin pisar trabajo local.
- Exponer validación estructural reproducible para usuarios y CI.

## Estructura

```text
tools/lufy-cli-go/
  cmd/lufy-ai/main.go        # entrypoint delgado
  internal/cli/              # parser, dispatch y exit codes
  internal/core/domain/      # modelos neutrales: tiers, roles, methodology_by_tier
  internal/adapters/         # adapters de tool y metodología
  internal/instructions/     # role contracts, skills y render neutral
  internal/assets/           # catálogo, go:embed, policies y SHA-256
  internal/state/            # .lufy-ai/install-state.json schema v1/v2
  internal/installer/        # install planner/apply idempotente
  internal/uninstaller/      # uninstall planner/apply con backup y drift guard
  internal/syncer/           # sync conservador por manifest/hash
  internal/status/           # estado humano/JSON y drift
  internal/governance/       # info/doctor/pin/unpin operativo
  internal/verify/           # verify estructural y deep checks
  internal/memory/           # memoria Obsidian init/status/validate/search
  internal/backup/           # backup/restore multiasset
  internal/config/           # merge conservador de opencode.json
  internal/projectconfig/    # init/rescan de .lufy/project.yaml
  internal/opsx/             # resolución OpenSpec PATH/cache/embedded
  internal/platform/         # path safety, locks y resolución portable
  internal/version/          # metadata de release
```

## Build y test local

Desde `tools/lufy-cli-go/`:

```bash
mkdir -p bin
go build -o bin/lufy-ai ./cmd/lufy-ai
go test ./...
go run ./cmd/lufy-ai version
```

Desde la raíz del repo:

```bash
scripts/validate.sh
```

`scripts/validate.sh` es el gate agrupado preferido. No hay toolchain Node/TS global en la raíz.

## Comandos de usuario

| Comando | Propósito | Flags principales |
| --- | --- | --- |
| `lufy-ai init` | Genera `.lufy/project.yaml` stack-aware/surface-aware y abre selector Bubble Tea cuando hay TTY. | `--target`, `--force`, `--rescan`, `--interactive` |
| `lufy-ai install` | Instala assets gestionados, mergea configs user-owned y escribe manifest SHA-256. | `--target`, `--scope`, `--tool`, `--methodology-tier`, `--dry-run`, `--yes`, `--backup` |
| `lufy-ai uninstall` | Remueve assets gestionados sin drift, crea backup, preserva user-owned y quita solo la referencia Lufy de `AGENTS.md`. | `--target`, `--dry-run`, `--yes`, `--keep-state` |
| `lufy-ai verify` | Valida manifest, hashes, estructura, JSON merge-managed y referencias críticas. | `--target`, `--scope`, `--tool`, `--json`, `--quiet`, `--verbose`, `--deep` |
| `lufy-ai memory init` | Crea `.lufy/memory` y completa defaults de memoria/paralelismo en `.lufy/project.yaml`. | `--target`, `--json` |
| `lufy-ai memory status` | Resume estructura, notas, drafts y backlinks rotos. | `--target`, `--json` |
| `lufy-ai memory validate` | Valida schema de notas Obsidian, decisiones y backlinks. | `--target`, `--json` |
| `lufy-ai memory search` | Busca en `knowledge/` y `maps/` con `rg` cuando está disponible. | `--target`, `--json`, `<query>` |
| `lufy-ai status` | Resume instalación, drift, faltantes, frozen assets y `.lufy-new` pendiente. | `--target`, `--scope`, `--json`, `--verbose` |
| `lufy-ai info` | Muestra catálogo efectivo, manifest, stacks, surfaces y conteos operativos sin mutar. | `--target`, `--scope`, `--json` |
| `lufy-ai doctor` | Diagnostica `.lufy/project.yaml`, manifest, drift y conflictos pendientes sin mutar. | `--target`, `--scope`, `--json` |
| `lufy-ai pin` | Congela un asset gestionado para que `sync` lo preserve sin modificar. | `--target`, `--reason` |
| `lufy-ai unpin` | Remueve el freeze de un asset gestionado. | `--target` |
| `lufy-ai sync` | Reaplica assets gestionados cuando el source cambió y el target no tiene drift local. | `--target`, `--scope`, `--tool`, `--dry-run`, `--yes`, |
| `lufy-ai merge` | Reconcilia `.lufy-new` con edits locales usando ancestor seguro. | `--target`, `--accept-theirs`, `--accept-ours` |
| `lufy-ai backup` | Captura assets gestionados en `.lufy-ai/backups/<timestamp>/manifest.json`. | `--target` |
| `lufy-ai restore` | Restaura desde backup validando target, paths seguros y hashes. | `--target`, `--backup`, `--dry-run`, `--yes`, `--list` |
| `lufy-ai opsx render` | Renderiza un change OpenSpec a HTML offline/autocontenido para revisión humana. | `--target`, `--change`, `--format`, `--theme`, `--output` |
| `lufy-ai upgrade` | Actualiza el binario a una versión fija con checksum. | `--to`, `--dry-run` |
| `lufy-ai version` | Muestra versión, commit, build date, GOOS y GOARCH. | n/a |

## Selección de harness

El adapter escribible actual es `opencode`.

```bash
lufy-ai install --target <repo> --yes
lufy-ai install --target <repo> --tool opencode --yes
```

Adapters no escribibles todavía:

- `codex`;
- `claude-code`.

Ambos existen como dry-run/preview para modelar capabilities y superficies futuras, pero los comandos mutantes los bloquean.

## Methodology por tier

`install` acepta overrides repetibles:

```bash
lufy-ai install --target <repo> --methodology-tier T3:none --yes
lufy-ai install --target <repo> --methodology-tier T2:openspec/lite --methodology-tier T3:none --yes
lufy-ai install --target <repo> --methodology-tier T2:lufy-sdd/lite --yes
```

Reglas actuales:

- `openspec` puede instalar superficie full/lite;
- `lufy-sdd` instala `.lufy/sdd/` como superficie inicial;
- `none` se permite donde la policy lo habilita;
- `T1:none` y `T2:none` están bloqueados en comandos mutantes.

`verify --tool opencode`, `status --json` y `verify --json` exponen `tool`, `schemaVersion` y `methodologyByTier`.

## Project profile

`init` y `scan` escriben `project_profile.surfaces` para separar stack técnico de superficie de producto. Ambos abren el selector Bubble Tea por default cuando hay TTY; usa `--interactive=false` para conservar solo la detección automática. Cada superficie puede incluir `architecture` con `detected`, `preferred`, `options`, `review_required` y `structural_expectations`.

Cuando una superficie se detecta o selecciona como `frontend` o `fullstack`, el `agent_lens` incluye reglas para estructura feature-driven:

- colocation por funcionalidad en `src/features/<feature>/`;
- `components/`, `hooks/`, `services/` y `types.ts` dentro de cada feature cuando son exclusivos de esa funcionalidad;
- `index.ts` como barril público y frontera de acoplamiento de la feature;
- `src/pages/` solo para routing/layouts;
- `src/components`, `src/hooks`, `src/services` y `src/utils` reservados para piezas globales compartidas.

Esto queda persistido en `agent_lens.structural_expectations` para que implementers, validators y reviewers traten la estructura solicitada como aceptación obligatoria. Si el usuario pide carpetas concretas, el harness debe auditar por feature y bloquear `validated`/readiness cuando páginas, hooks o utilidades quedan en la raíz de la feature sin confirmación explícita del usuario.

Para `backend`, las opciones de arquitectura son:

- `controller_service_repository`: default mínimo con controllers/handlers delgados, servicios para reglas de negocio y repositories para persistencia/adapters;
- `clean_architecture`: capas de dominio/casos de uso/infraestructura cuando el proyecto ya lo usa o el usuario lo elige;
- `hexagonal`: ports/adapters alrededor del dominio cuando el proyecto ya lo usa o el usuario lo elige.

El scanner detecta señales existentes como `controllers` + `services` + `repositories`, `domain` + `usecase/application` + `infrastructure`, o `ports` + `adapters`. En modo interactivo, Bubble Tea permite cambiar la arquitectura preferida antes de escribir `.lufy/project.yaml`. La arquitectura elegida persiste `architecture.structural_expectations`: controller/service/repository exige handlers delgados, servicios con reglas de negocio y repositorios aislando persistencia; clean architecture exige capas domain/usecase-or-application/infrastructure; hexagonal exige dominio, ports y adapters.

En `fullstack`, la surface de flujo mantiene frontend feature-driven; la arquitectura clean/hexagonal/controller-service-repository aplica solo al backend y se lee desde la surface backend conectada.

## Memoria Obsidian

`init` y `scan` escriben defaults de memoria en `.lufy/project.yaml`:

```yaml
memory:
  provider: obsidian
  root: .lufy/memory
  git_policy: ignored
  schema_version: 1
  search: rg
  backlinks_index: .lufy/memory/index/backlinks.json
```

`lufy-ai memory init` crea una estructura idempotente:

```text
.lufy/memory/
  MEMORY.md
  inbox/
  knowledge/
  maps/_app-profile.md
  index/backlinks.json
  .gitignore
```

`doctor` reporta memoria faltante, drafts y backlinks rotos sin bloquear instalación normal. `verify --deep` valida memoria cuando existe. `sync` gestiona los comandos, skills, hooks y templates de memoria, pero no registra ni sobrescribe contenido privado dentro de `.lufy/memory/inbox` o `.lufy/memory/knowledge`.

## Paralelismo gobernado

`init` y `scan` también escriben:

```yaml
parallel_execution:
  enabled: true
  strategy: independent_review_slices
  max_parallel_agents: 3
  requires_independent_files: true
  requires_merge_plan: true
  validation_mode: grouped_after_join
```

El CLI solo persiste la política. La decisión operacional queda en `sdd-router`: recomienda paralelismo únicamente para `review_slices` independientes con archivos separados y plan de merge; bloquea delivery, migraciones, contratos compartidos o cambios sobre los mismos archivos.

## OpenSpec helpers

`opsx render` es un helper opcional, tool-agnostic y no bloqueante. Toma artifacts OpenSpec ya generados y produce un HTML autocontenido/offline para revisión humana:

```bash
lufy-ai opsx render --target <repo> --change <name> --format html --theme notion-dark
lufy-ai opsx render --target <repo> --change <name> --format html --theme notion-dark --output /tmp/lufy-opsx-overview.html
```

La salida default es `openspec/changes/<name>/change-overview.html`; con `--output` puede escribirse en otra ruta, por ejemplo `/tmp/lufy-opsx-overview.html`. El render incluye solo Markdown top-level del change: `proposal.md`, `design.md`, `plan.md`, `tasks.md` y otros `.md` directos cuando existen. Excluye `specs/**`.

El HTML usa diseño dark con tabs y no carga recursos remotos. No muestra los badges/textos removidos `Notion dark`, `Offline HTML`, `Artifacts disponibles` ni `Sin recursos remotos`.

El parser Markdown local soporta headings, listas, checkboxes deshabilitados/checked, fenced code, inline code, bold con `**texto**` y links seguros (`http://`, `https://`, `mailto:`). HTML crudo y links inseguros como `javascript:` quedan escapados y no se convierten en anchors. Si un artifact falta o está vacío, se marca como `No disponible` y muestra `Este artefacto no existe o está vacío.`

## Managed assets

El catálogo gestiona assets completos y assets merge-managed.

Assets completos típicos:

- `.opencode/agents`;
- `.opencode/commands`;
- `.opencode/skills`;
- `.opencode/templates`;
- `.opencode/policies`;
- `.opencode/plugins`;
- `.opencode/agent-observatory`;
- `lufy-ia.harness.md`;
- `tui.json`;
- `openspec/`;
- `.lufy/sdd/` cuando aplica.

Assets user-owned o merge-managed:

- `AGENTS.md`: solo referencia `@lufy-ia.harness.md`;
- `opencode.json`: merge conservador;
- `.lufy/project.yaml`: creado por `init`, no sincronizado por hash.
- `.lufy/memory`: creado por `memory init`; notas privadas user-owned, no sincronizadas por hash.

`.lufy-ai/install-state.json` schema v2 registra:

- `tool`;
- `methodologyByTier`;
- ownership por asset;
- policy;
- scope;
- source/target SHA-256;
- ancestors cuando aplica.

## Install

`install`:

1. resuelve target y source;
2. construye plan;
3. respeta `--dry-run`;
4. crea backups cuando corresponde;
5. copia o actualiza assets gestionados sin drift;
6. mergea `opencode.json`;
7. inserta referencia en `AGENTS.md`;
8. escribe `.lufy-ai/install-state.json`;
9. ejecuta verify estructural posterior.

Si hay drift bloqueante, no sobrescribe aunque `--yes` esté presente.

## Uninstall

`uninstall`:

1. lee `.lufy-ai/install-state.json`;
2. planifica remoción de assets gestionados;
3. bloquea si algún asset tiene drift local;
4. crea backup previo;
5. elimina archivos managed y ancestors sin drift;
6. remueve solo `@lufy-ia.harness.md` de `AGENTS.md`;
7. preserva `opencode.json`;
8. elimina `install-state.json` salvo `--keep-state`;
9. limpia directorios vacíos sin borrar directorios con contenido user-owned.

Ejemplo:

```bash
lufy-ai uninstall --target <repo> --dry-run
lufy-ai uninstall --target <repo> --yes
lufy-ai install --target <repo> --yes
lufy-ai verify --target <repo> --quiet
```

## Sync

`sync` reaplica assets gestionados ya registrados.

- Requiere manifest existente.
- Requiere `--yes` para mutaciones reales.
- Crea backup antes de updates.
- Bloquea estado ausente/corrupto, drift local y paths inseguros.
- Reporta `pinned-skip` para assets frozen y preserva sus hashes registrados sin tocarlos.
- Trata `opencode.json` como `merge-json`.

## Verify y status

`verify` valida:

- manifest parseable y schema compatible;
- target root;
- paths críticos;
- assets registrados;
- hashes SHA-256;
- JSON de `opencode.json`, `tui.json`, `.opencode/package*.json` y `openspec/UPSTREAM.json`;
- referencia `@lufy-ia.harness.md`;
- tool esperada cuando se pasa `--tool`;
- referencias de plugins con `--deep`.

`status` resume lo mismo con foco operativo y puede emitir JSON. También expone `pinned` y `conflictsPending`; `doctor` falla cuando quedan `.lufy-new` pendientes y reporta frozen assets como información.

## Governance

`pin` y `unpin` son mutaciones solo de manifest. No editan el asset target.

```bash
lufy-ai pin --target <repo> --reason "override local" lufy-ia.harness.md
lufy-ai sync --target <repo> --dry-run --yes
lufy-ai unpin --target <repo> lufy-ia.harness.md
```

Un asset pinned/frozen queda registrado con `pinned`, `pinnedAt` y `pinnedReason` en `.lufy-ai/install-state.json`. Mientras siga frozen, `sync` lo preserva aunque el catálogo cambie.

## `.lufy/project.yaml`

`lufy-ai init` crea configuración stack-aware user-managed.

Comportamiento:

- si no existe, crea `.lufy/project.yaml`;
- si existe, falla sin `--force`;
- `--force` reemplaza;
- `--rescan` preserva overrides y agrega evidencia nueva;
- `--interactive=false` desactiva el selector Bubble Tea;
- `workflow_limits` es la fuente canónica de límites.

Stacks v1: Go, JavaScript/TypeScript, React, Next, Remix, Vue, Svelte, Python, Java/Kotlin. Stacks no soportados se reportan como `supported: false`.

## Release y bootstrap

La CLI se distribuye como binario standalone con assets embebidos por `go:embed`.

Release:

- artifacts por OS/arch;
- checksums SHA-256;
- SBOM/provenance/firma cuando el workflow corre;
- tags `v*` solo sobre commits alcanzables desde `origin/main`.

Bootstrap:

```bash
curl -fsSL https://raw.githubusercontent.com/adrotech/lufy-ai/v0.6.11/scripts/bootstrap.sh -o /tmp/lufy-bootstrap.sh
bash /tmp/lufy-bootstrap.sh --version v0.6.11 --install-dir "$HOME/.local/bin"
```

El bootstrap instala solo el binario. No toca repositorios destino.


## Wrapper Bash

`scripts/install.sh`:

- delega en `lufy-ai install`;
- resuelve primero `tools/lufy-cli-go/bin/lufy-ai`;
- luego busca `lufy-ai` en `PATH`;
- reenvía flags de install;
- falla con instrucción de build si no encuentra binario;
- no descarga releases;
- no tiene fallback legacy.

Prueba:

```bash
cd tools/lufy-cli-go
mkdir -p bin
go build -o bin/lufy-ai ./cmd/lufy-ai
cd ../..
./scripts/install.sh --target "$(mktemp -d)" --dry-run --yes
```

## Validación esperada

Desde la raíz:

```bash
scripts/validate.sh
git diff --check origin/develop...HEAD
git diff --check
```

Desde `tools/lufy-cli-go/` para diagnóstico puntual:

```bash
go test ./...
go build ./cmd/lufy-ai
scripts/smoke-install.sh
```

`scripts/validate.sh` es preferido para cierre de bloques porque agrupa whitespace, Actions pinning, YAML, shell lint si existe, Go tests con coverage, `go vet` y build.
