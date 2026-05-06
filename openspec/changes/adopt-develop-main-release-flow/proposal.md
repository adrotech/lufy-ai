## Why

El repositorio necesita una política explícita para separar trabajo diario, integración y producción antes de publicar releases estables. Adoptar `develop` como base de integración y `main` como rama productiva reduce ambigüedad en PRs, protección de ramas y generación de releases.

## What Changes

- Documentar el flujo canónico `feature/*` → `develop` → `main` → tag `v*` → GitHub Release.
- Actualizar la política de delivery para que `develop` sea la base por defecto de cambios normales y `main` quede reservada para promoción/release u hotfix explícitamente autorizado.
- Ajustar workflows para validar PRs/pushes contra `develop` y `main` y mantener releases solo desde tags `v*`.
- Agregar un guard en release para impedir publicación de assets si el tag `v*` no apunta a un commit alcanzable desde `main`.
- Actualizar documentación pública y operativa para explicar que el bootstrap estable consume releases publicadas; no se implementa canal snapshot/main en este cambio.
- Documentar pasos manuales o automatizables de configuración GitHub: default branch `develop` y protección de `develop`/`main`.

## Capabilities

### New Capabilities
- `release-branch-flow`: cubre reglas de ramas, PRs, promoción a producción, tags `v*`, releases y configuración GitHub esperada.

### Modified Capabilities
- `go-cli-install-ci`: cambia los requisitos de CI para ejecutar validación en PR/push hacia `develop` y `main`.
- `versioned-binary-distribution`: restringe la publicación estable a tags `v*` alcanzables desde `main`.
- `current-state-documentation`: alinea README, getting started, README de la CLI y roadmap con el modelo `develop`/`main` y el bootstrap estable basado en releases publicadas.

## Impact

- Archivos de política y guía: `.opencode/policies/delivery.md`, `AGENTS.md`.
- Workflows GitHub Actions: `.github/workflows/go-cli-install.yml`, `.github/workflows/release.yml`.
- Documentación: `README.md`, `docs/getting-started.md`, `tools/lufy-cli-go/README.md`, `docs/roadmap.md` y documentación nueva de configuración GitHub si aplica.
- Sin cambios de runtime, API pública, esquema de datos ni fallback legacy del instalador.
