## ADDED Requirements

### Requirement: Structural hardening checks in CI
The Go installer CI SHALL include checks that protect structural hardening guarantees for assets, paths, state metadata and atomic writes.

#### Scenario: Catalog parity covered by tests
- **WHEN** CI runs Go tests for `tools/lufy-cli-go/`
- **THEN** tests fail if the root managed asset catalog and embedded catalog drift for target paths, policies or source hashes

#### Scenario: Windows traversal semantics covered
- **WHEN** CI runs path safety tests
- **THEN** tests include traversal inputs with Windows separators or mixed separators and verify they are rejected

#### Scenario: State metadata covered
- **WHEN** CI runs install/sync tests that write `.lufy-ai/install-state.json`
- **THEN** tests verify tool metadata and source fingerprint are populated from runtime/catalog sources rather than hardcoded proposal-era constants

#### Scenario: Atomic copy behavior covered
- **WHEN** CI runs install/sync/backup tests
- **THEN** tests cover that managed payload writes use the atomic write path or verify equivalent behavior through the copy helper
