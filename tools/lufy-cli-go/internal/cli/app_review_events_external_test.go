package cli_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/cli"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/application"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

func invokeReviewCLI(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := cli.Run(args, cli.Dependencies{Stdout: &stdout, Stderr: &stderr})
	return code, stdout.String(), stderr.String()
}

func savedReviewEvents(t *testing.T, target string) []runledger.Event {
	t.Helper()
	store, err := runledger.NewFileStore(target, runledger.Options{})
	if err != nil {
		t.Fatal(err)
	}
	events, err := store.LoadRun(context.Background(), "run-review")
	if err != nil {
		t.Fatal(err)
	}
	return events
}

func evidenceFile(t *testing.T, status string) string {
	t.Helper()
	result := "passed"
	if status == "blocked" || status == "escalated" {
		result = "failed"
	}
	body, err := json.Marshal(map[string]any{"checks": []map[string]string{{"name": "verificación observada", "result": result, "evidence": "PRIVATE-REVIEW-BODY-CANARY"}}})
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "review.json")
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func startReviewCLI(t *testing.T, target string) {
	t.Helper()
	code, stdout, stderr := invokeReviewCLI("review", "start", "--target", target, "--run", "run-review", "--review", "review-one", "--json")
	if code != cli.ExitOK {
		t.Fatalf("start without result/evidence rejected: %d %s", code, stderr)
	}
	if !json.Valid([]byte(stdout)) {
		t.Fatal("start must return structured JSON")
	}
}

func TestReviewCLIExplicitStartCompleteFeedsObservedMetrics(t *testing.T) {
	target := t.TempDir()
	startReviewCLI(t, target)
	code, stdout, stderr := invokeReviewCLI("review", "complete", "--target", target, "--run", "run-review", "--review", "review-one", "--result", "passed", "--evidence", evidenceFile(t, "ready"), "--json")
	if code != cli.ExitOK {
		t.Fatalf("valid reviewer contract rejected: %d %s", code, stderr)
	}
	if !json.Valid([]byte(stdout)) || strings.Contains(stdout+stderr, "PRIVATE-REVIEW-BODY-CANARY") {
		t.Fatal("CLI result must be JSON without evidence text")
	}
	metrics := application.NewService().Metrics(target)
	if metrics.Status != "available" || metrics.CompletedReviews != 1 || metrics.ReviewDurationMillis == nil || *metrics.ReviewDurationMillis < 0 {
		t.Fatalf("explicit review pair not observed by Metrics: %#v", metrics)
	}
	for _, event := range savedReviewEvents(t, target) {
		body, _ := json.Marshal(event)
		if event.Kind != runledger.KindEvidence || event.Checkpoint != nil || strings.Contains(string(body), "PRIVATE-REVIEW-BODY-CANARY") {
			t.Fatal("manual review metadata must not expose body or advance gates")
		}
	}
}

func TestReviewCLIRejectsBlockedContractClaimedAsPassedWithoutAppend(t *testing.T) {
	target := t.TempDir()
	startReviewCLI(t, target)
	code, stdout, stderr := invokeReviewCLI("review", "complete", "--target", target, "--run", "run-review", "--review", "review-one", "--result", "passed", "--evidence", evidenceFile(t, "blocked"), "--json")
	if code != cli.ExitResultRejected || len(savedReviewEvents(t, target)) != 1 {
		t.Fatal("blocked reviewer cannot be recorded as passed")
	}
	if strings.Contains(stdout+stderr, "PRIVATE-REVIEW-BODY-CANARY") {
		t.Fatal("rejected evidence body leaked")
	}
}

