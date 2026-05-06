## ADDED Requirements

### Requirement: Standalone asset source
The CLI Go SHALL support installation from a distributed binary without requiring access to the source repository checkout.

#### Scenario: Embedded assets install without clone
- **WHEN** a release binary includes managed assets embedded in the binary
- **THEN** `lufy-ai install --target <dir>` can install the managed OpenCode/OpenSpec assets without reading from the repository source tree

#### Scenario: Bundle assets install without clone
- **WHEN** a release uses a versioned asset bundle instead of embedded assets
- **THEN** the CLI or bootstrap verifies the bundle integrity before using it as the asset source for installation

#### Scenario: Source checkout remains development path only
- **WHEN** the CLI runs from a developer checkout
- **THEN** it may use local assets for development workflows, but public installation documentation does not require cloning once standalone assets are implemented

### Requirement: Release binary preserves installer safety
The release-distributed `lufy-ai` binary SHALL preserve existing install, verify, backup, restore and sync safety semantics.

#### Scenario: Distributed install remains idempotent
- **WHEN** the user runs a release binary installation twice against the same target
- **THEN** the second run reports unchanged managed assets without overwriting local drift or unmanaged user files

#### Scenario: Distributed verify uses same structural checks
- **WHEN** the user runs `lufy-ai verify --target <dir> --no-engram` from a release binary
- **THEN** it validates structure, JSON, manifest and SHA-256 managed asset hashes with the same contract as the local build

#### Scenario: Wrapper remains strict
- **WHEN** `scripts/install.sh` is retained after release distribution exists
- **THEN** it continues to delegate to `lufy-ai install` and does not implement its own remote download fallback
