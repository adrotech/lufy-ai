package application

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/domain"
)

func TestBuildWritesDeterministicArtifacts(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "go.mod"), "module example.com/demo\n")
	mustWrite(t, filepath.Join(root, "main.go"), "package main\n\nimport \"fmt\"\n\ntype User struct{}\n\nfunc main() { fmt.Println(\"ok\") }\n")
	mustWrite(t, filepath.Join(root, "README.md"), "# Demo\n\nVer tools/lufy-cli-go/main.go\n")
	mustWrite(t, filepath.Join(root, "config.yaml"), "agent:\n  path: .opencode/agents/explorer.md\n")
	mustWrite(t, filepath.Join(root, "data.json"), `{"skill":{"path":".opencode/skills/demo/SKILL.md"}}`)
	mustWrite(t, filepath.Join(root, ".env"), "SECRET_TOKEN=redacted\n")

	res, err := NewService().Build(root)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if res.Status != "ready" || res.Sources != 5 || res.Nodes == 0 || res.Edges == 0 || res.Health.SkippedFiles != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}
	if _, err := os.Stat(filepath.Join(root, ".lufy", "context", "graph.json")); err != nil {
		t.Fatalf("graph missing: %v", err)
	}
	graphBody, err := os.ReadFile(filepath.Join(root, ".lufy", "context", "graph.json"))
	if err != nil || !strings.Contains(string(graphBody), `"schema": "lufy-context-graph"`) || strings.Contains(string(graphBody), "lufy-context-graph/v1") {
		t.Fatalf("unexpected graph schema body=%s err=%v", string(graphBody), err)
	}
	if _, err := os.Stat(filepath.Join(root, ".lufy", "context", "graph-summary.md")); err != nil {
		t.Fatalf("summary missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".lufy", "context", "GRAPH_REPORT.md")); err != nil {
		t.Fatalf("report missing: %v", err)
	}
	second, err := NewService().Build(root)
	if err != nil {
		t.Fatalf("second Build() error = %v", err)
	}
	if second.Changed || second.CacheHits == 0 {
		t.Fatalf("second build should be idempotent and use cache: %+v", second)
	}
}

func TestBuildResolvesWorkflowReferencesAndPreservesDiagnosticsAcrossCache(t *testing.T) {
	root := t.TempDir()
	const (
		validTarget  = "scenario:cache-demo/traceability#explicit-links/valid-target"
		brokenTarget = "scenario:cache-demo/traceability#explicit-links/missing-target"
		taskID       = "task:cache-demo#1.1"
	)
	mustWrite(t, filepath.Join(root, ".lufy/workflows/sdd/changes/cache-demo/specs/traceability/spec.md"), strings.Join([]string{
		"# Traceability Specification",
		"### Requirement: Explicit links",
		"#### Scenario: Valid target",
		"- **WHEN** extraction runs",
		"- **THEN** the target resolves.",
		"",
	}, "\n"))
	mustWrite(t, filepath.Join(root, ".lufy/workflows/sdd/changes/cache-demo/tasks.md"), strings.Join([]string{
		"# Tasks: cache-demo",
		"- [ ] 1.1 Valid cross-file marker <!-- lufy:implements " + validTarget + " -->",
		"- [ ] 1.2 Broken cross-file marker <!-- lufy:implements " + brokenTarget + " -->",
		"",
	}, "\n"))

	service := NewService()
	first, err := service.Build(root)
	if err != nil {
		t.Fatalf("first Build() error = %v", err)
	}
	if len(first.Diagnostics) != 1 || first.Diagnostics[0].TargetID != brokenTarget {
		t.Fatalf("first build diagnostics = %#v, want bounded broken reference", first.Diagnostics)
	}
	firstGraph, err := service.store.LoadGraph(root, ".lufy/context")
	if err != nil {
		t.Fatalf("load first graph: %v", err)
	}
	if !hasGraphEdge(firstGraph.Edges, taskID, "implements", validTarget) {
		t.Fatalf("valid cross-file edge missing from graph: %#v", firstGraph.Edges)
	}
	if hasGraphEdge(firstGraph.Edges, "task:cache-demo#1.2", "implements", brokenTarget) {
		t.Fatalf("broken edge persisted in graph: %#v", firstGraph.Edges)
	}
	rawCache, err := service.store.LoadCache(root, ".lufy/context")
	if err != nil {
		t.Fatalf("load raw cache: %v", err)
	}
	if !cacheHasEdge(rawCache, "task:cache-demo#1.2", "implements", brokenTarget) {
		t.Fatalf("raw cache lost unresolved candidate: %#v", rawCache.Entries)
	}

	second, err := service.Build(root)
	if err != nil {
		t.Fatalf("second Build() error = %v", err)
	}
	if second.Changed || second.CacheHits < 2 || !reflect.DeepEqual(first.Diagnostics, second.Diagnostics) ||
		first.Nodes != second.Nodes || first.Edges != second.Edges {
		t.Fatalf("cached build is not equivalent: first=%+v second=%+v", first, second)
	}
}

