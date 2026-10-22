package reviewevents

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/application"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

const evidenceDigest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func fixture(t *testing.T) (string, *runledger.FileStore, Service) {
	t.Helper()
	root := t.TempDir()
	store, err := runledger.NewFileStore(root, runledger.Options{Now: func() time.Time { return time.Unix(1000, 0).UTC() }})
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(store)
	svc.Now = func() time.Time { return time.Unix(10, 0).UTC() }
	return root, store, svc
}

func record(t *testing.T, svc Service, options Options) runledger.AppendResult {
	t.Helper()
	result, err := svc.Record(context.Background(), options)
	if err != nil {
		t.Fatalf("Record(%s) failed: %v", options.Action, err)
	}
	return result
}

func startOptions() Options {
	return Options{RunID: "run-review", ReviewID: "review-one", Action: "start"}
}

func completeOptions(operation string) Options {
	return Options{RunID: "run-review", ReviewID: "review-one", Action: "complete", OperationID: operation, Result: "passed", EvidenceSHA256: evidenceDigest}
}

func events(t *testing.T, store *runledger.FileStore) []runledger.Event {
	t.Helper()
	loaded, err := store.LoadRun(context.Background(), "run-review")
	if err != nil {
		t.Fatal(err)
	}
	return loaded
}

func TestRecordPairFeedsMetricsWithoutAdvancingGates(t *testing.T) {
	root, store, svc := fixture(t)
	start := record(t, svc, startOptions())
	svc.Now = func() time.Time { return time.Unix(40, 0).UTC() }
	end := record(t, svc, completeOptions("complete-one"))
	if start.Status != runledger.AppendRecorded || end.Status != runledger.AppendRecorded {
		t.Fatal("new review pair must be recorded")
	}
	metrics := application.NewService().Metrics(root)
	if metrics.Status != "available" || metrics.CompletedReviews != 1 || metrics.ReviewDurationMillis == nil || *metrics.ReviewDurationMillis != 30000 {
		t.Fatalf("review pair must produce observed 30-second duration: %#v", metrics)
	}
	for _, event := range events(t, store) {
		if event.Kind != runledger.KindEvidence || event.Checkpoint != nil || event.TaskRef != runledger.HashReference("review-one") {
			t.Fatalf("review observations must be content-free evidence, never gate transitions: %#v", event)
		}
	}
}

func TestRecordRetryPreservesOriginalTimestampWithAdvancedNow(t *testing.T) {
	_, store, svc := fixture(t)
	firstStart := record(t, svc, startOptions())
	svc.Now = func() time.Time { return time.Unix(20, 0).UTC() }
	explicitDefault := startOptions()
	explicitDefault.OperationID = explicitDefault.ReviewID
	retryStart := record(t, svc, explicitDefault)
	if retryStart.Status != runledger.AppendDuplicateNoop || !retryStart.Event.OccurredAt.Equal(firstStart.Event.OccurredAt) {
		t.Fatal("default operation and explicit review ID retry must preserve original start")
	}
	firstEnd := record(t, svc, completeOptions("complete-one"))
	svc.Now = func() time.Time { return time.Unix(90, 0).UTC() }
	retryEnd := record(t, svc, completeOptions("complete-one"))
	if retryEnd.Status != runledger.AppendDuplicateNoop || retryEnd.Event.EventID != firstEnd.Event.EventID || !retryEnd.Event.OccurredAt.Equal(firstEnd.Event.OccurredAt) || len(events(t, store)) != 2 {
		t.Fatal("completion retry must not append or change observed time")
	}
}

func TestRecordRejectsInvalidInputsWithoutAppending(t *testing.T) {
	cases := []struct {
		name      string
		options   Options
		seedStart bool
	}{
		{"unknown_action", Options{RunID: "run-review", ReviewID: "review-one", Action: "ship"}, false},
		{"unsafe_run", Options{RunID: "../escape", ReviewID: "review-one", Action: "start"}, false},
		{"unsafe_review", Options{RunID: "run-review", ReviewID: "raw review", Action: "start"}, false},
		{"unsafe_operation", Options{RunID: "run-review", ReviewID: "review-one", OperationID: "op/raw", Action: "start"}, false},
		{"completion_without_start", completeOptions("complete-one"), false},
		{"rework_without_start", Options{RunID: "run-review", ReviewID: "review-one", Action: "rework", EvidenceSHA256: evidenceDigest}, false},
		{"reopen_without_start", Options{RunID: "run-review", ReviewID: "review-one", Action: "reopen", EvidenceSHA256: evidenceDigest}, false},
		{"bad_fingerprint", Options{RunID: "run-review", ReviewID: "review-one", Action: "complete", Result: "passed", EvidenceSHA256: "not-hex"}, true},
		{"missing_fingerprint", Options{RunID: "run-review", ReviewID: "review-one", Action: "complete", Result: "passed"}, true},
		{"bad_result", Options{RunID: "run-review", ReviewID: "review-one", Action: "complete", Result: "delivered", EvidenceSHA256: evidenceDigest}, true},
		{"start_with_evidence", Options{RunID: "run-review", ReviewID: "review-one", Action: "start", EvidenceSHA256: evidenceDigest}, false},
		{"start_with_result", Options{RunID: "run-review", ReviewID: "review-one", Action: "start", Result: "passed"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, store, svc := fixture(t)
			want := 0
			if tc.seedStart {
				record(t, svc, startOptions())
				want = 1
			}
			if _, err := svc.Record(context.Background(), tc.options); err == nil {
				t.Fatal("invalid observation must fail before append")
			}
			if got := len(events(t, store)); got != want {
				t.Fatalf("invalid observation changed ledger count: %d != %d", got, want)
			}
		})
	}
}

