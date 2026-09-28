package extractors

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/domain"
)

const (
	workflowChangeID      = "change:trace-demo"
	workflowSpecID        = "spec:trace-demo/traceability"
	workflowRequirementID = "requirement:trace-demo/traceability#preserve-explicit-traceability"
	workflowScenarioID    = "scenario:trace-demo/traceability#preserve-explicit-traceability/explicit-references-are-indexed"
	workflowSimilarID     = "scenario:trace-demo/traceability#preserve-explicit-traceability/similar-names-do-not-prove-traceability"
	workflowTaskID        = "task:trace-demo#1.1"
	workflowDependentID   = "task:trace-demo#1.2"
	workflowDecisionID    = "decision:trace-demo#explicit-references-only"
	workflowTestID        = "test:pkg/trace_test.go#TestExplicitReferencesAreIndexed"
	workflowSimilarTestID = "test:pkg/trace_test.go#TestSimilarNamesDoNotProveTraceability"
	workflowIssueID       = "issue:github.com/acme/lufy#221"
	workflowPRID          = "pr:github.com/acme/lufy#42"
)

var workflowCorpusPaths = []string{
	".lufy/workflows/sdd/changes/legacy-change/proposal.md",
	".lufy/workflows/sdd/changes/trace-demo/proposal.md",
	".lufy/workflows/sdd/changes/trace-demo/specs/traceability/spec.md",
	".lufy/workflows/sdd/changes/trace-demo/tasks.md",
	".lufy/workflows/sdd/changes/trace-demo/design.md",
	"pkg/trace.go",
	"pkg/trace_test.go",
}

func TestWorkflowExtractorRecognizesDeterministicSemanticNodes(t *testing.T) {
	root := t.TempDir()
	writeWorkflowCorpus(t, root, "\n")
	result := collectWorkflowResults(root, workflowCorpusPaths)

	for _, want := range []struct {
		id  string
		typ string
	}{
		{id: "change:legacy-change", typ: "workflow_change"},
		{id: workflowChangeID, typ: "workflow_change"},
		{id: workflowSpecID, typ: "workflow_spec"},
		{id: workflowRequirementID, typ: "workflow_requirement"},
		{id: workflowScenarioID, typ: "workflow_scenario"},
		{id: workflowSimilarID, typ: "workflow_scenario"},
		{id: workflowTaskID, typ: "workflow_task"},
		{id: workflowDependentID, typ: "workflow_task"},
		{id: workflowDecisionID, typ: "workflow_decision"},
		{id: workflowTestID, typ: "workflow_test"},
		{id: workflowSimilarTestID, typ: "workflow_test"},
		{id: workflowIssueID, typ: "github_issue"},
		{id: workflowPRID, typ: "github_pull_request"},
	} {
		node := findNode(result.Nodes, want.id)
		if node == nil {
			t.Errorf("missing semantic node %s (%s)", want.id, want.typ)
			continue
		}
		if node.Type != want.typ || node.Path == "" || node.Span == nil || node.Span.Line < 1 {
			t.Errorf("semantic node %s lost type/path/span: %#v", want.id, node)
		}
		if node.Attrs["provenance"] == "" {
			t.Errorf("semantic node %s missing explicit provenance: %#v", want.id, node)
		}
	}
}

func TestWorkflowExtractorCreatesOnlyExplicitAllowlistedEdges(t *testing.T) {
	root := t.TempDir()
	writeWorkflowCorpus(t, root, "\n")
	result := collectWorkflowResults(root, workflowCorpusPaths)

	for _, want := range []struct {
		from string
		typ  string
		to   string
	}{
		{from: workflowTaskID, typ: "implements", to: workflowScenarioID},
		{from: "file:pkg/trace.go#func:BuildTrace", typ: "implements", to: workflowTaskID},
		{from: workflowTestID, typ: "verifies", to: workflowScenarioID},
		{from: workflowDependentID, typ: "depends_on", to: workflowTaskID},
		{from: workflowChangeID, typ: "caused_by", to: workflowIssueID},
		{from: workflowPRID, typ: "reviewed_by", to: workflowIssueID},
		{from: workflowChangeID, typ: "supersedes", to: "change:legacy-change"},
	} {
		edge := findWorkflowEdge(result.Edges, want.from, want.typ, want.to)
		if edge == nil {
			t.Errorf("missing explicit edge %s --%s--> %s", want.from, want.typ, want.to)
			continue
		}
		if edge.Reason == "" || edge.Attrs["provenance"] != "explicit_marker" {
			t.Errorf("edge lacks reason/provenance: %#v", edge)
		}
	}

	for _, edge := range result.Edges {
		if (edge.Type == "implements" || edge.Type == "verifies") &&
			(edge.From == workflowDependentID || edge.From == workflowSimilarTestID ||
				edge.From == "file:pkg/trace.go#func:IndexExplicitReferences" || edge.To == workflowSimilarID) {
			t.Errorf("similar names created fuzzy evidence edge: %#v", edge)
		}
	}
}

