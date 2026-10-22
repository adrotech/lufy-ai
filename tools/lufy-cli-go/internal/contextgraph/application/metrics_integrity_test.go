package application

import (
	"strings"
	"testing"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/adapters"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

func TestMetricsIntegrityPairsIndependentTasksWithinSameRun(t *testing.T) {
	root := t.TempDir()
	startA := metricEvent("start-a", "review.started", time.Unix(10, 0).UTC())
	startB := metricEvent("start-b", "review.started", time.Unix(20, 0).UTC())
	endA := metricEvent("end-a", "review.completed", time.Unix(40, 0).UTC())
	endB := metricEvent("end-b", "review.completed", time.Unix(60, 0).UTC())
	startA.TaskRef, endA.TaskRef = "task-a", "task-a"
	startB.TaskRef, endB.TaskRef = "task-b", "task-b"
	appendMetricEvents(t, root, startA, startB, endA, endB)

	result := NewService().Metrics(root)
	if result.Status != "available" || result.CompletedReviews != 2 || result.ReviewDurationMillis == nil || *result.ReviewDurationMillis != 70000 || len(result.Missing) != 0 {
		t.Fatalf("independent task reviews must total 30s + 40s without crossing pairs: %#v", result)
	}
}

func TestMetricsIntegrityDuplicateStartDoesNotOverwriteAmbiguousPair(t *testing.T) {
	root := t.TempDir()
	start := metricEvent("start-original", "review.started", time.Unix(10, 0).UTC())
	duplicate := metricEvent("start-ambiguous", "review.started", time.Unix(20, 0).UTC())
	end := metricEvent("end", "review.completed", time.Unix(30, 0).UTC())
	start.TaskRef, duplicate.TaskRef, end.TaskRef = "task-a", "task-a", "task-a"
	appendMetricEvents(t, root, start, duplicate, end)

	result := NewService().Metrics(root)
	if result.Status != "partial" || result.CompletedReviews != 0 || result.ReviewDurationMillis != nil || len(result.Missing) == 0 {
		t.Fatalf("ambiguous starts must not fabricate a duration from either candidate: %#v", result)
	}
}

func TestMetricsIntegrityMissingObservationTimestampRetainsCountsAndDegrades(t *testing.T) {
	for _, eventName := range []string{"review.rework", "defect.reopened"} {
		t.Run(eventName, func(t *testing.T) {
			// El store normaliza timestamps vacíos. La proyección legacy permite
			// verificar el read-side sin adulterar eventos persistidos ni defaults.
			projection := adapters.WorkflowProjection{Availability: "available", Events: []adapters.WorkflowLedgerEvent{
				{RunID: "run-metrics", EventName: "review.started", OccurredAt: time.Unix(10, 0).UTC()},
				{RunID: "run-metrics", EventName: "review.completed", OccurredAt: time.Unix(30, 0).UTC()},
				{RunID: "run-metrics", EventName: eventName},
			}}
			result := metricsFromProjection(projection)
			if result.Status != "partial" || result.CompletedReviews != 1 || result.ReviewDurationMillis == nil || *result.ReviewDurationMillis != 20000 {
				t.Fatalf("missing observation timestamp must degrade an otherwise valid review pair: %#v", result)
			}
			if eventName == "review.rework" && (result.ReworkCount != 1 || result.ReopenedDefectCount != 0) {
				t.Fatalf("rework count must retain one observed event without inferring reopened defects: %#v", result)
			}
			if eventName == "defect.reopened" && (result.ReopenedDefectCount != 1 || result.ReworkCount != 0) {
				t.Fatalf("reopened count must retain one observed event without inferring rework: %#v", result)
			}
			if !strings.Contains(strings.Join(result.Missing, " "), "timestamp") {
				t.Fatalf("missing evidence must identify timestamp incompleteness: %#v", result.Missing)
			}
		})
	}
}

func TestMetricsIntegrityLegacyEmptyTaskPairRemainsAvailable(t *testing.T) {
	root := t.TempDir()
	appendMetricEvents(t, root,
		metricEvent("legacy-start", "review.started", time.Unix(10, 0).UTC()),
		metricEvent("legacy-end", "review.completed", time.Unix(40, 0).UTC()),
	)
	result := NewService().Metrics(root)
	if result.Status != "available" || result.CompletedReviews != 1 || result.ReviewDurationMillis == nil || *result.ReviewDurationMillis != 30000 {
		t.Fatalf("simple legacy pair with empty task must remain supported: %#v", result)
	}
}

func TestMetricsIntegrityOrphanAndInvertedPairsNeverFabricateDuration(t *testing.T) {
	cases := []struct {
		name   string
		events []runledger.EventDraft
	}{
		{"orphan_start", []runledger.EventDraft{metricEvent("start", "review.started", time.Unix(10, 0).UTC())}},
		{"orphan_end", []runledger.EventDraft{metricEvent("end", "review.completed", time.Unix(40, 0).UTC())}},
		{"inverted_timestamp", []runledger.EventDraft{
			metricEvent("start", "review.started", time.Unix(40, 0).UTC()),
			metricEvent("end", "review.completed", time.Unix(10, 0).UTC()),
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			appendMetricEvents(t, root, tc.events...)
			result := NewService().Metrics(root)
			if result.Status != "partial" || result.CompletedReviews != 0 || result.ReviewDurationMillis != nil || len(result.Missing) == 0 {
				t.Fatalf("invalid review pair must remain partial without invented duration: %#v", result)
			}
		})
	}
}
