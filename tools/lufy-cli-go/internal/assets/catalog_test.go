package assets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildCatalogExpandsManagedAssetsAndExcludesOpenSpecChanges(t *testing.T) {
	source := minimalSource(t)
	catalog, err := BuildCatalog(source)
	if err != nil {
		t.Fatalf("BuildCatalog() error = %v", err)
	}

	want := map[string]bool{
		"AGENTS.md": false,
		filepath.Join(".opencode", "agents", "orchestrator.md"): false,
		filepath.Join("openspec", "config.yaml"):                false,
	}
	for _, asset := range catalog.Assets {
		if strings.HasPrefix(asset.TargetRel, filepath.Join("openspec", "changes")) {
			t.Fatalf("catalog included openspec changes asset: %s", asset.TargetRel)
		}
		if _, ok := want[asset.TargetRel]; ok {
			want[asset.TargetRel] = true
			if asset.Kind == KindFile && asset.SourceSHA256 == "" {
				t.Fatalf("file asset %s missing hash", asset.TargetRel)
			}
		}
	}
	for path, found := range want {
		if !found {
			t.Fatalf("catalog missing %s", path)
		}
	}
}

func TestBuildCatalogRejectsSourceSymlink(t *testing.T) {
	source := minimalSource(t)
	if err := os.Remove(filepath.Join(source, ".opencode", "agents", "orchestrator.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(source, "tui.json"), filepath.Join(source, ".opencode", "agents", "orchestrator.md")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	if _, err := BuildCatalog(source); err == nil {
		t.Fatal("BuildCatalog() expected symlink error")
	}
}

func TestBuildEmbeddedCatalogIncludesManagedAssetsAndExcludesOpenSpecChanges(t *testing.T) {
	catalog, err := BuildEmbeddedCatalog()
	if err != nil {
		t.Fatalf("BuildEmbeddedCatalog() error = %v", err)
	}
	if catalog.SourceRoot != EmbeddedSourceRoot {
		t.Fatalf("SourceRoot = %q, want %q", catalog.SourceRoot, EmbeddedSourceRoot)
	}

	want := map[string]bool{
		"AGENTS.md": false,
		filepath.Join(".opencode", "agents", "orchestrator.md"): false,
		filepath.Join("openspec", "config.yaml"):                false,
	}
	for _, asset := range catalog.Assets {
		if strings.HasPrefix(asset.TargetRel, filepath.Join("openspec", "changes")) {
			t.Fatalf("embedded catalog included openspec changes asset: %s", asset.TargetRel)
		}
		if _, ok := want[asset.TargetRel]; ok {
			want[asset.TargetRel] = true
			if asset.Kind == KindFile && asset.SourceSHA256 == "" {
				t.Fatalf("embedded file asset %s missing hash", asset.TargetRel)
			}
		}
	}
	for path, found := range want {
		if !found {
			t.Fatalf("embedded catalog missing %s", path)
		}
	}
}

func minimalSource(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"AGENTS.md":                                                    "agents root\n",
		"AGENTS.md.template":                                           "agents template\n",
		"tui.json":                                                     "{}\n",
		filepath.Join(".opencode", ".gitignore"):                       "node_modules\n",
		filepath.Join(".opencode", "README.md"):                        "readme\n",
		filepath.Join(".opencode", "package.json"):                     "{}\n",
		filepath.Join(".opencode", "package-lock.json"):                "{}\n",
		filepath.Join(".opencode", "agents", "orchestrator.md"):        "orchestrator\n",
		filepath.Join(".opencode", "commands", "opsx-apply.md"):        "apply\n",
		filepath.Join(".opencode", "skills", "sdd-workflow", "x.md"):   "skill\n",
		filepath.Join(".opencode", "policies", "delivery.md"):          "delivery\n",
		filepath.Join(".opencode", "plugins", "agent-observatory.tsx"): "plugin\n",
		filepath.Join(".opencode", "agent-observatory", "state.ts"):    "state\n",
		filepath.Join("openspec", "config.yaml"):                       "config\n",
		filepath.Join("openspec", "README.md"):                         "openspec\n",
		filepath.Join("openspec", "specs", ".gitkeep"):                 "",
		filepath.Join("tools", "lufy-cli-go", "go.mod"):                "module github.com/adrianrojas/lufy-ai/tools/lufy-cli-go\n",
		filepath.Join("openspec", "changes", "active", "proposal.md"):  "must not copy\n",
	}
	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
