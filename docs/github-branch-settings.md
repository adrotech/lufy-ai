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

## Releases

- Crear tags estables `v*` solo después de promover el commit a `main`.
- El workflow `.github/workflows/release.yml` valida que el commit del tag sea alcanzable desde `origin/main` antes de publicar assets.
- No publicar GitHub Releases estables desde commits que existan únicamente en `develop`.

## Operaciones Git/GitHub

Cambios reales de settings, commits, pushes, PRs y tags quedan fuera del rol `implementer`. Usar `delivery` con autorización explícita o ejecutar manualmente como maintainer.
