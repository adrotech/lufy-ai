package adapters

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

func TestLoadWorkflowProjectionIsContentFreeDeterministicAndCausal(t *testing.T) {
	root := t.TempDir()
	store, err := runledger.NewFileStore(root, runledger.Options{Now: func() time.Time { return time.Unix(100, 0).UTC() }})
	if err != nil {
		t.Fatal(err)
	}
	parent := appendWorkflowEvent(t, store, runledger.EventDraft{
		EventID: "event-parent", RunID: "run-parent", Kind: runledger.KindStart,
		Source:  runledger.Source{Adapter: "test", EventName: "review.started"},
		TaskRef: "phase4-task-5.1", OccurredAt: time.Unix(10, 0).UTC(),
	}, "parent")
	appendWorkflowEvent(t, store, runledger.EventDraft{
		EventID: "event-child", RunID: "run-child", ParentRunID: "run-parent", CausedByEventID: parent.Event.EventID,
		Kind: runledger.KindEvidence, Source: runledger.Source{Adapter: "test", EventName: "review.completed"},
		TaskRef: "phase4-task-5.1", OccurredAt: time.Unix(20, 0).UTC(),
		EvidenceRefs: []runledger.EvidenceRef{{Category: "tests", Result: "passed", PathSHA256: runledger.HashReference("PRIVATE_RUNTIME_PATH_CANARY")}},
	}, "child")

	first := LoadWorkflowProjection(root)
	second := LoadWorkflowProjection(root)
	if first.Availability != "available" || first.Digest == "" || first.Digest != second.Digest || len(first.Events) != 2 {
		t.Fatalf("projection = %#v; second digest=%q", first, second.Digest)
	}
	body, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, forbidden := range []string{"PRIVATE_RUNTIME_PATH_CANARY", ".lufy/runtime", "prompt", "output", "summary"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("projection leaked %q: %s", forbidden, text)
		}
	}
	var child WorkflowLedgerEvent
	for _, event := range first.Events {
		if event.EventID == "event-child" {
			child = event
		}
	}
	if child.ParentRunID != "run-parent" || child.CausedByEventID != "event-parent" || len(child.Evidence) != 1 {
		t.Fatalf("causal metadata missing: %#v", child)
	}
}

func TestLoadWorkflowProjectionMissingLedgerIsUnavailable(t *testing.T) {
	projection := LoadWorkflowProjection(t.TempDir())
	if projection.Availability != "unavailable" || projection.Digest != "" || len(projection.Events) != 0 || projection.Recovery == "" {
		t.Fatalf("missing ledger projection = %#v", projection)
	}
}

func TestLoadWorkflowProjectionDisabledOrInvalidConfigIsUnavailable(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		root := t.TempDir()
		cfg, err := projectconfig.Scan(root, time.Unix(0, 0).UTC())
		if err != nil {
			t.Fatal(err)
		}
		disabled := false
		cfg.RunLedger.Enabled = &disabled
		body, err := projectconfig.Marshal(cfg)
		if err != nil {
			t.Fatal(err)
		}
		writeWorkflowConfig(t, root, body)
		if projection := LoadWorkflowProjection(root); projection.Availability != "unavailable" || projection.Recovery == "" {
			t.Fatalf("disabled projection = %#v", projection)
		}
	})
	t.Run("invalid", func(t *testing.T) {
		root := t.TempDir()
		writeWorkflowConfig(t, root, []byte("run_ledger:\n  enabled: [\n"))
		if projection := LoadWorkflowProjection(root); projection.Availability != "unavailable" || projection.Recovery == "" {
			t.Fatalf("invalid projection = %#v", projection)
		}
	})
}

func writeWorkflowConfig(t *testing.T, root string, body []byte) {
	t.Helper()
	path := filepath.Join(root, projectconfig.ProjectConfigPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
}

func appendWorkflowEvent(t *testing.T, store *runledger.FileStore, draft runledger.EventDraft, key string) runledger.AppendResult {
	t.Helper()
	result, err := store.Append(t.Context(), runledger.AppendRequest{Draft: draft, IdempotencyKey: key})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
