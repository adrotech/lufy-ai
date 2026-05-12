## Why

El flujo de release ya construye y publica binarios desde tags `v*`, pero todavía requiere que un maintainer cree manualmente cada tag estable después de promover a `main`. Automatizar el tag patch al mergear un PR hacia `main` reduce fricción y mantiene la política existente de publicar releases solo desde commits alcanzables desde `origin/main`.

## What Changes

- Agregar un workflow de GitHub Actions que se ejecuta al cerrar PRs hacia `main` y solo actúa cuando `merged == true`.
- Calcular el siguiente tag patch semver simple a partir del último tag `vMAJOR.MINOR.PATCH`; si no existe ningún tag válido, crear `v0.1.0`.
- Crear y pushear un tag anotado sobre el merge commit final del PR e invocar explícitamente `.github/workflows/release.yml` mediante `workflow_dispatch` apuntando a ese tag.
- Mantener el build y publicación de binarios exclusivamente en `release.yml`; el nuevo workflow solo crea el tag y solicita el workflow de release mediante un mecanismo soportado por Actions.
- Documentar el comportamiento automático, la idempotencia elegida y los límites de seguridad.

## Capabilities

### New Capabilities

### Modified Capabilities
- `versioned-binary-distribution`: añade creación automática de tags patch estables al mergear PRs hacia `main`, conservando la publicación de artefactos en el workflow de release existente.

## Impact

- Nuevo workflow `.github/workflows/auto-release-tag.yml` con permisos mínimos `actions: write`, `contents: write` y `pull-requests: read`.
- Actualización documental en `README.md` y `docs/github-branch-settings.md` sobre releases automáticos desde merges a `main`.
- Delta OpenSpec para `versioned-binary-distribution`.
