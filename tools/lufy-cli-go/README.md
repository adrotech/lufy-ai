# lufy-cli-go

CLI Go del instalador de `lufy-ai`, ubicada en carpeta dedicada para separar runtime/infra del resto de assets del kit. El binario instala el harness OpenCode/OpenSpec vigente, incluyendo `sdd-router`, templates T2/result, Review Workload Harness, policies y comandos gestionados.

## Propósito

- Reemplazar la lógica de `scripts/install.sh` con una implementación tipada y testeable.
- Mantener compatibilidad de entrada durante la transición: Bash queda como wrapper estricto de `lufy-ai install`, sin fallback legacy.

## Estructura

```text
tools/lufy-cli-go/
  cmd/lufy-ai/main.go        # entrypoint delgado
  internal/cli/              # parser/dispatch y códigos de salida
  internal/assets/           # catálogo de assets gestionados y hashing SHA-256
  internal/core/domain/      # contratos neutrales de harness, tiers, roles y methodology_by_tier
  internal/adapters/         # adapters iniciales de tool/metodología
  internal/instructions/     # role contracts, bindings de skills y renderer neutral
  internal/installer/        # plan y ejecución idempotente de install
  internal/platform/         # resolución portable (source, target, engram)
  internal/projectconfig/    # scanner stack-aware y .opencode/project.yaml
  internal/state/            # .lufy-ai/install-state.json versionado
  internal/backup/           # backup/restore multiasset con manifest.json
  internal/syncer/           # sync seguro de assets gestionados con hash/backup
  internal/verify/           # verify estructural de manifest, hashes y configs merge-managed
  internal/config/           # planificación/merge/validación de opencode.json
  internal/version/          # metadata de release y comando lufy-ai version
```

## Comandos locales

Ejecutar desde `tools/lufy-cli-go/`:

- Build:
  - Recomendado para generar el binario consumido por el wrapper: `mkdir -p bin && go build -o bin/lufy-ai ./cmd/lufy-ai`
  - Validación rápida tolerada: `go build ./cmd/lufy-ai` (genera `./lufy-ai`, ignorado como artefacto local)
- Test:
  - `go test ./...`
- Run (ejemplo dry-run seguro):
  - `go run ./cmd/lufy-ai install --target . --dry-run --yes --no-engram`
- Version local:
  - `go run ./cmd/lufy-ai version` (sin metadata de linker reporta `development build`)
- Init stack-aware:
  - `go run ./cmd/lufy-ai init --target <proyecto>`
- Sync (revisar plan sin escribir):
  - `go run ./cmd/lufy-ai sync --target <proyecto-instalado> --dry-run --yes --no-engram`

## Releases binarios y bootstrap

La CLI puede compilarse como artifact standalone con assets gestionados embebidos vía `go:embed`; esto permite que `lufy-ai install --target <dir>` funcione desde un binario distribuido sin leer el checkout fuente. Si el binario se ejecuta dentro de un checkout válido, puede usar assets locales para desarrollo; fuera de ese contexto usa el catálogo embebido.

El flujo de publicación está preparado en `.github/workflows/release.yml` y requiere un contexto autorizado:

- `develop` es la rama normal de integración para cambios de la CLI;
- `main` es la rama estable/productiva y recibe promociones desde `develop`;
- tags `v*` creados sobre commits alcanzables desde `origin/main` construyen artifacts versionados y pueden publicar GitHub Release;
- tags `v*` sobre commits no promovidos a `main` fallan antes de publicar assets;
- mientras no exista un tag/release `v*`, no hay release pública consumible por usuarios.

Scripts relacionados:

