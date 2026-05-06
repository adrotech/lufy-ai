## ADDED Requirements

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
