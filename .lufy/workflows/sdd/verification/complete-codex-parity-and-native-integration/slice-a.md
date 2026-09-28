# Verification: Slice A - Canonical harness integrity

## Scope

- Change: `complete-codex-parity-and-native-integration`
- Slice: A
- Gate state: `validated`
- Runtime surface: Go CLI harness resolution, project config, install, sync, skills, setup, info, verify and doctor.

## Requirement Evidence

| Requirement | Evidence |
| --- | --- |
| Explicit Codex selection is persisted consistently | CLI and installer tests compare `project.yaml` with install state after `--tool codex --methodology-tier ...`. |
| Harness resolution has explicit precedence and provenance | `internal/harnessconfig` tests cover explicit, project config, install state and default sources, plus unresolved drift. |
| Verify and doctor reconcile intent with installed reality | Verify and governance tests fail on tool drift; resolver tests identify methodology drift by tier. |
| Project config and install state writes are failure-safe | Project config tests cover reversible atomic merge; installer tests restore the original config when apply fails. |

## RED Evidence

Command:

```text
go test ./internal/harnessconfig ./internal/projectconfig ./internal/installer ./internal/verify ./internal/governance
```

Initial result: failed because `Resolve`, provenance sources and `MergeHarnessSelection` did not exist; doctor accepted `project=codex` with `install-state=opencode`.

## GREEN Evidence

Focused command:

```text
GOCACHE=/private/tmp/lufy-go-cache go test ./internal/harnessconfig ./internal/projectconfig ./internal/installer ./internal/syncer ./internal/skillregistry ./internal/setup ./internal/verify ./internal/governance ./internal/cli
```

Result: passed for every package.

Grouped command:

```text
GOCACHE=/private/tmp/lufy-go-cache LUFY_AI_VALIDATE_BASE=main scripts/validate.sh
```

Result: passed outside the filesystem/network sandbox required by loopback `httptest` fixtures.

- PR guard: passed for 16 tracked paths.
- Workflow YAML, action pinning, release checks, harness coupling and format-dispatch smoke: passed.
- Go tests and `go vet`: passed.
- Coverage: `80.2%`, threshold `80.0%`.
- Go build: passed.
- Shellcheck: `not_available`; the repository gate reported the local omission explicitly.

## Residual Risks

- Switching between already-installed tool adapters can leave retired files outside the new effective catalog; cleanup semantics remain outside Slice A.
- Cross-file consistency is rollback-safe for reported errors, while abrupt process termination is detected later by verify/doctor rather than committed transactionally across two files.
- Delivery, sync of Full SDD specs and archive remain pending.
