package adapters

import (
	"errors"
	"os"
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