| Script | Uso |
| --- | --- |
| `scripts/build-release-artifacts.sh <version>` | Construye artifacts determinísticos para darwin/linux/windows soportados e inyecta `Version`, `Commit` y `BuildDate` con linker flags. |
| `scripts/smoke-release-artifacts.sh` | Genera fixtures locales, recalcula checksums, ejecuta `lufy-ai version`, `install --dry-run`, instalación temporal y `verify` con el artifact del runner. |
| `../../scripts/bootstrap.sh` | Bootstrap remoto seguro: detecta OS/arch, descarga artifact/checksums de una release, verifica SHA-256 e instala solo el binario en el directorio elegido. |
| `scripts/smoke-bootstrap.sh` | Valida bootstrap con fixtures `file://`, incluyendo dry-run, checksum correcto y bloqueo por checksum incorrecto. |

El bootstrap soporta `--version vX.Y.Z` o `LUFY_AI_VERSION`; `--version latest` existe como conveniencia no reproducible. Para automatización se recomienda siempre una versión fija. También soporta `--install-dir`/`LUFY_AI_INSTALL_DIR` y falla con mensaje accionable si el destino no es escribible. Si el directorio instalado no está en `PATH`, muestra instrucciones para bash/zsh y fish sin modificar archivos de shell automáticamente. No ejecuta comandos destructivos como `lufy-ai install` contra un target de proyecto. Ver la guía de usuario en [`../../docs/installation.md`](../../docs/installation.md).

## Comandos de usuario

La CLI expone estos comandos en el slice actual:

| Comando | Propósito | Flags principales |
| --- | --- | --- |
| `lufy-ai init` | Genera `.opencode/project.yaml` con stacks, comandos y reglas editables detectadas del repo destino. | `--target`, `--force`, `--rescan` |
| `lufy-ai install` | Instala assets gestionados, escribe estado con SHA-256, mergea `opencode.json` de forma conservadora y evita sobrescribir drift local. | `--target`, `--scope`, `--tool`, `--methodology-tier`, `--dry-run`, `--yes`, `--no-engram`, `--backup` |
| `lufy-ai verify` | Verificador canónico de instalación: valida categorías críticas, `.lufy-ai/install-state.json`, manifest, existencia de assets gestionados, hashes SHA-256 registrados y estructura merge-managed de `opencode.json`. | `--target`, `--scope`, `--tool`, `--no-engram`, `--json`, `--quiet`, `--verbose`, `--deep` |
| `lufy-ai backup` | Captura assets gestionados en `.lufy-ai/backups/<timestamp>/manifest.json`. | `--target` |
| `lufy-ai restore` | Restaura desde un backup validando `targetRoot`, paths seguros y hashes. | `--target`, `--backup`, `--dry-run`, `--yes` |
| `lufy-ai sync` | Reaplica assets gestionados cuando el source cambió y el target no tiene drift local; aplica `merge-json` para `opencode.json` cuando corresponde. | `--target`, `--scope`, `--tool`, `--dry-run`, `--yes`, `--no-engram` |
| `lufy-ai version` | Muestra versión semántica, commit, fecha de build, GOOS y GOARCH; los builds sin metadata se marcan como `development build`. | n/a |

`lufy-ai init` detecta stacks y genera configuración local editable, pero no instala templates por stack ni cambia todavía el comportamiento de agentes consumidores. Los templates instalables actuales siguen siendo templates de proceso del harness: `.opencode/templates/sdd-lite.md` y `.opencode/templates/result-contract.md`.

### Selección de harness

El adapter escribible actual es `opencode`. Estos comandos son equivalentes:

```bash
lufy-ai install --target <repo> --yes --no-engram
lufy-ai install --target <repo> --tool opencode --yes --no-engram
```

`install` acepta overrides repetibles de metodología por tier:

```bash
lufy-ai install --target <repo> --methodology-tier T3:none --yes --no-engram
lufy-ai install --target <repo> --methodology-tier T2:openspec/lite --methodology-tier T3:openspec/full --yes --no-engram
```

El parser bloquea `--tool codex`, `--tool claude-code`, `--methodology-tier T1:none`, `--methodology-tier T2:none` y `lufy-sdd` para comandos mutantes. `codex` existe en el registry solo como adapter dry-run: expone capabilities conservadoras y preview conceptual para `AGENTS.md`, pero no instala assets. `verify --tool opencode` valida que el manifest instalado use el adapter esperado; `status --json` y `verify --json` exponen `tool`, `schemaVersion` y `methodologyByTier`.

