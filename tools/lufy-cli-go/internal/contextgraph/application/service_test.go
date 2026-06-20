package application

import (
	"os"
	"os/exec"
	"path/filepath"
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

	res, err := NewService().Build(root)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if res.Status != "ready" || res.Sources != 4 || res.Nodes == 0 || res.Edges == 0 {
		t.Fatalf("unexpected result: %+v", res)
	}
	if _, err := os.Stat(filepath.Join(root, ".lufy", "context", "graph.json")); err != nil {
		t.Fatalf("graph missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".lufy", "context", "graph-summary.md")); err != nil {
		t.Fatalf("summary missing: %v", err)
	}

	second, err := NewService().Build(root)
	if err != nil {
		t.Fatalf("second Build() error = %v", err)
	}
	if second.Changed {
		t.Fatalf("second build should be idempotent: %+v", second)
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
	if res.Status != "not_available" || res.Recovery != "lufy-ai context build" {
		t.Fatalf("unexpected status: %+v", res)
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
	if ready.Status != "ready" || ready.Recovery != "" || ready.GraphPath == "" {
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
