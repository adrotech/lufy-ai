package runledger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestFileStoreAppendIsIdempotentAndConflictsDoNotOverwrite(t *testing.T) {
	store := newTestStore(t, Options{})
	ctx := context.Background()
	draft := validDraft("run-root", KindCheckpoint)

	first, err := store.Append(ctx, AppendRequest{Draft: draft, IdempotencyKey: "codex:turn-1:stop"})
	if err != nil {
		t.Fatalf("first Append() error = %v", err)
	}
	if first.Status != AppendRecorded {
		t.Fatalf("first status = %q", first.Status)
	}

	retry, err := store.Append(ctx, AppendRequest{Draft: draft, IdempotencyKey: "codex:turn-1:stop"})
	if err != nil {
		t.Fatalf("retry Append() error = %v", err)
	}
	if retry.Status != AppendDuplicateNoop || retry.Event.EventID != first.Event.EventID {
		t.Fatalf("retry = %#v, first = %#v", retry, first)
	}

	draft.Checkpoint.Status = "blocked"
	conflict, err := store.Append(ctx, AppendRequest{Draft: draft, IdempotencyKey: "codex:turn-1:stop"})
	if !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("conflict error = %v", err)
	}
	if conflict.Status != AppendConflict || conflict.Event.EventID != first.Event.EventID {
		t.Fatalf("conflict = %#v", conflict)
	}

	events, err := store.LoadRun(ctx, "run-root")
	if err != nil {
		t.Fatalf("LoadRun() error = %v", err)
	}
	if len(events) != 1 || events[0].Checkpoint.Status != "implemented" {
		t.Fatalf("events after conflict = %#v", events)
	}
}

func TestFileStoreAssignsCausalLamportClock(t *testing.T) {
	store := newTestStore(t, Options{})
	ctx := context.Background()

	root, err := store.Append(ctx, AppendRequest{
		Draft:          validDraft("run-root", KindStart),
		IdempotencyKey: "root:start",
	})
	if err != nil {
		t.Fatal(err)
	}
	childDraft := validDraft("run-child", KindStart)
	childDraft.ParentRunID = "run-root"
	childDraft.CausedByEventID = root.Event.EventID
	child, err := store.Append(ctx, AppendRequest{
		Draft:           childDraft,
		IdempotencyKey:  "child:start",
		ProposedLamport: root.Event.LamportClock,
	})
	if err != nil {
		t.Fatal(err)
	}
	if child.Event.LamportClock != root.Event.LamportClock+1 {
		t.Fatalf("child clock = %d, root = %d", child.Event.LamportClock, root.Event.LamportClock)
	}
	if child.Event.LocalSequence != 1 || child.Event.ParentRunID != "run-root" {
		t.Fatalf("child causal metadata = %#v", child.Event)
	}
}

func TestFileStoreRejectsMissingParentRun(t *testing.T) {
	store := newTestStore(t, Options{})
	draft := validDraft("run-orphan-child", KindStart)
	draft.ParentRunID = "run-missing-parent"

	_, err := store.Append(context.Background(), AppendRequest{
		Draft:          draft,
		IdempotencyKey: "orphan-child:start",
	})
	if err == nil || !strings.Contains(err.Error(), "parent_run_id") {
		t.Fatalf("Append() error = %v", err)
	}
}

func TestFileStoreSerializesConcurrentWriters(t *testing.T) {
	store := newTestStore(t, Options{LockTimeout: 3 * time.Second})
	ctx := context.Background()
	const writers = 24
	var wg sync.WaitGroup
	errCh := make(chan error, writers)

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			draft := validDraft("run-concurrent", KindEvidence)
			draft.TaskRef = fmt.Sprintf("task-%02d", i)
			_, err := store.Append(ctx, AppendRequest{
				Draft:          draft,
				IdempotencyKey: fmt.Sprintf("writer-%02d", i),
			})
			errCh <- err
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent Append() error = %v", err)
		}
	}

	events, err := store.LoadRun(ctx, "run-concurrent")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != writers {
		t.Fatalf("events = %d, want %d", len(events), writers)
	}
	clocks := make([]int, 0, writers)
	sequences := make([]int, 0, writers)
	for _, event := range events {
		clocks = append(clocks, int(event.LamportClock))
		sequences = append(sequences, int(event.LocalSequence))
	}
	sort.Ints(clocks)
	sort.Ints(sequences)
	for i := 0; i < writers; i++ {
		if clocks[i] != i+1 || sequences[i] != i+1 {
			t.Fatalf("clocks=%v sequences=%v", clocks, sequences)
		}
	}
}

