package application

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/domain"
)

const (
	coverageRequirementID = "requirement:coverage-demo/traceability#workflow-coverage"
	coveredScenarioID     = "scenario:coverage-demo/traceability#workflow-coverage/covered-scenario"
	missingTaskScenarioID = "scenario:coverage-demo/traceability#workflow-coverage/missing-task"
	missingTestScenarioID = "scenario:coverage-demo/traceability#workflow-coverage/missing-test"
	coveredTaskID         = "task:coverage-demo#1.1"
	missingTestTaskID     = "task:coverage-demo#1.2"
	coveredTestID         = "test:pkg/coverage_test.go#TestCoveredScenario"
	missingTaskTestID     = "test:pkg/coverage_test.go#TestMissingTask"
	tracedFileID          = "file:pkg/coverage.go"
	tracedFunctionID      = "file:pkg/coverage.go#func:ImplementCoverage"
)

func TestCoverageReadyDistinguishesCoveredAndMissingRelationships(t *testing.T) {
	root := t.TempDir()
	writeCoverageFixture(t, root)
	service := NewService()
	if _, err := service.Build(root); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	result := service.Coverage(root)
	if result.Status != "ready" || result.GraphStatus != "ready" || result.Recovery != "" {
		t.Fatalf("Coverage() readiness = %#v", result)
	}
	if result.TotalScenarios != 3 || result.Truncated {
		t.Fatalf("Coverage() bounds = total %d truncated=%t", result.TotalScenarios, result.Truncated)
	}

	covered := requireScenarioCoverage(t, result.Scenarios, coveredScenarioID)
	if covered.Status != "covered" || len(covered.Missing) != 0 || covered.Recovery != "" {
		t.Fatalf("covered scenario = %#v", covered)
	}
	assertSupportingEdge(t, covered.SupportingEdges, coveredTaskID, "implements", coveredScenarioID)
	assertSupportingEdge(t, covered.SupportingEdges, coveredTestID, "verifies", coveredScenarioID)

	missingTask := requireScenarioCoverage(t, result.Scenarios, missingTaskScenarioID)
	assertCoverageGap(t, missingTask.Status, missingTask.Missing, "missing_task", missingTask.Path, missingTask.Span, missingTask.Recovery)
	assertSupportingEdge(t, missingTask.SupportingEdges, missingTaskTestID, "verifies", missingTaskScenarioID)

	missingTest := requireScenarioCoverage(t, result.Scenarios, missingTestScenarioID)
	assertCoverageGap(t, missingTest.Status, missingTest.Missing, "missing_test", missingTest.Path, missingTest.Span, missingTest.Recovery)
	assertSupportingEdge(t, missingTest.SupportingEdges, missingTestTaskID, "implements", missingTestScenarioID)
}

func TestCoverageResultIsBoundedWithoutHidingTotal(t *testing.T) {
	root := t.TempDir()
	var spec strings.Builder
	spec.WriteString("# Traceability Specification\n### Requirement: Bounded coverage\n")
	for index := 1; index <= 140; index++ {
		fmt.Fprintf(&spec, "#### Scenario: Bounded scenario %03d\n- **WHEN** coverage runs\n- **THEN** output stays bounded.\n", index)
	}
	mustWrite(t, filepath.Join(root, ".lufy/workflows/sdd/changes/bounded/specs/traceability/spec.md"), spec.String())
	service := NewService()
	if _, err := service.Build(root); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	result := service.Coverage(root)
	if result.Status != "ready" || result.TotalScenarios != 140 || !result.Truncated {
		t.Fatalf("bounded coverage metadata = %#v", result)
	}
	if len(result.Scenarios) == 0 || len(result.Scenarios) > 128 {
		t.Fatalf("bounded coverage returned %d scenarios, want 1..128", len(result.Scenarios))
	}
}

func TestCoverageAndTraceDegradeToUnknownForMissingOrStaleGraph(t *testing.T) {
	for _, tc := range []struct {
		name        string
		graphStatus string
		prepare     func(*testing.T, string, Service)
	}{
		{
			name: "missing", graphStatus: "not_available",
			prepare: func(t *testing.T, root string, service Service) {},
		},
		{
			name: "stale", graphStatus: "stale",
			prepare: func(t *testing.T, root string, service Service) {
				writeCoverageFixture(t, root)
				if _, err := service.Build(root); err != nil {
					t.Fatalf("Build() error = %v", err)
				}
				mustWrite(t, filepath.Join(root, "stale.md"), "# Changed after build\n")
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			service := NewService()
			tc.prepare(t, root, service)

			coverage := service.Coverage(root)
			if coverage.Status != "unknown" || coverage.GraphStatus != tc.graphStatus ||
				coverage.Recovery != "lufy-ai context build" || len(coverage.Scenarios) != 0 {
				t.Fatalf("Coverage(%s) = %#v", tc.graphStatus, coverage)
			}

			trace := service.Trace(root, tracedFileID)
			if trace.Status != "unknown" || trace.GraphStatus != tc.graphStatus || trace.Traced ||
				trace.Recovery != "lufy-ai context build" || len(trace.Path) != 0 {
				t.Fatalf("Trace(%s) = %#v", tc.graphStatus, trace)
			}
		})
	}
}

