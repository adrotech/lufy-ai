## Context

`lufy-ai` ya publica artifacts binarios con `.github/workflows/release.yml` cuando recibe un push humano/manual de tag `v*`. Ese workflow valida que el commit taggeado sea alcanzable desde `origin/main`, por lo que la política de releases estables ya existe en el gate de publicación. Los tags creados con `GITHUB_TOKEN` no disparan otros workflows por evento `push`, así que el flujo automático debe invocar `release.yml` explícitamente con `workflow_dispatch`.

El punto manual pendiente es crear el tag después de mergear una promoción o hotfix hacia `main`. El nuevo workflow debe operar únicamente sobre eventos `pull_request.closed`, no debe reconstruir binarios y debe usar el merge commit final del PR como target del tag.

## Goals / Non-Goals

**Goals:**
- Crear automáticamente el siguiente tag patch semver simple al mergear un PR hacia `main`.
- Usar tags `vMAJOR.MINOR.PATCH`, ignorando prereleases y otros tags no compatibles para el cálculo.
- Crear un tag anotado que apunte al merge commit del PR, pushearlo e invocar `release.yml` mediante `workflow_dispatch` sobre ese tag.
- Verificar antes de taggear que el target commit es alcanzable desde `origin/main`.
- Ser seguro ante eventos no aplicables y ante tags existentes.

**Non-Goals:**
- No cambiar la estrategia de release binaria ni duplicar build/publicación fuera de `release.yml`.
- No implementar selección automática de minor/major ni prereleases.
- No cambiar branch protection ni permisos del repositorio desde código.

## Decisions

- **Workflow separado `auto-release-tag.yml`:** mantiene la responsabilidad acotada a crear tags y a solicitar el release mediante `workflow_dispatch`, dejando el build de binarios en `release.yml`. Alternativa descartada: depender de que el push del tag con `GITHUB_TOKEN` dispare `release.yml`, porque GitHub Actions suprime esos eventos encadenados.
- **Evento `pull_request.closed` con condición de job:** el job solo corre cuando `github.event.pull_request.merged == true` y `base.ref == 'main'`. Alternativa descartada: disparar en `push` a `main`, porque perdería contexto confiable del PR y podría taggear pushes directos no deseados.
- **Patch semver automático:** se selecciona el mayor tag simple `vX.Y.Z` mediante ordenamiento versionado y se incrementa `PATCH`; si no hay tags válidos, el primer tag es `v0.1.0`. Esto conserva un flujo simple y predecible sin inferir breaking changes.
- **Tag anotado:** el workflow crea tags anotados con mensaje que referencia el PR y el merge commit. Se prefiere sobre lightweight porque deja metadata auditable del release automático.
- **Idempotencia como no-op explícito:** si el tag calculado ya existe local/remotamente, el workflow escribe un mensaje claro, no invoca `release.yml` y termina exitosamente sin intentar sobrescribir. Es más seguro que fallar, recrear tags o duplicar releases, porque evita romper ejecuciones repetidas sin mutar historia publicada.
- **Guard de alcanzabilidad:** antes de crear el tag se hace `git fetch origin main` y `git merge-base --is-ancestor <merge_commit> origin/main`. Esto replica la política de release estable antes de solicitar el workflow de publicación.
- **Dispatch seguro de `release.yml`:** `release.yml` conserva el trigger por push de tags `v*` para tags manuales y agrega `workflow_dispatch` con input `tag`; en ambos casos valida formato `v*`, resuelve el commit del tag y exige alcanzabilidad desde `origin/main` antes de publicar.

## Risks / Trade-offs

- **Carrera entre merges simultáneos a `main`:** dos ejecuciones podrían calcular el mismo siguiente tag. → Mitigación: `concurrency` por workflow y verificación de tag existente antes de pushear; si ya existe, no-op explícito.
- **Tags prerelease existentes:** tags como `v1.2.3-rc.1` no participan del cálculo. → Mitigación: documentar que el cálculo usa solo semver simple `vMAJOR.MINOR.PATCH`.
- **Dependencia de `sort -V` en runner Ubuntu:** el workflow corre en `ubuntu-latest`, donde coreutils provee `sort -V`. → Mitigación: el script valida el formato con regex y no introduce dependencias externas ni descargas.
- **`merge_commit_sha` ausente o inválido:** el evento debería proveerlo para PRs mergeados. → Mitigación: el script falla con mensaje claro si no está disponible o no resuelve a commit.
