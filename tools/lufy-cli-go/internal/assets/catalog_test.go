package assets

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
)

func TestBuildCatalogExpandsManagedAssetsAndExcludesOpenSpecChanges(t *testing.T) {
	source := minimalSource(t)
	catalog, err := BuildCatalog(source)
	if err != nil {
		t.Fatalf("BuildCatalog() error = %v", err)
	}

	want := map[string]bool{
		"lufy-ia.harness.md": false,
		filepath.Join(".opencode", "agents", "orchestrator.md"):                           false,
		filepath.Join(".opencode", "commands", "opsx-sync.md"):                            false,
		filepath.Join(".opencode", "commands", "opsx-version.md"):                         false,
		filepath.Join(".opencode", "hooks", "format-dispatch.sh"):                         false,
		filepath.Join(".opencode", "skills", "sdd-workflow", "openspec-sync", "SKILL.md"): false,
		filepath.Join(".opencode", "templates", "result-contract.md"):                     false,
		filepath.Join(".opencode", "templates", "sdd-lite.md"):                            false,
		filepath.Join("openspec", "config.yaml"):                                          false,
		filepath.Join("openspec", "UPSTREAM.json"):                                        false,
		filepath.Join(".lufy", "README.md"):                                               false,
		filepath.Join(".lufy", "workflows", "sdd", "README.md"):                           false,
		filepath.Join(".lufy", "workflows", "sdd", "changes", ".gitkeep"):                 false,
	}
	for _, asset := range catalog.Assets {
		if strings.HasPrefix(asset.TargetRel, filepath.Join("openspec", "changes")) {
			t.Fatalf("catalog included openspec changes asset: %s", asset.TargetRel)
		}
		if strings.HasPrefix(filepath.ToSlash(asset.TargetRel), ".lufy/memory/") {
			t.Fatalf("catalog included user-owned memory asset: %s", asset.TargetRel)
		}
		if _, ok := want[asset.TargetRel]; ok {
			want[asset.TargetRel] = true
			if asset.Kind == KindFile && asset.SourceSHA256 == "" {
				t.Fatalf("file asset %s missing hash", asset.TargetRel)
			}
			if asset.Tool != domain.ToolInitialDefault || asset.Component == "" {
				t.Fatalf("asset %s missing ownership metadata: %#v", asset.TargetRel, asset)
			}
			if strings.HasPrefix(filepath.ToSlash(asset.TargetRel), "openspec/") && asset.Methodology != domain.MethodologySpecWorkflow {
				t.Fatalf("openspec asset %s methodology = %s", asset.TargetRel, asset.Methodology)
			}
			if strings.HasPrefix(filepath.ToSlash(asset.TargetRel), ".lufy/workflows/sdd/") && asset.Methodology != domain.MethodologyLufyWorkflow {
				t.Fatalf("lufy-sdd asset %s methodology = %s", asset.TargetRel, asset.Methodology)
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
		"lufy-ia.harness.md": false,
		filepath.Join(".opencode", "agents", "orchestrator.md"):                           false,
		filepath.Join(".opencode", "commands", "opsx-sync.md"):                            false,
		filepath.Join(".opencode", "commands", "opsx-version.md"):                         false,
		filepath.Join(".opencode", "hooks", "format-dispatch.sh"):                         false,
		filepath.Join(".opencode", "skills", "sdd-workflow", "openspec-sync", "SKILL.md"): false,
		filepath.Join(".opencode", "templates", "result-contract.md"):                     false,
		filepath.Join(".opencode", "templates", "sdd-lite.md"):                            false,
		filepath.Join("openspec", "config.yaml"):                                          false,
		filepath.Join("openspec", "UPSTREAM.json"):                                        false,
		filepath.Join(".lufy", "README.md"):                                               false,
		filepath.Join(".lufy", "workflows", "sdd", "README.md"):                           false,
		filepath.Join(".lufy", "workflows", "sdd", "changes", ".gitkeep"):                 false,
	}
	for _, asset := range catalog.Assets {
		if strings.HasPrefix(asset.TargetRel, filepath.Join("openspec", "changes")) {
			t.Fatalf("embedded catalog included openspec changes asset: %s", asset.TargetRel)
		}
		if strings.HasPrefix(filepath.ToSlash(asset.TargetRel), ".lufy/memory/") {
			t.Fatalf("embedded catalog included user-owned memory asset: %s", asset.TargetRel)
		}
		if _, ok := want[asset.TargetRel]; ok {
			want[asset.TargetRel] = true
			if asset.Kind == KindFile && asset.SourceSHA256 == "" {
				t.Fatalf("embedded file asset %s missing hash", asset.TargetRel)
			}
			if asset.Tool != domain.ToolInitialDefault || asset.Component == "" {
				t.Fatalf("embedded asset %s missing ownership metadata: %#v", asset.TargetRel, asset)
			}
		}
	}
	for path, found := range want {
		if !found {
			t.Fatalf("embedded catalog missing %s", path)
		}
	}
}

func TestCatalogFingerprintIsStableForSameFileAssets(t *testing.T) {
	catalog := Catalog{Assets: []Asset{
		{TargetRel: "b.txt", Kind: KindFile, SourceSHA256: "bbb"},
		{TargetRel: "dir", Kind: KindDir},
		{TargetRel: "a.txt", Kind: KindFile, SourceSHA256: "aaa"},
	}}
	reordered := Catalog{Assets: []Asset{
		{TargetRel: "a.txt", Kind: KindFile, SourceSHA256: "aaa"},
		{TargetRel: "b.txt", Kind: KindFile, SourceSHA256: "bbb"},
	}}
	one, err := catalog.Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	two, err := reordered.Fingerprint()
	if err != nil {
		t.Fatal(err)
	}
	if one == "" || one != two {
		t.Fatalf("fingerprints mismatch: %q != %q", one, two)
	}
}

func TestEmbeddedCatalogMatchesRepositoryAssets(t *testing.T) {
	root := repoRoot(t)
	rootCatalog, err := BuildCatalog(root)
	if err != nil {
		t.Fatalf("BuildCatalog() error = %v", err)
	}
	embeddedCatalog, err := BuildEmbeddedCatalog()
	if err != nil {
		t.Fatalf("BuildEmbeddedCatalog() error = %v", err)
	}
	rootAssets := comparableAssets(rootCatalog)
	embeddedAssets := comparableAssets(embeddedCatalog)
	if !reflect.DeepEqual(rootAssets, embeddedAssets) {
		t.Fatalf("root and embedded catalogs drifted\nroot=%#v\nembedded=%#v", rootAssets, embeddedAssets)
	}
}

type comparableAsset struct {
	TargetRel    string
	Kind         Kind
	Policy       Policy
	Scope        Scope
	SourceSHA256 string
}

func comparableAssets(c Catalog) []comparableAsset {
	out := make([]comparableAsset, 0, len(c.Assets))
	for _, asset := range c.Assets {
		out = append(out, comparableAsset{TargetRel: filepath.ToSlash(asset.TargetRel), Kind: asset.Kind, Policy: asset.Policy, Scope: asset.Scope, SourceSHA256: asset.SourceSHA256})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TargetRel < out[j].TargetRel })
	return out
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if fileExists(filepath.Join(dir, "AGENTS.md")) && fileExists(filepath.Join(dir, "tools", "lufy-cli-go", "go.mod")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func minimalSource(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"AGENTS.md":                                                                       "agents root\n",
		"AGENTS.md.template":                                                              "agents template\n",
		"lufy-ia.harness.md":                                                              "agents template\n",
		"tui.json":                                                                        "{}\n",
		filepath.Join(".opencode", ".gitignore"):                                          "node_modules\n",
		filepath.Join(".opencode", "README.md"):                                           "readme\n",
		filepath.Join(".opencode", "package.json"):                                        "{}\n",
		filepath.Join(".opencode", "package-lock.json"):                                   "{}\n",
		filepath.Join(".opencode", "agents", "orchestrator.md"):                           "orchestrator\n",
		filepath.Join(".opencode", "commands", "opsx-apply.md"):                           "apply\n",
		filepath.Join(".opencode", "commands", "opsx-sync.md"):                            "sync\n",
		filepath.Join(".opencode", "commands", "opsx-version.md"):                         "version\n",
		filepath.Join(".opencode", "hooks", "format-dispatch.sh"):                         "hook\n",
		filepath.Join(".opencode", "skills", "sdd-workflow", "x.md"):                      "skill\n",
		filepath.Join(".opencode", "skills", "sdd-workflow", "openspec-sync", "SKILL.md"): "sync skill\n",
		filepath.Join(".opencode", "templates", "sdd-lite.md"):                            "lite\n",
		filepath.Join(".opencode", "templates", "result-contract.md"):                     "result\n",
		filepath.Join(".opencode", "policies", "delivery.md"):                             "delivery\n",
		filepath.Join(".opencode", "plugins", "agent-observatory.tsx"):                    "plugin\n",
		filepath.Join(".opencode", "agent-observatory", "state.ts"):                       "state\n",
		filepath.Join("openspec", "config.yaml"):                                          "config\n",
		filepath.Join("openspec", "UPSTREAM.json"):                                        "{}\n",
		filepath.Join("openspec", "README.md"):                                            "openspec\n",
		filepath.Join("openspec", "specs", ".gitkeep"):                                    "",
		filepath.Join(".lufy", "README.md"):                                               "layout\n",
		filepath.Join(".lufy", "sdd", "README.md"):                                        "lufy-sdd\n",
		filepath.Join(".lufy", "sdd", "changes", ".gitkeep"):                              "",
		filepath.Join(".lufy", "sdd", "decisions", ".gitkeep"):                            "",
		filepath.Join(".lufy", "sdd", "specs", ".gitkeep"):                                "",
		filepath.Join(".lufy", "sdd", "verification", ".gitkeep"):                         "",
		filepath.Join(".lufy", "memory", "knowledge", "private.md"):                       "private\n",
		filepath.Join("tools", "lufy-cli-go", "go.mod"):                                   "module github.com/adrianrojas/lufy-ai/tools/lufy-cli-go\n",
		filepath.Join("openspec", "changes", "active", "proposal.md"):                     "must not copy\n",
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
