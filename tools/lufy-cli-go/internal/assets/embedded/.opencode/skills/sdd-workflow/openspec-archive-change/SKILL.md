---
name: openspec-archive-change
description: Archive a completed change in the experimental workflow. Use when the user wants to finalize and archive a change after implementation is complete.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.2.0"
---

Archive a completed change in the experimental workflow.

**Input**: Optionally specify a change name. If omitted, check if it can be inferred from conversation context. If vague or ambiguous you MUST prompt for available changes.

**Steps**

1. **If no change name provided, prompt for selection**

   Run `openspec list --json` to get available changes. Use the **AskUserQuestion tool** to let the user select.

   Show only active changes (not already archived).
   Include the schema used for each change if available.

    **IMPORTANT**: Do NOT guess or auto-select a change. Always let the user choose.

   Repo-specific context:
   - Current active/focus spec is `install-managed-assets-with-hash-idempotency` (managed assets, SHA-256, manifest, idempotency, backup/restore, structural verify).
   - `migrate-installer-to-go-cli` must not be archived while any tasks remain incomplete.
   - Installer architecture: CLI Go lives at `tools/lufy-cli-go`; `scripts/install.sh` is a wrapper estricto with no legacy fallback.

2. **Check artifact completion status**

   Run `openspec status --change "<name>" --json` to check artifact completion.

   Parse the JSON to understand:
   - `schemaName`: The workflow being used
   - `artifacts`: List of artifacts with their status (`done` or other)

   **If any artifacts are not `done`:**
   - Display warning listing incomplete artifacts
   - Return `blocked` with the exact incomplete artifacts and required recovery

3. **Check task completion status**

   Read the tasks file (typically `tasks.md`) to check for incomplete tasks.

   Count tasks marked with `- [ ]` (incomplete) vs `- [x]` (complete).

   **If incomplete tasks found:**
   - Display warning showing count of incomplete tasks
   - Return `blocked`; tasks incompletas are not archivable
   - If the change is `migrate-installer-to-go-cli`, explicitly state the repo policy block: no archive until all tasks are complete

   **If no tasks file exists:** Return `blocked` unless the schema explicitly has no tasks artifact.

4. **Assess delta spec sync state**

   Check for delta specs at `openspec/changes/<name>/specs/`. If none exist, proceed without sync prompt.

   **If delta specs exist:**
   - Compare each delta spec with its corresponding main spec at `openspec/specs/<capability>/spec.md`
   - Determine what changes would be applied (adds, modifications, removals, renames)
   - Show a combined summary before prompting
   - If changes are still needed, return `blocked` and instruct the user to run `/opsx-sync <name>`; do not archive unsynced deltas

   **Prompt options:**
   - If already synced: "Archive now", "Sync anyway", "Cancel"
   - If changes are needed: "Run /opsx-sync now", "Cancel"

   If the user chooses sync, use the installed concrete sync skill and then re-check all artifact, task and sync gates before archive.

5. **Perform the archive**

   Create the archive directory if it doesn't exist:
   ```bash
   mkdir -p openspec/changes/archive
   ```

   Generate target name using current date: `YYYY-MM-DD-<change-name>`

   **Check if target already exists:**
   - If yes: Fail with error, suggest renaming existing archive or using different date
   - If no: Move the change directory to archive

   ```bash
   mv openspec/changes/<name> openspec/changes/archive/YYYY-MM-DD-<name>
   ```

6. **Display summary**

   Show archive completion summary including:
   - Change name
   - Schema that was used
   - Archive location
   - Whether specs were synced (if applicable)
   - Note that incomplete artifacts/tasks would have blocked archive

**Output On Success**

```
## Archive Complete

**Change:** <change-name>
**Schema:** <schema-name>
**Archived to:** openspec/changes/archive/YYYY-MM-DD-<name>/
**Specs:** ✓ Synced to main specs (or "No delta specs" or "Sync skipped")

All artifacts complete. All tasks complete.
```

**Guardrails**
- Always prompt for change selection if not provided
- Use artifact graph (openspec status --json) for completion checking
- Block archive on incomplete artifacts/tasks; do not override this with user confirmation
- Preserve .openspec.yaml when moving to archive (it moves with the directory)
- Show clear summary of what happened
- If sync is requested, use available concrete OpenSpec sync tooling; do not reference generic workflow modes.
- If delta specs exist, always run the sync assessment and show the combined summary before prompting
- Never archive a change with unsynced delta specs, even if the user asks to skip sync.
- Use validación agrupada evidence from the completed block/proposal; do not run tests constantly solely for archive.