func TestRecordConflictingCompletionNeverAppends(t *testing.T) {
	_, store, svc := fixture(t)
	record(t, svc, startOptions())
	record(t, svc, completeOptions("complete-one"))
	for _, changed := range []Options{
		completeOptions("different-operation"),
		{RunID: "run-review", ReviewID: "review-one", Action: "complete", OperationID: "complete-one", Result: "failed", EvidenceSHA256: evidenceDigest},
		{RunID: "run-review", ReviewID: "review-one", Action: "complete", OperationID: "complete-one", Result: "passed", EvidenceSHA256: strings.Repeat("b", 64)},
	} {
		if _, err := svc.Record(context.Background(), changed); err == nil {
			t.Fatal("different completion operation or payload must conflict")
		}
		if len(events(t, store)) != 2 {
			t.Fatal("conflicting completion appended an extra event")
		}
	}
}

func TestRecordPrivacyDoesNotPersistRawReviewOrOperation(t *testing.T) {
	root, store, svc := fixture(t)
	start := Options{RunID: "run-review", ReviewID: "PRIVATE-REVIEW-CANARY", OperationID: "PRIVATE-OPERATION-CANARY", Action: "start"}
	record(t, svc, start)
	loaded := events(t, store)
	if loaded[0].TaskRef != runledger.HashReference(start.ReviewID) {
		t.Fatal("task reference must be review hash")
	}
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || filepath.Ext(path) != ".json" {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, canary := range []string{start.ReviewID, start.OperationID} {
			if strings.Contains(string(body), canary) {
				t.Fatal("raw review or operation identifier leaked into persisted JSON")
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestRecordConcurrentExactCompletionHasOneEventAndDuplicateNoops(t *testing.T) {
	_, store, svc := fixture(t)
	record(t, svc, startOptions())
	barrier := make(chan struct{})
	type outcome struct {
		result runledger.AppendResult
		err    error
	}
	results := make(chan outcome, 4)
	for index := 0; index < 4; index++ {
		go func() {
			<-barrier
			result, err := svc.Record(context.Background(), completeOptions("same-operation"))
			results <- outcome{result, err}
		}()
	}
	close(barrier)
	recorded, duplicates := 0, 0
	for index := 0; index < 4; index++ {
		result := <-results
		if result.err != nil {
			t.Fatalf("exact concurrent retry failed: %v", result.err)
		}
		switch result.result.Status {
		case runledger.AppendRecorded:
			recorded++
		case runledger.AppendDuplicateNoop:
			duplicates++
		default:
			t.Fatalf("unexpected status: %s", result.result.Status)
		}
	}
	if recorded != 1 || duplicates != 3 || len(events(t, store)) != 2 {
		t.Fatal("exact concurrent completion must append once and preserve retry noops")
	}
}

func TestRecordConcurrentDifferentCompletionOperationsAppendOnlyOne(t *testing.T) {
	_, store, svc := fixture(t)
	record(t, svc, startOptions())
	barrier := make(chan struct{})
	results := make(chan error, 2)
	for _, operation := range []string{"operation-a", "operation-b"} {
		go func(operation string) {
			<-barrier
			_, err := svc.Record(context.Background(), completeOptions(operation))
			results <- err
		}(operation)
	}
	close(barrier)
	successes := 0
	for index := 0; index < 2; index++ {
		if <-results == nil {
			successes++
		}
	}
	if successes != 1 || len(events(t, store)) != 2 {
		t.Fatal("racing distinct completion operations must not both append")
	}
}

func TestRecordReworkAndReopenAreOperationScopedEvidenceOnly(t *testing.T) {
	_, store, svc := fixture(t)
	record(t, svc, startOptions())
	for _, action := range []string{"rework", "reopen"} {
		for _, operation := range []string{"operation-a", "operation-b"} {
			options := Options{RunID: "run-review", ReviewID: "review-one", Action: action, OperationID: operation, EvidenceSHA256: evidenceDigest}
			first := record(t, svc, options)
			retry := record(t, svc, options)
			if first.Status != runledger.AppendRecorded || retry.Status != runledger.AppendDuplicateNoop {
				t.Fatal("rework/reopen retries must be idempotent but distinct operations independent")
			}
		}
	}
	loaded := events(t, store)
	if len(loaded) != 5 {
		t.Fatalf("expected start plus four observations, got %d", len(loaded))
	}
	for _, event := range loaded {
		if event.Kind != runledger.KindEvidence || event.Checkpoint != nil {
			t.Fatal("review events must never act as lifecycle checkpoint, finish, or gate advancement")
		}
	}
}

func TestRecordRejectsObservationClockEarlierThanStart(t *testing.T) {
	_, store, svc := fixture(t)
	record(t, svc, startOptions())
	svc.Now = func() time.Time { return time.Unix(9, 0).UTC() }
	if _, err := svc.Record(context.Background(), completeOptions("complete-one")); err == nil {
		t.Fatal("completion clock before review start must be rejected")
	}
	if len(events(t, store)) != 1 {
		t.Fatal("invalid clock appended an observation")
	}
}
