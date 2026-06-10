# current-state-documentation Specification

## Purpose
Keep public and operational documentation aligned with the repository's real capabilities, supported workflows, validation commands and release model.

## Requirements
### Requirement: README centrado en estado real
El README SHALL funcionar como landing page operativa del repositorio y MUST describir solo capacidades actuales o explícitamente marcadas como trabajo en curso/futuro.

#### Scenario: Entrada principal honesta
- **WHEN** una persona lee `README.md`
- **THEN** encuentra qué entrega realmente el repo, cómo instalar o validar el kit, y enlaces a documentación detallada sin que templates o subagentes futuros aparezcan como instalables actuales

#### Scenario: Banner y navegación preservados
- **WHEN** se reestructura `README.md`
- **THEN** el banner existente y los enlaces relevantes a documentación local se preservan o se reemplazan por navegación equivalente y vigente

### Requirement: Quickstart basado en CLI Go actual
La documentación pública SHALL describir la instalación actual mediante la CLI Go `lufy-ai` y el wrapper estricto `scripts/install.sh`, sin reintroducir instrucciones legacy como camino principal.

#### Scenario: Instalación documentada con comandos reales
- **WHEN** una persona sigue el quickstart en `README.md` o `docs/getting-started.md`
- **THEN** los comandos usan rutas, binario y flags reales como `tools/lufy-cli-go`, `lufy-ai install`, `scripts/install.sh`, `--target`, `--dry-run` y `--yes`

#### Scenario: Verificación documentada con límites reales
- **WHEN** la documentación menciona validación local
- **THEN** usa comandos disponibles para el alcance actual, como validación Go desde `tools/lufy-cli-go/`, `lufy-ai verify` o revisión OpenSpec, y MUST NOT inventar comandos Node/TS de raíz

### Requirement: Separación de futuro y templates
La documentación SHALL ubicar templates, stacks, subagentes futuros y decisiones estratégicas en `docs/roadmap.md` o docs específicas, no como capacidades disponibles en la entrada principal.

#### Scenario: Templates no instalables no se prometen
- **WHEN** `README.md` menciona templates por stack, detección de stack o subagentes futuros
- **THEN** los presenta únicamente como roadmap/enlace futuro o los omite del README operativo

#### Scenario: Roadmap conserva contexto futuro
- **WHEN** se mueve contenido futuro fuera del README
- **THEN** `docs/roadmap.md` o una doc enlazada conserva el contexto relevante y aclara que no son assets instalables hasta que existan y se validen

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

### Requirement: Validación documental del cambio
La implementación de esta change SHALL incluir evidencia mínima de revisión documental y whitespace sin ejecutar toolchains no relacionados.

#### Scenario: Validación estática suficiente
- **WHEN** se completa la limpieza documental
- **THEN** se reporta revisión de rutas/enlaces locales, coherencia con roadmap/specs y resultado real de `git diff --check` u otra validación estática disponible

### Requirement: Documentation migrates only after implementation
Public documentation SHALL describe the clone-free release installer only after release artifacts, checksums, bootstrap and standalone assets are implemented and validated.

#### Scenario: Proposal stage does not update quickstart as current state
- **WHEN** this proposal exists but runtime implementation is not complete
- **THEN** `README.md` and `docs/getting-started.md` do not present clone-free remote installation as an available current capability

#### Scenario: Final docs describe no-clone path
- **WHEN** the clone-free release installer is implemented and validated
- **THEN** `README.md`, `docs/getting-started.md` and `tools/lufy-cli-go/README.md` describe the no-clone install path with version pinning, checksum verification and `lufy-ai verify`

### Requirement: Obsolete clone/build docs removed at completion
Documentation SHALL remove or demote obsolete clone/build instructions once the release installer is the supported primary path.

#### Scenario: Clone no longer primary install path
- **WHEN** standalone release installation is validated
- **THEN** README/getting-started no longer require cloning the repository as the primary user install flow

#### Scenario: Development build remains scoped
- **WHEN** clone/build instructions remain useful for contributors
- **THEN** they are clearly scoped as development/contributor workflow rather than end-user installation

### Requirement: Roadmap marks release installer as planned
The roadmap SHALL record this distribution roadmap as planned future work and MUST NOT describe it as already implemented before runtime completion.

#### Scenario: Roadmap planned block
- **WHEN** a reader opens `docs/roadmap.md` during this proposal stage
- **THEN** it includes a planned block for versioned binary releases, bootstrap installation and standalone assets without claiming them as current capabilities

### Requirement: Documentacion del workflow sistemico
La documentacion operativa SHALL describir el workflow sistemico vigente como una practica de analisis inicial, implementacion por bloques, relectura final acotada y validacion final agrupada.

#### Scenario: Documentacion refleja fases reales
- **WHEN** una persona revisa guias operativas como `AGENTS.md`, `.opencode/policies/delivery.md` u OpenSpec docs relevantes
- **THEN** encuentra que el analisis de archivos existentes ocurre al inicio, las relecturas se evitan salvo justificacion, y tests/coverage se agrupan al final de la propuesta cuando apliquen

#### Scenario: Documentacion conserva limites de toolchain
- **WHEN** la documentacion menciona tests, coverage o validacion final
- **THEN** aclara que solo se ejecutan comandos reales disponibles para el alcance y que las limitaciones se reportan explicitamente
