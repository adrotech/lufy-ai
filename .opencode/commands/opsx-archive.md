---
description: Archive a completed OpenSpec change.
agent: orchestrator
---

Archive completed change and update tracking.

## Command behavior

- Verify all tasks are done.
- If tasks are incomplete, return `blocked`; tasks incompletas are not archivable and must not be overridden.
- `migrate-installer-to-go-cli` is explicitly blocked from archive while any task remains incomplete.
- Run final grouped validation evidence when available; do not run tests constantly.
- Update artifacts.
- Call project sync with status Done.
- Preserve repo context when relevant: CLI Go lives in `tools/lufy-cli-go`; `scripts/install.sh` is a wrapper estricto with no legacy fallback.
- Current active/focus spec: `install-managed-assets-with-hash-idempotency`.

## Recommended execution

1. Verify 100% tasks complete
2. Use skill `openspec-archive-change` only after completion gates pass
3. If a downstream project installed `project-sync`, run `skill project-sync --change <name> --status Done`
