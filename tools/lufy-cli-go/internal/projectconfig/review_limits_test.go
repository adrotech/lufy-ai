package projectconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanAndMarshalProvideCanonicalReviewDefaults(t *testing.T) {
	cfg, err := Scan(t.TempDir(), fixedTime())
	if err != nil {
		t.Fatal(err)
	}
	assertReviewLimits(t, cfg.WorkflowLimits.Review, 8, 800, 3, 2)

	body, err := Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, want := range []string{"review:", "max_files_per_slice: 8", "max_churn_lines_per_slice: 800", "max_concurrent_slices: 3", "min_evidence_items: 2"} {
		if !strings.Contains(text, want) {
			t.Fatalf("canonical review config missing %q:\n%s", want, text)
		}
	}
}

func TestLoadPartialReviewLimitsKeepsMissingUnavailableAndExtras(t *testing.T) {
	path := filepath.Join(t.TempDir(), "project.yaml")
	body := []byte("schema_version: 1\nworkflow_limits:\n  review:\n    max_files_per_slice: 5\n    future_review_policy: keep\nmax_churn_lines_per_slice: 17\n")
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	assertReviewLimits(t, cfg.WorkflowLimits.Review, 5, 0, 0, 0)
	if cfg.WorkflowLimits.Review.Extra["future_review_policy"] != "keep" {
		t.Fatalf("review extras lost: %#v", cfg.WorkflowLimits.Review.Extra)
	}
	if cfg.WorkflowLimits.Review.MaxChurnLinesPerSlice == 17 {
		t.Fatal("legacy top-level max_churn_lines_per_slice must not feed canonical workflow_limits.review")
	}
}

func TestRescanCompletesReviewDefaultsAndPreservesOverrides(t *testing.T) {
	current := ProjectConfig{WorkflowLimits: WorkflowLimits{Review: WorkflowReviewLimits{
		MaxFilesPerSlice: 5,
		Extra:            map[string]any{"future_review_policy": "keep"},
	}}}
	detected, err := Scan(t.TempDir(), fixedTime())
	if err != nil {
		t.Fatal(err)
	}
	merged := MergeRescan(current, detected)
	assertReviewLimits(t, merged.WorkflowLimits.Review, 5, 800, 3, 2)
	if merged.WorkflowLimits.Review.Extra["future_review_policy"] != "keep" {
		t.Fatalf("review extras lost on rescan: %#v", merged.WorkflowLimits.Review.Extra)
	}
}

func TestLoadRejectsNegativeReviewLimitsWithCanonicalPath(t *testing.T) {
	fields := []string{"max_files_per_slice", "max_churn_lines_per_slice", "max_concurrent_slices", "min_evidence_items"}
	for _, field := range fields {
		t.Run(field, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "project.yaml")
			body := []byte("schema_version: 1\nworkflow_limits:\n  review:\n    " + field + ": -1\n")
			if err := os.WriteFile(path, body, 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := Load(path)
			if err == nil || !strings.Contains(err.Error(), "workflow_limits.review."+field) {
				t.Fatalf("Load() error = %v, want canonical field path", err)
			}
		})
	}
}

func assertReviewLimits(t *testing.T, got WorkflowReviewLimits, files, churn, concurrent, evidence int) {
	t.Helper()
	if got.MaxFilesPerSlice != files || got.MaxChurnLinesPerSlice != churn || got.MaxConcurrentSlices != concurrent || got.MinEvidenceItems != evidence {
		t.Fatalf("review limits = %#v, want files=%d churn=%d concurrent=%d evidence=%d", got, files, churn, concurrent, evidence)
	}
}
