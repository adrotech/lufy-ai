---
name: lufy-context-search
description: Use lufy-ai context as a secondary local index for compact architecture, impact, and path hints without replacing file, diff, or command evidence.
---

# Lufy Context Search

Use when a task needs quick orientation across files, symbols, docs, or OpenSpec and `.lufy/context/graph.json` may exist.

1. Check availability with `lufy-ai context status --target <repo> --json`.
2. If status is `not_available` or `stale`, report that status and continue with normal repository inspection. Recovery is `lufy-ai context build`.
3. Search compact hints with `lufy-ai context query --target <repo> --json "<term>"`.
4. For diff impact, use `lufy-ai context diff --target <repo> --json --base <ref>` when allowed by the role.
5. Explain a node or edge with `lufy-ai context explain --target <repo> --json <node-or-edge>` before relying on it as a hint.

Rules:

- Treat context graph output as a secondary index only.
- Do not treat graph inference as stronger evidence than current files, diffs, tests, logs, or validation commands.
- Return compact hints: node, path, kind, reason, status, relevance.
- Do not run `context build` unless the user or workflow explicitly allows mutating `.lufy/context/`.
- If the CLI or graph is unavailable, degrade to `not_available` and continue.
