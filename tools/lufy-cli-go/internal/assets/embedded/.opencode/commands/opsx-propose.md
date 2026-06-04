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
- After `openspec-propose` returns, actively read the expected generated files under `openspec/changes/<change>/` before reporting ready:
  - `proposal.md`
  - `tasks.md`
  - `specs/**/spec.md`
  - `design.md` when required by the active schema/status
- Ensure change specs use explicit delta sections: `## ADDED Requirements`, `## MODIFIED Requirements` or `## REMOVED Requirements`.
- Ensure each added or modified requirement has at least one `#### Scenario:` with `WHEN` and `THEN`; `GIVEN` is optional.
- If any expected artifact is missing, empty, lacks required delta markers or lacks testable scenarios, STOP with a blocked result that names the exact path and recovery action; do not route to `/opsx-apply`.
- If Engram MCP is enabled and available, verify the proposal/delta trace record after artifact creation; if enabled but unavailable, report the traceability limitation explicitly.
- Write proposal, design, tasks, specs in Spanish by default; keep filenames unchanged.
- After artifacts are ready, offer an optional HTML overview: ask whether the user wants to run `lufy-ai opsx render --change <change> --format html --theme notion-dark`. If accepted, run it and report the generated path; if skipped or unavailable, do not block `/opsx-apply`.
- If GitHub Project tracking enabled, call sync with status Ready.

## Recommended execution

1. Use skill `openspec-propose`
2. If a downstream project installed `project-sync`, run `skill project-sync --change <name> --status Ready` (optional)