func TestRunLockReleaseRetriesWhileOwnerFileIsInUse(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "lock")
	if err := os.Mkdir(lockPath, 0o700); err != nil {
		t.Fatal(err)
	}
	owner := lockOwner{Token: "evt_release_owner"}
	ownerPath := filepath.Join(lockPath, "owner.json")
	if err := writeJSONAtomic(ownerPath, owner); err != nil {
		t.Fatal(err)
	}

	held, err := os.Open(ownerPath)
	if err != nil {
		t.Fatal(err)
	}
	closed := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = held.Close()
		close(closed)
	}()

	lock := &runLock{
		path:             lockPath,
		token:            owner.Token,
		cleanupTimeout:   time.Second,
		cleanupPollDelay: 5 * time.Millisecond,
	}
	if err := lock.release(); err != nil {
		t.Fatalf("release() with transient owner handle error = %v", err)
	}
	<-closed
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("lock dir remains after release, err=%v", err)
	}
}

func TestFileStoreRecoversEventPublishedBeforeReceipt(t *testing.T) {
	target := t.TempDir()
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	injected := errors.New("receipt unavailable")
	failing, err := NewFileStore(target, Options{
		Now:                func() time.Time { return now },
		BeforeReceiptWrite: func(Event) error { return injected },
	})
	if err != nil {
		t.Fatal(err)
	}
	draft := validDraft("run-crash", KindCheckpoint)
	_, err = failing.Append(context.Background(), AppendRequest{Draft: draft, IdempotencyKey: "crash-key"})
	if !errors.Is(err, injected) {
		t.Fatalf("injected error = %v", err)
	}

	recovered, err := NewFileStore(target, Options{Now: func() time.Time { return now.Add(time.Second) }})
	if err != nil {
		t.Fatal(err)
	}
	result, err := recovered.Append(context.Background(), AppendRequest{Draft: draft, IdempotencyKey: "crash-key"})
	if err != nil {
		t.Fatalf("recovery Append() error = %v", err)
	}
	if result.Status != AppendDuplicateNoop {
		t.Fatalf("recovery status = %q", result.Status)
	}
	report, err := recovered.VerifyRun(context.Background(), "run-crash")
	if err != nil {
		t.Fatal(err)
	}
	if !report.Healthy || len(report.Issues) != 0 {
		t.Fatalf("verification after recovery = %#v", report)
	}
}

func TestVerifyRunDetectsOrphanReceiptWithoutLeakingContent(t *testing.T) {
	store := newTestStore(t, Options{})
	ctx := context.Background()
	result, err := store.Append(ctx, AppendRequest{
		Draft:          validDraft("run-verify", KindEvidence),
		IdempotencyKey: "verify-key",
	})
	if err != nil {
		t.Fatal(err)
	}
	eventPath := filepath.Join(store.runsRoot, "run-verify", "events", eventFilename(result.Event))
	if err := os.Remove(eventPath); err != nil {
		t.Fatal(err)
	}

	report, err := store.VerifyRun(ctx, "run-verify")
	if err != nil {
		t.Fatal(err)
	}
	if report.Healthy || !hasIssue(report, "orphan_receipt") {
		t.Fatalf("report = %#v", report)
	}
	serialized := fmt.Sprintf("%#v", report)
	if strings.Contains(serialized, "verify-key") {
		t.Fatalf("verification leaked idempotency key: %s", serialized)
	}
}

