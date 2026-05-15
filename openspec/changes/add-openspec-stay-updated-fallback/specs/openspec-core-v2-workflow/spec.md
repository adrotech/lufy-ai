## ADDED Requirements

### Requirement: Baseline participates in stay-updated resolution
The installed OpenSpec workflow SHALL use `openspec/UPSTREAM.json` as the local baseline input for stay-updated resolution.

#### Scenario: Baseline includes resolver metadata
- **WHEN** a target is installed or synced with stay-updated assets
- **THEN** `openspec/UPSTREAM.json` includes enough metadata to compare effective version, minimum compatible version and source type

#### Scenario: Baseline remains offline-readable
- **WHEN** the user is offline and no cache or PATH source is available
- **THEN** the installed baseline remains sufficient for local workflow commands to report the fallback version

### Requirement: Opsx version reports resolved source
The installed workflow SHALL report the resolved OpenSpec source layer, not only static baseline metadata.

#### Scenario: Version report shows resolver layer
- **WHEN** the user runs `opsx-version` after stay-updated support is installed
- **THEN** the output identifies `PATH`, cache or embedded baseline as the effective source

#### Scenario: Resolver failures are actionable
- **WHEN** resolver metadata, cache or baseline files are invalid
- **THEN** `opsx-version` reports the failing layer and the next recovery action instead of inventing a version