### `.opencode/project.yaml`

El archivo generado por `lufy-ai init` es configuración user-managed del repositorio destino. No se registra como asset completo en `.lufy-ai/install-state.json` ni se sincroniza por hash como parte de `install`/`sync`.

Comportamiento:

- `lufy-ai init --target <repo>` crea `.opencode/project.yaml` si no existe.
- Si el archivo existe, el comando falla sin sobrescribir.
- `--force` reemplaza el archivo con una detección nueva.
- `--rescan` compara la configuración existente con la evidencia actual, reporta drift por stack/tooling/CI/stale, fusiona solo campos detectados seguros y preserva overrides como `coverage_threshold`, `anti_patterns`, `workflow_limits` y campos desconocidos; si detecta `loc_budget` o `delivery_strategy` top-level los reporta como legacy no canónico; si no hay drift, no reescribe el archivo.
- Stacks soportados v1: Go, JavaScript/TypeScript con frameworks React/Next/Remix/Vue/Svelte, Python y Java/Kotlin.
- Stacks conocidos no soportados, como Rust, se emiten con `supported: false` y placeholders editables.

## Assets gestionados, SHA-256 e idempotencia

`install`, `verify` y `sync` consumen el catálogo de assets gestionados y el estado `.lufy-ai/install-state.json`. El catálogo incluye `.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/templates`, `.opencode/policies`, `.opencode/plugins`, `.opencode/agent-observatory`, `.opencode/README.md`, `AGENTS.md`, `tui.json` y `openspec`. `lufy-ai verify` es el verificador canónico; no existe ni se planea un `scripts/verify-install.sh` paralelo. El estado usa schema v2: registra `tool`, `methodologyByTier` y ownership por asset (`tool`, `methodology`, `component`, `scope`) además de hashes SHA-256 de source/target para distinguir estos casos:

- `skip`: el target ya coincide con el estado gestionado.
- `create`: el asset gestionado aún no existe y puede crearse.
- `update-managed`: el source cambió, el target sigue sin drift local y se puede actualizar con backup previo.
- `conflict`: existe contenido sin estado previo o el target no coincide con el último hash gestionado; la CLI no sobrescribe aunque `--yes` esté presente.

Las escrituras bloquean paths relativos inseguros y symlinks en rutas gestionadas para evitar escapes fuera del target.

### `opencode.json` como `merge-json`

`opencode.json` no se trata como asset completo gestionado por hash porque puede contener proveedores, modelos, MCPs o claves locales del proyecto destino. En `install` y `sync`, la CLI usa una estrategia especial `merge-json`:

- crea el archivo cuando falta con estructura OpenCode mínima;
- preserva claves desconocidas existentes del usuario;
- agrega/actualiza solo claves gestionadas por `lufy-ai` (`$schema`, `plugin` e integración Engram cuando aplica);
- falla sin sobrescribir si el JSON existente es inválido, para que el usuario lo corrija o respalde explícitamente;
- no registra `opencode.json` en `.lufy-ai/install-state.json` como asset completo con `targetSHA256`.

`lufy-ai verify` valida que `opencode.json` sea JSON parseable y contenga la estructura merge-managed mínima, pero no exige una entrada de hash para ese archivo en el manifest de instalación.

## Sync de assets gestionados

`lufy-ai sync` reaplica assets del catálogo gestionado sobre un target ya instalado usando `.lufy-ai/install-state.json` y hashes SHA-256. Está diseñado para actualizar solo archivos previamente gestionados y sin drift local.

