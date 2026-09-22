package cli

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunContextReviewHumanAndJSON(t *testing.T) {
	root, base := prepareCLIReviewRepo(t, true)
	var stdout, stderr bytes.Buffer
	args := []string{"context", "review", "--target", root, "--base", base, "--concurrent-slices", "1", "--evidence-items", "2"}
	if code := Run(args, Dependencies{Stdout: &stdout, Stderr: &stderr}); code != ExitOK || !strings.Contains(stdout.String(), "context review: proceed") || !strings.Contains(stdout.String(), "files=1") {
		t.Fatalf("review human code/output = %s stderr=%s", stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	args = append(args, "--json")
	if code := Run(args, Dependencies{Stdout: &stdout, Stderr: &stderr}); code != ExitOK {
		t.Fatalf("review JSON code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var got struct {
		Action       string                     `json:"action"`
		GraphStatus  string                     `json:"graph_status"`
		Observations struct{ Files, Churn int } `json:"observations"`
		TracedFiles  []string                   `json:"traced_files"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("invalid review JSON: %v\n%s", err, stdout.String())
	}
	if got.Action != "proceed" || got.GraphStatus != "ready" || got.Observations.Files != 1 || got.Observations.Churn == 0 || len(got.TracedFiles) != 1 {
		t.Fatalf("review JSON = %#v", got)
	}
}

func TestRunContextReviewUnknownUsageHelpAndGitError(t *testing.T) {
	root, base := prepareCLIReviewRepo(t, false)
	var stdout, stderr bytes.Buffer
	code := Run([]string{"context", "review", "--target", root, "--base", base, "--json"}, Dependencies{Stdout: &stdout, Stderr: &stderr})
	if code != ExitOK {
		t.Fatalf("unknown graph must be structured ExitOK: code=%d stderr=%s", code, stderr.String())
	}
	var unknown struct {
		Action       string `json:"action"`
		GraphStatus  string `json:"graph_status"`
		Traceability string `json:"traceability"`
		Observations struct {
			Files int `json:"files"`
		} `json:"observations"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &unknown); err != nil || unknown.Action != "escalate" || unknown.GraphStatus != "not_available" || unknown.Traceability != "unknown" || unknown.Observations.Files != 1 {
		t.Fatalf("unknown graph result = %#v err=%v body=%s", unknown, err, stdout.String())
	}

	for _, args := range [][]string{
		{"context", "review"},
		{"context", "review", "--base", "HEAD", "extra"},
		{"context", "review", "--base", "HEAD", "--concurrent-slices", "-1"},
		{"context", "review", "--base", "HEAD", "--evidence-items", "-1"},
	} {
		stdout.Reset()
		stderr.Reset()
		if code := Run(args, Dependencies{Stdout: &stdout, Stderr: &stderr}); code != ExitUsageErr {
			t.Fatalf("Run(%v) code=%d stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
		}
	}
	for _, args := range [][]string{{"context", "--help"}, {"context", "review", "--help"}} {
		stdout.Reset()
		stderr.Reset()
		if code := Run(args, Dependencies{Stdout: &stdout, Stderr: &stderr}); code != ExitOK || !strings.Contains(stdout.String(), "review") {
			t.Fatalf("Run(%v) help code=%d stdout=%s", args, code, stdout.String())
		}
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"context", "review", "--target", root, "--base", "missing-base", "--json"}, Dependencies{Stdout: &stdout, Stderr: &stderr}); code != ExitRuntimeErr {
		t.Fatalf("invalid Git base code=%d want=%d stdout=%s stderr=%s", code, ExitRuntimeErr, stdout.String(), stderr.String())
	}
}

func prepareCLIReviewRepo(t *testing.T, buildGraph bool) (string, string) {
	t.Helper()
	root := t.TempDir()
	writeContextCLITestFile(t, filepath.Join(root, ".lufy/config/project.yaml"), strings.Join([]string{
		"schema_version: 1", "workflow_limits:", "  review:", "    max_files_per_slice: 8",
		"    max_churn_lines_per_slice: 800", "    max_concurrent_slices: 3", "    min_evidence_items: 2", "",
	}, "\n"))
	writeContextCLITestFile(t, filepath.Join(root, ".lufy/workflows/sdd/changes/review-cli/tasks.md"), "# Tasks\n- [ ] 1.1 Review CLI\n")
	writeContextCLITestFile(t, filepath.Join(root, "pkg/review.go"), "package pkg\n\n// lufy:implements task:review-cli#1.1\nfunc Review() {}\n")
	gitCLIReview(t, root, "init")
	gitCLIReview(t, root, "config", "user.email", "test@example.com")
	gitCLIReview(t, root, "config", "user.name", "Test User")
	gitCLIReview(t, root, "add", ".")
	gitCLIReview(t, root, "commit", "-m", "base")
	base := strings.TrimSpace(gitCLIReview(t, root, "rev-parse", "HEAD"))
	writeContextCLITestFile(t, filepath.Join(root, "pkg/review.go"), "package pkg\n\n// lufy:implements task:review-cli#1.1\nfunc Review() {}\n// changed\n")
	if buildGraph {
		var out, errOut bytes.Buffer
		if code := Run([]string{"context", "build", "--target", root}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitOK {
			t.Fatalf("context build code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
		}
	}
	return root, base
}

func gitCLIReview(t *testing.T, root string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}
