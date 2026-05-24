# CI quality gates

Este repositorio valida el producto real desde la CLI Go en `tools/lufy-cli-go`. No existe suite Node/TS global en la raíz.

## Validación local agrupada

Ejecutar desde la raíz:

```bash
scripts/validate.sh
```

El comando ejecuta:

- whitespace check con rango/base de PR;
- pinning de GitHub Actions en workflows;
- sintaxis YAML de workflows;
- ShellCheck para scripts versionados cuando `shellcheck` está instalado;
- tests Go con coverage de módulo completo (`-coverpkg=./...`) y threshold objetivo;
- `go vet ./...`;
- build de `./cmd/lufy-ai`.

Si `shellcheck` no está instalado localmente, el gate lo reporta explícitamente como omitido. En CI Ubuntu se instala antes de ejecutar `scripts/validate.sh`, por lo que allí el shell lint es obligatorio.

## Coverage

El threshold objetivo es `80.0%`, medido con `go test ./... -coverpkg=./...` para incluir cobertura ejercitada por tests de integración entre paquetes del módulo.

Puede ajustarse temporalmente con:

```bash
LUFY_AI_COVERAGE_MIN=75.0 scripts/quality-go.sh
```

## Multi-plataforma

El workflow `Go CLI install` separa:

- `quality`: quality gates completos en Ubuntu;
- `platform-go`: `go test ./...` y `go build ./cmd/lufy-ai` en Linux, macOS y Windows;
- `go-cli-install`: smokes POSIX/installer completos en Ubuntu.

Los smokes de instalación y wrapper permanecen en Ubuntu porque dependen de scripts Bash y rutas POSIX del flujo actual.

## E2E post-release

El workflow `post-release-e2e.yml` valida artifacts ya publicados para un tag `v*`. No corre en PRs normales porque depende de GitHub Releases y red externa.

Ejecución manual local, si existe el release publicado y `gh` está autenticado:

```bash
tools/lufy-cli-go/scripts/e2e-release-artifact.sh vX.Y.Z
```
