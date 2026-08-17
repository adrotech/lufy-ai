## Why

El release estable publico ya avanzo hasta `v0.6.21`, pero `RELEASE_VERSION`, README, guia de instalacion y changelog siguen anclados en `v0.6.11`. El gate actual solo compara documentos contra el archivo local, por lo que puede quedar verde aunque todo el repositorio este atrasado respecto del tag que se publica.

Lufy vende gobernanza, reproducibilidad y evidencia. La version declarada por el commit de release debe coincidir con el tag calculado, la documentacion copiable, el changelog y la metadata del binario antes de crear o publicar un release.

## What Changes

- Convertir `RELEASE_VERSION` en la unica version de release planificada dentro del source tree y agregar un comando seguro para propagarla a la documentacion copiable.
- Exigir que el tag calculado por `auto-release-tag.yml` coincida con `RELEASE_VERSION` y con una entrada explicita del changelog en el merge commit.
- Exigir nuevamente esa consistencia dentro de `release.yml` antes de construir/publicar artifacts.
- Verificar que el artifact nativo construido y el artifact publicado reporten exactamente el tag esperado.
- Agregar pruebas shell para escenarios correctos, version esperada divergente, documentacion stale, changelog faltante y binario divergente.
- Corregir la configuracion de cache de `actions/setup-go` para el modulo anidado en `tools/lufy-cli-go`.
- Preparar `v0.6.22` como siguiente version coherente sobre el ultimo tag estable `v0.6.21`.

## Capabilities

### Modified Capabilities

- `release-version-governance`: agrega consistencia obligatoria entre source tree, tag, changelog, documentacion y binario.

## Impact

- Workflows: `.github/workflows/auto-release-tag.yml`, `.github/workflows/release.yml`, `.github/workflows/go-cli-install.yml` y workflows Go auxiliares.
- Scripts: checker de version documental, setter canonico, pruebas y verificacion del artifact.
- Documentacion: README principal, guia de instalacion, README de la CLI y changelog.
- Riesgo operacional: una promocion a `main` fallara de forma intencional si no prepara la version exacta esperada.

## Review Slices

### Slice 1: Source tree y pruebas locales

- Objetivo: establecer una version canonica actualizable y validable sin depender de GitHub Actions.
- Criterio: WHEN se cambia la version con el setter, THEN todos los documentos copiables usan la nueva version y el checker detecta cualquier drift.

### Slice 2: Gates remotos y artifacts

- Objetivo: impedir tags/releases incoherentes y probar la version embebida en el binario.
- Criterio: WHEN un PR mergeado a `main` calcula el siguiente tag, THEN el workflow solo crea el tag si source, changelog y artifact coinciden.

## Validation

- `scripts/test-release-version.sh`
- `scripts/validate.sh`
- revision estatica de los workflows y del rango contra `origin/develop`
- `openspec validate "enforce-release-version-consistency" --strict` cuando el CLI OpenSpec este disponible