func TestBuildReevaluatesCachedWorkflowCandidateWhenTargetAppears(t *testing.T) {
	root := t.TempDir()
	const target = "scenario:late-target/traceability#cache-resolution/added-later"
	tasksPath := filepath.Join(root, ".lufy/workflows/sdd/changes/late-target/tasks.md")
	mustWrite(t, tasksPath, strings.Join([]string{
		"# Tasks: late-target",
		"- [ ] 1.1 Waiting marker <!-- lufy:implements " + target + " -->",
		"",
	}, "\n"))

	service := NewService()
	first, err := service.Build(root)
	if err != nil {
		t.Fatalf("first Build() error = %v", err)
	}
	if len(first.Diagnostics) != 1 || first.Diagnostics[0].TargetID != target {
		t.Fatalf("first build should report unresolved target: %#v", first.Diagnostics)
	}
	mustWrite(t, filepath.Join(root, ".lufy/workflows/sdd/changes/late-target/specs/traceability/spec.md"), strings.Join([]string{
		"# Traceability Specification",
		"### Requirement: Cache resolution",
		"#### Scenario: Added later",
		"- **WHEN** the target source appears",
		"- **THEN** a cached marker resolves.",
		"",
	}, "\n"))

	second, err := service.Build(root)
	if err != nil {
		t.Fatalf("second Build() error = %v", err)
	}
	if !second.Changed || second.CacheHits != 1 || second.CacheMisses != 1 || len(second.Diagnostics) != 0 {
		t.Fatalf("second build should combine cached marker with new target: %+v", second)
	}
	graph, err := service.store.LoadGraph(root, ".lufy/context")
	if err != nil {
		t.Fatalf("load updated graph: %v", err)
	}
	if !hasGraphEdge(graph.Edges, "task:late-target#1.1", "implements", target) {
		t.Fatalf("cached marker did not resolve after target appeared: %#v", graph.Edges)
	}
}

func TestLimitGraphDiagnosticsIsBounded(t *testing.T) {
	diagnostics := make([]domain.Diagnostic, maxGraphDiagnostics+5)
	for index := range diagnostics {
		diagnostics[index] = domain.Diagnostic{Code: "test", Path: "fixture", Line: index + 1, Recovery: "fix fixture"}
	}
	limited := limitGraphDiagnostics(diagnostics)
	if len(limited) != maxGraphDiagnostics || limited[0].Line != 1 || limited[len(limited)-1].Line != maxGraphDiagnostics {
		t.Fatalf("bounded diagnostics = %#v", limited)
	}
}

func TestBuildExcludesManagedStateByDefault(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "README.md"), "# Demo\n")
	mustWrite(t, filepath.Join(root, ".lufy", "managed-state", "backups", "2026", "README.md"), "# Backup\n")
	mustWrite(t, filepath.Join(root, ".lufy", "managed-state", "ancestors", "README.md"), "# Ancestor\n")

	res, err := NewService().Build(root)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if res.Sources != 1 {
		t.Fatalf("expected only workspace README source, got %+v", res)
	}
	query, err := NewService().Query(root, "Backup Ancestor")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	for _, match := range query.Matches {
		if strings.Contains(match.Node.Path, ".lufy/managed-state/") {
			t.Fatalf("query returned excluded managed-state path: %+v", match)
		}
	}
}

func TestExcludedPathPatterns(t *testing.T) {
	cfg := NewService().config(t.TempDir())
	if !excludedPath(".lufy/managed-state/backups/2026/manifest.json", cfg) {
		t.Fatal("expected default backup path exclusion")
	}
	cfg.Exclude = []string{"docs/private.md", "generated/*.json"}
	if !excludedPath("docs/private.md", cfg) || !excludedPath("generated/report.json", cfg) {
		t.Fatalf("expected exact and glob exclusions")
	}
	if excludedPath("docs/public.md", cfg) {
		t.Fatal("unexpected exclusion for public path")
	}
}

func TestQueryPathAndExplain(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "main.go"), "package main\n\ntype User struct{}\nfunc Run() {}\n")
	if _, err := NewService().Build(root); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	query, err := NewService().Query(root, "User")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(query.Matches) == 0 {
		t.Fatal("expected query match")
	}
	if query.TokenSavings == "" || query.Matches[0].Score == 0 || query.Matches[0].Rank != 1 || query.Matches[0].Confidence == "" {
		t.Fatalf("expected ranked token-saving hints: %+v", query)
	}
	if len(query.Matches[0].MatchedSignals) == 0 || query.Matches[0].Relevance == "" || len(query.NextCommands) == 0 || query.Confidence == "" {
		t.Fatalf("expected actionable query diagnostics: %+v", query)
	}
	from := "file:main.go"
	to := "file:main.go#type:User"
	path, err := NewService().Path(root, from, to)
	if err != nil {
		t.Fatalf("Path() error = %v", err)
	}
	if !path.Found {
		t.Fatalf("expected path from %s to %s", from, to)
	}
	explain, err := NewService().Explain(root, to)
	if err != nil {
		t.Fatalf("Explain() error = %v", err)
	}
	if !strings.Contains(explain.Explanation, "go type") {
		t.Fatalf("unexpected explanation: %+v", explain)
	}
}

