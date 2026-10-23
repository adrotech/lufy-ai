package reviewevents_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/reviewevents"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

func adversarialFixture(t *testing.T) (*runledger.FileStore, reviewevents.Service) {
	t.Helper()
	store, err := runledger.NewFileStore(t.TempDir(), runledger.Options{Now: func() time.Time { return time.Unix(1000, 0).UTC() }})
	if err != nil {
		t.Fatal(err)
	}
	svc := reviewevents.NewService(store)
	svc.Now = func() time.Time { return time.Unix(10, 0).UTC() }
	return store, svc
}

func mustRecord(t *testing.T, svc reviewevents.Service, opts reviewevents.Options) runledger.AppendResult {
	t.Helper()
	result, err := svc.Record(context.Background(), opts)
	if err != nil {
		t.Fatalf("Record rejected distinct valid semantic operation: %v", err)
	}
	return result
}

func TestReviewSemanticTupleWithColonIDsCannotCollide(t *testing.T) {
	store, svc := adversarialFixture(t)
	mustRecord(t, svc, reviewevents.Options{RunID: "run-review", ReviewID: "rv:start", Action: "start", OperationID: "seed-start"})
	svc.Now = func() time.Time { return time.Unix(20, 0).UTC() }
	complete := mustRecord(t, svc, reviewevents.Options{RunID: "run-review", ReviewID: "rv:start", Action: "complete", OperationID: "zz", Result: "passed", EvidenceSHA256: strings.Repeat("a", 64)})
	otherStart := mustRecord(t, svc, reviewevents.Options{RunID: "run-review", ReviewID: "rv", Action: "start", OperationID: "complete:zz"})
	if complete.Status != runledger.AppendRecorded || otherStart.Status != runledger.AppendRecorded || complete.Event.IdempotencyKeyHash == otherStart.Event.IdempotencyKeyHash {
		t.Fatal("colon-delimited identifiers must not alias different semantic tuples")
	}
	events, err := store.LoadRun(context.Background(), "run-review")
	if err != nil || len(events) != 3 {
		t.Fatalf("expected three independent observations, got %d: %v", len(events), err)
	}
}

func TestReviewZeroClockAndNilStoreFailWithoutEvents(t *testing.T) {
	store, svc := adversarialFixture(t)
	opts := reviewevents.Options{RunID: "run-review", ReviewID: "review-one", Action: "start"}
	svc.Now = func() time.Time { return time.Time{} }
	if _, err := svc.Record(context.Background(), opts); err == nil {
		t.Fatal("zero clock must not fabricate occurrence time")
	}
	empty := reviewevents.NewService(nil)
	if _, err := empty.Record(context.Background(), opts); err == nil {
		t.Fatal("nil source must fail safely")
	}
	events, err := store.LoadRun(context.Background(), opts.RunID)
	if err != nil || len(events) != 0 {
		t.Fatal("invalid source/clock must not append")
	}
}

func TestReviewEvidencePreservesExistingWorkflowCheckpoint(t *testing.T) {
	store, svc := adversarialFixture(t)
	ctx := context.Background()
	checkpoint, err := store.Append(ctx, runledger.AppendRequest{IdempotencyKey: "workflow-checkpoint", Draft: runledger.EventDraft{
		RunID: "run-review", Kind: runledger.KindCheckpoint, OccurredAt: time.Unix(1, 0).UTC(),
		Source:     runledger.Source{Adapter: "test", EventName: "workflow.checkpoint"},
		Checkpoint: &runledger.Checkpoint{Status: "blocked", Gate: "validation"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	mustRecord(t, svc, reviewevents.Options{RunID: "run-review", ReviewID: "review-one", Action: "start"})
	mustRecord(t, svc, reviewevents.Options{RunID: "run-review", ReviewID: "review-one", Action: "complete", Result: "passed", EvidenceSHA256: strings.Repeat("a", 64)})
	summary, err := runledger.NewProjector(store).Build(ctx, "run-review")
	if err != nil || summary.Status != "blocked" || summary.Terminal {
		t.Fatalf("review cannot change existing workflow state: %#v %v", summary, err)
	}
	events, err := store.LoadRun(ctx, "run-review")
	if err != nil || len(events) != 3 {
		t.Fatal("unexpected ledger count")
	}
	if events[0].Fingerprint != checkpoint.Event.Fingerprint {
		t.Fatal("workflow checkpoint was changed")
	}
	for _, event := range events[1:] {
		if event.Kind != runledger.KindEvidence || event.Checkpoint != nil {
			t.Fatal("review observation must never finish or advance a gate")
		}
	}
}