- `--target <dir>` apunta al proyecto instalado; por defecto usa `.`.
- `--dry-run` usa el planner real y no crea backups, no copia archivos ni escribe estado.
- `--yes` es obligatorio para aplicar mutaciones reales (`backup` + `update-managed`).
- `--no-engram` mantiene el flujo portable sin requerir Engram instalado.
- Archivos no gestionados, drift local, estado ausente/corrupto y symlinks/escapes de path bloquean la mutación.
- `opencode.json` se planifica como `merge-json`: puede requerir backup si existe y será mergeado, pero no aparece como `copy`/`update-managed` ni se registra por hash completo.
- Antes de actualizar assets gestionados, sync crea backup bajo `.lufy-ai/backups/<timestamp>/manifest.json` con causa `sync`; si falla después del backup, el error incluye guía de `restore`.

## Cómo validar localmente

El gate mínimo de CI para esta CLI no usa comandos Node/TS de raíz (`npm test`, `tsc`, etc.) ni requiere Engram instalado. Los smokes pasan `--no-engram` para mantener la validación portable en GitHub Actions y en entornos locales.

Ejecutar desde `tools/lufy-cli-go/`:

1. Tests unitarios:
   - `go test ./...`
2. Compilación del binario principal:
   - Recomendado: `mkdir -p bin && go build -o bin/lufy-ai ./cmd/lufy-ai`
   - Validación rápida tolerada: `go build ./cmd/lufy-ai` (genera `./lufy-ai`, ignorado como artefacto local)
3. Smoke end-to-end de la CLI Go contra directorios temporales:
    - `scripts/smoke-install.sh`
4. Smokes de release/bootstrap con fixtures locales, sin descargar internet:
   - `scripts/smoke-release-artifacts.sh`
   - `scripts/smoke-bootstrap.sh`

Desde la raíz del repo, ejecutar además el smoke del wrapper estricto:

```bash
tools/lufy-cli-go/scripts/smoke-wrapper.sh
```

El smoke de CLI cubre:

- `install --dry-run --yes --no-engram` sin mutaciones.
- `install --yes --no-engram` real y `verify --no-engram`.
- Idempotencia básica con segunda instalación y comparación de hashes de `.lufy-ai/install-state.json` y `AGENTS.md`.
- Merge conservador de `opencode.json`, preservando configuración local y excluyéndolo del estado por hash completo.
- `backup`, `restore --dry-run --yes`, restore real con `--yes` y errores accionables cuando `install`/`restore` se ejecutan sin `--yes` ante mutaciones reales.

El workflow `.github/workflows/go-cli-install.yml` existe en esta rama y ejecuta el set mínimo en PRs/pushes a `develop` y `main`: tests Go, build Go, binario local `tools/lufy-cli-go/bin/lufy-ai`, smoke de CLI, smoke de `scripts/install.sh`, sanity OpenSpec con `openspec list --json` cuando la CLI `openspec` esté disponible, y `git diff --check`.

El workflow `.github/workflows/release.yml` agrega el gate de artifacts versionados: tests Go, build, artifacts/checksums, smoke release, smoke bootstrap, verificación de checksums y publicación de assets solo desde tag `v*` autorizado y alcanzable desde `origin/main`.

`shellcheck scripts/install.sh` es una validación útil cuando `shellcheck` existe localmente, pero no forma parte de este gate mínimo inicial para no añadir una dependencia extra al runner; el wrapper queda cubierto funcionalmente por `tools/lufy-cli-go/scripts/smoke-wrapper.sh`.

Si `go` no está instalado en el entorno, estos pasos quedan pendientes y deben correrse en una máquina con toolchain Go disponible.

## Integración con `scripts/install.sh`

- Estado actual: el wrapper Bash solo ejecuta `lufy-ai install`.
- Orden de resolución: primero `tools/lufy-cli-go/bin/lufy-ai`, luego `lufy-ai` en `PATH`.
- Si no encuentra binario, falla con una instrucción explícita de build local:
  - `cd tools/lufy-cli-go && mkdir -p bin && go build -o bin/lufy-ai ./cmd/lufy-ai`
