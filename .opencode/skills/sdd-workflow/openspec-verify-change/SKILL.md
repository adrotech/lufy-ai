---
name: openspec-verify-change
description: Verify implementation matches change artifacts. Use when the user wants to validate that implementation is complete, correct, and coherent before archiving.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.2.0"
---

Verify that an implementation matches the change artifacts (specs, tasks, design).

**Input**: Optionally specify a change name. If omitted, check if it can be inferred from conversation context. If vague or ambiguous you MUST prompt for available changes.

**Steps**

1. **If no change name provided, prompt for selection**

   Run `openspec list --json` to get available changes. Use the **AskUserQuestion tool** to let the user select.

   Show changes that have implementation tasks (tasks artifact exists).
    Include the schema used for each change if available.
    Mark changes with incomplete tasks as "(In Progress)".

    **IMPORTANT**: Do NOT guess or auto-select a change. Always let the user choose.

   Repo-specific context:
   - Current active/focus spec is `install-managed-assets-with-hash-idempotency` (managed assets, SHA-256, manifest, idempotency, backup/restore, structural verify).
   - `migrate-installer-to-go-cli` must be reported as not ready for archive while tasks are incomplete.
   - Installer architecture: CLI Go lives at `tools/lufy-cli-go`; `scripts/install.sh` is a wrapper estricto with no legacy fallback.

2. **Check status to understand the schema**
   ```bash
   openspec status --change "<name>" --json
   ```
   Parse the JSON to understand:
   - `schemaName`: The workflow being used (e.g., "spec-driven")
   - Which artifacts exist for this change

3. **Get the change directory and load artifacts**

   ```bash
   openspec instructions apply --change "<name>" --json
   ```

   This returns the change directory and context files. Read all available artifacts from `contextFiles`.

4. **Initialize verification report structure**

   Create a report structure with four dimensions:
   - **Completeness**: Track tasks and spec coverage
   - **Correctness**: Track requirement implementation and scenario coverage
   - **Coherence**: Track design adherence and pattern consistency
   - **Gate State**: Track validation evidence, delivery/sync needs, and whether the verified unit is `validated`, `delivery_pending`, `blocked`, or `closed`

   Each dimension can have CRITICAL, WARNING, or SUGGESTION issues.

5. **Verify Completeness**

   **Task Completion**:
   - If tasks.md exists in contextFiles, read it
   - Parse checkboxes: `- [ ]` (incomplete) vs `- [x]` (complete)
   - Count complete vs total tasks
    - If incomplete tasks exist:
      - Add CRITICAL issue for each incomplete task
      - Recommendation: "Complete task: <description>" or "Mark as done if already implemented"
      - Archive assessment: `blocked`; tasks incompletas are never archivable
   - Treat completed task checkboxes as necessary but not sufficient for `closed` or archive-ready.

   **Spec Coverage**:
   - If delta specs exist in `openspec/changes/<name>/specs/`:
      - Verify every delta spec contains at least one `## ADDED Requirements`, `## MODIFIED Requirements` or `## REMOVED Requirements` section
      - Verify each added or modified requirement has at least one `#### Scenario:` with `WHEN` and `THEN`; `GIVEN` is optional
      - Extract all requirements (marked with "### Requirement:")
      - For each requirement:
        - Search codebase for keywords related to the requirement
        - Assess if implementation likely exists
     - If requirements appear unimplemented:
       - Add CRITICAL issue: "Requirement not found: <requirement name>"
       - Recommendation: "Implement requirement X: <description>"

   **Structural Acceptance**:
   - Extract folder/layer expectations from the user request, proposal, design, tasks and specs.
   - For frontend/fullstack feature-driven changes, verify each affected feature has the requested directories such as `components/`, `pages/` or normalized route directory, `hooks/`, `utils/`, `constants/`, `services/`, `types.ts` and `index.ts` when requested or profile-required.
   - Search for pages, hooks, utilities or constants left in the feature root when the requested structure requires subdirectories.
   - For backend changes, read `.lufy/project.yaml` when available and audit the affected backend surface against `project_profile.surfaces[*].architecture.preferred` and `architecture.structural_expectations`.
   - If required directories/layers are missing, or forbidden root files remain, add a CRITICAL issue and set the gate state to `blocked` or `needs_revision` unless the artifacts include explicit user confirmation that the missing structure is an accepted follow-up.

