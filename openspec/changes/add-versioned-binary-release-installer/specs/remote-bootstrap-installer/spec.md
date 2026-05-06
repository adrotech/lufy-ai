## ADDED Requirements

### Requirement: Remote bootstrap installer
The system SHALL provide a remote bootstrap installer that installs `lufy-ai` without requiring the user to clone the repository.

#### Scenario: OS and architecture detection
- **WHEN** the user runs the bootstrap installer on a supported platform
- **THEN** the installer detects OS and architecture and selects the matching versioned release artifact

#### Scenario: Unsupported platform fails safely
- **WHEN** the bootstrap installer runs on an unsupported OS/arch combination
- **THEN** it exits without downloading or installing a mismatched binary and reports supported alternatives

### Requirement: Checksum-before-install
The bootstrap installer MUST verify the SHA-256 checksum of every downloaded binary artifact before installing or executing it.

#### Scenario: Matching checksum permits installation
- **WHEN** the downloaded artifact hash matches the checksum published for the selected version
- **THEN** the bootstrap installer may install the binary into the selected destination

#### Scenario: Mismatched checksum blocks installation
- **WHEN** the downloaded artifact hash does not match the published checksum
- **THEN** the bootstrap installer deletes or quarantines the downloaded artifact, exits non-zero and does not install or execute the binary

### Requirement: Version pinning
The bootstrap installer SHALL support installing a specific release version and SHALL make that mode suitable for automation.

#### Scenario: Explicit version selected
- **WHEN** the user passes a version such as `vX.Y.Z` through a flag or documented environment variable
- **THEN** the bootstrap installer downloads artifacts and checksums only for that version

#### Scenario: Latest version is opt-in convenience
- **WHEN** the user omits an explicit version and chooses a latest/stable mode
- **THEN** the installer reports the resolved version before installing and documentation describes the reproducibility trade-off

### Requirement: Inspectable curl workflow
The documentation for remote installation SHALL present `curl | bash` only with an inspectable alternative.

#### Scenario: Direct pipe documented with warning
- **WHEN** documentation shows a direct `curl | bash` command
- **THEN** the same section also shows how to download the script, inspect it locally and execute it with an explicit version

#### Scenario: No destructive auto-run by default
- **WHEN** the bootstrap installer installs the binary
- **THEN** it does not run destructive commands such as `lufy-ai install` against a project target unless the user passes an explicit flag for that action

### Requirement: PATH installation controls
The bootstrap installer SHALL install the verified binary into a user-controllable destination that can be placed on `PATH`.

#### Scenario: Destination selected explicitly
- **WHEN** the user passes an install directory flag or variable
- **THEN** the bootstrap installer places the binary in that directory and reports any required PATH update

#### Scenario: Destination requires privilege
- **WHEN** the selected install directory is not writable by the current user
- **THEN** the bootstrap installer fails with an actionable message or requests an explicit privileged path strategy rather than silently escalating