func TestTraceFollowsDefinesAndExplicitWorkflowEdgesOrReportsGap(t *testing.T) {
	root := t.TempDir()
	writeCoverageFixture(t, root)
	mustWrite(t, filepath.Join(root, "pkg/untraced.go"), "package pkg\n\nfunc CoveredScenario() {}\n")
	service := NewService()
	if _, err := service.Build(root); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	traced := service.Trace(root, tracedFileID)
	if traced.Status != "ready" || traced.GraphStatus != "ready" || !traced.Traced || traced.Recovery != "" {
		t.Fatalf("traced result = %#v", traced)
	}
	wantPath := []string{tracedFileID, tracedFunctionID, coveredTaskID, coveredScenarioID}
	if strings.Join(traced.Path, "\x00") != strings.Join(wantPath, "\x00") {
		t.Fatalf("trace path = %#v, want %#v", traced.Path, wantPath)
	}
	assertSupportingEdge(t, traced.SupportingEdges, tracedFileID, "defines", tracedFunctionID)
	assertSupportingEdge(t, traced.SupportingEdges, tracedFunctionID, "implements", coveredTaskID)
	assertSupportingEdge(t, traced.SupportingEdges, coveredTaskID, "implements", coveredScenarioID)

	untraced := service.Trace(root, "file:pkg/untraced.go")
	if untraced.Status != "gap" || untraced.GraphStatus != "ready" || untraced.Traced ||
		untraced.Reason != "untraced_file" || untraced.Recovery == "" || len(untraced.Path) != 0 {
		t.Fatalf("untraced result = %#v", untraced)
	}
}

func writeCoverageFixture(t *testing.T, root string) {
	t.Helper()
	mustWrite(t, filepath.Join(root, ".lufy/workflows/sdd/changes/coverage-demo/specs/traceability/spec.md"), strings.Join([]string{
		"# Traceability Specification",
		"### Requirement: Workflow coverage",
		"#### Scenario: Covered scenario",
		"- **WHEN** explicit task and test markers exist",
		"- **THEN** coverage is complete.",
		"#### Scenario: Missing task",
		"- **WHEN** only a test marker exists",
		"- **THEN** coverage reports missing_task.",
		"#### Scenario: Missing test",
		"- **WHEN** only a task marker exists",
		"- **THEN** coverage reports missing_test.",
		"",
	}, "\n"))
	mustWrite(t, filepath.Join(root, ".lufy/workflows/sdd/changes/coverage-demo/tasks.md"), strings.Join([]string{
		"# Tasks: coverage-demo",
		"- [ ] 1.1 Implement covered scenario <!-- lufy:implements " + coveredScenarioID + " -->",
		"- [ ] 1.2 Implement missing-test scenario <!-- lufy:implements " + missingTestScenarioID + " -->",
		"",
	}, "\n"))
	mustWrite(t, filepath.Join(root, "pkg/coverage.go"), strings.Join([]string{
		"package pkg",
		"",
		"// lufy:implements " + coveredTaskID,
		"func ImplementCoverage() {}",
		"",
	}, "\n"))
	mustWrite(t, filepath.Join(root, "pkg/coverage_test.go"), strings.Join([]string{
		"package pkg",
		"",
		"import \"testing\"",
		"",
		"// lufy:verifies " + coveredScenarioID,
		"func TestCoveredScenario(t *testing.T) {}",
		"",
		"// lufy:verifies " + missingTaskScenarioID,
		"func TestMissingTask(t *testing.T) {}",
		"",
	}, "\n"))
}

func requireScenarioCoverage(t *testing.T, scenarios []ScenarioCoverage, id string) ScenarioCoverage {
	t.Helper()
	for _, scenario := range scenarios {
		if scenario.ScenarioID == id {
			return scenario
		}
	}
	t.Fatalf("missing coverage for %s in %#v", id, scenarios)
	return ScenarioCoverage{}
}

func assertCoverageGap(t *testing.T, status string, missing []string, want, path string, span *domain.Span, recovery string) {
	t.Helper()
	if status != "gap" || !containsCoverageString(missing, want) || path == "" || span == nil || span.Line < 1 || recovery == "" {
		t.Fatalf("coverage gap status=%s missing=%#v path=%q span=%#v recovery=%q", status, missing, path, span, recovery)
	}
}

func assertSupportingEdge(t *testing.T, edges []domain.Edge, from, typ, to string) {
	t.Helper()
	for _, edge := range edges {
		if edge.From == from && edge.Type == typ && edge.To == to {
			return
		}
	}
	t.Fatalf("missing supporting edge %s --%s--> %s in %#v", from, typ, to, edges)
}

func containsCoverageString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
