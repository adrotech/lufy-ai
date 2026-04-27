---
description: Implement approved OpenSpec tasks.
agent: orchestrator
---

Run unified OpenSpec workflow in `apply` mode.

## Command behavior

- Verify change artifacts exist before implementation.
- Use `implementer` for bounded code changes.
- Run validation during iteration.
- Do not create PRs, delegate to `delivery`.

## Recommended execution

1. `skill sdd-workflow` (mode=apply)
2. Run validation (typecheck, tests, coverage)