---
name: lufy-close
description: Close a Lufy-governed change by checking validation, sync, delivery state, and remaining blockers before reporting final status.
---

# Lufy Close

Use this skill when the user asks to close, finish, archive, or hand off a Lufy-governed change.

1. Read `AGENTS.md` and the current `.lufy/managed-state/install-state.json` when present.
2. Confirm validation evidence from real commands or explicit manual review.
3. Confirm delivery was explicitly authorized before any commit, push, PR, merge, or branch cleanup.
4. Report blockers instead of marking work closed when validation, sync, delivery, or remote checks are missing.
5. Use the Result Contract envelope when the result is substantive.