- Contrato preservado: `scripts/install.sh [target-project-dir]`, mapeado a `lufy-ai install --target <target-project-dir>`.
- Flags reenviados: `--target`, `--scope`, `--tool`, `--methodology-tier`, `--dry-run`, `--yes`, `--no-engram`, `--backup`.
- No existe fallback legacy de copia, detección de stack, Engram o `copy_files` en Bash.
- Tampoco existe fallback remoto: las descargas versionadas viven en `scripts/bootstrap.sh`, no en el wrapper local.

Para probar el wrapper desde la raíz del repo:

```bash
cd tools/lufy-cli-go && mkdir -p bin && go build -o bin/lufy-ai ./cmd/lufy-ai
cd ../..
./scripts/install.sh --target "$(mktemp -d)" --dry-run --yes --no-engram
```

## Estado actual del slice

- Implementado parser base con comandos `install`, `verify`, `backup`, `restore`, `sync` y `version`.
- `install --dry-run` construye plan e imprime resultado sin mutaciones.
- `install` real copia assets gestionados del catálogo (`.opencode`, `AGENTS.md`, `tui.json`, `openspec` base), escribe `.lufy-ai/install-state.json` con hashes SHA-256 y en segunda ejecución reporta `skip` sin reescribir archivos ni estado.
- `opencode.json` se crea o mergea con estrategia `merge-json`, preserva claves desconocidas, falla ante JSON inválido y queda fuera del manifest de assets completos por hash.
- Si un archivo gestionado cambió upstream y el target no tiene drift local, `install` crea backup bajo `.lufy-ai/backups/<timestamp>/` antes de `update-managed`.
- Si un archivo existe sin estado previo o su hash actual no coincide con el último hash gestionado, `install` reporta `conflict` y no sobrescribe aunque `--yes` esté presente.
- Resolución de Engram portable por `PATH` (`exec.LookPath("engram")`), sin hardcode de `/opt/homebrew/bin/engram`.
- `verify` valida `install-state.json`, categorías críticas (`.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/plugins`, `.opencode/policies`), archivos críticos (`.opencode/plugins/agent-observatory.tsx`, `AGENTS.md`, `tui.json`, `openspec/config.yaml`), manifest y hashes de assets listados; reporta Engram como warning no bloqueante u omitido con `--no-engram`.
- `sync` puede aplicar `merge-json` a `opencode.json` sin copiarlo como asset completo ni registrarlo con SHA-256, manteniendo backup previo si el archivo existente será modificado.
- `backup`/`restore` respaldan múltiples assets gestionados con `manifest.json`, hashes, tamaño, timestamp y validación de `targetRoot`.
- Antes de un `restore` real que sobrescribe archivos existentes, la CLI crea un backup de recovery `pre-restore-recovery`; si la restauración falla parcialmente, el error incluye la ruta de ese backup.
- `restore` rechaza manifests de otro target o con paths que escapan del target; `verify` falla si el manifest está corrupto o si `targetRoot` indica que la instalación fue movida.
- Las escrituras rechazan paths relativos inseguros y symlinks en rutas gestionadas para evitar escapes fuera del target.
- Los artifacts standalone pueden instalar assets embebidos cuando no se encuentra un checkout fuente válido; el checkout fuente ahora requiere el marcador adicional `tools/lufy-cli-go/go.mod` para evitar falsos positivos en repositorios destino ya instalados.
- Cuando cambian assets del harness, genera un nuevo installer local con `mkdir -p bin && go build -o bin/lufy-ai ./cmd/lufy-ai` desde `tools/lufy-cli-go/` y valida con `scripts/validate.sh` desde la raíz.
- La publicación pública requiere promover a `main` y crear un tag `v*` sobre un commit alcanzable desde `origin/main`; esta rama no implica que exista ya una release publicada.

## Próximos pasos

1. Ampliar la cobertura de docs/smokes cuando se agreguen nuevas claves gestionadas al contrato `merge-json` de `opencode.json`.
