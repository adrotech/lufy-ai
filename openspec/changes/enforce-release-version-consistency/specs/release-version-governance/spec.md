## ADDED Requirements

### Requirement: Source tree declares the planned release version
El commit candidato a release SHALL declarar una unica version semver estable en `RELEASE_VERSION`, y la documentacion copiable SHALL usar esa misma version.

#### Scenario: Canonical version is propagated
- **WHEN** un maintainer ejecuta el setter con una version `vMAJOR.MINOR.PATCH` valida
- **THEN** `RELEASE_VERSION` y todas las referencias copiables gestionadas SHALL usar exactamente esa version
- **AND** el setter SHALL NOT reescribir entradas historicas del changelog

#### Scenario: Documentation drift is rejected
- **WHEN** un documento gestionado contiene una version distinta o no contiene la version canonica
- **THEN** el checker SHALL fallar con el path y la divergencia detectada

### Requirement: Tag creation matches release source metadata
El workflow automatico SHALL crear un tag solo cuando la version calculada coincide con la metadata versionada del merge commit.

#### Scenario: Calculated tag matches source and changelog
- **WHEN** el siguiente tag calculado coincide con `RELEASE_VERSION` y `CHANGELOG.md` contiene un heading exacto para esa version
- **THEN** el workflow puede continuar con la creacion segura del tag

#### Scenario: Calculated tag differs from source
- **WHEN** el siguiente tag calculado no coincide con `RELEASE_VERSION`
- **THEN** el workflow SHALL fallar antes de crear o mover un tag
- **AND** SHALL indicar la version calculada y la declarada

#### Scenario: Tag race changes the expected version
- **WHEN** aparece un tag remoto durante los reintentos y la nueva version calculada ya no coincide con el source tree
- **THEN** el workflow SHALL detenerse sin etiquetar el commit con una version no declarada

### Requirement: Release publication revalidates version consistency
El workflow de publicacion SHALL volver a validar source, tag, changelog y metadata del binario antes de publicar artifacts.

#### Scenario: Tag checkout is internally consistent
- **WHEN** `release.yml` procesa un tag valido y alcanzable desde `main`
- **THEN** SHALL comprobar que el tag coincide con `RELEASE_VERSION` y con una entrada del changelog antes del build

#### Scenario: Built artifact reports exact tag
- **WHEN** se construye el artifact nativo para el runner
- **THEN** ejecutar `lufy-ai version` desde ese artifact SHALL reportar exactamente el tag del release
- **AND** una divergencia SHALL bloquear la publicacion

#### Scenario: Published artifact reports exact tag
- **WHEN** el E2E descarga un artifact desde GitHub Releases
- **THEN** su salida de version SHALL contener exactamente el tag solicitado antes de probar install y verify
