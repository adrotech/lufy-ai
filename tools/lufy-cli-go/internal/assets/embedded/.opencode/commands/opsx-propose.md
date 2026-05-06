---
description: Create a new OpenSpec change and generate proposal artifacts.
agent: orchestrator
---

Run the concrete OpenSpec proposal skill.

## Command behavior

- Before creating a new change, check if there's a pending PR from a completed change.
- If pending, block or pause until PR is closed or user explicitly authorizes.
- Resolve change name from command argument or user request.
- Invoke concrete skill `openspec-propose`.
- Ensure generated artifacts are ready for implementation.
- Write proposal, design, tasks, specs in Spanish by default; keep filenames unchanged.
- If GitHub Project tracking enabled, call sync with status Ready.

## Recommended execution

1. Use skill `openspec-propose`
2. If a downstream project installed `project-sync`, run `skill project-sync --change <name> --status Ready` (optional)
