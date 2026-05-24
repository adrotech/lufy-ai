## 1. Command and Skill Skeleton

- [x] 1.1 Create `.opencode/commands/lufy.timereport.md` as a Spanish slash-command wrapper that documents inputs, default output path and delegation to the `lufy.timereport` skill.
- [x] 1.2 Create `.opencode/skills/lufy.timereport/SKILL.md` with purpose, usage, required privacy boundaries, data sources, degradation rules, output contract and validation guidance.
- [x] 1.3 Add any skill-local templates or references needed for the HTML report without introducing remote assets or network dependencies.

## 2. Data Collection Design in Skill

- [x] 2.1 Define the read-only OpenCode SQLite query strategy for `project`, `workspace`, `session`, `message`, `part` and `event`, including repository-directory filtering and schema-mismatch handling.
- [x] 2.2 Define Git read-only collection for commit count and net LOC in the selected report range, including non-Git degradation.
- [x] 2.3 Define optional `.opencode/project.yaml` stack loading and fallback heuristic/`No configurado` behavior when the file is absent or invalid.
- [x] 2.4 Explicitly exclude JSONL and `session_diff` from default data collection, noting future opt-in only if later specified.

## 3. Metrics and Heuristics

- [x] 3.1 Implement or document deterministic calculations for wall-clock time from included activity bounds.
- [x] 3.2 Implement or document AI working time estimation from assistant/tool/event intervals with documented gap caps.
- [x] 3.3 Implement or document human active time estimation from user-interaction intervals with documented gap caps.
- [x] 3.4 Implement or document aggregation for tool calls, top tools, subagents, skills and phase timeline without using prompts, outputs or payload bodies.

## 4. HTML Report Generation

- [x] 4.1 Generate a single self-contained HTML file with embedded styles and no remote resources.
- [x] 4.2 Include required sections: wall-clock, AI working time, human active time, net LOC, commits, tool calls, top tools, subagents, skills, phases, stack and limitations.
- [x] 4.3 Support default temporary output path and explicit user-provided output path with clear success/failure reporting.
- [x] 4.4 Include a visible methodology/limitations section describing sources, missing-source degradations and time heuristics.

## 5. Privacy and Safety Guards

- [x] 5.1 Enforce an allowlist of report fields so prompts, assistant outputs, full tool arguments, tool outputs, file contents, diffs and `session_diff` are not emitted by default.
- [x] 5.2 Ensure all data-source operations are local/read-only and do not mutate OpenCode DB, Git history or `.opencode/project.yaml`.
- [x] 5.3 Add actionable error/degradation messages for missing DB, missing Git, missing/invalid project config and unavailable timestamps.

## 6. Validation

- [x] 6.1 Add sanitized or synthetic fixtures for OpenCode-like sessions/messages/parts/events that cover required metrics without sensitive content.
- [x] 6.2 Add degradation fixtures or scenarios for missing OpenCode DB, non-Git directory and missing/invalid `.opencode/project.yaml`.
- [x] 6.3 Validate that generated HTML contains required metric sections and does not contain fixture prompts, outputs, diffs or full tool payloads.
- [x] 6.4 Run `openspec validate add-lufy-timereport --strict` and the available project validation commands appropriate to the implemented files.
- [x] 6.5 Document exact validation commands and results in the implementation handoff without marking delivery complete.
