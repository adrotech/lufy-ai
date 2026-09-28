package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

func TestRunLedgerCLIRecordCheckpointStatusSummaryAndVerify(t *testing.T) {
	target := t.TempDir()
	draft := runledger.EventDraft{
		EventID:    "evt_cli_root_start",
		RunID:      "run-cli-root",
		OccurredAt: time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC),
		Kind:       runledger.KindStart,
		Source:     runledger.Source{Adapter: "manual", EventName: "Start"},
		TaskRef:    "task-cli",
		Checkpoint: &runledger.Checkpoint{
			Status: "implemented", Gate: "implementation", NextOwner: "validator",
		},
	}
	body, err := json.Marshal(draft)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"run", "record", "--target", target, "--idempotency-key", "cli:start", "--json"}, Dependencies{
		Stdin: bytes.NewReader(body), Stdout: &stdout, Stderr: &stderr,
	})
	if code != ExitOK {
		t.Fatalf("record code=%d stderr=%s", code, stderr.String())
	}
	var recorded runledger.AppendResult
	if err := json.Unmarshal(stdout.Bytes(), &recorded); err != nil {
		t.Fatalf("record JSON invalid: %v\n%s", err, stdout.String())
	}
	if recorded.SchemaVersion != runledger.AppendResultSchemaVersion || recorded.Status != runledger.AppendRecorded || recorded.Event.EventID != draft.EventID {
		t.Fatalf("recorded = %#v", recorded)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"run", "checkpoint", "--target", target, "--run", "run-cli-root", "--status", "closed",
		"--gate", "closure", "--idempotency-key", "cli:closed", "--json",
	}, Dependencies{Stdout: &stdout, Stderr: &stderr})
	if code != ExitOK {
		t.Fatalf("checkpoint code=%d stderr=%s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"run", "checkpoint", "--target", target, "--run", "run-cli-root", "--status", "closed",
		"--gate", "closure", "--idempotency-key", "cli:closed", "--json",
	}, Dependencies{Stdout: &stdout, Stderr: &stderr})
	if code != ExitOK {
		t.Fatalf("checkpoint retry code=%d stderr=%s", code, stderr.String())
	}
	var checkpointRetry runledger.AppendResult
	if err := json.Unmarshal(stdout.Bytes(), &checkpointRetry); err != nil {
		t.Fatal(err)
	}
	if checkpointRetry.Status != runledger.AppendDuplicateNoop {
		t.Fatalf("checkpoint retry = %#v", checkpointRetry)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"run", "status", "--target", target, "--run", "run-cli-root", "--json"}, Dependencies{Stdout: &stdout, Stderr: &stderr})
	if code != ExitOK {
		t.Fatalf("status code=%d stderr=%s", code, stderr.String())
	}
	var status runStatusView
	if err := json.Unmarshal(stdout.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status.SchemaVersion != "lufy-run-status/v1" || status.Status != "closed" || !status.Terminal || !status.SourceHealthy || status.ProjectionPresent {
		t.Fatalf("status = %#v", status)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"run", "summary", "--target", target, "--run", "run-cli-root", "--json"}, Dependencies{Stdout: &stdout, Stderr: &stderr})
	if code != ExitOK {
		t.Fatalf("summary code=%d stderr=%s", code, stderr.String())
	}
	var summary runledger.RunSummary
	if err := json.Unmarshal(stdout.Bytes(), &summary); err != nil {
		t.Fatal(err)
	}
	if summary.SchemaVersion != runledger.SummarySchemaVersion || summary.EventCount != 2 || summary.SourceDigest == "" {
		t.Fatalf("summary = %#v", summary)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"run", "verify", "--target", target, "--run", "run-cli-root", "--repair", "--json"}, Dependencies{Stdout: &stdout, Stderr: &stderr})
	if code != ExitOK {
		t.Fatalf("verify code=%d stderr=%s", code, stderr.String())
	}
	var verification runledger.RunVerification
	if err := json.Unmarshal(stdout.Bytes(), &verification); err != nil {
		t.Fatal(err)
	}
	if !verification.SourceHealthy || !verification.ProjectionFresh || !verification.Repaired {
		t.Fatalf("verification = %#v", verification)
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy", "runtime", "runs", "run-cli-root", "projections", "summary.json")); err != nil {
		t.Fatalf("projection missing: %v", err)
	}

	eventsDir := filepath.Join(target, ".lufy", "runtime", "runs", "run-cli-root", "events")
	entries, err := os.ReadDir(eventsDir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("events unavailable: %v", err)
	}
	if err := os.Remove(filepath.Join(eventsDir, entries[0].Name())); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"run", "verify", "--target", target, "--run", "run-cli-root", "--json"}, Dependencies{Stdout: &stdout, Stderr: &stderr})
	if code != ExitRuntimeErr {
		t.Fatalf("corrupt verify code=%d stderr=%s", code, stderr.String())
	}
	verification = runledger.RunVerification{}
	if err := json.Unmarshal(stdout.Bytes(), &verification); err != nil {
		t.Fatal(err)
	}
	if verification.SourceHealthy {
		t.Fatalf("corrupt verification = %#v", verification)
	}
}

func TestRunLedgerCLIPruneRequiresExplicitMode(t *testing.T) {
	target := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"run", "prune", "--target", target}, Dependencies{Stdout: &stdout, Stderr: &stderr}); code != ExitUsageErr {
		t.Fatalf("prune without mode code=%d", code)
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"run", "prune", "--target", target, "--dry-run", "--json"}, Dependencies{Stdout: &stdout, Stderr: &stderr}); code != ExitOK {
		t.Fatalf("prune dry-run code=%d stderr=%s", code, stderr.String())
	}
	var report runledger.PruneReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.SchemaVersion != "lufy-run-prune/v1" || !report.DryRun || len(report.Candidates) != 0 {
		t.Fatalf("prune report = %#v", report)
	}
}

func TestRunLedgerCLIHonorsDisabledProjectConfig(t *testing.T) {
	target := t.TempDir()
	disabled := false
	cfg := projectconfig.ProjectConfig{RunLedger: projectconfig.DefaultRunLedgerConfig()}
	cfg.RunLedger.Enabled = &disabled
	body, err := projectconfig.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(target, ".lufy", "config", "project.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{"run", "status", "--target", target, "--run", "run-disabled"}, Dependencies{Stdout: &stdout, Stderr: &stderr})
	if code != ExitRuntimeErr || !strings.Contains(stderr.String(), "run ledger deshabilitado") {
		t.Fatalf("disabled code=%d stderr=%s", code, stderr.String())
	}
}

func TestRunLedgerCLIUsageAndHelp(t *testing.T) {
	cases := []struct {
		args []string
		code int
	}{
		{[]string{"run"}, ExitUsageErr},
		{[]string{"run", "help"}, ExitOK},
		{[]string{"run", "unknown"}, ExitUsageErr},
		{[]string{"run", "record"}, ExitUsageErr},
		{[]string{"run", "checkpoint"}, ExitUsageErr},
		{[]string{"run", "status"}, ExitUsageErr},
		{[]string{"run", "verify"}, ExitUsageErr},
	}
	for _, tc := range cases {
		var stdout, stderr bytes.Buffer
		if code := Run(tc.args, Dependencies{Stdout: &stdout, Stderr: &stderr}); code != tc.code {
			t.Errorf("Run(%v) code=%d want=%d stderr=%s", tc.args, code, tc.code, stderr.String())
		}
	}
}
