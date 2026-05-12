## 1. Catalog And State Model

- [x] 1.1 Extend `assets.Policy` with stable string values for `managed`, `no-replace`, `merge-block`, `merge-json` and `metadata`.
- [x] 1.2 Add catalog scope metadata for project/global/both without changing project-only behavior yet.
- [x] 1.3 Extend `state.AssetState` and install-state read/write paths with `policy`, `scope` and ancestor metadata.
- [x] 1.4 Add silent migration tests for current install-state JSON into the new state model.
- [x] 1.5 Update catalog and state JSON tests to reject unknown policies/scopes with actionable errors.

## 2. Policy-Driven Planning

- [x] 2.1 Refactor install planning to resolve actions by policy before apply.
- [x] 2.2 Refactor sync planning to resolve actions by policy before apply.
- [x] 2.3 Preserve current blocking behavior for `managed` assets with local drift.
- [x] 2.4 Implement `no-replace` planning for clean updates and drifted `.lufy-new` output.
- [x] 2.5 Keep `merge-json` behavior for `opencode.json` compatible with existing tests and state exclusions.

## 3. Ancestors And Lufy-New Writes

- [x] 3.1 Add safe ancestor path mapping under `.lufy-ai/ancestors/` with traversal tests.
- [x] 3.2 Record ancestor content after successful install/sync writes for drift-resolvable assets.
- [x] 3.3 Implement atomic `.lufy-new` writes for `no-replace` drift cases.
- [x] 3.4 Add verify/status reporting for existing `.lufy-new` files and recommended next action.
- [x] 3.5 Ensure backups still run before original target overwrites and do not treat `.lufy-new` as destructive target mutation.

## 4. MergeBlock For AGENTS.md

- [x] 4.1 Convert `AGENTS.md.template` to lufy-managed blocks using `<!-- LUFY:BEGIN <id> -->` and `<!-- LUFY:END <id> -->` markers.
- [x] 4.2 Implement stdlib-only merge-block parser and renderer with duplicate, nested and unclosed marker detection.
- [x] 4.3 Integrate `merge-block` into install planning/apply for `AGENTS.md`.
- [x] 4.4 Integrate `merge-block` into sync planning/apply for `AGENTS.md`.
- [x] 4.5 Synchronize embedded assets after template changes and keep asset parity tests green.

## 5. Scope-Aware Operations

- [x] 5.1 Add `--scope=project|global|both` parsing and validation for relevant CLI commands.
- [x] 5.2 Resolve global OpenCode config root in a testable way that can use a temporary HOME/XDG config during tests.
- [x] 5.3 Apply scope metadata in install and sync plan generation.
- [x] 5.4 Update verify/status to report the effective scope and root paths.
- [x] 5.5 Add tests proving `--scope=project` preserves current behavior.

## 6. Merge And Restore UX

- [x] 6.1 Add `lufy-ai merge <path>` command dispatch and help text.
- [x] 6.2 Validate merge prerequisites: target file, ancestor and `.lufy-new` must exist and be confined to the target.
- [x] 6.3 Invoke configurable merge tool via `LUFY_MERGE_TOOL`, preserving files if the tool fails.
- [x] 6.4 Add restore backup list mode with IDs, timestamps and manifest paths.
- [x] 6.5 Preserve existing `restore --backup <manifest-or-dir>` behavior while adding ID lookup.

## 7. Documentation And Migration

- [x] 7.1 Document Drift Resolution behavior in CLI docs and user-facing install/sync/verify guidance.
- [x] 7.2 Add migration notes for install-state schema and scope behavior.
- [x] 7.3 Update `docs/status.md` and `docs/roadmap.md` only where implementation status changes.
- [x] 7.4 Decide and document the release default for `--scope` before marking the proposal complete.

## 8. Validation And Delivery

- [x] 8.1 Run targeted Go tests for changed packages while implementing policy/state/merge behavior.
- [x] 8.2 Run `gofmt` on changed Go files.
- [x] 8.3 Run `tools/lufy-cli-go/scripts/smoke-install.sh` after installer behavior changes.
- [x] 8.4 Run `scripts/validate.sh` from repository root after all tasks are complete.
- [x] 8.5 Run `git diff --check origin/develop` before delivery.
- [x] 8.6 Validate OpenSpec status for `resolve-install-drift-policies` before delivery.
