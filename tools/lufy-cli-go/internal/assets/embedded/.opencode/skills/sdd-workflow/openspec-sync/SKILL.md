---
name: openspec-sync
description: Apply validated OpenSpec change deltas to main specs without archiving the change.
license: MIT
compatibility: Requires openspec CLI and core v2 delta specs.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.3.1"
---

Sync an OpenSpec change into the main specs after implementation or before archive.

**Input**: Optionally specify a change name. If omitted, infer it only when a single active change exists or the conversation clearly names it; otherwise ask the user to choose.

## Steps

1. **Select the change**

   - If a change name is provided, use it.
   - Otherwise run `openspec list --json` and select only when unambiguous.
   - Announce: `Using change: <name>`. To override, run `/opsx-sync <other>`.

2. **Load and validate change artifacts**

   ```bash
   openspec status --change "<name>" --json
   openspec validate "<name>"
   ```

   If validation fails, stop. Do not mutate main specs.

3. **Inspect delta specs**

   - Read `openspec/changes/<name>/specs/**/spec.md`.
   - Each delta spec must contain at least one of:
     - `## ADDED Requirements`
     - `## MODIFIED Requirements`
     - `## REMOVED Requirements`
   - Added and modified requirements must include `### Requirement:` sections.
   - Each added or modified requirement must include at least one `#### Scenario:` with `WHEN` and `THEN`; `GIVEN` is optional.
   - Removed requirements must include a reason or migration note in the removed requirement body.

4. **Plan target updates**

   For each delta spec, map it to `openspec/specs/<capability>/spec.md`.

   - `ADDED`: append new requirements that do not already exist.
   - `MODIFIED`: replace the complete matching main requirement by title.
   - `REMOVED`: remove the matching main requirement by title only when the delta explains why and the target exists.

   Show a concise summary of additions, replacements and removals before editing.
   Keep this plan as the source of truth for post-sync verification.

5. **Apply deltas without archiving**

   - Create missing capability spec files only for `ADDED` requirements.
   - Preserve unrelated main spec content.
   - Do not move `openspec/changes/<name>/`.
   - If a target requirement is missing, duplicated or ambiguous, stop before editing that file and report the blocker.

6. **Verify sync result**

   Actively read every affected `openspec/specs/<capability>/spec.md` after applying deltas.
   - For `ADDED`, verify the target contains the added requirement title and scenario content.
   - For `MODIFIED`, verify the target contains the updated requirement body and no stale duplicate of the old requirement remains.
   - For `REMOVED`, verify the target no longer contains the removed requirement title.
   - If any expected target file or requirement state is missing, STOP with `status: blocked`, cite the path/requirement, and do not recommend archive.
   - If Engram MCP is enabled and an Engram tool is available, verify the sync/delta trace record for `<name>` exists. If Engram is enabled but unavailable, report that traceability limitation explicitly.

   ```bash
   openspec validate "<name>"
   openspec validate --all
   ```

   If validation commands are unavailable or fail for unrelated reasons, report the exact limitation and any static checks performed.

## Guardrails

- Do not archive changes; `/opsx-archive` is separate.
- Do not silently normalize ambiguous deltas.
- Do not apply deltas that lack required markers or testable scenarios.
- Do not overwrite unrelated main spec content.
- Keep output focused on changed capabilities and validation evidence.
