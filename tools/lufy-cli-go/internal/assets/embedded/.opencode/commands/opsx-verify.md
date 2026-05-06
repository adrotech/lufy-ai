---
description: Verify implementation matches change artifacts.
agent: orchestrator
---

Verify implementation against OpenSpec change artifacts.

## Command behavior

- Check artifacts exist (proposal, design, tasks, specs).
- Run grouped validation evidence for the block/proposal when applicable; avoid constant tests unless blocked, risky, or diagnosing.
- Report findings.
- Report incomplete tasks as archive blockers.
- `migrate-installer-to-go-cli` must be `blocked` for archive while tasks are incomplete.
- Current active/focus spec: `install-managed-assets-with-hash-idempotency` (managed assets, SHA-256, manifest, idempotency, backup/restore, structural verify).
- Installer context: CLI Go lives in `tools/lufy-cli-go`; `scripts/install.sh` is a wrapper estricto with no legacy fallback.

## Recommended execution

- Use skill `openspec-verify-change` for artifact/implementation verification.
- Use `validator` for command validation evidence when needed.
