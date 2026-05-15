## ADDED Requirements

### Requirement: OpenSpec resolver uses three fallback layers
La CLI Go SHALL resolver la fuente efectiva de OpenSpec usando capas ordenadas y verificables.

#### Scenario: PATH openspec has priority when compatible
- **WHEN** `openspec` existe en `PATH` y reporta una versión compatible con el baseline mínimo
- **THEN** el resolver lo selecciona como fuente efectiva y reporta su ruta y versión

#### Scenario: Cache is used when PATH is unavailable
- **WHEN** `openspec` no existe en `PATH` o no cumple la versión mínima y `.lufy-ai/openspec-cache/<version>/manifest.json` es válido
- **THEN** el resolver selecciona la cache local versionada como fuente efectiva

#### Scenario: Embedded baseline is offline fallback
- **WHEN** ni `PATH` ni cache local proveen una fuente válida
- **THEN** el resolver usa el baseline embebido y reporta que opera en modo fallback offline

### Requirement: OpenSpec cache is versioned and manifest-backed
La cache OpenSpec SHALL persistir versiones bajo `.lufy-ai/openspec-cache/<version>/` con manifiesto machine-readable.

#### Scenario: Cache manifest records provenance
- **WHEN** una versión OpenSpec se guarda en cache
- **THEN** `.lufy-ai/openspec-cache/<version>/manifest.json` registra versión, fuente, timestamps, assets y hashes disponibles

#### Scenario: Invalid cache is rejected
- **WHEN** el manifiesto de cache falta, no parsea o no coincide con los assets esperados
- **THEN** el resolver rechaza esa cache y continúa con la siguiente capa sin mutar el target

### Requirement: Cache and manifests use atomic writes
Las escrituras de cache y manifiestos OpenSpec SHALL ser atómicas y confinadas al target.

#### Scenario: Cache write is atomic
- **WHEN** la CLI escribe o actualiza una entrada de `.lufy-ai/openspec-cache/<version>/`
- **THEN** escribe mediante archivo temporal y rename, sin dejar manifiestos parciales como fuente válida

#### Scenario: Path traversal is rejected
- **WHEN** una versión, asset o path de manifiesto intenta escapar de `.lufy-ai/openspec-cache/`
- **THEN** la CLI rechaza la operación y no escribe archivos fuera del target

### Requirement: Upstream sync opens pull requests only
El workflow de actualización OpenSpec SHALL proponer bumps mediante PR y SHALL NOT modificar ramas protegidas directamente.

#### Scenario: Compatible upstream creates PR
- **WHEN** el workflow `sync-openspec.yml` detecta una versión upstream compatible más nueva que `openspec/UPSTREAM.json`
- **THEN** abre o actualiza un PR con cambios de baseline/manifiesto y evidencia de validación

#### Scenario: Workflow never automerges
- **WHEN** el workflow crea o actualiza un PR de baseline
- **THEN** no hace merge automático, no crea tags y no publica releases

### Requirement: Resolver reports actionable diagnostics
El resolver OpenSpec SHALL reportar la fuente efectiva y errores de fallback de forma accionable.

#### Scenario: Version report includes resolver layer
- **WHEN** el usuario consulta la versión efectiva del workflow OpenSpec
- **THEN** el reporte indica si la fuente proviene de `PATH`, cache local o baseline embebida

#### Scenario: All layers invalid fails clearly
- **WHEN** PATH, cache y baseline embebida son inválidos o incompatibles
- **THEN** la CLI falla con instrucciones concretas para reinstalar assets, limpiar cache o revisar `openspec/UPSTREAM.json`
