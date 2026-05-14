---
description: Sync validated OpenSpec change deltas into main specs.
agent: orchestrator
---

Run the concrete OpenSpec sync skill.

## Command behavior

- Resolve change name from command argument or active change context.
- Validate the change before mutating main specs: run `openspec validate <change>` when available.
- Require change specs under `openspec/changes/<change>/specs/` to use delta sections `## ADDED Requirements`, `## MODIFIED Requirements` or `## REMOVED Requirements`.
- Require each added or modified requirement to include at least one `#### Scenario:` with `WHEN` and `THEN`; `GIVEN` is optional for setup context.
- Apply validated deltas to matching main specs under `openspec/specs/<capability>/spec.md` without moving the change to archive.
- If any delta is ambiguous, conflicting or invalid, stop with an actionable error and do not mutate main specs.
- Preserve repo context: CLI Go lives in `tools/lufy-cli-go`; `scripts/install.sh` is a wrapper estricto with no legacy fallback.

## Recommended execution

1. Use skill `openspec-sync`.
2. After sync, use `/opsx-verify <change>` before `/opsx-archive <change>`.
