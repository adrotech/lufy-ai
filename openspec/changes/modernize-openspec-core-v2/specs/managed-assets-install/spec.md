## ADDED Requirements

### Requirement: Catálogo incluye assets OpenSpec core v2
La CLI Go SHALL instalar los assets OpenSpec core v2 como parte del catálogo gestionado cuando estén presentes en la fuente de `lufy-ai`.

#### Scenario: Nuevos comandos y skills se instalan
- **WHEN** la CLI construye el catálogo desde una fuente con OpenSpec core v2
- **THEN** incluye `/opsx-sync`, `openspec-sync`, `opsx-version`, `openspec/config.yaml` v2 y `openspec/UPSTREAM.json` como assets gestionados

#### Scenario: Assets embebidos conservan paridad
- **WHEN** la CLI se compila como binario standalone
- **THEN** los assets OpenSpec core v2 embebidos coinciden con los assets raíz usados en desarrollo

### Requirement: Verify cubre baseline OpenSpec core v2
`lufy-ai verify` SHALL reportar el estado de los assets OpenSpec core v2 requeridos por la instalación gestionada.

#### Scenario: Baseline faltante falla verify
- **WHEN** un target instalado con catálogo core v2 no contiene `openspec/UPSTREAM.json`
- **THEN** `lufy-ai verify` reporta fallo accionable para el asset faltante

#### Scenario: Sync repara assets core v2 sin pisar drift
- **WHEN** un target existente carece de un nuevo asset OpenSpec core v2 y no hay conflicto local
- **THEN** `lufy-ai sync` planifica copiarlo y preserva las policies de drift existentes para cualquier archivo modificado por el usuario
