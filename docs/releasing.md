# Preparar una nueva release

Esta guía describe el flujo existente; ejecutarlo requiere autorización explícita. El modelo inline integrado en `develop` todavía no se publicó. Actualizar documentación no crea una versión, tag o promoción.

## Preparación del candidato

1. Revisar el diff respecto de la línea estable, los cambios incompatibles y la evidencia pertinente. Para inline, explicar formatos actuales, ausencia de migrador y recuperación conjunta.
2. Elegir el incremento de versión y el label de promoción de forma coherente con el cambio. Actualizar `RELEASE_VERSION`, las referencias canónicas en README/instalación/README de la CLI y una entrada del changelog para esa versión.
3. Ejecutar `scripts/validate.sh` y `scripts/check-doc-release-version.sh --require-changelog`. La validación de PR no publica artifacts.
4. Revisar distribución y recuperación con los scripts de artifacts/smokes disponibles. Atribuir cada resultado a su candidato y plataforma; no llamar validación de Windows a un smoke Bash en macOS.
5. Con autorización, promover el candidato desde `develop` a `main` por PR y observar los checks requeridos.

La versión preparada debe coincidir con el tag que calculará el workflow desde los tags remotos existentes. Los incrementos por labels no actualizan por sí solos `RELEASE_VERSION` ni la documentación. Una discrepancia bloquea el pipeline de artifacts incluso si el tag ya fue creado.

## Automatización del remoto

Al mergear un PR hacia `main`, `auto-release-tag.yml` calcula un tag anotado sobre el merge commit:

| Label | Efecto |
| --- | --- |
| `release:patch` o sin label | Incremento patch |
| `release:minor` | Incremento minor y patch a cero |
| `release:major` | Incremento major, minor/patch a cero |
| `release:skip` | Sin tag ni publicación automática |

Sólo se admite un label de incremento; `release:skip` no se combina con ellos. El workflow no sobrescribe tags y reintenta de forma acotada ante carreras. Un merge normal a `develop` no dispara esta publicación.

El tag creado dispara explícitamente `release.yml` mediante `workflow_dispatch`. El workflow también admite push de tags y dispatch manual para un tag existente. Exige que el commit del tag sea alcanzable desde `origin/main`, versión consistente, tests/build y artifacts válidos antes de publicar.

## Evidencia de publicación

`release.yml` construye archives de seis combinaciones OS/arquitectura, checksums, SBOM, provenance y bundles cosign. Ejecuta smokes y, después de publicar en GitHub Releases, descarga el artifact publicado y comprueba versión, instalación y verify en un target temporal. Un tag o upload por sí solo no prueba que el pipeline terminó.

Observar la conclusión del workflow de release y su E2E. `post-release-e2e.yml` permite un backstop manual para artifacts existentes; no corre en cada PR. Consultá [seguridad de releases](release-security.md) para autenticidad, [gates](ci-quality-gates.md) para cobertura y [ramas](github-branch-settings.md) para settings esperados.

Si la publicación falla, informar el tag/etapa y la recuperación necesaria. No declarar una nueva versión disponible por haber preparado el changelog ni reemplazar un tag publicado para ocultar el fallo.