6. **Verify Correctness**

   **Requirement Implementation Mapping**:
   - For each requirement from delta specs:
     - Search codebase for implementation evidence
     - If found, note file paths and line ranges
     - Assess if implementation matches requirement intent
     - If divergence detected:
       - Add WARNING: "Implementation may diverge from spec: <details>"
       - Recommendation: "Review <file>:<lines> against requirement X"

   **Scenario Coverage**:
   - For each scenario in delta specs (marked with "#### Scenario:"):
     - Check if conditions are handled in code
     - Check if tests exist covering the scenario
      - If scenario appears uncovered:
        - Add WARNING: "Scenario not covered: <scenario name>"
        - Recommendation: "Add test or implementation for scenario: <description>"

    **Sync Readiness**:
    - If delta specs exist, compare their requirement titles against corresponding main specs in `openspec/specs/<capability>/spec.md`
    - If deltas appear unapplied after implementation, add CRITICAL issue for archive readiness: "Run `/opsx-sync <change>` before archive"
    - Do not require archive during verification; only report whether archive is blocked by unsynced specs.
    - If sync remains required, final state is `sync_pending` or `blocked`, not `closed`.

    **Delivery Readiness**:
    - If validation passes but commit, push, PR, issue/project sync, or other Git/GH delivery is required and not explicitly authorized/completed, report `validated` with next state `delivery_pending` or `blocked`.
    - Do not perform delivery; recommend `delivery` only after explicit user authorization.

7. **Verify Coherence**

   **Design Adherence**:
   - If design.md exists in contextFiles:
     - Extract key decisions (look for sections like "Decision:", "Approach:", "Architecture:")
     - Verify implementation follows those decisions
     - If contradiction detected:
       - Add WARNING: "Design decision not followed: <decision>"
       - Recommendation: "Update implementation or revise design.md to match reality"
   - If no design.md: Skip design adherence check, note "No design.md to verify against"

   **Code Pattern Consistency**:
    - Review new code for consistency with project patterns
    - Check file naming, directory structure, coding style
    - Review changed/affected old files or diffs for coherence with initial analysis, dependencies, feedback paths, and expected structure/behavior
    - Verify carried structural acceptance criteria before treating code pattern consistency as satisfied
    - If significant deviations found:
      - Add SUGGESTION: "Code pattern deviation: <details>"
      - Recommendation: "Consider following project pattern: <example>"

8. **Generate Verification Report**

   **Summary Scorecard**:
   ```
   ## Verification Report: <change-name>

   ### Summary
   | Dimension    | Status           |
   |--------------|------------------|
   | Completeness | X/Y tasks, N reqs|
   | Correctness  | M/N reqs covered |
   | Coherence    | Followed/Issues  |
   | Gate State   | validated / delivery_pending / sync_pending / blocked / closed |
   ```

   **Issues by Priority**:

   1. **CRITICAL** (Must fix before archive):
      - Incomplete tasks
      - Missing requirement implementations
      - Missing mandatory structural acceptance or forbidden root files contradicting the requested structure
      - Each with specific, actionable recommendation

   2. **WARNING** (Should fix):
      - Spec/design divergences
      - Missing scenario coverage
      - Each with specific recommendation

   3. **SUGGESTION** (Nice to fix):
      - Pattern inconsistencies
      - Minor improvements
      - Each with specific recommendation

    **Final Assessment**:
    - If CRITICAL issues: "X critical issue(s) found. Fix before archiving."
    - If structural acceptance is missing: "State: `blocked`; complete requested structure or get explicit user confirmation for follow-up."
    - If only warnings and delivery/sync remains: "No critical issues. State: `validated`; next: `delivery_pending`/`sync_pending`."
    - If all validation checks pass but delivery is not authorized/completed: "State: `validated`; next: `delivery_pending` or `blocked` for explicit delivery authorization."
    - If all validation, delivery, sync, and closure gates pass: "State: `closed`; archive may proceed."
    - If verifying `migrate-installer-to-go-cli` and any task remains incomplete: "Archive blocked by repo policy until all tasks are complete."

   Validation preference:
   - Use systemic workflow: verify that implementation reflects initial analysis, interconnections, dependencies, feedback loops, and structure/behavior expectations.
   - Use validación agrupada at the end of a coherent block/proposal after all coherent implementation tasks are complete; do not validate each micro-checkbox.
   - Include tests and coverage in final evidence when real commands exist for the scope; otherwise report the limitation explicitly.
   - Do not run tests constantly during verification unless diagnosing or handling a risky/blocking change.

**Verification Heuristics**

- **Completeness**: Focus on objective checklist items (checkboxes, requirements list)
- **Correctness**: Use keyword search, file path analysis, reasonable inference - don't require perfect certainty
- **Coherence**: Look for glaring inconsistencies, don't nitpick style
- **Systemic fit**: Confirm changed/affected old files align with dependencies and expected behavior without rereading unrelated old files
- **Gate semantics**: Task checkboxes are not enough for archive; assess proportional validation, delivery authorization/execution, sync, and blockers separately
- **False Positives**: When uncertain, prefer SUGGESTION over WARNING, WARNING over CRITICAL
- **Actionability**: Every issue must have a specific recommendation with file/line references where applicable

**Graceful Degradation**

- If only tasks.md exists: verify task completion only, skip spec/design checks
- If tasks + specs exist: verify completeness and correctness, skip design
- If full artifacts: verify all three dimensions
- Always note which checks were skipped and why

**Output Format**

Use clear markdown with:
- Table for summary scorecard
- Grouped lists for issues (CRITICAL/WARNING/SUGGESTION)
- Code references in format: `file.ts:123`
- Specific, actionable recommendations
- No vague suggestions like "consider reviewing"
