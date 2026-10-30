package harnesscatalog

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
)

func TestInlineCatalogSelectsPrincipalPolicyWithoutMethodologicalAssets(t *testing.T) {
	base, err := assets.BuildEmbeddedCatalog()
	if err != nil {
		t.Fatal(err)
	}
	effective, err := Effective(base, domain.DefaultHarnessConfig())
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, asset := range effective.Assets {
		path := filepath.ToSlash(asset.TargetRel)
		found[path] = true
		for _, forbidden := range []string{"openspec", "/sdd/", "sdd-workflow", "lufy-sdd", ".opencode", "sdd-router", "proportional-workflow"} {
			if strings.Contains(path, forbidden) {
				t.Errorf("asset retirado instalado: %s", path)
			}
		}
	}
	for _, required := range []string{".lufy/contracts/work.md", ".codex/config.toml"} {
		if !found[required] {
			t.Errorf("falta %s", required)
		}
	}
}
