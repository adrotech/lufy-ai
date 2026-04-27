---
description: Create a new OpenSpec change and generate proposal artifacts.
agent: orchestrator
---

Run unified OpenSpec workflow in `propose` mode.

## Command behavior

- Before creating a new change, check if there's a pending PR from a completed change.
- If pending, block or pause until PR is closed or user explicitly authorizes.
- Resolve change name from command argument or user request.
- Invoke skill with mode `propose`.
- Ensure generated artifacts are ready for implementation.
- Write proposal, design, tasks, specs in Spanish by default; keep filenames unchanged.
- If GitHub Project tracking enabled, call sync with status Ready.

## Recommended execution

1. `skill sdd-workflow` (mode=propose)
2. `skill project-sync --change <name> --status Ready` (optional)