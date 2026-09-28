package application

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/adapters"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

func TestMetricsDerivesRecognizedContentFreeReviewEvents(t *testing.T) {
	root := t.TempDir()
	appendMetricEvents(t, root,
		metricEvent("review-start", "review.started", time.Unix(10, 0).UTC()),
		metricEvent("review-end", "review.completed", time.Unix(130, 0).UTC()),
		metricEvent("rework", "review.rework", time.Unix(140, 0).UTC()),
		metricEvent("reopened", "defect.reopened", time.Unix(150, 0).UTC()),
	)
	result := NewService().Metrics(root)
	if result.Status != "available" || result.ReviewDurationMillis == nil || *result.ReviewDurationMillis != 120000 || result.CompletedReviews != 1 || result.ReworkCount != 1 || result.ReopenedDefectCount != 1 {
		t.Fatalf("metrics = %#v", result)
	}
}

func TestMetricsPartialAndUnavailableDoNotSynthesizeValues(t *testing.T) {
	partialRoot := t.TempDir()
	appendMetricEvents(t, partialRoot, metricEvent("review-start", "review.started", time.Time{}))
	partial := NewService().Metrics(partialRoot)
	if partial.Status != "partial" || partial.ReviewDurationMillis != nil || partial.CompletedReviews != 0 || len(partial.Missing) == 0 {
		t.Fatalf("partial metrics = %#v", partial)
	}
	unavailable := NewService().Metrics(t.TempDir())
	if unavailable.Status != "unavailable" || unavailable.ReviewDurationMillis != nil || unavailable.ReworkCount != 0 || unavailable.ReopenedDefectCount != 0 || unavailable.Recovery == "" {
		t.Fatalf("unavailable metrics = %#v", unavailable)
	}
}

func TestBuildProjectsLedgerMetadataWithoutRuntimeContent(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, ".lufy/workflows/sdd/changes/demo/tasks.md"), "# Tasks\n- [ ] 1.1 Demo\n")
	appendMetricEvents(t, root, runledger.EventDraft{
		EventID: "evidence-event", RunID: "run-demo", Kind: runledger.KindEvidence,
		Source:  runledger.Source{Adapter: "test", EventName: "review.completed"},
		TaskRef: "demo-1.1", OccurredAt: time.Unix(20, 0).UTC(),
		EvidenceRefs: []runledger.EvidenceRef{{Category: "tests", Result: "passed", PathSHA256: runledger.HashReference("PRIVATE_PROMPT_OUTPUT_SECRET")}},
	})
	if _, err := NewService().Build(root); err != nil {
		t.Fatal(err)
	}
	graph, err := adapters.NewStore().LoadGraph(root, NewService().config(root).Root)
	if err != nil {
		t.Fatal(err)
	}
	var hasRun, hasEvidence, hasVerify bool
	for _, node := range graph.Nodes {
		hasRun = hasRun || node.Type == "workflow_run"
		hasEvidence = hasEvidence || node.Type == "workflow_evidence"
	}
	for _, edge := range graph.Edges {
		hasVerify = hasVerify || edge.Type == "verifies"
	}
	body, err := json.Marshal(graph)
	if err != nil {
		t.Fatal(err)
	}
	if !hasRun || !hasEvidence || !hasVerify {
		t.Fatalf("ledger graph projection missing: nodes=%#v edges=%#v", graph.Nodes, graph.Edges)
	}
	for _, forbidden := range []string{"PRIVATE_PROMPT_OUTPUT_SECRET", ".lufy/runtime/runs/", "prompt", "output", "summary"} {
		if strings.Contains(string(body), forbidden) {
			t.Fatalf("graph leaked %q", forbidden)
		}
	}
}

func TestLedgerChangeMakesPersistedGraphStale(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, ".lufy/workflows/sdd/changes/demo/tasks.md"), "# Tasks\n- [ ] 1.1 Demo\n")
	appendMetricEvents(t, root, metricEvent("review-start", "review.started", time.Unix(10, 0).UTC()))
	if _, err := NewService().Build(root); err != nil {
		t.Fatal(err)
	}
	if status := NewService().Status(root); status.Status != "ready" {
		t.Fatalf("status before ledger change = %#v", status)
	}
	appendMetricEvents(t, root, metricEvent("review-end", "review.completed", time.Unix(20, 0).UTC()))
	if status := NewService().Status(root); status.Status != "stale" {
		t.Fatalf("status after ledger change = %#v", status)
	}
}

func metricEvent(id, name string, occurred time.Time) runledger.EventDraft {
	return runledger.EventDraft{EventID: id, RunID: "run-metrics", Kind: runledger.KindCheckpoint, Source: runledger.Source{Adapter: "test", EventName: name}, OccurredAt: occurred}
}

func appendMetricEvents(t *testing.T, root string, drafts ...runledger.EventDraft) {
	t.Helper()
	store, err := runledger.NewFileStore(root, runledger.Options{Now: func() time.Time { return time.Unix(1000, 0).UTC() }})
	if err != nil {
		t.Fatal(err)
	}
	for index, draft := range drafts {
		if _, err := store.Append(t.Context(), runledger.AppendRequest{Draft: draft, IdempotencyKey: draft.EventID + string(rune('a'+index))}); err != nil {
			t.Fatal(err)
		}
	}
}
