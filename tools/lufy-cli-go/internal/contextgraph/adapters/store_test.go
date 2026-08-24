package adapters

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/domain"
)

func TestStorePathsAndSaveLoad(t *testing.T) {
	root := t.TempDir()
	if got := ContextDir(root); got != filepath.Join(root, ".lufy", "context") {
		t.Fatalf("ContextDir = %s", got)
	}

	graph := domain.Graph{
		Schema:  domain.SchemaVersion,
		Root:    domain.Root{Name: "repo"},
		Sources: []domain.Source{{Path: "main.go", Hash: "abc", Parser: "go/parser", Status: "ok"}},
		Nodes:   []domain.Node{{ID: "file:main.go", Type: "file", Label: "main.go", Path: "main.go"}},
		Edges:   []domain.Edge{{From: "file:main.go", Type: "defines", To: "file:main.go#package:main"}},
		Manifest: domain.Manifest{
			Schema:           domain.SchemaVersion,
			ExtractorVersion: "test/v1",
			SourcesHash:      "sources",
		},
		Extensions: map[string]interface{}{},
	}
	store := NewStore()
	if err := store.Save(root, graph, "# summary\n", "# report\n", ".lufy/context"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	loaded, err := store.LoadGraph(root)
	if err != nil {
		t.Fatalf("LoadGraph() error = %v", err)
	}
	if loaded.Schema != domain.SchemaVersion || len(loaded.Nodes) != 1 {
		t.Fatalf("unexpected graph: %+v", loaded)
	}
	manifest, err := store.LoadManifest(root)
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}
	if manifest.ExtractorVersion != "test/v1" {
		t.Fatalf("unexpected manifest: %+v", manifest)
	}
	data, err := os.ReadFile(SummaryPath(root))
	if err != nil || string(data) != "# summary\n" {
		t.Fatalf("summary = %q err=%v", string(data), err)
	}
}

func TestStoreCacheAndCustomPaths(t *testing.T) {
	root := t.TempDir()
	contextRoot := ".cache/contextgraph"
	reportRel := "docs/context-report.md"

	if got := ReportPath(root); got != filepath.Join(root, ".lufy", "context", "GRAPH_REPORT.md") {
		t.Fatalf("ReportPath = %s", got)
	}
	if got := ManifestPath(root); got != filepath.Join(root, ".lufy", "context", "manifest.json") {
		t.Fatalf("ManifestPath = %s", got)
	}
	if got := CachePath(root); got != filepath.Join(root, ".lufy", "context", "cache", "extract.json") {
		t.Fatalf("CachePath = %s", got)
	}
	if got := ReportPathFor(root, contextRoot, reportRel); got != filepath.Join(root, "docs", "context-report.md") {
		t.Fatalf("ReportPathFor custom = %s", got)
	}
	if got := ReportPathFor(root, contextRoot, ""); got != filepath.Join(root, ".cache", "contextgraph", "GRAPH_REPORT.md") {
		t.Fatalf("ReportPathFor default = %s", got)
	}

	store := NewStore()
	cache := domain.Cache{
		Schema:           domain.SchemaVersion,
		ExtractorVersion: "test/cache",
		Entries: []domain.CacheEntry{{
			Source: domain.Source{Path: "main.go", Hash: "abc", Parser: "go/parser", Status: "ok"},
			Nodes:  []domain.Node{{ID: "file:main.go", Type: "file", Label: "main.go"}},
			Edges:  []domain.Edge{{From: "file:main.go", Type: "defines", To: "file:main.go#package:main"}},
		}},
	}
	if err := store.SaveCache(root, contextRoot, cache); err != nil {
		t.Fatalf("SaveCache() error = %v", err)
	}
	loaded, err := store.LoadCache(root, contextRoot)
	if err != nil {
		t.Fatalf("LoadCache() error = %v", err)
	}
	if loaded.ExtractorVersion != "test/cache" || len(loaded.Entries) != 1 {
		t.Fatalf("unexpected cache: %+v", loaded)
	}
}

func TestChangedFilesReturnsSortedDiff(t *testing.T) {
	repo := t.TempDir()
	gitStoreTest(t, repo, "init")
	gitStoreTest(t, repo, "config", "user.email", "test@example.com")
	gitStoreTest(t, repo, "config", "user.name", "Test User")
	writeStoreFile(t, filepath.Join(repo, "README.md"), "base\n")
	gitStoreTest(t, repo, "add", "README.md")
	gitStoreTest(t, repo, "commit", "-m", "base")
	base := strings.TrimSpace(gitOutputStoreTest(t, repo, "rev-parse", "HEAD"))

	writeStoreFile(t, filepath.Join(repo, "zeta.txt"), "z\n")
	writeStoreFile(t, filepath.Join(repo, "alpha.txt"), "a\n")
	gitStoreTest(t, repo, "add", "zeta.txt", "alpha.txt")
	files, err := ChangedFiles(repo, base)
	if err != nil {
		t.Fatalf("ChangedFiles() error = %v", err)
	}
	if strings.Join(files, ",") != "alpha.txt,zeta.txt" {
		t.Fatalf("ChangedFiles() = %#v", files)
	}
}

func TestLoadGraphNotAvailableForMissingInvalidOrWrongSchema(t *testing.T) {
	store := NewStore()
	root := t.TempDir()
	if _, err := store.LoadGraph(root); !errors.Is(err, ErrGraphNotAvailable) {
		t.Fatalf("missing graph err = %v", err)
	}

	writeStoreFile(t, GraphPath(root), `{`)
	if _, err := store.LoadGraph(root); !errors.Is(err, ErrGraphNotAvailable) {
		t.Fatalf("invalid json err = %v", err)
	}

	writeStoreFile(t, GraphPath(root), `{"schema":"wrong"}`)
	if _, err := store.LoadGraph(root); !errors.Is(err, ErrGraphNotAvailable) {
		t.Fatalf("wrong schema err = %v", err)
	}
}

func TestChangedFilesReportsGitErrors(t *testing.T) {
	_, err := ChangedFiles(t.TempDir(), "HEAD")
	if err == nil || !strings.Contains(err.Error(), "git diff --name-only HEAD -- fallo") {
		t.Fatalf("expected git diff error, got %v", err)
	}
}

func writeStoreFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func gitStoreTest(t *testing.T, repo string, args ...string) {
	t.Helper()
	_ = gitOutputStoreTest(t, repo, args...)
}

func gitOutputStoreTest(t *testing.T, repo string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
	return string(out)
}
