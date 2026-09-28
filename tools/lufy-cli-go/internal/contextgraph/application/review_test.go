package application

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
)

func TestReviewProceedUsesDirectDiffCanonicalLimitsAndExplicitTrace(t *testing.T) {
	root, base := prepareReviewRepo(t, reviewFixture{buildGraph: true, tracedFiles: 1})
	result, err := NewService().Review(root, ReviewOptions{Base: base, ConcurrentSlices: 1, EvidenceItems: 2})
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "proceed" || result.GraphStatus != "ready" || result.Traceability != "complete" {
		t.Fatalf("Review() = %#v", result)
	}
	if result.Observations.Files != 1 || result.Observations.Churn == 0 || result.Observations.ConcurrentSlices != 1 || result.Observations.EvidenceItems != 2 {
		t.Fatalf("direct observations = %#v", result.Observations)
	}
	limits := []struct {
		name      string
		available bool
		value     int
		source    string
	}{
		{name: "files", available: result.Limits.MaxFilesPerSlice.Available, value: result.Limits.MaxFilesPerSlice.Value, source: result.Limits.MaxFilesPerSlice.Source},
		{name: "churn", available: result.Limits.MaxChurnLinesPerSlice.Available, value: result.Limits.MaxChurnLinesPerSlice.Value, source: result.Limits.MaxChurnLinesPerSlice.Source},
		{name: "concurrency", available: result.Limits.MaxConcurrentSlices.Available, value: result.Limits.MaxConcurrentSlices.Value, source: result.Limits.MaxConcurrentSlices.Source},
		{name: "evidence", available: result.Limits.MinEvidenceItems.Available, value: result.Limits.MinEvidenceItems.Value, source: result.Limits.MinEvidenceItems.Source},
	}
	for _, limit := range limits {
		if !limit.available || limit.value == 0 || limit.source != "workflow_limits.review" {
			t.Fatalf("canonical %s limit = available:%t value:%d source:%q", limit.name, limit.available, limit.value, limit.source)
		}
	}
	if strings.Join(result.TracedFiles, ",") != "pkg/review-00.go" || len(result.UntracedFiles) != 0 || len(result.Violations) != 0 {
		t.Fatalf("trace details = traced:%#v untraced:%#v violations:%#v", result.TracedFiles, result.UntracedFiles, result.Violations)
	}
}

func TestReviewDecisionPrecedence(t *testing.T) {
	tests := []struct {
		name       string
		fx         reviewFixture
		opts       ReviewOptions
		wantAction string
		wantGraph  string
	}{
		{name: "missing config", fx: reviewFixture{omitConfig: true, buildGraph: true, tracedFiles: 1}, opts: ReviewOptions{ConcurrentSlices: 1, EvidenceItems: 2}, wantAction: "escalate", wantGraph: "ready"},
		{name: "missing limit", fx: reviewFixture{buildGraph: true, tracedFiles: 1, missingEvidenceLimit: true}, opts: ReviewOptions{ConcurrentSlices: 1, EvidenceItems: 2}, wantAction: "escalate", wantGraph: "ready"},
		{name: "missing graph", fx: reviewFixture{tracedFiles: 1}, opts: ReviewOptions{ConcurrentSlices: 1, EvidenceItems: 2}, wantAction: "escalate", wantGraph: "not_available"},
		{name: "untraced", fx: reviewFixture{buildGraph: true, untracedFiles: 1}, opts: ReviewOptions{ConcurrentSlices: 1, EvidenceItems: 2}, wantAction: "escalate", wantGraph: "ready"},
		{name: "concurrency", fx: reviewFixture{buildGraph: true, tracedFiles: 1}, opts: ReviewOptions{ConcurrentSlices: 4, EvidenceItems: 2}, wantAction: "escalate", wantGraph: "ready"},
		{name: "evidence", fx: reviewFixture{buildGraph: true, tracedFiles: 1}, opts: ReviewOptions{ConcurrentSlices: 1, EvidenceItems: 1}, wantAction: "escalate", wantGraph: "ready"},
		{name: "files only", fx: reviewFixture{buildGraph: true, tracedFiles: 2, maxFiles: 1}, opts: ReviewOptions{ConcurrentSlices: 1, EvidenceItems: 2}, wantAction: "split", wantGraph: "ready"},
		{name: "churn only", fx: reviewFixture{buildGraph: true, tracedFiles: 1, maxChurn: 1}, opts: ReviewOptions{ConcurrentSlices: 1, EvidenceItems: 2}, wantAction: "split", wantGraph: "ready"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, base := prepareReviewRepo(t, tc.fx)
			tc.opts.Base = base
			result, err := NewService().Review(root, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if result.Action != tc.wantAction || result.GraphStatus != tc.wantGraph || result.Observations.Files == 0 || len(result.Violations) == 0 {
				t.Fatalf("Review(%s) = %#v", tc.name, result)
			}
			if tc.name == "missing config" && (result.Limits.MaxFilesPerSlice.Available || result.Limits.MaxChurnLinesPerSlice.Available || result.Limits.MaxConcurrentSlices.Available || result.Limits.MinEvidenceItems.Available) {
				t.Fatalf("missing config must make every canonical limit unavailable: %#v", result.Limits)
			}
			if tc.name == "missing limit" && result.Limits.MinEvidenceItems.Available {
				t.Fatalf("zero/missing canonical field must remain unavailable: %#v", result.Limits.MinEvidenceItems)
			}
		})
	}
}

