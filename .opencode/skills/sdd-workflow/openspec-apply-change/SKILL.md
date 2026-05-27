---
name: openspec-apply-change
description: Implement tasks from an OpenSpec change. Use when the user wants to start implementing, continue implementation, or work through tasks.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.2.0"
---

Implement tasks from an OpenSpec change.

**Input**: Optionally specify a change name. If omitted, check if it can be inferred from conversation context. If vague or ambiguous you MUST prompt for available changes.

**Steps**

1. **Select the change**

   If a name is provided, use it. Otherwise:
   - Infer from conversation context if the user mentioned a change
   - Auto-select if only one active change exists
   - If ambiguous, run `openspec list --json` to get available changes and use the **AskUserQuestion tool** to let the user select

   Always announce: "Using change: <name>" and how to override (e.g., `/opsx-apply <other>`).

2. **Check status to understand the schema**
   ```bash
   openspec status --change "<name>" --json
   ```
   Parse the JSON to understand:
   - `schemaName`: The workflow being used (e.g., "spec-driven")
   - Which artifact contains the tasks (typically "tasks" for spec-driven, check status for others)

3. **Get apply instructions**

   ```bash
   openspec instructions apply --change "<name>" --json
   ```

   This returns:
   - Context file paths (varies by schema - could be proposal/specs/design/tasks or spec/tests/implementation/docs)
   - Progress (total, complete, remaining)
   - Task list with status
   - Dynamic instruction based on current state

   **Handle states:**
   - If `state: "blocked"` (missing artifacts): show message, suggest using openspec-continue-change
   - If `state: "all_done"`: report that task checkboxes are complete, then verify validation/delivery/sync gates before suggesting archive
   - Otherwise: proceed to implementation

4. **Read context files**

   Read the files listed in `contextFiles` from the apply instructions output.
   The files depend on the schema being used:
   - **spec-driven**: proposal, specs, design, tasks
   - Other schemas: follow the contextFiles from CLI output

   For core v2 change specs, confirm delta sections are present before implementation:
   - `## ADDED Requirements`
   - `## MODIFIED Requirements`
   - `## REMOVED Requirements`

   Added and modified requirements must include scenarios with `WHEN` and `THEN`; `GIVEN` is optional.

5. **Show current progress**

   Display:
   - Schema being used
   - Progress: "N/M task checkboxes complete"
   - Remaining tasks overview
   - Dynamic instruction from CLI
   - Repo context when relevant: CLI Go is in `tools/lufy-cli-go`; `scripts/install.sh` is a wrapper estricto and must not use legacy fallback paths.
   - Active/focus spec context: `install-managed-assets-with-hash-idempotency` covers managed assets, SHA-256, manifest, idempotency, backup/restore, and structural verify.
   - Task/block gate context: micro-checkboxes are internal progress only; `implemented`, `validated`, `delivery_pending`, `delivered`, and `closed` are distinct states.

6. **Implement tasks (loop until done or blocked)**

   For each pending task:
   - Show which task is being worked on
   - Make the code changes required
   - Keep changes minimal and focused
   - Mark task complete in the tasks file: `- [ ]` → `- [x]` only when the coherent implementation task/block is actually done
   - Continue to next task
   - If implementation changes requirement behavior, update the change delta spec rather than editing main specs directly; use `/opsx-sync` after validation to apply deltas.

   Validation preference:
   - Use systemic workflow: analyze old files, dependencies, interconnections, feedback paths, and structure/behavior risks at the beginning of the coherent block/proposal.
   - For planning-only or OpenSpec/docs-only micro-slices that touch 1-2 artifacts and have no runtime/app changes, use the fast path: read only the relevant OpenSpec/docs files, apply the bounded edit, and validate with `openspec validate "<change>" --strict` plus static checkbox/file review when applicable.
   - Do not reread old files repeatedly during normal implementation after the initial analysis.
   - Reread old files only if they were modified/affected, conflict with changes, new evidence appears, scope changes, a blocker appears, or risk requires confirmation.
   - Use validación agrupada at the end of all tasks in a coherent block/proposal, including tests/coverage only when real commands exist for the scope.
   - Do not run tests constantly during normal implementation.
   - Run focused validation earlier only for blockers, risky changes, feedback loops, or failure diagnosis.
   - Finishing edits usually means `implemented`; use `validated` only with proportional evidence, and never report `closed` or archive-ready solely from task checkboxes.

   **Pause if:**
   - Task is unclear → ask for clarification
   - Implementation reveals a design issue → suggest updating artifacts
   - Error or blocker encountered → report and wait for guidance
   - User interrupts

7. **On completion or pause, show status**

   Display:
   - Tasks completed this session
   - Overall progress: "N/M task checkboxes complete"
   - If all done: report `implemented` or validation-pending state and recommend `/opsx-verify`; suggest archive only after validation, sync, and delivery gates are resolved
   - If paused: explain why and wait for guidance

**Output During Implementation**

```
## Implementing: <change-name> (schema: <schema-name>)

Working on task 3/7: <task description>
[...implementation happening...]
✓ Task complete

Working on task 4/7: <task description>
[...implementation happening...]
✓ Task complete
```

**Output On Completion**

```
## Implementation Complete

**Change:** <change-name>
**Schema:** <schema-name>
**Progress:** 7/7 task checkboxes complete ✓
**State:** implemented / validation pending unless evidence proves otherwise

### Completed This Session
- [x] Task 1
- [x] Task 2
...

Task checkboxes complete. Run `/opsx-verify <change-name>` and resolve validation, sync, and delivery gates before archive.
```

**Output On Pause (Issue Encountered)**

```
## Implementation Paused

**Change:** <change-name>
**Schema:** <schema-name>
**Progress:** 4/7 task checkboxes complete

### Issue Encountered
<description of the issue>

**Options:**
1. <option 1>
2. <option 2>
3. Other approach

What would you like to do?
```

**Guardrails**
- Keep going through tasks until done or blocked
- Always read context files before starting (from the apply instructions output)
- Perform the initial systemic analysis once before implementation, then avoid repeated old-file rereads unless justified by modification, conflict, blocker, new evidence, scope change, or explicit risk.
- Use the planning/OpenSpec-only fast path when the assigned slice is documentation-only, bounded to 1-2 artifacts, and acceptance criteria are already clear; do not request `explorer` only to formalize an already sufficient handoff.
- If task is ambiguous, pause and ask before implementing
- If implementation reveals issues, pause and suggest artifact updates
- Keep code changes minimal and scoped to each task
- Update task checkbox immediately after completing each task
- Before final validation, review changed/affected old files or diffs for coherence with the initial analysis.
- Pause on errors, blockers, or unclear requirements - don't guess
- Use contextFiles from CLI output, don't assume specific file names
- Do not suggest archive for `migrate-installer-to-go-cli` while tasks are incomplete; tasks incompletas mean `blocked`, no archive.

**Fluid Workflow Integration**

This skill supports the "actions on a change" model:

- **Can be invoked anytime**: Before all artifacts are done (if tasks exist), after partial implementation, interleaved with other actions
- **Allows artifact updates**: If implementation reveals design issues, suggest updating artifacts - not phase-locked, work fluidly
