package runledger

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAdaptiveMetadataIsCanonicalAndContentFree(t *testing.T) {
	t.Parallel()
	draft := validAdaptiveDraft(KindAssignment)
	first, err := CanonicalFingerprint(draft)
	if err != nil {
		t.Fatal(err)
	}
	second, err := CanonicalFingerprint(draft)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("adaptive fingerprint changed: %s != %s", first, second)
	}
	body, err := json.Marshal(draft)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"prompt", "response", "summary", "transcript", "secret", "diff", "hypothesis_text"} {
		if strings.Contains(strings.ToLower(string(body)), forbidden) {
			t.Fatalf("adaptive event leaked forbidden field %q: %s", forbidden, body)
		}
	}
}

func TestValidateAdaptiveMetadataRejectsInvalidKindsAndReferences(t *testing.T) {
	t.Parallel()
	tests := []func(*EventDraft){
		func(d *EventDraft) { d.Adaptive = nil },
		func(d *EventDraft) { d.Adaptive.SchemaVersion = "lufy-run-adaptive/v2" },
		func(d *EventDraft) { d.Adaptive.ActorRef = "raw-agent" },
		func(d *EventDraft) { d.Adaptive.LeaseTokenDigest = "raw-token" },
		func(d *EventDraft) { d.Adaptive.ExpectedVersion = 0 },
		func(d *EventDraft) { d.Adaptive.LeaseExpiresAt = time.Time{} },
		func(d *EventDraft) { d.Adaptive.RoleHint = "bad role" },
	}
	for index, mutate := range tests {
		draft := validAdaptiveDraft(KindAssignment)
		mutate(&draft)
		if err := ValidateDraft(draft); err == nil {
			t.Fatalf("case %d: expected validation error", index)
		}
	}
	demand := validAdaptiveDraft(KindDemand)
	demand.Adaptive.Reason = "blocked"
	if err := ValidateDraft(demand); err == nil {
		t.Fatal("demand accepted yield-only metadata")
	}
}

func TestFileStoreExpectedSequenceFencesStaleWriter(t *testing.T) {
	store := newTestStore(t, Options{})
	expected := uint64(0)
	first, err := store.Append(t.Context(), AppendRequest{
		Draft: validAdaptiveDraft(KindDemand), IdempotencyKey: "demand:first", ExpectedLocalSequence: &expected,
	})
	if err != nil || first.Status != AppendRecorded {
		t.Fatalf("first append = %#v err=%v", first, err)
	}
	stale := uint64(0)
	secondDraft := validAdaptiveDraft(KindDemand)
	secondDraft.Adaptive.DemandID = "demand-2"
	_, err = store.Append(t.Context(), AppendRequest{
		Draft: secondDraft, IdempotencyKey: "demand:second", ExpectedLocalSequence: &stale,
	})
	if !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("stale append error = %v", err)
	}
	events, err := store.LoadRun(t.Context(), "run-adaptive")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events=%d, want 1", len(events))
	}
}

func TestFileStoreRecoversAdaptiveEventPublishedBeforeReceipt(t *testing.T) {
	target := t.TempDir()
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	injected := errors.New("adaptive receipt unavailable")
	failing, err := NewFileStore(target, Options{
		Now:                func() time.Time { return now },
		BeforeReceiptWrite: func(Event) error { return injected },
	})
	if err != nil {
		t.Fatal(err)
	}
	draft := validAdaptiveDraft(KindDemand)
	_, err = failing.Append(context.Background(), AppendRequest{Draft: draft, IdempotencyKey: "adaptive:crash"})
	if !errors.Is(err, injected) {
		t.Fatalf("injected error = %v", err)
	}

	recovered, err := NewFileStore(target, Options{Now: func() time.Time { return now.Add(time.Second) }})
	if err != nil {
		t.Fatal(err)
	}
	result, err := recovered.Append(context.Background(), AppendRequest{Draft: draft, IdempotencyKey: "adaptive:crash"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != AppendDuplicateNoop || result.Event.Adaptive == nil || result.Event.Adaptive.DemandID != "demand-1" {
		t.Fatalf("adaptive recovery = %#v", result)
	}
	report, err := recovered.VerifyRun(context.Background(), "run-adaptive")
	if err != nil {
		t.Fatal(err)
	}
	if !report.Healthy || report.EventCount != 1 || len(report.Issues) != 0 {
		t.Fatalf("adaptive verification after recovery = %#v", report)
	}
}

func validAdaptiveDraft(kind EventKind) EventDraft {
	now := time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
	metadata := &AdaptiveMetadata{
		SchemaVersion:             AdaptiveMetadataSchemaVersion,
		DemandID:                  "demand-1",
		AssignmentID:              "assignment-1",
		ActorRef:                  strings.Repeat("a", 64),
		RoleHint:                  "implementer",
		PolicyVersion:             "deterministic-v1",
		RecommendationFingerprint: strings.Repeat("b", 64),
		LeaseTokenDigest:          strings.Repeat("c", 64),
		LeaseExpiresAt:            now.Add(15 * time.Minute),
		ExpectedVersion:           1,
		Priority:                  80,
		RequiredBudget:            40,
		Score:                     900,
	}
	if kind == KindDemand {
		metadata.AssignmentID = ""
		metadata.ActorRef = ""
		metadata.RoleHint = ""
		metadata.PolicyVersion = ""
		metadata.RecommendationFingerprint = ""
		metadata.LeaseTokenDigest = ""
		metadata.LeaseExpiresAt = time.Time{}
		metadata.ExpectedVersion = 0
		metadata.Score = 0
	}
	return EventDraft{
		RunID: "run-adaptive", Kind: kind, OccurredAt: now,
		Source:  Source{Adapter: "lufy", EventName: "adaptive_" + string(kind)},
		TaskRef: "task-1", Adaptive: metadata,
	}
}
