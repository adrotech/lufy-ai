package runledger

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestProjectorBuildsDeterministicCausalTree(t *testing.T) {
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	store, err := NewFileStore(t.TempDir(), Options{Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	duration := int64(1200)
	tokens := int64(340)
	rootDraft := validDraft("run-root", KindStart)
	rootDraft.EventID = "evt_projector_root_start"
	rootDraft.Checkpoint = &Checkpoint{Status: "implemented", Gate: "implementation", NextOwner: "validator"}
	rootDraft.Metrics = &Metrics{Availability: "available", DurationMillis: &duration, Tokens: &tokens}
	root, err := store.Append(ctx, AppendRequest{Draft: rootDraft, IdempotencyKey: "root:start"})
	if err != nil {
		t.Fatal(err)
	}

	now = now.Add(time.Second)
	childDraft := validDraft("run-child", KindBlocked)
	childDraft.EventID = "evt_projector_child_blocked"
	childDraft.ParentRunID = "run-root"
	childDraft.CausedByEventID = root.Event.EventID
	childDraft.TaskRef = "task-child"
	childDraft.Checkpoint = &Checkpoint{Status: "blocked", Gate: "validation", NextOwner: "implementer"}
	childDraft.Metrics = &Metrics{Availability: "unavailable"}
	child, err := store.Append(ctx, AppendRequest{Draft: childDraft, IdempotencyKey: "child:blocked", ProposedLamport: root.Event.LamportClock})
	if err != nil {
		t.Fatal(err)
	}

	now = now.Add(time.Second)
	finish := validDraft("run-root", KindFinish)
	finish.EventID = "evt_projector_root_finish"
	finish.Checkpoint = &Checkpoint{Status: "closed", Gate: "closure"}
	if _, err := store.Append(ctx, AppendRequest{Draft: finish, IdempotencyKey: "root:finish"}); err != nil {
		t.Fatal(err)
	}

	projector := NewProjector(store)
	summary, err := projector.Build(ctx, "run-root")
	if err != nil {
		t.Fatal(err)
	}
	if summary.SchemaVersion != SummarySchemaVersion || summary.Status != "closed" || summary.Terminal {
		t.Fatalf("root summary = %#v", summary)
	}
	if len(summary.Children) != 1 || summary.Children[0].RunID != "run-child" || summary.Children[0].Status != "blocked" {
		t.Fatalf("children = %#v", summary.Children)
	}
	if len(summary.Children[0].BlockerEventIDs) != 1 || summary.Children[0].BlockerEventIDs[0] != child.Event.EventID {
		t.Fatalf("child blockers = %#v", summary.Children[0].BlockerEventIDs)
	}
	if summary.Metrics.Availability != "partial" || summary.Metrics.DurationMillis == nil || *summary.Metrics.DurationMillis != duration {
		t.Fatalf("metrics = %#v", summary.Metrics)
	}
	if summary.SourceDigest == "" || summary.LastObservedAt.IsZero() {
		t.Fatalf("source metadata = %#v", summary)
	}

	firstJSON, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	second, err := projector.Build(ctx, "run-root")
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.MarshalIndent(second, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("projection is not deterministic:\n%s\n%s", firstJSON, secondJSON)
	}
	golden, err := os.ReadFile(filepath.Join("testdata", "summary.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	goldenLF := bytes.ReplaceAll(golden, []byte("\r\n"), []byte("\n"))
	goldenCRLF := bytes.ReplaceAll(goldenLF, []byte("\n"), []byte("\r\n"))
	var compactGolden, compactActual bytes.Buffer
	if err := json.Compact(&compactGolden, goldenCRLF); err != nil {
		t.Fatalf("invalid summary golden: %v", err)
	}
	if err := json.Compact(&compactActual, firstJSON); err != nil {
		t.Fatalf("invalid projected summary: %v", err)
	}
	if compactGolden.String() != compactActual.String() {
		t.Fatalf("summary differs from golden:\n%s", firstJSON)
	}
}

func TestProjectorVerifyRepairsOnlyDerivedSummary(t *testing.T) {
	store := newTestStore(t, Options{})
	ctx := context.Background()
	draft := validDraft("run-repair", KindFinish)
	draft.Checkpoint = &Checkpoint{Status: "closed", Gate: "closure"}
	if _, err := store.Append(ctx, AppendRequest{Draft: draft, IdempotencyKey: "repair:finish"}); err != nil {
		t.Fatal(err)
	}
	projector := NewProjector(store)

	before, err := projector.Verify(ctx, "run-repair", false)
	if err != nil {
		t.Fatal(err)
	}
	if !before.SourceHealthy || before.ProjectionPresent || before.ProjectionFresh {
		t.Fatalf("before repair = %#v", before)
	}
	repaired, err := projector.Verify(ctx, "run-repair", true)
	if err != nil {
		t.Fatal(err)
	}
	if !repaired.SourceHealthy || !repaired.ProjectionPresent || !repaired.ProjectionFresh || !repaired.Repaired {
		t.Fatalf("repair = %#v", repaired)
	}
	after, err := projector.Verify(ctx, "run-repair", false)
	if err != nil {
		t.Fatal(err)
	}
	if !after.ProjectionFresh || after.Repaired {
		t.Fatalf("after repair = %#v", after)
	}
}

func TestProjectorTracksConcurrentChildWritersAndPartialMetrics(t *testing.T) {
	store := newTestStore(t, Options{LockTimeout: 3 * time.Second})
	ctx := context.Background()
	rootDraft := validDraft("run-hierarchy-root", KindStart)
	rootDraft.Checkpoint = nil
	root, err := store.Append(ctx, AppendRequest{Draft: rootDraft, IdempotencyKey: "hierarchy:root:start"})
	if err != nil {
		t.Fatal(err)
	}

	duration := int64(250)
	children := []struct {
		runID   string
		metrics *Metrics
	}{
		{runID: "run-hierarchy-child-a", metrics: &Metrics{Availability: "available", DurationMillis: &duration}},
		{runID: "run-hierarchy-child-b", metrics: &Metrics{Availability: "unavailable"}},
	}
	var wg sync.WaitGroup
	errCh := make(chan error, len(children))
	for _, child := range children {
		child := child
		wg.Add(1)
		go func() {
			defer wg.Done()
			draft := validDraft(child.runID, KindStart)
			draft.ParentRunID = root.Event.RunID
			draft.CausedByEventID = root.Event.EventID
			draft.Checkpoint = nil
			draft.Metrics = child.metrics
			_, appendErr := store.Append(ctx, AppendRequest{
				Draft: draft, IdempotencyKey: child.runID + ":start", ProposedLamport: root.Event.LamportClock,
			})
			errCh <- appendErr
		}()
	}
	wg.Wait()
	close(errCh)
	for appendErr := range errCh {
		if appendErr != nil {
			t.Fatal(appendErr)
		}
	}

	closeRun := func(runID string) {
		draft := validDraft(runID, KindFinish)
		draft.ParentRunID = root.Event.RunID
		draft.Checkpoint = nil
		if _, closeErr := store.Append(ctx, AppendRequest{Draft: draft, IdempotencyKey: runID + ":finish"}); closeErr != nil {
			t.Fatal(closeErr)
		}
	}
	closeRun(children[0].runID)
	rootFinish := validDraft(root.Event.RunID, KindFinish)
	rootFinish.Checkpoint = nil
	if _, err := store.Append(ctx, AppendRequest{Draft: rootFinish, IdempotencyKey: "hierarchy:root:finish"}); err != nil {
		t.Fatal(err)
	}

	projector := NewProjector(store)
	incomplete, err := projector.Build(ctx, root.Event.RunID)
	if err != nil {
		t.Fatal(err)
	}
	if incomplete.Terminal || len(incomplete.Children) != 2 {
		t.Fatalf("root must wait for every child: %#v", incomplete)
	}
	if incomplete.Children[0].Metrics.Availability != "partial" || incomplete.Children[1].Metrics.Availability != "unavailable" {
		t.Fatalf("partial child metrics lost: %#v", incomplete.Children)
	}

	closeRun(children[1].runID)
	complete, err := projector.Build(ctx, root.Event.RunID)
	if err != nil {
		t.Fatal(err)
	}
	if !complete.Terminal {
		t.Fatalf("closed hierarchy must be terminal: %#v", complete)
	}
}