func TestStatusNotAvailable(t *testing.T) {
	res := NewService().Status(t.TempDir())
	if res.Status != "not_available" || res.Recovery != "lufy-ai context build" || len(res.NextCommands) == 0 {
		t.Fatalf("unexpected status: %+v", res)
	}
}

func TestQueryMarksNoisyLowConfidence(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "a.md"), "# Alpha\n")
	mustWrite(t, filepath.Join(root, "b.md"), "# Beta\n")
	if _, err := NewService().Build(root); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	query, err := NewService().Query(root, "a")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if !query.Noise || query.Confidence != "low" || query.Reason == "" {
		t.Fatalf("expected noisy low-confidence diagnostics: %+v", query)
	}
	if len(query.NextCommands) == 0 || !strings.Contains(strings.Join(query.NextCommands, "\n"), "<narrower path symbol issue spec>") {
		t.Fatalf("expected narrowing next command: %+v", query.NextCommands)
	}
}

func TestScanStatusReadyStaleAndRecoveries(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "main.go"), "package main\nfunc Run() {}\n")

	scan, err := NewService().Scan(root)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if scan.Status != "scan" || scan.Changed || scan.Sources != 1 {
		t.Fatalf("unexpected scan result: %+v", scan)
	}
	if _, err := os.Stat(filepath.Join(root, ".lufy", "context", "graph.json")); !os.IsNotExist(err) {
		t.Fatalf("scan should not persist graph, stat err=%v", err)
	}

	if _, err := NewService().Build(root); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	ready := NewService().Status(root)
	if ready.Status != "ready" || ready.Recovery != "" || ready.GraphPath == "" || len(ready.CandidatePaths) == 0 || len(ready.NextCommands) == 0 {
		t.Fatalf("expected ready status: %+v", ready)
	}

	mustWrite(t, filepath.Join(root, "other.md"), "# Added\n")
	stale := NewService().Status(root)
	if stale.Status != "stale" || stale.Recovery != "lufy-ai context build" {
		t.Fatalf("expected stale status after input change: %+v", stale)
	}
}

func TestExplainEdgeFallbackAndMissingGraphErrors(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "main.go"), "package main\ntype User struct{}\n")
	if _, err := NewService().Build(root); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	edgeID := "file:main.go->file:main.go#type:User"
	explain, err := NewService().Explain(root, edgeID)
	if err != nil {
		t.Fatalf("Explain(edge) error = %v", err)
	}
	if explain.Edge == nil || !strings.Contains(explain.Explanation, "extracted") {
		t.Fatalf("unexpected edge explanation: %+v", explain)
	}

	missing, err := NewService().Explain(root, "file:missing.go")
	if err != nil {
		t.Fatalf("Explain(missing) error = %v", err)
	}
	if missing.Explanation != "node_or_edge_not_found" {
		t.Fatalf("unexpected missing explanation: %+v", missing)
	}

	noGraph := t.TempDir()
	if _, err := NewService().Query(noGraph, "main"); err == nil {
		t.Fatal("Query() without graph should error")
	}
	if _, err := NewService().Path(noGraph, "a", "b"); err == nil {
		t.Fatal("Path() without graph should error")
	}
	if _, err := NewService().Explain(noGraph, "a"); err == nil {
		t.Fatal("Explain() without graph should error")
	}
	diff, err := NewService().Diff(noGraph, "HEAD")
	if err == nil || diff.Status != "not_available" || diff.Recovery != "lufy-ai context build" {
		t.Fatalf("Diff() without graph = %+v err=%v", diff, err)
	}
}

func TestDiffMapsChangedFilesToImpact(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Test User")
	mustWrite(t, filepath.Join(root, "main.go"), "package main\nfunc Run() {}\n")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "initial")
	if _, err := NewService().Build(root); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	mustWrite(t, filepath.Join(root, "main.go"), "package main\nfunc Run() {}\nfunc Added() {}\n")

	res, err := NewService().Diff(root, "HEAD")
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}
	if res.Status != "ready" || len(res.ChangedFiles) != 1 || res.ChangedFiles[0] != "main.go" || len(res.Impact) == 0 {
		t.Fatalf("unexpected diff result: %+v", res)
	}
}

func TestHelpersForCoverage(t *testing.T) {
	if recoveryIf("ready") != "" || recoveryIf("stale") != "lufy-ai context build" {
		t.Fatal("unexpected recovery helper")
	}
	if got := explanationNode(domain.Node{Path: "README.md"}); got != "node extracted deterministically from README.md" {
		t.Fatalf("unexpected fallback explanation: %q", got)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
}

func hasGraphEdge(edges []domain.Edge, from, relation, to string) bool {
	for _, edge := range edges {
		if edge.From == from && edge.Type == relation && edge.To == to {
			return true
		}
	}
	return false
}

func cacheHasEdge(cache domain.Cache, from, relation, to string) bool {
	for _, entry := range cache.Entries {
		if hasGraphEdge(entry.Edges, from, relation, to) {
			return true
		}
	}
	return false
}
