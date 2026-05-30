## MODIFIED Requirements

### Requirement: Adapter-owned paths
Los paths concretos de configuración SHALL pertenecer al adapter de tool, no al core ni a componentes metodológicos.

#### Scenario: Component requests skills path
- **WHEN** un componente necesita instalar skills
- **THEN** SHALL consultar el adapter efectivo para obtener el directorio de skills en vez de hardcodear `.opencode/skills` u otro path

#### Scenario: Use cases request project config through tool runtime
- **WHEN** install, sync o verify necesitan planificar, aplicar o validar configuracion project-level de la tool efectiva
- **THEN** SHALL hacerlo mediante una capa runtime/adaptador de tool
- **AND** SHALL NOT invocar directamente servicios especificos de OpenCode desde el caso de uso

#### Scenario: Use cases request global config root through tool runtime
- **WHEN** install, sync, verify o status necesitan resolver config global por scope
- **THEN** SHALL hacerlo mediante una capa runtime/adaptador de tool
- **AND** SHALL preservar el path global actual para `opencode`

#### Scenario: Non writable tool runtime is explicit
- **WHEN** la capa runtime recibe `codex`, `claude-code` u otra tool sin escritura real autorizada
- **THEN** SHALL retornar un error explicito sin resolver paths OpenCode por fallback implicito
