## Why

La CI actual cubre tests/build/smokes mínimos, pero aún no mide cobertura, linting, shell scripts ni portabilidad multi-OS de forma explícita. Antes de seguir ampliando distribución y UX conviene elevar la evidencia automática sin inventar toolchains raíz ni convertir la validación en un bloqueo frágil.

## What Changes

- Agregar coverage Go con `-coverprofile` y un threshold inicial bajo/realista para evitar regresiones silenciosas.
- Incorporar lint Go mínimo con `golangci-lint` o una alternativa equivalente documentada, manteniendo configuración acotada.
- Incorporar ShellCheck para scripts shell críticos (`scripts/*.sh`, `tools/lufy-cli-go/scripts/*.sh`) cuando esté disponible en CI.
- Ampliar CI a una matriz de plataformas para tests/build Go y smokes compatibles, cubriendo Linux, macOS y Windows cuando aplique.
- Definir un gate E2E post-release real contra artifacts publicados como validación separada de releases estables.
- Añadir tests de salida golden para planes y un test runtime de `cmd/lufy-ai/main.go` cuando el alcance lo permita.
- Mantener `scripts/validate.sh` como entrypoint local agrupado, sin comandos Node/TS globales inventados.

## Capabilities

### New Capabilities
- `ci-quality-gates`: gates de coverage, lint, shellcheck, matriz multi-OS, E2E post-release y pruebas de regresión de output/runtime.

### Modified Capabilities
- `go-cli-install-ci`: la CI mínima del instalador Go se amplía para integrar quality gates sin perder reproducibilidad local ni portabilidad.

## Impact

- `.github/workflows/go-cli-install.yml`: matriz multi-OS, jobs o pasos adicionales de coverage/lint/shellcheck según disponibilidad.
- Posibles nuevos workflows de release/E2E si se separa el gate post-release.
- `scripts/validate.sh` y scripts auxiliares de validación local.
- `tools/lufy-cli-go/`: posibles tests golden/runtime y configuración lint/coverage.
- Documentación de comandos reales y limitaciones cuando herramientas opcionales no estén disponibles localmente.
