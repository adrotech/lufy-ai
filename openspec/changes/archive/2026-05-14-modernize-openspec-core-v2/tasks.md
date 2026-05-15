## 1. Core Assets

- [x] 1.1 Define `openspec/UPSTREAM.json` with baseline version, profile and source metadata.
- [x] 1.2 Update `openspec/config.yaml` to the core v2 action-based schema.
- [x] 1.3 Add `/opsx-sync` command asset under `.opencode/commands/`.
- [x] 1.4 Add `openspec-sync` skill asset under `.opencode/skills/sdd-workflow/`.
- [x] 1.5 Add `opsx-version` command asset that reports effective baseline metadata from `UPSTREAM.json`.

## 2. Workflow Enforcement

- [x] 2.1 Update OpenSpec proposal/apply/verify/archive guidance to require delta markers for change specs.
- [x] 2.2 Update workflow guidance to require testable scenarios with `WHEN`/`THEN` and optional `GIVEN`.
- [x] 2.3 Implement `/opsx-sync` behavior to validate deltas and apply them to main specs without archiving.
- [x] 2.4 Ensure archive guidance requires synced specs before moving a change to archive.

## 3. Installer Integration

- [x] 3.1 Add new core v2 assets to the managed catalog if explicit catalog entries are needed.
- [x] 3.2 Sync all new or changed OpenSpec core v2 assets into `tools/lufy-cli-go/internal/assets/embedded/`.
- [x] 3.3 Update verify/status behavior only where needed so missing core v2 baseline assets are reported clearly.
- [x] 3.4 Add or adjust tests that prove root and embedded assets stay in parity.

## 4. Documentation And Migration

- [x] 4.1 Document the core v2 workflow in project docs without claiming expanded/stay-updated features.
- [x] 4.2 Add migration notes for targets moving from `v0.2.0` assets to core v2 assets.
- [x] 4.3 Update README/getting-started only for behavior implemented by this change.

## 5. Validation

- [x] 5.1 Run `openspec validate modernize-openspec-core-v2`.
- [x] 5.2 Run `openspec validate --all`.
- [x] 5.3 Run `scripts/validate.sh`.
- [x] 5.4 Run `git diff --check origin/develop`.
- [x] 5.5 Run sandbox smokes for greenfield install and brownfield sync from `v0.2.0` with customized `AGENTS.md`.
