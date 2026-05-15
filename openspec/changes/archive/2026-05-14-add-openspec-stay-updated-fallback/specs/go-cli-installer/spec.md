## ADDED Requirements

### Requirement: CLI provides OpenSpec resolver package
La CLI Go SHALL incluir un paquete interno para resolver la fuente efectiva de OpenSpec sin acoplarla al catálogo de assets gestionados.

#### Scenario: Resolver package is isolated
- **WHEN** se inspecciona la implementación de stay-updated
- **THEN** la lógica de resolución, manifiestos y cache vive en un paquete interno dedicado y no dentro de `cmd/lufy-ai/main.go`

#### Scenario: Resolver remains stdlib-only
- **WHEN** se compila la CLI Go después de agregar el resolver
- **THEN** no requiere dependencias externas nuevas salvo decisión explícita documentada

### Requirement: CLI uses embedded baseline as fallback
La CLI Go SHALL conservar instalación standalone usando baseline embebida cuando no haya fuente OpenSpec externa válida.

#### Scenario: Release binary resolves offline baseline
- **WHEN** un binario release se ejecuta sin checkout fuente, sin red y sin `openspec` en `PATH`
- **THEN** el resolver selecciona la baseline embebida y los comandos locales siguen funcionando

#### Scenario: Resolver does not mutate during install by default
- **WHEN** el usuario ejecuta `lufy-ai install` o `lufy-ai sync`
- **THEN** la CLI no descarga OpenSpec remoto ni modifica cache salvo que una acción explícita de update/cache lo solicite

### Requirement: CLI validates OpenSpec cache safely
La CLI Go SHALL validar cache OpenSpec por manifiesto y paths seguros antes de usarla.

#### Scenario: Corrupt cache falls back safely
- **WHEN** la cache local existe pero su manifiesto es inválido
- **THEN** la CLI ignora esa cache, reporta warning accionable y usa la siguiente capa válida

#### Scenario: Unsafe cache paths are blocked
- **WHEN** el manifiesto de cache contiene paths absolutos, traversal o symlinks inseguros
- **THEN** la CLI rechaza la cache y no lee ni escribe fuera del target