func TestWorkflowExtractorReportsBrokenReferenceWithoutEvidenceEdge(t *testing.T) {
	root := t.TempDir()
	writeWorkflowCorpus(t, root, "\n")
	const missing = "scenario:trace-demo/traceability#preserve-explicit-traceability/missing-scenario"
	path := ".lufy/workflows/sdd/changes/trace-demo/tasks.md"
	writeExtractorFixture(t, root, path, strings.Join([]string{
		"# Tasks: trace-demo",
		"- [ ] 1.1 Broken reference <!-- lufy:implements " + missing + " -->",
		"",
	}, "\n"))

	result := collectWorkflowResults(root, []string{
		".lufy/workflows/sdd/changes/trace-demo/specs/traceability/spec.md",
		path,
	})
	if edge := findWorkflowEdge(result.Edges, workflowTaskID, "implements", missing); edge != nil {
		t.Fatalf("broken reference created evidence edge: %#v", edge)
	}
	diagnostics := workflowDiagnostics(t, result)
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics=%#v, want one bounded unresolved reference", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Code != "unresolved_workflow_reference" || diagnostic.Path != path || diagnostic.Line != 2 ||
		diagnostic.Relation != "implements" || diagnostic.TargetID != missing || diagnostic.Recovery == "" {
		t.Fatalf("broken reference diagnostic=%#v", diagnostic)
	}
}

func TestWorkflowExtractorOrderingAndIdentityAreStableAcrossCRLF(t *testing.T) {
	lfRoot := t.TempDir()
	crlfRoot := t.TempDir()
	writeWorkflowCorpus(t, lfRoot, "\n")
	writeWorkflowCorpus(t, crlfRoot, "\r\n")

	lf := collectWorkflowResults(lfRoot, workflowCorpusPaths)
	crlf := collectWorkflowResults(crlfRoot, workflowCorpusPaths)
	assertWorkflowOrdering(t, lf)
	assertWorkflowOrdering(t, crlf)
	if !reflect.DeepEqual(lf.Nodes, crlf.Nodes) || !reflect.DeepEqual(lf.Edges, crlf.Edges) ||
		!reflect.DeepEqual(lf.Diagnostics, crlf.Diagnostics) {
		t.Fatalf("workflow extraction differs for CRLF")
	}
	repeated := collectWorkflowResults(lfRoot, workflowCorpusPaths)
	if !reflect.DeepEqual(lf.Nodes, repeated.Nodes) || !reflect.DeepEqual(lf.Edges, repeated.Edges) ||
		!reflect.DeepEqual(lf.Diagnostics, repeated.Diagnostics) {
		t.Fatal("workflow extraction is not repeatable")
	}
}

func TestWorkflowExtractorKeepsCanariesAndRuntimePathsOutOfMetadata(t *testing.T) {
	const (
		secretCanary = "SECRET_CANARY_workflow_private"
		promptCanary = "PROMPT_CANARY_workflow_ignore_previous"
		outputCanary = "OUTPUT_CANARY_workflow_stdout"
		runtimePath  = ".lufy/runtime/runs/private-run/events/000001.json"
	)
	root := t.TempDir()
	writeWorkflowCorpus(t, root, "\n")
	path := ".lufy/workflows/sdd/changes/trace-demo/tasks.md"
	writeExtractorFixture(t, root, path, strings.Join([]string{
		"# Tasks: trace-demo",
		"- [ ] 1.1 Safe task <!-- lufy:implements scenario:" + secretCanary + " -->",
		"Prose only: " + promptCanary + " " + outputCanary + " " + runtimePath,
		"",
	}, "\n"))

	result := ResolveWorkflowReferences(Extract(root, path))
	metadata := workflowMetadataJSON(t, result)
	for _, forbidden := range []string{secretCanary, promptCanary, outputCanary, runtimePath} {
		if strings.Contains(metadata, forbidden) {
			t.Fatalf("workflow metadata leaked %q: %s", forbidden, metadata)
		}
	}
}

