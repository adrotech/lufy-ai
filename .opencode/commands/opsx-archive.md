---
description: Archive a completed OpenSpec change.
agent: orchestrator
---

Archive completed change and update tracking.

## Command behavior

- Verify all tasks are done.
- Run final validation.
- Update artifacts.
- Call project sync with status Done.

## Recommended execution

1. Verify 100% tasks complete
2. If a downstream project installed `project-sync`, run `skill project-sync --change <name> --status Done`
