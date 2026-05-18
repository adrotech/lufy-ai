---
description: Archive a completed OpenSpec change.
agent: orchestrator
---

Archive completed change and update tracking.

## Command behavior

- Verify all tasks are done, then verify closure gates; task checkboxes alone are not archive-ready.
- Verify delta specs have already been synced into `openspec/specs/` with `/opsx-sync`; if not, return `blocked` instead of archiving.
- If tasks are incomplete, return `blocked`; tasks incompletas are not archivable and must not be overridden.
- If validation, explicit delivery authorization/execution, issue/project sync, or any block gate remains unresolved, return `blocked`, `delivery_pending`, or `sync_pending` instead of archiving.
- `migrate-installer-to-go-cli` is explicitly blocked from archive while any task remains incomplete.
- Run final grouped validation evidence when available; do not run tests constantly.
- Update artifacts.
- Call project sync with status Done.
- Preserve repo context when relevant: CLI Go lives in `tools/lufy-cli-go`; `scripts/install.sh` is a wrapper estricto with no legacy fallback.
- Current active/focus spec: `install-managed-assets-with-hash-idempotency`.

## Recommended execution

1. Verify 100% task checkboxes complete
2. Verify proportional validation and delivery/sync closure gates are satisfied or explicitly not required
3. Verify specs are synced or run `/opsx-sync <change>` first
4. Use skill `openspec-archive-change` only after closure gates pass
5. If a downstream project installed `project-sync`, run `skill project-sync --change <name> --status Done`
