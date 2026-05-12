## Why

`lufy-ai` currently treats most installed files as fully managed assets and blocks install/sync when users make expected local edits, especially in `AGENTS.md`, `.opencode/policies/*` or project OpenSpec files. That protects data, but it also prevents safe upgrades and creates a poor brownfield UX: the installer cannot deliver new defaults without forcing manual conflict resolution first.

This change introduces an explicit drift-resolution model so upgrades can continue without overwriting user work: lufy-owned assets can update with backup, user-owned defaults receive `.lufy-new`, mixed files can update only managed blocks, and JSON config keeps merge semantics.

## What Changes

- Add declarative asset update policies: `managed`, `no-replace`, `merge-block`, `merge-json` and `metadata`.
- Extend install state to record each asset policy, scope and ancestor metadata with silent migration from the current state schema.
- Change install/sync planning so drift behavior is policy-driven instead of always becoming a blocking conflict.
- Add `.lufy-new` output for `no-replace` assets when a newer source exists but the local file has drift.
- Add `merge-block` support for files such as `AGENTS.md`, preserving user text outside `<!-- LUFY:BEGIN <id> -->` / `<!-- LUFY:END <id> -->` blocks.
- Preserve and expose ancestors under `.lufy-ai/ancestors/` so future or current merge workflows have the last clean lufy baseline.
- Extend CLI UX with merge/restore/status/verify reporting for drift, `.lufy-new`, policy and ancestor state.
- Add install scope support for global/project/both so OpenCode shared assets can live globally while project-local files remain local.
- **BREAKING / behavior change candidate**: the default install scope may move away from project-only after RC validation. Until that decision is finalized, `--scope=project` must preserve current behavior.

## Capabilities

### New Capabilities
- `install-drift-resolution`: Policy-driven upgrade behavior, `.lufy-new`, ancestors, merge-block semantics and merge/restore UX for drifted assets.

### Modified Capabilities
- `managed-assets-install`: Existing install/sync/verify behavior changes from binary managed/conflict semantics to declarative per-asset policies, scope-aware catalogs and richer state.
- `go-cli-installer`: CLI surface changes to include policy/scope-aware flags and user-facing merge/restore/status/reporting behavior.

## Impact

- Affected Go packages: `tools/lufy-cli-go/internal/assets`, `state`, `installer`, `syncer`, `verify`, `status`, `backup`, `platform`, `cli` and new merge helpers as needed.
- Affected embedded/root assets: `AGENTS.md.template`, `.opencode/*`, `openspec/*` and `tools/lufy-cli-go/internal/assets/embedded/*` mirror.
- Affected persistent data: `.lufy-ai/install-state.json`, `.lufy-ai/backups/*`, new `.lufy-ai/ancestors/*` and `.lufy-new` files in targets.
- Affected docs/tests: CLI docs, roadmap/status docs, Go tests, smoke install/sync/verify scripts and sandbox upgrade scenarios.
- No new external Go dependencies should be introduced; implementation remains stdlib-only unless explicitly approved later.
