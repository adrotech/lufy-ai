## ADDED Requirements

### Requirement: Offline regression pipeline
La validación local y CI SHALL ejecutar las regresiones offline del piloto, propagar fallos y SHALL NOT iniciar modelos, usar credenciales ni convertir una dependencia ausente en validación exitosa.

#### Scenario: Pilot paths trigger CI
- **WHEN** cambia un archivo bajo tools/codex-behavior-pilot en un PR o push a las ramas configuradas
- **THEN** el workflow incluye ese cambio y ejecuta la suite offline mediante Quality

#### Scenario: Offline validation is mandatory
- **WHEN** se ejecuta scripts/validate.sh
- **THEN** unittest ejecuta las regresiones sin --live y un fallo o Python ausente devuelve un código no exitoso con diagnóstico
