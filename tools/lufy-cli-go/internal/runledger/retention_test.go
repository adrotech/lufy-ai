package runledger

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPruneDryRunAndApplyPreserveActiveRuns(t *testing.T) {
	target := t.TempDir()
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	clock := now.Add(-60 * 24 * time.Hour)
	store, err := NewFileStore(target, Options{Now: func() time.Time { return clock }})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	appendRunState(t, store, ctx, "run-old-terminal", "closed", KindFinish)
	appendRunState(t, store, ctx, "run-old-active", "implemented", KindCheckpoint)
	clock = now.Add(-24 * time.Hour)
	appendRunState(t, store, ctx, "run-recent-terminal", "closed", KindFinish)

	policy := RetentionPolicy{MaxAge: 30 * 24 * time.Hour, MaxTerminalRuns: 10, MaxBytes: 1 << 30}
	dryRun, err := store.Prune(ctx, policy, now, false)
	if err != nil {
		t.Fatal(err)
	}
	if !dryRun.DryRun || len(dryRun.Candidates) != 1 || dryRun.Candidates[0].RunID != "run-old-terminal" {
		t.Fatalf("dry-run report = %#v", dryRun)
	}
	if len(dryRun.ProtectedActive) != 1 || dryRun.ProtectedActive[0] != "run-old-active" {
		t.Fatalf("protected active = %#v", dryRun.ProtectedActive)
	}
	if _, err := os.Stat(filepath.Join(store.runsRoot, "run-old-terminal")); err != nil {
		t.Fatalf("dry-run deleted terminal run: %v", err)
	}

	applied, err := store.Prune(ctx, policy, now, true)
	if err != nil {
		t.Fatal(err)
	}
	if applied.DryRun || len(applied.Deleted) != 1 || applied.Deleted[0] != "run-old-terminal" || applied.ReclaimedBytes == 0 {
		t.Fatalf("applied report = %#v", applied)
	}
	if _, err := os.Stat(filepath.Join(store.runsRoot, "run-old-terminal")); !os.IsNotExist(err) {
		t.Fatalf("old terminal run remains, err=%v", err)
	}
	for _, runID := range []string{"run-old-active", "run-recent-terminal"} {
		if _, err := os.Stat(filepath.Join(store.runsRoot, runID)); err != nil {
			t.Fatalf("protected run %s missing: %v", runID, err)
		}
	}
}

func TestPruneUsesCountLimitDeterministically(t *testing.T) {
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	clock := now.Add(-3 * time.Hour)
	store, err := NewFileStore(t.TempDir(), Options{Now: func() time.Time { return clock }})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for i, runID := range []string{"run-first", "run-second", "run-third"} {
		clock = now.Add(time.Duration(i-3) * time.Hour)
		appendRunState(t, store, ctx, runID, "closed", KindFinish)
	}
	report, err := store.Prune(ctx, RetentionPolicy{MaxAge: 365 * 24 * time.Hour, MaxTerminalRuns: 1, MaxBytes: 1 << 30}, now, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Candidates) != 2 || report.Candidates[0].RunID != "run-first" || report.Candidates[1].RunID != "run-second" {
		t.Fatalf("count candidates = %#v", report.Candidates)
	}
}

func appendRunState(t *testing.T, store *FileStore, ctx context.Context, runID, status string, kind EventKind) {
	t.Helper()
	draft := validDraft(runID, kind)
	draft.Checkpoint = &Checkpoint{Status: status, Gate: "implementation"}
	if _, err := store.Append(ctx, AppendRequest{Draft: draft, IdempotencyKey: runID + ":state"}); err != nil {
		t.Fatal(err)
	}
}
