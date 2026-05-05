## 1. CLI Surface

- [x] 1.1 Inspect `tools/lufy-cli-go` command parsing and identify existing shared command/flag patterns for `install`, `verify`, `backup` and `restore`.
- [x] 1.2 Add `lufy-ai sync` help/usage with `--target`, `--dry-run`, `--yes` and `--no-engram` semantics aligned to existing commands.
- [x] 1.3 Route `sync` from `cmd/lufy-ai/main.go` to internal packages without placing planning/copy logic in `main.go`.

## 2. Sync Planning

- [x] 2.1 Reuse the managed asset catalog and safe source/target path resolution for sync.
- [x] 2.2 Load and validate `.lufy-ai/install-state.json` for sync, treating absent, corrupt or unsupported state as blocking for unsafe overwrites.
- [x] 2.3 Compute sync actions from source current hash, target current hash, recorded source hash and recorded target hash.
- [x] 2.4 Classify unchanged managed assets as `skip` without rewriting content or changing timestamps.
- [x] 2.5 Classify upstream-changed assets without local drift as `backup` plus `update-managed`.
- [x] 2.6 Classify local drift, non-managed existing files and unsafe path/symlink escapes as `conflict` or actionable failure without mutation.
- [x] 2.7 Ensure `--dry-run` uses the real planner and performs zero filesystem mutations, including backups and state repair.
- [x] 2.8 Classify previously managed assets removed from the current source catalog as `retired`, preserving target content and state when hashes still match, or blocking on drift.

## 3. Sync Apply And State

- [x] 3.1 Apply sync plans only after conflict checks and required confirmation behavior are satisfied.
- [x] 3.2 Create a multiasset backup under `.lufy-ai/backups/<timestamp>/` before any sync `update-managed` writes.
- [x] 3.3 Write backup `manifest.json` with relative paths, previous hashes, action cause `sync`, capture status and backup locations.
- [x] 3.4 Copy updated managed assets from source to target only for safe `update-managed` actions.
- [x] 3.5 Update `.lufy-ai/install-state.json` after successful sync with refreshed source and target hashes for synchronized assets.
- [x] 3.6 Report backup manifest path and restore guidance if a sync mutation fails after backup creation.

## 4. Preservation And Integration

- [x] 4.1 Verify sync does not copy or register source files outside the managed asset catalog.
- [x] 4.2 Verify sync does not delete, modify or register target files outside managed catalog scope.
- [x] 4.3 Keep `scripts/install.sh` as a strict wrapper for `lufy-ai install` with no sync fallback or legacy asset-copy logic.
- [x] 4.4 Ensure sync output includes explainable actions with relative paths, reasons and hashes where available.

## 5. Tests And Validation

- [x] 5.1 Add or update Go tests for sync planning: skip unchanged, update upstream-changed/no-drift, conflict on local drift and conflict on untracked existing file.
- [x] 5.2 Add or update Go tests for dry-run no-mutation behavior, including no backups and no state writes.
- [x] 5.3 Add or update Go tests for backup manifest contents and state hash updates after successful sync.
- [x] 5.4 Add or update Go tests for target confinement and symlink/path escape rejection during sync.
- [x] 5.5 Add or update Go tests for retired managed assets that disappeared from the source catalog.
- [x] 5.6 Run `go test ./...` from `tools/lufy-cli-go/` and record the result.
- [x] 5.7 Run `go build ./cmd/lufy-ai` from `tools/lufy-cli-go/` and record the result.
- [x] 5.8 Run a temp-dir `install`, `sync --dry-run`, real `sync` and `verify --target <temp>` smoke flow when the compiled CLI is available, and record the result.
