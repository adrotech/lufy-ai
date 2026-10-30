# Validación local y gates de CI

Este repositorio valida el producto real desde la CLI Go en `tools/lufy-cli-go`. No existe suite Node/TS global en la raíz.

## Validación local agrupada

Ejecutar desde la raíz:

```bash
scripts/validate.sh
```

El comando ejecuta:

- whitespace check con rango/base de PR;
- pinning de GitHub Actions en workflows;
- consistencia de `RELEASE_VERSION`, referencias documentales y pruebas del guard de versiones;
- regresión de metadata: los comentarios de checksums no se incluyen como artifacts en el provenance;
- sintaxis YAML de workflows mediante helper Go local;
- ShellCheck para scripts versionados cuando `shellcheck` está instalado;
- regresiones Python offline, paridad del payload y tests Go con coverage y threshold objetivo;
- `go vet ./...`;
- build de `./cmd/lufy-ai`.

Si `shellcheck` no está instalado localmente, el gate lo reporta explícitamente como omitido. En CI Ubuntu se instala antes de ejecutar `scripts/validate.sh`, por lo que allí el shell lint es obligatorio.

## Coverage

El threshold objetivo es `80.0%`, medido con `go test ./... -coverprofile=<archivo>` y agregado por `go tool cover`.

El flujo inline conserva el umbral de 80%. Las regresiones Python son offline: no invocan modelos ni ejecutan ensayos live. El script no comprueba formato Go con gofmt; si cambia Go, revisá su formato aparte.

## Multi-plataforma

El workflow `Go CLI install` separa:

- `quality`: quality gates completos en Ubuntu;
- `platform-go`: `go test ./...` y `go build ./cmd/lufy-ai` en Linux, macOS y Windows;
- `go-cli-install`: smokes POSIX/installer completos en Ubuntu.

Los smokes de instalación y wrapper permanecen en Ubuntu porque dependen de scripts Bash y rutas POSIX del flujo actual.

## E2E post-release

El workflow `release.yml` valida el artifact publicado como parte del pipeline automático de release. Después de publicar los assets en GitHub Releases, descarga el artifact correspondiente al runner, verifica checksums y ejecuta `version`, `install --dry-run`, `install` y `verify` contra un target temporal.

El workflow `post-release-e2e.yml` queda disponible como validación manual o backstop para artifacts ya publicados de un tag `v*`. No corre en PRs normales porque depende de GitHub Releases y red externa.

Ejecución manual local, si existe el release publicado y `gh` está autenticado:

```bash
tools/lufy-cli-go/scripts/e2e-release-artifact.sh vX.Y.Z
```
