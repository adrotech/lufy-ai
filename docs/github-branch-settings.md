# Configuración GitHub de ramas

Esta guía documenta el estado esperado del remoto. No aplica cambios automáticamente; requiere un maintainer o `delivery` con autorización explícita.

## Ramas

- `develop`: default branch y base normal de PRs de trabajo (`feature/*`, `fix/*`, `chore/*` o equivalentes).
- `main`: rama productiva/estable para promociones `develop` → `main`, releases y hotfixes explícitamente autorizados.

## Default branch

1. En GitHub, ir a **Settings → Branches → Default branch**.
2. Cambiar la default branch a `develop`.
3. Confirmar que las reglas, workflows y documentación de onboarding apuntan a `develop` como base normal.

## Branch protection recomendada

Configurar reglas para `develop` y `main`:

- requerir PR antes de merge;
- requerir status checks relevantes antes de merge, incluyendo el workflow `.github/workflows/go-cli-install.yml` cuando aplique;
- bloquear force pushes;
- bloquear deletes;
- exigir ramas actualizadas antes de merge si el repositorio lo requiere;
- restringir quién puede pushear directo a `main` si el plan del equipo lo permite.

Para que el release automático funcione, permitir que GitHub Actions cree y pushee tags e invoque workflows desde `.github/workflows/auto-release-tag.yml`. El repositorio debe tener habilitado **Settings → Actions → General → Workflow permissions → Read and write permissions** o una configuración equivalente que permita `contents: write` y `actions: write` al `GITHUB_TOKEN` del workflow. No requiere permisos de escritura sobre PRs; el workflow usa `pull-requests: read`.

## Releases

- Los tags estables `v*` se crean automáticamente al mergear un PR hacia `main`: si no hay tags simples `vMAJOR.MINOR.PATCH`, el primer tag es `v0.1.0`; si ya existen, se incrementa `PATCH` sobre el mayor tag simple e ignora prereleases.
- El tag automático es anotado, apunta al merge commit del PR y no se sobrescribe si el tag calculado ya existe localmente o en `origin`; en ese caso el workflow reporta un no-op explícito.
- El workflow `.github/workflows/auto-release-tag.yml` no construye binarios ni publica GitHub Releases; pushea el tag y luego invoca explícitamente `.github/workflows/release.yml` mediante `workflow_dispatch` apuntando al tag creado. Si el tag ya existía, no dispara release duplicada.
- El workflow `.github/workflows/release.yml` conserva el trigger por push de tags `v*` para tags manuales/humanos y también acepta `workflow_dispatch` con input `tag`; siempre valida que el tag sea `v*` y que su commit sea alcanzable desde `origin/main` antes de publicar assets.
- No publicar GitHub Releases estables desde commits que existan únicamente en `develop`.

## Operaciones Git/GitHub

Cambios reales de settings, commits, pushes, PRs y tags quedan fuera del rol `implementer`. Usar `delivery` con autorización explícita o ejecutar manualmente como maintainer.
