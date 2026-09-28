package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	contextapp "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/application"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

func TestRunContextMetricsHumanJSONHelpAndUsage(t *testing.T) {
	root := t.TempDir()
	store, err := runledger.NewFileStore(root, runledger.Options{Now: func() time.Time { return time.Unix(1000, 0).UTC() }})
	if err != nil {
		t.Fatal(err)
	}
	for index, draft := range []runledger.EventDraft{
		{EventID: "review-start", RunID: "run-cli-metrics", Kind: runledger.KindCheckpoint, Source: runledger.Source{Adapter: "test", EventName: "review.started"}, OccurredAt: time.Unix(10, 0).UTC()},
		{EventID: "review-end", RunID: "run-cli-metrics", Kind: runledger.KindCheckpoint, Source: runledger.Source{Adapter: "test", EventName: "review.completed"}, OccurredAt: time.Unix(20, 0).UTC()},
	} {
		if _, err := store.Append(t.Context(), runledger.AppendRequest{Draft: draft, IdempotencyKey: draft.EventID + string(rune('a'+index))}); err != nil {
			t.Fatal(err)
		}
	}

	var stdout, stderr bytes.Buffer
	if code := Run([]string{"context", "metrics", "--target", root}, Dependencies{Stdout: &stdout, Stderr: &stderr}); code != ExitOK || !strings.Contains(stdout.String(), "context metrics: available") {
		t.Fatalf("human code/output = %d %q stderr=%q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"context", "metrics", "--target", root, "--json"}, Dependencies{Stdout: &stdout, Stderr: &stderr}); code != ExitOK {
		t.Fatalf("json code=%d stderr=%q", code, stderr.String())
	}
	var result contextapp.MetricsResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result.Status != "available" || result.CompletedReviews != 1 {
		t.Fatalf("json result=%#v err=%v body=%q", result, err, stdout.String())
	}
	stdout.Reset()
	if code := Run([]string{"context", "metrics", "--help"}, Dependencies{Stdout: &stdout, Stderr: &stderr}); code != ExitOK || !strings.Contains(stdout.String(), "context metrics") {
		t.Fatalf("help code/output=%d %q", code, stdout.String())
	}
	stdout.Reset()
	if code := Run([]string{"context", "metrics", "extra"}, Dependencies{Stdout: &stdout, Stderr: &stderr}); code != ExitUsageErr {
		t.Fatalf("usage code=%d", code)
	}
}
