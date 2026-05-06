## MODIFIED Requirements

### Requirement: Consistencia documental entre README y docs auxiliares
`README.md`, `docs/getting-started.md`, `tools/lufy-cli-go/README.md` y `docs/roadmap.md` SHALL describir un estado compatible entre sí sobre CLI Go, assets gestionados, backup/restore, verify, sync, CI, releases y el flujo de ramas `develop`/`main`.

#### Scenario: Getting started en español y sincronizado
- **WHEN** una persona lee `docs/getting-started.md`
- **THEN** el contenido humano está en español y coincide con el quickstart y límites descritos en `README.md`

#### Scenario: README de CLI refleja alcance técnico
- **WHEN** una persona lee `tools/lufy-cli-go/README.md`
- **THEN** entiende los comandos implementados o en curso, la validación local real, el estado de assets gestionados y la regla de releases `v*` desde `main` sin contradicción con el README raíz

#### Scenario: Flujo de ramas documentado
- **WHEN** la documentación pública u operativa describe contribución, CI o releases
- **THEN** explica que el trabajo normal entra por PR a `develop`, `main` es productiva, la promoción es `develop` → `main` y los tags `v*` se crean desde commits alcanzables desde `main`

#### Scenario: Configuración GitHub documentada
- **WHEN** un maintainer necesita aplicar settings remotos
- **THEN** existe una guía local breve para default branch `develop` y protección de `develop`/`main`

#### Scenario: Trabajo en curso claramente marcado
- **WHEN** la documentación menciona CI o sync si aún dependen de proposals/implementaciones activas
- **THEN** los marca como en curso o propuesta activa, no como capacidad final completada
