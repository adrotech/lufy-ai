# release-branch-flow Specification

## Purpose
Define the canonical branch and release flow where normal integration targets `develop`, production promotion targets `main`, and stable `v*` tags come from `main`.

## Requirements
### Requirement: Flujo canónico develop-main-release
El repositorio SHALL usar `develop` como rama normal de integración y `main` como rama productiva/estable.

#### Scenario: PR normal hacia develop
- **WHEN** una rama de trabajo como `feature/*`, `fix/*`, `chore/*` o equivalente está lista para integración normal
- **THEN** el PR se abre contra `develop` como base por defecto

#### Scenario: Main reservada para producción
- **WHEN** un cambio necesita llegar a `main`
- **THEN** llega mediante promoción desde `develop` o hotfix/release explícitamente autorizado, no como base normal de trabajo diario

#### Scenario: Promoción develop hacia main
- **WHEN** `develop` acumula cambios listos para producción
- **THEN** `delivery` o un maintainer autorizado promueve `develop` hacia `main` mediante PR de release/promoción y validación correspondiente

#### Scenario: Tags v* desde main
- **WHEN** se crea un tag estable `v*` para publicar una release
- **THEN** el tag apunta a un commit alcanzable desde `origin/main`

### Requirement: Configuración GitHub esperada
El repositorio SHALL documentar la configuración esperada de ramas en GitHub sin aplicarla automáticamente desde roles sin autorización.

#### Scenario: Default branch develop
- **WHEN** se prepara la configuración del repositorio remoto
- **THEN** `develop` queda documentada como default branch para PRs normales y trabajo diario

#### Scenario: Protecciones para develop y main
- **WHEN** se configuran branch protection rules
- **THEN** `develop` y `main` quedan documentadas como ramas protegidas, con PR requerido y checks aplicables antes de merge

#### Scenario: Delivery autorizado
- **WHEN** se necesite cambiar settings reales de GitHub, crear tags, commits, pushes o PRs
- **THEN** esas operaciones quedan fuera del rol `implementer` y requieren autorización explícita para `delivery` o acción manual del maintainer
