## Why

El instalador Go ya concentra la lógica real de instalación, verificación, backup y restore, pero no existe un gate automatizado que proteja esos flujos antes de mergear cambios. El roadmap prioriza `RM-006`/`RM-007`: validación continua mínima antes de avanzar hacia `sync`, templates o productización.

## What Changes

- Añadir un workflow de GitHub Actions mínimo para validar `tools/lufy-cli-go` sin asumir tooling Node/TS en la raíz.
- Ejecutar validación Go básica: `go test ./...` y `go build ./cmd/lufy-ai` desde `tools/lufy-cli-go/`.
- Ejecutar un smoke de instalación en directorio temporal que cubra dry-run sin mutaciones, install real, `verify`, idempotencia básica, `backup` y `restore`.
- Validar el wrapper estricto `scripts/install.sh` contra un target temporal usando el binario Go local construido.
- Documentar cómo correr localmente el mismo set mínimo de validaciones.
- Mantener fuera de alcance comandos futuros como `sync`, `update`, publicación de releases o empaquetado.

## Capabilities

### New Capabilities
- `go-cli-install-ci`: validación continua mínima para compilar, probar y smoke-testear el instalador Go y su wrapper Bash.

### Modified Capabilities
- `go-cli-installer`: añade expectativa contractual de que la CLI Go sea validable en CI mediante tests, build y smoke de instalación temporal.

## Impact

- `.github/workflows/`: nuevo workflow de CI mínimo.
- `tools/lufy-cli-go/`: posible script/helper de smoke si evita duplicar lógica compleja en YAML.
- `scripts/install.sh`: validado por CI sin reintroducir fallback legacy.
- `README.md` o `tools/lufy-cli-go/README.md`: documentación breve de validación local.
- OpenSpec: nueva capacidad `go-cli-install-ci` y delta de `go-cli-installer` para formalizar el gate.