func writeWorkflowCorpus(t *testing.T, root, eol string) {
	t.Helper()
	files := map[string]string{
		".lufy/workflows/sdd/changes/legacy-change/proposal.md": strings.Join([]string{
			"# Proposal: legacy-change",
			"",
		}, "\n"),
		".lufy/workflows/sdd/changes/trace-demo/proposal.md": strings.Join([]string{
			"# Proposal: trace-demo",
			"<!-- lufy:caused_by " + workflowIssueID + " -->",
			"<!-- lufy:supersedes change:legacy-change -->",
			"Issue: https://github.com/acme/lufy/issues/221",
			"Pull request: https://github.com/acme/lufy/pull/42 <!-- lufy:reviewed_by " + workflowIssueID + " -->",
			"",
		}, "\n"),
		".lufy/workflows/sdd/changes/trace-demo/specs/traceability/spec.md": strings.Join([]string{
			"# Traceability Specification",
			"",
			"### Requirement: Preserve explicit traceability",
			"",
			"#### Scenario: Explicit references are indexed",
			"- **WHEN** extraction runs",
			"- **THEN** the graph contains typed edges.",
			"",
			"#### Scenario: Similar names do not prove traceability",
			"- **WHEN** no marker exists",
			"- **THEN** no evidence edge is created.",
			"",
		}, "\n"),
		".lufy/workflows/sdd/changes/trace-demo/tasks.md": strings.Join([]string{
			"# Tasks: trace-demo",
			"- [ ] 1.1 Index explicit references <!-- lufy:implements " + workflowScenarioID + " -->",
			"- [ ] 1.2 Similar names do not prove traceability <!-- lufy:depends_on " + workflowTaskID + " -->",
			"",
		}, "\n"),
		".lufy/workflows/sdd/changes/trace-demo/design.md": strings.Join([]string{
			"# Design: trace-demo",
			"",
			"### Decision: Explicit references only",
			"",
		}, "\n"),
		"pkg/trace.go": strings.Join([]string{
			"package trace",
			"",
			"// lufy:implements " + workflowTaskID,
			"func BuildTrace() {}",
			"",
			"func IndexExplicitReferences() {}",
			"",
		}, "\n"),
		"pkg/trace_test.go": strings.Join([]string{
			"package trace",
			"",
			"import \"testing\"",
			"",
			"// lufy:verifies " + workflowScenarioID,
			"func TestExplicitReferencesAreIndexed(t *testing.T) {}",
			"",
			"func TestSimilarNamesDoNotProveTraceability(t *testing.T) {}",
			"",
		}, "\n"),
	}
	for path, content := range files {
		if eol != "\n" {
			content = strings.ReplaceAll(content, "\n", eol)
		}
		writeExtractorFixture(t, root, path, content)
	}
}

func collectWorkflowResults(root string, paths []string) domain.ExtractResult {
	var combined domain.ExtractResult
	for _, path := range paths {
		result := Extract(root, path)
		combined.Nodes = append(combined.Nodes, result.Nodes...)
		combined.Edges = append(combined.Edges, result.Edges...)
		combined.Diagnostics = append(combined.Diagnostics, result.Diagnostics...)
	}
	return ResolveWorkflowReferences(combined)
}

func findWorkflowEdge(edges []domain.Edge, from, typ, to string) *domain.Edge {
	for index := range edges {
		if edges[index].From == from && edges[index].Type == typ && edges[index].To == to {
			return &edges[index]
		}
	}
	return nil
}

func assertWorkflowOrdering(t *testing.T, result domain.ExtractResult) {
	t.Helper()
	if !sort.SliceIsSorted(result.Nodes, func(i, j int) bool { return result.Nodes[i].ID < result.Nodes[j].ID }) {
		t.Fatalf("nodes are not deterministically ordered: %#v", result.Nodes)
	}
	if !sort.SliceIsSorted(result.Edges, func(i, j int) bool {
		left := result.Edges[i].From + "\x00" + result.Edges[i].Type + "\x00" + result.Edges[i].To
		right := result.Edges[j].From + "\x00" + result.Edges[j].Type + "\x00" + result.Edges[j].To
		return left < right
	}) {
		t.Fatalf("edges are not deterministically ordered: %#v", result.Edges)
	}
}

type workflowDiagnosticContract struct {
	Code     string `json:"code"`
	Path     string `json:"path"`
	Line     int    `json:"line"`
	Relation string `json:"relation"`
	TargetID string `json:"target_id"`
	Recovery string `json:"recovery"`
}

func workflowDiagnostics(t *testing.T, result domain.ExtractResult) []workflowDiagnosticContract {
	t.Helper()
	field := reflect.ValueOf(result).FieldByName("Diagnostics")
	if !field.IsValid() {
		t.Fatal("domain.ExtractResult must expose Diagnostics []domain.Diagnostic")
	}
	body, err := json.Marshal(field.Interface())
	if err != nil {
		t.Fatal(err)
	}
	var diagnostics []workflowDiagnosticContract
	if err := json.Unmarshal(body, &diagnostics); err != nil {
		t.Fatalf("decode diagnostics contract: %v", err)
	}
	return diagnostics
}

func workflowMetadataJSON(t *testing.T, result domain.ExtractResult) string {
	t.Helper()
	metadata := struct {
		Nodes       []map[string]any             `json:"nodes"`
		Edges       []map[string]any             `json:"edges"`
		Diagnostics []workflowDiagnosticContract `json:"diagnostics"`
	}{Diagnostics: workflowDiagnostics(t, result)}
	for _, node := range result.Nodes {
		metadata.Nodes = append(metadata.Nodes, map[string]any{"attrs": node.Attrs, "reason": node.Reason})
	}
	for _, edge := range result.Edges {
		metadata.Edges = append(metadata.Edges, map[string]any{"attrs": edge.Attrs, "reason": edge.Reason})
	}
	body, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}
