## ADDED Requirements

### Requirement: Use-case services stay thin and composable
La CLI Go SHALL mantener los comandos publicos delegando a casos de uso composables, y los casos de uso SHALL coordinar componentes especializados en vez de concentrar toda la logica en una funcion grande.

#### Scenario: CLI command delegates without business branching
- **WHEN** `internal/cli` parsea flags de un comando soportado
- **THEN** SHALL delegar la ejecucion a un caso de uso o servicio de aplicacion
- **AND** SHALL limitarse a parsing, salida de errores, exit codes y ayuda del comando

#### Scenario: Installer plan and apply are separated
- **WHEN** `install` construye un plan
- **THEN** la planificacion SHALL poder probarse sin aplicar mutaciones reales
- **AND** la aplicacion SHALL poder probarse con un plan ya construido

#### Scenario: Verify checks and presentation are separated
- **WHEN** `verify` genera checks estructurales
- **THEN** SHALL construir checks y conteos en un modelo antes de presentarlos como salida humana o JSON
- **AND** la salida JSON SHALL depender del mismo modelo que la salida humana

### Requirement: Public CLI compatibility during architecture refactor
El refactor de boundaries SHALL conservar comandos, flags, defaults y layout instalado del preset actual.

#### Scenario: Existing install command remains stable
- **WHEN** el usuario ejecuta `lufy-ai install --target <dir> --yes --no-engram`
- **THEN** el comando SHALL instalar el preset `opencode` + `openspec` compatible
- **AND** SHALL preservar la semantica de `AGENTS.md`, `lufy-ia.harness.md`, `opencode.json`, `tui.json` y manifest

#### Scenario: Existing validation command remains stable
- **WHEN** el usuario ejecuta `lufy-ai verify --target <dir> --no-engram`
- **THEN** el comando SHALL validar manifest, hashes, directorios criticos, JSON parseable y referencia de `AGENTS.md` con el mismo contrato observable actual

#### Scenario: Wrapper remains strict
- **WHEN** se refactoriza la CLI Go
- **THEN** `scripts/install.sh` SHALL permanecer como wrapper estricto del binario Go
- **AND** SHALL NOT reintroducir fallback legacy ni logica propia de copia/configuracion