func TestFileStoreRecoversExpiredLeaseWithOwnerMetadata(t *testing.T) {
	target := t.TempDir()
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	store, err := NewFileStore(target, Options{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	runDir := filepath.Join(store.runsRoot, "run-stale")
	lockDir := filepath.Join(runDir, "lock")
	if err := os.MkdirAll(lockDir, 0o700); err != nil {
		t.Fatal(err)
	}
	owner := lockOwner{
		Token:          "evt_expired_lock_owner",
		PID:            999999,
		CreatedAt:      now.Add(-time.Minute),
		LeaseExpiresAt: now.Add(-time.Second),
	}
	if err := writeJSONAtomic(filepath.Join(lockDir, "owner.json"), owner); err != nil {
		t.Fatal(err)
	}

	result, err := store.Append(context.Background(), AppendRequest{
		Draft:          validDraft("run-stale", KindStart),
		IdempotencyKey: "stale-lock-recovery",
	})
	if err != nil {
		t.Fatalf("Append() after stale lock error = %v", err)
	}
	if result.Status != AppendRecorded {
		t.Fatalf("status = %q", result.Status)
	}
	if _, err := os.Stat(lockDir); !os.IsNotExist(err) {
		t.Fatalf("lock dir remains after append, err=%v", err)
	}
}

func TestFileStoreTimesOutOnActiveLeaseEvenWithFixedDomainClock(t *testing.T) {
	target := t.TempDir()
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	store, err := NewFileStore(target, Options{
		Now:              func() time.Time { return now },
		LockTimeout:      40 * time.Millisecond,
		LockPollInterval: 5 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	runDir := filepath.Join(store.runsRoot, "run-active-lock")
	lockDir := filepath.Join(runDir, "lock")
	if err := os.MkdirAll(lockDir, 0o700); err != nil {
		t.Fatal(err)
	}
	owner := lockOwner{
		Token:          "evt_active_lock_owner",
		PID:            os.Getpid(),
		CreatedAt:      now,
		LeaseExpiresAt: now.Add(time.Minute),
	}
	if err := writeJSONAtomic(filepath.Join(lockDir, "owner.json"), owner); err != nil {
		t.Fatal(err)
	}

	started := time.Now()
	_, err = store.Append(context.Background(), AppendRequest{
		Draft:          validDraft("run-active-lock", KindStart),
		IdempotencyKey: "active-lock:start",
	})
	if err == nil || !strings.Contains(err.Error(), "timeout esperando lock") {
		t.Fatalf("Append() error = %v", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("lock timeout took too long: %s", elapsed)
	}
}

func TestFileStoreFailsWithoutPublishingWhenRuntimePathIsUnavailable(t *testing.T) {
	target := t.TempDir()
	store, err := NewFileStore(target, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, ".lufy"), []byte("not-a-directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	const privateKey = "PRIVATE-RETRY-KEY"
	_, err = store.Append(context.Background(), AppendRequest{
		Draft:          validDraft("run-unavailable", KindStart),
		IdempotencyKey: privateKey,
	})
	if err == nil {
		t.Fatal("Append() expected unavailable filesystem error")
	}
	if strings.Contains(err.Error(), privateKey) {
		t.Fatalf("filesystem error leaked idempotency key: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(target, ".lufy", "runtime", "runs", "run-unavailable")); !os.IsNotExist(statErr) && statErr == nil {
		t.Fatal("failed append published a run directory")
	}
}

func TestFileStoreDoesNotPublishIntoReadOnlyRuntime(t *testing.T) {
	target := t.TempDir()
	runtimeDir := filepath.Join(target, ".lufy", "runtime")
	if err := os.MkdirAll(runtimeDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(runtimeDir, 0o500); err != nil {
		t.Skipf("chmod unsupported: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(runtimeDir, 0o700) })
	store, err := NewFileStore(target, Options{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Append(context.Background(), AppendRequest{
		Draft:          validDraft("run-read-only", KindStart),
		IdempotencyKey: "read-only:start",
	})
	if err == nil {
		t.Skip("filesystem does not enforce read-only directory permissions for this process")
	}
	if _, statErr := os.Stat(filepath.Join(store.runsRoot, "run-read-only", "events")); statErr == nil {
		t.Fatal("read-only append published event directory")
	}
}

func TestVerifyRunDetectsInterruptedAtomicFile(t *testing.T) {
	store := newTestStore(t, Options{})
	ctx := context.Background()
	if _, err := store.Append(ctx, AppendRequest{
		Draft:          validDraft("run-partial", KindStart),
		IdempotencyKey: "partial:start",
	}); err != nil {
		t.Fatal(err)
	}
	partial := filepath.Join(store.runsRoot, "run-partial", "events", ".write-interrupted.tmp")
	if err := os.WriteFile(partial, []byte("partial"), 0o600); err != nil {
		t.Fatal(err)
	}

	report, err := store.VerifyRun(ctx, "run-partial")
	if err != nil {
		t.Fatal(err)
	}
	if report.Healthy || !hasIssue(report, "partial_file") {
		t.Fatalf("report = %#v", report)
	}
}

func TestFileStorePersistsOnlyDigestsForExternalReferences(t *testing.T) {
	target := t.TempDir()
	store, err := NewFileStore(target, Options{
		Now: func() time.Time { return time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatal(err)
	}
	const rawSession = "session-private-123"
	const rawPath = "/Users/example/private/source.go"
	const rawKey = "adapter-retry-private-key"
	draft := validDraft("run-private", KindEvidence)
	draft.Source.SessionRefHash = HashReference(rawSession)
	draft.ArtifactRefs = []ArtifactRef{{
		Kind:          "source",
		PathSHA256:    HashReference(rawPath),
		ContentSHA256: HashReference("synthetic-content"),
	}}
	if _, err := store.Append(context.Background(), AppendRequest{Draft: draft, IdempotencyKey: rawKey}); err != nil {
		t.Fatal(err)
	}

	var durable strings.Builder
	err = filepath.WalkDir(target, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		body, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		durable.Write(body)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{rawSession, rawPath, rawKey} {
		if strings.Contains(durable.String(), raw) {
			t.Fatalf("durable ledger leaked raw value %q", raw)
		}
	}
	for _, digest := range []string{HashReference(rawSession), HashReference(rawPath), HashReference(rawKey)} {
		if !strings.Contains(durable.String(), digest) {
			t.Fatalf("durable ledger missing expected digest %q", digest)
		}
	}
}

func TestVerifyRunClassifiesCorruptDerivedAndCausalState(t *testing.T) {
	store := newTestStore(t, Options{})
	ctx := context.Background()
	result, err := store.Append(ctx, AppendRequest{
		Draft:          validDraft("run-corrupt", KindStart),
		IdempotencyKey: "corrupt:start",
	})
	if err != nil {
		t.Fatal(err)
	}
	runDir := filepath.Join(store.runsRoot, "run-corrupt")

	event := result.Event
	event.EventID = "evt_duplicate_order"
	event.IdempotencyKeyHash = HashReference("missing-receipt")
	event.CausedByEventID = "evt_missing_cause"
	if err := writeJSONAtomic(filepath.Join(runDir, "events", eventFilename(event)), event); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "events", "bad.json"), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	receiptPath := filepath.Join(runDir, "receipts", result.Event.IdempotencyKeyHash+".json")
	value, found, err := readReceipt(receiptPath)
	if err != nil || !found {
		t.Fatalf("readReceipt() = %#v, %v, %v", value, found, err)
	}
	value.EventFile = "wrong.json"
	if err := writeJSONAtomic(receiptPath, value); err != nil {
		t.Fatal(err)
	}

	report, err := store.VerifyRun(ctx, "run-corrupt")
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{"invalid_event", "duplicate_local_sequence", "duplicate_lamport_clock", "missing_causal_event", "missing_receipt", "receipt_mismatch"} {
		if !hasIssue(report, code) {
			t.Errorf("report missing %s: %#v", code, report)
		}
	}
}

func TestStoreRejectsCancelledContextAndInvalidIdempotencyKey(t *testing.T) {
	store := newTestStore(t, Options{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := store.LoadRun(ctx, "run-cancelled"); !errors.Is(err, context.Canceled) {
		t.Fatalf("LoadRun() error = %v", err)
	}
	if _, err := store.Append(context.Background(), AppendRequest{
		Draft:          validDraft("run-invalid-key", KindStart),
		IdempotencyKey: "",
	}); err == nil {
		t.Fatal("Append() expected invalid idempotency key error")
	}
}

func TestNewFileStoreRejectsRuntimeOutsideIgnoredBoundary(t *testing.T) {
	if _, err := NewFileStore(t.TempDir(), Options{RuntimeRoot: ".private/runs"}); err == nil {
		t.Fatal("NewFileStore() expected unsafe runtime root error")
	}
	if _, err := NewFileStore(t.TempDir(), Options{RuntimeRoot: ".lufy/runtime/custom"}); err != nil {
		t.Fatalf("NewFileStore() safe nested root error = %v", err)
	}
}

func newTestStore(t *testing.T, options Options) *FileStore {
	t.Helper()
	if options.Now == nil {
		options.Now = func() time.Time { return time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC) }
	}
	store, err := NewFileStore(t.TempDir(), options)
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}
	return store
}

func hasIssue(report VerificationReport, code string) bool {
	for _, issue := range report.Issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}
