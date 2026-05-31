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
  internal/verify/           # verify estructural y deep checks
  internal/backup/           # backup/restore multiasset
  internal/config/           # merge conservador de opencode.json
  internal/projectconfig/    # init/rescan de .opencode/project.yaml
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
| `lufy-ai init` | Genera `.opencode/project.yaml` stack-aware y editable. | `--target`, `--force`, `--rescan` |
| `lufy-ai install` | Instala assets gestionados, mergea configs user-owned y escribe manifest SHA-256. | `--target`, `--scope`, `--tool`, `--methodology-tier`, `--dry-run`, `--yes`, `--no-engram`, `--backup` |
| `lufy-ai uninstall` | Remueve assets gestionados sin drift, crea backup, preserva user-owned y quita solo la referencia Lufy de `AGENTS.md`. | `--target`, `--dry-run`, `--yes`, `--keep-state` |
| `lufy-ai verify` | Valida manifest, hashes, estructura, JSON merge-managed y referencias críticas. | `--target`, `--scope`, `--tool`, `--no-engram`, `--json`, `--quiet`, `--verbose`, `--deep` |
| `lufy-ai status` | Resume instalación, drift, faltantes y errores. | `--target`, `--scope`, `--json`, `--verbose` |
| `lufy-ai sync` | Reaplica assets gestionados cuando el source cambió y el target no tiene drift local. | `--target`, `--scope`, `--tool`, `--dry-run`, `--yes`, `--no-engram` |
| `lufy-ai merge` | Reconcilia `.lufy-new` con edits locales usando ancestor seguro. | `--target` |
| `lufy-ai backup` | Captura assets gestionados en `.lufy-ai/backups/<timestamp>/manifest.json`. | `--target` |
| `lufy-ai restore` | Restaura desde backup validando target, paths seguros y hashes. | `--target`, `--backup`, `--dry-run`, `--yes`, `--list` |
| `lufy-ai upgrade` | Actualiza el binario a una versión fija con checksum. | `--to`, `--dry-run` |
| `lufy-ai version` | Muestra versión, commit, build date, GOOS y GOARCH. | n/a |

## Selección de harness

El adapter escribible actual es `opencode`.

```bash
lufy-ai install --target <repo> --yes --no-engram
lufy-ai install --target <repo> --tool opencode --yes --no-engram
```

Adapters no escribibles todavía:

- `codex`;
- `claude-code`.

Ambos existen como dry-run/preview para modelar capabilities y superficies futuras, pero los comandos mutantes los bloquean.

## Methodology por tier

`install` acepta overrides repetibles:

```bash
lufy-ai install --target <repo> --methodology-tier T3:none --yes --no-engram
lufy-ai install --target <repo> --methodology-tier T2:openspec/lite --methodology-tier T3:none --yes --no-engram
lufy-ai install --target <repo> --methodology-tier T2:lufy-sdd/lite --yes --no-engram
```

Reglas actuales:

- `openspec` puede instalar superficie full/lite;
- `lufy-sdd` instala `.lufy/sdd/` como superficie inicial;
- `none` se permite donde la policy lo habilita;
- `T1:none` y `T2:none` están bloqueados en comandos mutantes.

`verify --tool opencode`, `status --json` y `verify --json` exponen `tool`, `schemaVersion` y `methodologyByTier`.

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
- `.opencode/project.yaml`: creado por `init`, no sincronizado por hash.

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
lufy-ai install --target <repo> --yes --no-engram
lufy-ai verify --target <repo> --no-engram --quiet
```

## Sync

`sync` reaplica assets gestionados ya registrados.

- Requiere manifest existente.
- Requiere `--yes` para mutaciones reales.
- Crea backup antes de updates.
- Bloquea estado ausente/corrupto, drift local y paths inseguros.
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

`status` resume lo mismo con foco operativo y puede emitir JSON.

## `.opencode/project.yaml`

`lufy-ai init` crea configuración stack-aware user-managed.

Comportamiento:

- si no existe, crea `.opencode/project.yaml`;
- si existe, falla sin `--force`;
- `--force` reemplaza;
- `--rescan` preserva overrides y agrega evidencia nueva;
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
curl -fsSL https://raw.githubusercontent.com/adrotech/lufy-ai/v0.5.0/scripts/bootstrap.sh -o /tmp/lufy-bootstrap.sh
bash /tmp/lufy-bootstrap.sh --version v0.5.0 --install-dir "$HOME/.local/bin"
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
./scripts/install.sh --target "$(mktemp -d)" --dry-run --yes --no-engram
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
