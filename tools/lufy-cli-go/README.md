# lufy-cli-go

CLI Go del instalador de `lufy-ai`, ubicada en carpeta dedicada para separar runtime/infra del resto de assets del kit.

## Propósito

- Reemplazar la lógica de `scripts/install.sh` con una implementación tipada y testeable.
- Mantener compatibilidad de entrada durante la transición: Bash queda como wrapper estricto de `lufy-ai install`, sin fallback legacy.

## Estructura

```text
tools/lufy-cli-go/
  cmd/lufy-ai/main.go        # entrypoint delgado
  internal/cli/              # parser/dispatch y códigos de salida
  internal/assets/           # catálogo de assets gestionados y hashing SHA-256
  internal/installer/        # plan y ejecución idempotente de install
  internal/platform/         # resolución portable (source, target, engram)
  internal/state/            # .lufy-ai/install-state.json versionado
  internal/backup/           # backup/restore multiasset con manifest.json
  internal/syncer/           # sync seguro de assets gestionados con hash/backup
  internal/verify/           # verify estructural de manifest y hashes
  internal/config/           # placeholder
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
- Sync (revisar plan sin escribir):
  - `go run ./cmd/lufy-ai sync --target <proyecto-instalado> --dry-run --yes --no-engram`

## Comandos de usuario

La CLI expone estos comandos en el slice actual:

| Comando | Propósito | Flags principales |
| --- | --- | --- |
| `lufy-ai install` | Instala assets gestionados, escribe estado con SHA-256 y evita sobrescribir drift local. | `--target`, `--dry-run`, `--yes`, `--no-engram`, `--backup` |
| `lufy-ai verify` | Verificador canónico de instalación: valida categorías críticas, `.lufy-ai/install-state.json`, manifest, existencia de assets gestionados y hashes SHA-256 registrados. | `--target`, `--no-engram` |
| `lufy-ai backup` | Captura assets gestionados en `.lufy-ai/backups/<timestamp>/manifest.json`. | `--target` |
| `lufy-ai restore` | Restaura desde un backup validando `targetRoot`, paths seguros y hashes. | `--target`, `--backup`, `--dry-run`, `--yes` |
| `lufy-ai sync` | Reaplica assets gestionados cuando el source cambió y el target no tiene drift local. | `--target`, `--dry-run`, `--yes`, `--no-engram` |

No hay comandos de detección de stack ni instalación de templates por stack en esta CLI.

## Assets gestionados, SHA-256 e idempotencia

`install`, `verify` y `sync` consumen el catálogo de assets gestionados y el estado `.lufy-ai/install-state.json`. `lufy-ai verify` es el verificador canónico; no existe ni se planea un `scripts/verify-install.sh` paralelo. Cada asset registrado conserva hashes SHA-256 de source/target para distinguir estos casos:

- `skip`: el target ya coincide con el estado gestionado.
- `create`: el asset gestionado aún no existe y puede crearse.
- `update-managed`: el source cambió, el target sigue sin drift local y se puede actualizar con backup previo.
- `conflict`: existe contenido sin estado previo o el target no coincide con el último hash gestionado; la CLI no sobrescribe aunque `--yes` esté presente.

Las escrituras bloquean paths relativos inseguros y symlinks en rutas gestionadas para evitar escapes fuera del target.

## Sync de assets gestionados

`lufy-ai sync` reaplica assets del catálogo gestionado sobre un target ya instalado usando `.lufy-ai/install-state.json` y hashes SHA-256. Está diseñado para actualizar solo archivos previamente gestionados y sin drift local.

- `--target <dir>` apunta al proyecto instalado; por defecto usa `.`.
- `--dry-run` usa el planner real y no crea backups, no copia archivos ni escribe estado.
- `--yes` es obligatorio para aplicar mutaciones reales (`backup` + `update-managed`).
- `--no-engram` mantiene el flujo portable sin requerir Engram instalado.
- Archivos no gestionados, drift local, estado ausente/corrupto y symlinks/escapes de path bloquean la mutación.
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

Desde la raíz del repo, ejecutar además el smoke del wrapper estricto:

```bash
tools/lufy-cli-go/scripts/smoke-wrapper.sh
```

El smoke de CLI cubre:

- `install --dry-run --yes --no-engram` sin mutaciones.
- `install --yes --no-engram` real y `verify --no-engram`.
- Idempotencia básica con segunda instalación y comparación de hashes de `.lufy-ai/install-state.json` y `AGENTS.md`.
- `backup`, `restore --dry-run --yes`, restore real con `--yes` y errores accionables cuando `install`/`restore` se ejecutan sin `--yes` ante mutaciones reales.

El workflow `.github/workflows/go-cli-install.yml` existe en esta rama y ejecuta el mismo set mínimo: tests Go, build Go, binario local `tools/lufy-cli-go/bin/lufy-ai`, smoke de CLI, smoke de `scripts/install.sh`, sanity OpenSpec con `openspec list --json` cuando la CLI `openspec` esté disponible, y `git diff --check`. Esto documenta el workflow presente, no el archive ni delivery final de ninguna proposal OpenSpec.

`shellcheck scripts/install.sh` es una validación útil cuando `shellcheck` existe localmente, pero no forma parte de este gate mínimo inicial para no añadir una dependencia extra al runner; el wrapper queda cubierto funcionalmente por `tools/lufy-cli-go/scripts/smoke-wrapper.sh`.

Si `go` no está instalado en el entorno, estos pasos quedan pendientes y deben correrse en una máquina con toolchain Go disponible.

## Integración con `scripts/install.sh`

- Estado actual: el wrapper Bash solo ejecuta `lufy-ai install`.
- Orden de resolución: primero `tools/lufy-cli-go/bin/lufy-ai`, luego `lufy-ai` en `PATH`.
- Si no encuentra binario, falla con una instrucción explícita de build local:
  - `cd tools/lufy-cli-go && mkdir -p bin && go build -o bin/lufy-ai ./cmd/lufy-ai`
- Contrato preservado: `scripts/install.sh [target-project-dir]`, mapeado a `lufy-ai install --target <target-project-dir>`.
- Flags reenviados: `--target`, `--dry-run`, `--yes`, `--no-engram`, `--backup`.
- No existe fallback legacy de copia, detección de stack, Engram o `copy_files` en Bash.

Para probar el wrapper desde la raíz del repo:

```bash
cd tools/lufy-cli-go && mkdir -p bin && go build -o bin/lufy-ai ./cmd/lufy-ai
cd ../..
./scripts/install.sh --target "$(mktemp -d)" --dry-run --yes --no-engram
```

## Estado actual del slice

- Implementado parser base con comandos `install`, `verify`, `backup` y `restore`.
- `install --dry-run` construye plan e imprime resultado sin mutaciones.
- `install` real copia assets gestionados del catálogo (`.opencode`, `AGENTS.md`, `tui.json`, `openspec` base), escribe `.lufy-ai/install-state.json` con hashes SHA-256 y en segunda ejecución reporta `skip` sin reescribir archivos ni estado.
- Si un archivo gestionado cambió upstream y el target no tiene drift local, `install` crea backup bajo `.lufy-ai/backups/<timestamp>/` antes de `update-managed`.
- Si un archivo existe sin estado previo o su hash actual no coincide con el último hash gestionado, `install` reporta `conflict` y no sobrescribe aunque `--yes` esté presente.
- Resolución de Engram portable por `PATH` (`exec.LookPath("engram")`), sin hardcode de `/opt/homebrew/bin/engram`.
- `verify` valida `install-state.json`, categorías críticas (`.opencode/agents`, `.opencode/commands`, `.opencode/skills`, `.opencode/plugins`, `.opencode/policies`), archivos críticos (`.opencode/plugins/agent-observatory.tsx`, `AGENTS.md`, `tui.json`, `openspec/config.yaml`), manifest y hashes de assets listados; reporta Engram como warning no bloqueante u omitido con `--no-engram`.
- `backup`/`restore` respaldan múltiples assets gestionados con `manifest.json`, hashes, tamaño, timestamp y validación de `targetRoot`.
- Antes de un `restore` real que sobrescribe archivos existentes, la CLI crea un backup de recovery `pre-restore-recovery`; si la restauración falla parcialmente, el error incluye la ruta de ese backup.
- `restore` rechaza manifests de otro target o con paths que escapan del target; `verify` falla si el manifest está corrupto o si `targetRoot` indica que la instalación fue movida.
- Las escrituras rechazan paths relativos inseguros y symlinks en rutas gestionadas para evitar escapes fuera del target.

## Próximos pasos

1. Implementar merge conservador de `opencode.json` si se decide incluirlo como asset gestionado futuro.
