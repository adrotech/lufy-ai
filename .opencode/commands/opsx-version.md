---
description: Report installed OpenSpec workflow resolver metadata.
agent: orchestrator
---

Report the effective OpenSpec workflow source resolved by the local stay-updated metadata.

## Command behavior

- Read `openspec/UPSTREAM.json` from the current repository.
- Report `workflow`, `effectiveOpenSpecVersion`, `minimumCompatibleOpenSpecVersion`, `profile`, `source.type`, `source.repository` and resolver metadata.
- Resolve the effective source layer in this order when possible: `openspec` CLI in `PATH`, `.lufy/cache/openspec/<version>/manifest.json`, then embedded/local baseline metadata.
- Print the effective layer as `PATH`, `cache` or `embedded`, plus the selected version and path/source detail.
- If a layer is unavailable or invalid, report the failing layer and the next recovery action instead of inventing a version.
- If `openspec/UPSTREAM.json` is missing or invalid JSON, fail with an actionable message: run `lufy-ai sync` from a trusted `lufy-ai` source or reinstall the managed assets.
- Do not download remote OpenSpec assets, merge PRs, create tags or publish releases from this command.

## Recommended execution

1. Read `openspec/UPSTREAM.json`.
2. Check `openspec --version` from `PATH` and compare it with `minimumCompatibleOpenSpecVersion`.
3. If PATH is unavailable/incompatible, inspect `.lufy/cache/openspec/<version>/manifest.json` entries and reject corrupt manifests, unsafe paths or symlink escapes.
4. If no external source is valid, report the local embedded baseline as the offline fallback.
5. Print a compact human-readable report with diagnostics for skipped layers.
