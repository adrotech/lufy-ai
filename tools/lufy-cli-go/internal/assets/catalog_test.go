package assets

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestBuildCatalogRequiresNoHistoricalToolSources(t *testing.T) {
	source := minimalSource(t)
	for _, rel := range []string{".opencode", "openspec", ".lufy/sdd"} {
		if err := os.RemoveAll(filepath.Join(source, rel)); err != nil {
			t.Fatal(err)
		}
	}
	catalog, err := BuildCatalog(source)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Assets) == 0 {
		t.Fatal("empty catalog")
	}
	for _, a := range catalog.Assets {
		if strings.HasPrefix(a.TargetRel, ".opencode") || strings.HasPrefix(a.TargetRel, "openspec") {
			t.Fatalf("historical source: %s", a.TargetRel)
		}
		if a.Kind == KindFile && a.SourceSHA256 == "" {
			t.Fatalf("missing hash: %s", a.TargetRel)
		}
	}
}

func TestBuildCatalogRejectsSourceSymlink(t *testing.T) {
	source := minimalSource(t)
	if err := os.Remove(filepath.Join(source, ".codex", "config.toml")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(source, "AGENTS.md.template"), filepath.Join(source, ".codex", "config.toml")); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	if _, err := BuildCatalog(source); err == nil {
		t.Fatal("BuildCatalog() expected symlink error")
	}
}

func TestEmbeddedCatalogHasInlineSources(t *testing.T) {
	c, err := BuildEmbeddedCatalog()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, a := range c.Assets {
		if a.TargetRel == ".lufy/contracts/work.md" {
			found = true
		}
		if strings.HasPrefix(a.TargetRel, ".opencode") || strings.HasPrefix(a.TargetRel, "openspec") {
			t.Fatalf("historical source: %s", a.TargetRel)
		}
	}
	if !found {
		t.Fatal("missing work policy")
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
		limit := len(rootAssets)
		if len(embeddedAssets) < limit {
			limit = len(embeddedAssets)
		}
		for i := 0; i < limit; i++ {
			if rootAssets[i] != embeddedAssets[i] {
				t.Fatalf("root and embedded catalogs drifted at index %d\nroot=%#v\nembedded=%#v", i, rootAssets[i], embeddedAssets[i])
			}
		}
		t.Fatalf("root and embedded catalogs differ in length: root=%d embedded=%d", len(rootAssets), len(embeddedAssets))
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
		"AGENTS.md": "root\n", "AGENTS.md.template": "template\n", "lufy-ia.harness.md": "harness\n",
		".agents/skills/lufy-check/SKILL.md": "check\n", ".codex/config.toml": "config\n", ".lufy/contracts/work.md": "work\n", ".lufy/README.md": "readme\n",
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
