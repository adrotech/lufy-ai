---
description: Verify implementation matches change artifacts.
agent: orchestrator
---

Verify implementation against OpenSpec change artifacts.

## Command behavior

- Check artifacts exist (proposal, design, tasks, specs).
- Verify delta specs use `ADDED`, `MODIFIED` or `REMOVED` sections.
- Verify scenarios are testable with `WHEN` and `THEN`; `GIVEN` is optional.
- Report unsynced delta specs as blockers for archive and recommend `/opsx-sync <change>`.
- Run grouped validation evidence for the block/proposal when applicable; avoid constant tests unless blocked, risky, or diagnosing.
- Evaluate gate state at task/block/review-slice scope, not per micro-checkbox.
- If validation passes but Git/GH delivery or sync remains, report `validated` with `delivery_pending`, `sync_pending`, or `blocked`; do not perform delivery.
- Report findings.
- Report incomplete tasks as archive blockers.
- `migrate-installer-to-go-cli` must be `blocked` for archive while tasks are incomplete.
- Current active/focus spec: `install-managed-assets-with-hash-idempotency` (managed assets, SHA-256, manifest, idempotency, backup/restore, structural verify).
- Installer context: CLI Go lives in `tools/lufy-cli-go`; `scripts/install.sh` is a wrapper estricto with no legacy fallback.

## Recommended execution

- Use skill `openspec-verify-change` for artifact/implementation verification.
- Use `validator` for command validation evidence when needed.
