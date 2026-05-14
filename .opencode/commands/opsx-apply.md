---
description: Implement approved OpenSpec tasks.
agent: orchestrator
---

Run the concrete OpenSpec apply skill for approved tasks.

## Command behavior

- Verify change artifacts exist before implementation.
- Verify change specs use core v2 delta markers before relying on them for implementation.
- Treat scenarios as testable acceptance criteria: each added or modified requirement needs `WHEN` and `THEN`, with optional `GIVEN`.
- Use `implementer` for bounded code changes.
- Use validación agrupada at the end of a coherent block/proposal; do not run tests constantly unless blocked, risky, or diagnosing.
- Do not create PRs, delegate to `delivery`.
- Preserve repo context: CLI Go lives in `tools/lufy-cli-go`; `scripts/install.sh` is a wrapper estricto with no legacy fallback.
- Current active/focus spec: `install-managed-assets-with-hash-idempotency` (managed assets, SHA-256, manifest, idempotency, backup/restore, structural verify).
- Do not suggest archive for `migrate-installer-to-go-cli` while tasks are incomplete.

## Recommended execution

1. Use skill `openspec-apply-change`.
2. Run the repository's real grouped validation for the block/proposal when appropriate.
