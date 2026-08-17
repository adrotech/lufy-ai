## Context

`RELEASE_VERSION` ya existe y `scripts/check-doc-release-version.sh` comprueba que los semver encontrados en tres documentos sean iguales a ese archivo. La brecha es sistémica: el auto-tag calcula una version exclusivamente desde tags remotos y nunca la contrasta con el source tree; `release.yml` confia en el tag; el smoke de artifacts usa una version fixture; y el E2E publicado imprime la version sin verificarla.

## Goals / Non-Goals

**Goals:**

- Hacer observable y bloqueante cualquier drift entre version planificada, tag y binario.
- Mantener tags como historia publicada inmutable y `RELEASE_VERSION` como intencion del commit que sera promovido.
- Dar un unico comando local para actualizar las referencias copiables.
- Preservar el flujo de labels `release:patch|minor|major|skip` y los reintentos acotados existentes.

**Non-Goals:**

- No publicar `v0.6.22` desde esta rama.
- No modificar la estrategia semver ni crear tags desde `develop`.
- No reemplazar GitHub-generated release notes.
- No introducir Node, Python ni una dependencia nueva para versionar documentos.

## Decisions

### `RELEASE_VERSION` representa la version planificada del commit

El archivo deja de ser una referencia informal. El PR de promocion debe llevar la version que el label de release producira. Si otro release gana una carrera y cambia el siguiente tag, el workflow falla en vez de etiquetar el commit con una version distinta de la documentada.

### Setter Bash coordinado sin dependencias externas

`scripts/set-release-version.sh <vX.Y.Z>` valida primero todos los archivos gestionados y reemplaza literalmente la version canonica anterior solo en esas referencias copiables. No modifica el changelog porque una entrada de release requiere contenido y fecha revisables por una persona.

### Checker reutilizable

`scripts/check-doc-release-version.sh` acepta `--expected`, `--require-changelog` y `--binary`. Tambien acepta `LUFY_AI_RELEASE_ROOT` para pruebas aisladas. El modo sin argumentos conserva compatibilidad con CI local.

### Doble gate remoto

`auto-release-tag.yml` valida antes de crear el tag. `release.yml` vuelve a validar sobre el checkout del tag. La duplicacion es deliberada: el primer gate protege la historia Git y el segundo protege los artifacts publicos/manual dispatches.

### Artifact real, no fixture

Un script extrae el artifact correspondiente al `GOOS/GOARCH` del runner desde `dist`, ejecuta `lufy-ai version` y exige el tag exacto. El E2E publicado repite la asercion luego de descargar desde GitHub Releases.

## Failure Model

- `RELEASE_VERSION` invalido o distinto del tag: fallo antes de tag/build.
- Documento con otra version: fallo con path y versiones detectadas.
- Documento sin referencia canonica: fallo; evita checks verdes vacios.
- Changelog sin heading exacto: fallo solo en gates de release.
- Binario sin tag esperado: fallo antes de publicar o al verificar el release publicado.
- Carrera de tag: el workflow recalcula, vuelve a validar y no publica una version que el commit no declara.

## Compatibility

- El checker conserva su nombre y su invocacion actual en `scripts/validate.sh`.
- `scripts/bootstrap.sh` y el formato de artifacts no cambian.
- Los releases existentes permanecen inmutables.
- `v0.6.22` es una preparacion en source; la publicacion sigue requiriendo promocion `develop -> main` y tag alcanzable desde `main`.
