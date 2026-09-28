package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

const (
	cliCoveredScenario = "scenario:coverage-cli/traceability#workflow-coverage/covered-scenario"
	cliMissingTest     = "scenario:coverage-cli/traceability#workflow-coverage/missing-test"
	cliCoveredTask     = "task:coverage-cli#1.1"
	cliTraceFile       = "file:pkg/coverage.go"
)

func TestRunContextCoverageAndTraceHumanAndJSON(t *testing.T) {
	target := t.TempDir()
	writeContextWorkflowFixture(t, target)
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"context", "build", "--target", target}, Dependencies{Stdout: &stdout, Stderr: &stderr}); code != ExitOK {
		t.Fatalf("context build code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code := Run([]string{"context", "coverage", "--target", target}, Dependencies{Stdout: &stdout, Stderr: &stderr})
	if code != ExitOK || !strings.Contains(stdout.String(), "context coverage: ready") ||
		!strings.Contains(stdout.String(), cliCoveredScenario) || !strings.Contains(stdout.String(), "missing_test") {
		t.Fatalf("coverage human code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"context", "coverage", "--target", target, "--json"}, Dependencies{Stdout: &stdout, Stderr: &stderr})
	if code != ExitOK {
		t.Fatalf("coverage JSON code=%d stderr=%s", code, stderr.String())
	}
	var coverage struct {
		Status      string `json:"status"`
		GraphStatus string `json:"graph_status"`
		Scenarios   []struct {
			ScenarioID string   `json:"scenario_id"`
			Status     string   `json:"status"`
			Missing    []string `json:"missing"`
		} `json:"scenarios"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &coverage); err != nil {
		t.Fatalf("coverage JSON invalid: %v\n%s", err, stdout.String())
	}
	if coverage.Status != "ready" || coverage.GraphStatus != "ready" || len(coverage.Scenarios) != 2 {
		t.Fatalf("coverage JSON = %#v", coverage)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"context", "trace", "--target", target, cliTraceFile}, Dependencies{Stdout: &stdout, Stderr: &stderr})
	if code != ExitOK || !strings.Contains(stdout.String(), "context trace: traced") || !strings.Contains(stdout.String(), cliCoveredScenario) {
		t.Fatalf("trace human code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"context", "trace", "--target", target, "--json", cliTraceFile}, Dependencies{Stdout: &stdout, Stderr: &stderr})
	if code != ExitOK {
		t.Fatalf("trace JSON code=%d stderr=%s", code, stderr.String())
	}
	var trace struct {
		Status      string   `json:"status"`
		GraphStatus string   `json:"graph_status"`
		Traced      bool     `json:"traced"`
		Path        []string `json:"path"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &trace); err != nil {
		t.Fatalf("trace JSON invalid: %v\n%s", err, stdout.String())
	}
	if trace.Status != "ready" || trace.GraphStatus != "ready" || !trace.Traced || len(trace.Path) != 4 {
		t.Fatalf("trace JSON = %#v", trace)
	}
}

func TestRunContextCoverageAndTraceUnknownUsageAndHelp(t *testing.T) {
	target := t.TempDir()
	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "coverage missing graph", args: []string{"context", "coverage", "--target", target, "--json"}},
		{name: "trace missing graph", args: []string{"context", "trace", "--target", target, "--json", cliTraceFile}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(tc.args, Dependencies{Stdout: &stdout, Stderr: &stderr})
			if code != ExitOK {
				t.Fatalf("Run(%v) code=%d stdout=%s stderr=%s", tc.args, code, stdout.String(), stderr.String())
			}
			var result struct {
				Status      string `json:"status"`
				GraphStatus string `json:"graph_status"`
				Recovery    string `json:"recovery"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("unknown JSON invalid: %v\n%s", err, stdout.String())
			}
			if result.Status != "unknown" || result.GraphStatus != "not_available" || result.Recovery != "lufy-ai context build" {
				t.Fatalf("unknown result = %#v", result)
			}
		})
	}

	for _, args := range [][]string{
		{"context", "coverage", "extra"},
		{"context", "trace"},
		{"context", "trace", "one", "two"},
	} {
		var stdout, stderr bytes.Buffer
		if code := Run(args, Dependencies{Stdout: &stdout, Stderr: &stderr}); code != ExitUsageErr {
			t.Fatalf("Run(%v) code=%d want=%d stdout=%s stderr=%s", args, code, ExitUsageErr, stdout.String(), stderr.String())
		}
	}

	for _, tc := range []struct {
		args []string
		want []string
	}{
		{args: []string{"context", "--help"}, want: []string{"coverage", "trace"}},
		{args: []string{"context", "coverage", "--help"}, want: []string{"context coverage", "--target", "--json"}},
		{args: []string{"context", "trace", "--help"}, want: []string{"context trace", "--target", "--json", "<node>"}},
	} {
		var stdout, stderr bytes.Buffer
		if code := Run(tc.args, Dependencies{Stdout: &stdout, Stderr: &stderr}); code != ExitOK {
			t.Fatalf("Run(%v) help code=%d stdout=%s stderr=%s", tc.args, code, stdout.String(), stderr.String())
		}
		for _, want := range tc.want {
			if !strings.Contains(stdout.String(), want) {
				t.Fatalf("Run(%v) help missing %q: %s", tc.args, want, stdout.String())
			}
		}
	}
}

func writeContextWorkflowFixture(t *testing.T, root string) {
	t.Helper()
	writeContextCLITestFile(t, filepath.Join(root, ".lufy/workflows/sdd/changes/coverage-cli/specs/traceability/spec.md"), strings.Join([]string{
		"# Traceability Specification",
		"### Requirement: Workflow coverage",
		"#### Scenario: Covered scenario",
		"- **WHEN** task and test are explicit",
		"- **THEN** coverage is complete.",
		"#### Scenario: Missing test",
		"- **WHEN** only a task is explicit",
		"- **THEN** coverage reports missing_test.",
		"",
	}, "\n"))
	writeContextCLITestFile(t, filepath.Join(root, ".lufy/workflows/sdd/changes/coverage-cli/tasks.md"), strings.Join([]string{
		"# Tasks: coverage-cli",
		"- [ ] 1.1 Covered <!-- lufy:implements " + cliCoveredScenario + " -->",
		"- [ ] 1.2 Missing test <!-- lufy:implements " + cliMissingTest + " -->",
		"",
	}, "\n"))
	writeContextCLITestFile(t, filepath.Join(root, "pkg/coverage.go"), strings.Join([]string{
		"package pkg",
		"",
		"// lufy:implements " + cliCoveredTask,
		"func ImplementCoverage() {}",
		"",
	}, "\n"))
	writeContextCLITestFile(t, filepath.Join(root, "pkg/coverage_test.go"), strings.Join([]string{
		"package pkg",
		"",
		"import \"testing\"",
		"",
		"// lufy:verifies " + cliCoveredScenario,
		"func TestCoveredScenario(t *testing.T) {}",
		"",
	}, "\n"))
}