func TestReviewMissingOrStaleGraphKeepsDiffObservations(t *testing.T) {
	root, base := prepareReviewRepo(t, reviewFixture{buildGraph: true, tracedFiles: 1})
	mustWrite(t, filepath.Join(root, "pkg/review-00.go"), "package pkg\n// changed after graph build\nfunc Review00() {}\n")
	result, err := NewService().Review(root, ReviewOptions{Base: base, ConcurrentSlices: 1, EvidenceItems: 2})
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "escalate" || result.GraphStatus != "stale" || result.Traceability != "unknown" || result.Observations.Files != 1 || result.Observations.Churn == 0 {
		t.Fatalf("stale review lost direct diff evidence: %#v", result)
	}
}

func TestReviewBoundsUntracedFilesAndViolationsWithoutHidingTotals(t *testing.T) {
	root, base := prepareReviewRepo(t, reviewFixture{buildGraph: true, untracedFiles: 140})
	result, err := NewService().Review(root, ReviewOptions{Base: base, ConcurrentSlices: 1, EvidenceItems: 2})
	if err != nil {
		t.Fatal(err)
	}
	if result.Action != "escalate" || result.Observations.Files != 140 || !result.Truncated || len(result.UntracedFiles) > 128 || len(result.Violations) > 128 {
		t.Fatalf("bounded review = observations:%#v untraced:%d violations:%d truncated:%t", result.Observations, len(result.UntracedFiles), len(result.Violations), result.Truncated)
	}
}

type reviewFixture struct {
	omitConfig, buildGraph     bool
	missingEvidenceLimit       bool
	tracedFiles, untracedFiles int
	maxFiles, maxChurn         int
}

func prepareReviewRepo(t *testing.T, fx reviewFixture) (string, string) {
	t.Helper()
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, ".lufy/workflows/sdd/changes/review-demo/tasks.md"), "# Tasks\n- [ ] 1.1 Review demo\n")
	if !fx.omitConfig {
		cfg, err := projectconfig.Scan(root, time.Unix(0, 0).UTC())
		if err != nil {
			t.Fatal(err)
		}
		if fx.maxFiles > 0 {
			cfg.WorkflowLimits.Review.MaxFilesPerSlice = fx.maxFiles
		}
		if fx.maxChurn > 0 {
			cfg.WorkflowLimits.Review.MaxChurnLinesPerSlice = fx.maxChurn
		}
		if fx.missingEvidenceLimit {
			cfg.WorkflowLimits.Review.MinEvidenceItems = 0
		}
		body, err := projectconfig.Marshal(cfg)
		if err != nil {
			t.Fatal(err)
		}
		mustWrite(t, filepath.Join(root, ".lufy/config/project.yaml"), string(body))
	}
	for i := 0; i < fx.tracedFiles; i++ {
		mustWrite(t, filepath.Join(root, "pkg", fileReviewName(i)), "package pkg\n\n// lufy:implements task:review-demo#1.1\nfunc Review"+twoDigits(i)+"() {}\n")
	}
	for i := 0; i < fx.untracedFiles; i++ {
		mustWrite(t, filepath.Join(root, "pkg", "untraced-"+twoDigits(i)+".go"), "package pkg\n\nfunc Untraced"+twoDigits(i)+"() {}\n")
	}
	gitReview(t, root, "init")
	gitReview(t, root, "config", "user.email", "test@example.com")
	gitReview(t, root, "config", "user.name", "Test User")
	gitReview(t, root, "config", "gc.auto", "0")
	gitReview(t, root, "config", "maintenance.auto", "false")
	gitReview(t, root, "add", ".")
	gitReview(t, root, "commit", "-m", "base")
	base := strings.TrimSpace(gitReview(t, root, "rev-parse", "HEAD"))
	for i := 0; i < fx.tracedFiles; i++ {
		mustWrite(t, filepath.Join(root, "pkg", fileReviewName(i)), "package pkg\n\n// lufy:implements task:review-demo#1.1\nfunc Review"+twoDigits(i)+"() {}\n// direct diff\n// second direct diff line\n")
	}
	for i := 0; i < fx.untracedFiles; i++ {
		mustWrite(t, filepath.Join(root, "pkg", "untraced-"+twoDigits(i)+".go"), "package pkg\n\nfunc Untraced"+twoDigits(i)+"() {}\n// direct diff\n// second direct diff line\n")
	}
	if fx.buildGraph {
		if _, err := NewService().Build(root); err != nil {
			t.Fatal(err)
		}
	}
	return root, base
}

func fileReviewName(i int) string { return "review-" + twoDigits(i) + ".go" }
func twoDigits(i int) string      { return fmt.Sprintf("%02d", i) }

func gitReview(t *testing.T, root string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}