func TestReviewCLIRejectsFreeBodyAndInvalidIdentifiersWithoutLeaks(t *testing.T) {
	target := t.TempDir()
	startReviewCLI(t, target)
	canary := "PRIVATE FREEBODY CANARY"
	path := filepath.Join(t.TempDir(), "invalid-review.txt")
	if err := os.WriteFile(path, []byte(canary), 0o600); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := invokeReviewCLI("review", "complete", "--target", target, "--run", "run-review", "--review", "review-one", "--result", "passed", "--evidence", path, "--json")
	if code != cli.ExitResultRejected || len(savedReviewEvents(t, target)) != 1 {
		t.Fatal("free-body review input must not append")
	}
	if strings.Contains(stdout+stderr, canary) {
		t.Fatal("free-body canary leaked")
	}
	code, stdout, stderr = invokeReviewCLI("review", "start", "--target", target, "--run", "run-review", "--review", canary, "--json")
	if code == cli.ExitOK || len(savedReviewEvents(t, target)) != 1 || strings.Contains(stdout+stderr, canary) {
		t.Fatal("invalid identifiers must fail without mutation or raw echoes")
	}
}

func fingerprintTree(t *testing.T, target string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(target, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(target, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			result[relative] = "directory"
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		hash := sha256.Sum256(body)
		result[relative] = hex.EncodeToString(hash[:])
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestReviewCLIDisabledLedgerRejectsWithoutAnyWrites(t *testing.T) {
	target := t.TempDir()
	config := filepath.Join(target, ".lufy", "config", "project.yaml")
	if err := os.MkdirAll(filepath.Dir(config), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte("schema_version: 2\ntool: codex\nrun_ledger:\n  enabled: false\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	before := fingerprintTree(t, target)
	code, _, _ := invokeReviewCLI("review", "start", "--target", target, "--run", "run-review", "--review", "review-one", "--json")
	if code != cli.ExitRuntimeErr || !reflect.DeepEqual(before, fingerprintTree(t, target)) {
		t.Fatal("disabled ledger must reject before creating runtime artifacts")
	}
}

func TestReviewCLIIsVisibleInRootHelpAndStartNeedsNoResult(t *testing.T) {
	code, stdout, stderr := invokeReviewCLI("--help")
	if code != cli.ExitOK || !strings.Contains(stdout+stderr, "review") {
		t.Fatal("explicit review command must be discoverable in root help")
	}
	startReviewCLI(t, t.TempDir())
}

func TestReviewNeverReportsCriticalOrInconclusiveEvidenceAsPassed(t *testing.T) {
	for _, body := range []string{
		`{"checks":[{"name":"CLI","result":"passed","evidence":"test observado"}],"findings":[{"severity":"critical","summary":"PRIVATE-CRITICAL"}]}`,
		`{"checks":[{"name":"dato","result":"inconclusive","evidence":"consulta incompleta"}]}`,
		`{"checks":[{"name":"build","result":"not_applicable","reason":"es una redacción"}]}`,
	} {
		target := t.TempDir()
		startReviewCLI(t, target)
		path := filepath.Join(t.TempDir(), "checks.json")
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		code, out, errOut := invokeReviewCLI("review", "complete", "--target", target, "--run", "run-review", "--review", "review-one", "--result", "passed", "--evidence", path, "--json")
		if code != cli.ExitResultRejected || len(savedReviewEvents(t, target)) != 1 {
			t.Fatal("unsupported success was recorded")
		}
		if strings.Contains(out+errOut, "PRIVATE-CRITICAL") {
			t.Fatal("finding body leaked")
		}
	}
}
func TestReviewRetainsNotApplicableAsAnExplicitOutcome(t *testing.T) {
	target := t.TempDir()
	startReviewCLI(t, target)
	path := filepath.Join(t.TempDir(), "checks.json")
	if err := os.WriteFile(path, []byte(`{"checks":[{"name":"build","result":"not_applicable","reason":"No es código"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	args := []string{"review", "complete", "--target", target, "--run", "run-review", "--review", "review-one", "--result", "not_applicable", "--evidence", path, "--json"}
	code, _, errOut := invokeReviewCLI(args...)
	if code != cli.ExitOK {
		t.Fatal(errOut)
	}
	code, _, errOut = invokeReviewCLI(args...)
	if code != cli.ExitOK {
		t.Fatal(errOut)
	}
	events := savedReviewEvents(t, target)
	if len(events) != 2 || events[1].EvidenceRefs[0].Result != "not_applicable" {
		t.Fatalf("N/A became success or retry duplicated event: %#v", events)
	}
}
