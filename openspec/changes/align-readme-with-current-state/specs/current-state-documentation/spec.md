## ADDED Requirements

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
- **THEN** los comandos usan rutas, binario y flags reales como `tools/lufy-cli-go`, `lufy-ai install`, `scripts/install.sh`, `--target`, `--dry-run`, `--yes` y `--no-engram`

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
`README.md`, `docs/getting-started.md` y `tools/lufy-cli-go/README.md` SHALL describir un estado compatible entre sí sobre CLI Go, assets gestionados, backup/restore, verify, sync y CI.

#### Scenario: Getting started en español y sincronizado
- **WHEN** una persona lee `docs/getting-started.md`
- **THEN** el contenido humano está en español y coincide con el quickstart y límites descritos en `README.md`

#### Scenario: README de CLI refleja alcance técnico
- **WHEN** una persona lee `tools/lufy-cli-go/README.md`
- **THEN** entiende los comandos implementados o en curso, la validación local real, y el estado de assets gestionados sin contradicción con el README raíz

#### Scenario: Trabajo en curso claramente marcado
- **WHEN** la documentación menciona CI o sync si aún dependen de proposals/implementaciones activas
- **THEN** los marca como en curso o propuesta activa, no como capacidad final completada

### Requirement: Validación documental del cambio
La implementación de esta change SHALL incluir evidencia mínima de revisión documental y whitespace sin ejecutar toolchains no relacionados.

#### Scenario: Validación estática suficiente
- **WHEN** se completa la limpieza documental
- **THEN** se reporta revisión de rutas/enlaces locales, coherencia con roadmap/specs y resultado real de `git diff --check` u otra validación estática disponible
