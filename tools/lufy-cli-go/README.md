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
  internal/installer/        # plan y ejecución de install (slice inicial real)
  internal/platform/         # resolución portable (target, engram)
  internal/backup/           # backup/restore mínimo con manifest.json
  internal/verify/           # verify mínimo de estado instalado
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

## Cómo validar localmente

Ejecutar desde `tools/lufy-cli-go/`:

1. Tests unitarios:
   - `go test ./...`
2. Compilación del binario principal:
   - Recomendado: `mkdir -p bin && go build -o bin/lufy-ai ./cmd/lufy-ai`
   - Validación rápida tolerada: `go build ./cmd/lufy-ai` (genera `./lufy-ai`, ignorado como artefacto local)
3. Ejecución de smoke test dry-run (sin mutaciones):
   - `go run ./cmd/lufy-ai install --target . --dry-run --yes --no-engram`
4. Smoke end-to-end (temp dir):
   - `go build -o /tmp/lufy-ai ./cmd/lufy-ai`
   - `TMP_DIR="$(mktemp -d)"`
   - `cp /ruta/al/repo/AGENTS.md.template "$TMP_DIR/AGENTS.md.template"`
   - `/tmp/lufy-ai install --target "$TMP_DIR" --no-engram`
   - `/tmp/lufy-ai verify --target "$TMP_DIR" --no-engram`
   - `/tmp/lufy-ai backup --target "$TMP_DIR"`
   - `/tmp/lufy-ai restore --target "$TMP_DIR" --backup <backup-dir> --dry-run --yes`
   - `/tmp/lufy-ai restore --target "$TMP_DIR" --backup <backup-dir> --yes`

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
- `install` real (mínimo) crea `.lufy-ai/install-state.json` y crea `AGENTS.md` desde `AGENTS.md.template` cuando falta; en segunda ejecución reporta `skip` (idempotencia básica demostrable).
- Resolución de Engram portable por `PATH` (`exec.LookPath("engram")`), sin hardcode de `/opt/homebrew/bin/engram`.
- `verify` mínimo valida `install-state.json` parseable y reporta estado de Engram.
- `backup`/`restore` mínimos crean backup con `manifest.json`, soportan restore dry-run y restore real para archivos respaldados.

## Próximos pasos

1. Expandir instalación real a assets gestionados completos (`.opencode/...`, `tui.json`, etc.).
2. Endurecer idempotencia por contenido/hash y no solo por existencia.
3. Endurecer backup/restore con hashes por archivo y cobertura de más paths gestionados.
4. Implementar merge conservador de `opencode.json`.
