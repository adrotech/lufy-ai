---
description: Report installed OpenSpec workflow baseline metadata.
agent: orchestrator
---

Report the effective local OpenSpec workflow baseline.

## Command behavior

- Read `openspec/UPSTREAM.json` from the current repository.
- Report `workflow`, `effectiveOpenSpecVersion`, `profile`, `source.type`, `source.repository` and whether the source is local embedded metadata.
- If `openspec/UPSTREAM.json` is missing or invalid JSON, fail with an actionable message: run `lufy-ai sync` from a trusted `lufy-ai` source or reinstall the managed assets.
- Do not infer or invent a version from package metadata, git tags, network calls or remote OpenSpec state.

## Recommended execution

1. Read `openspec/UPSTREAM.json`.
2. Print the baseline fields in a compact human-readable report.
