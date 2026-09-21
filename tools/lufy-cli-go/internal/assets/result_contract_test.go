package assets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecutableResultContractAssetsStayInRootEmbeddedParity(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	embeddedRoot := filepath.Join(root, "tools", "lufy-cli-go", "internal", "assets", "embedded")
	cases := []struct {
		rel  string
		want []string
	}{
		{
			rel: filepath.Join(".lufy", "contracts", "result-contract.md"),
			want: []string{
				"result-contract/v1", "lufy-ai result validate", "lufy-ai result normalize",
				"result-transition/v1", "unknown", "duplicate", "canonical", "fingerprint",
			},
		},
		{
			rel: filepath.Join(".lufy", "contracts", "result-transition.md"),
			want: []string{
				"result-transition/v1", "expected_version", "previous_fingerprint", "idempotency_key",
				"lease", "join", "accepted", "duplicate_noop", "conflict",
			},
		},
		{
			rel:  filepath.Join(".lufy", "contracts", "run-ledger-producer.md"),
			want: []string{"contract_fingerprint", "decision_status", "content-free", "duplicate_noop", "conflict", "unavailable"},
		},
		{
			rel:  filepath.Join(".lufy", "contracts", "README.md"),
			want: []string{"result-contract.md", "result-transition.md", "run-ledger-producer.md"},
		},
		{
			rel:  filepath.Join(".opencode", "templates", "result-contract.md"),
			want: []string{"canonical", ".lufy/contracts/result-contract.md", "lufy-ai result validate"},
		},
		{
			rel:  "AGENTS.md.template",
			want: []string{"Result Contract", "lufy-ai result validate", "result-transition/v1"},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.rel, func(t *testing.T) {
			t.Parallel()
			rootBody := readAssetTestFile(t, filepath.Join(root, tc.rel))
			embeddedBody := readAssetTestFile(t, filepath.Join(embeddedRoot, tc.rel))
			if rootBody != embeddedBody {
				t.Fatalf("root and embedded asset drifted: %s", tc.rel)
			}
			lower := strings.ToLower(rootBody)
			for _, marker := range tc.want {
				if !strings.Contains(lower, strings.ToLower(marker)) {
					t.Fatalf("%s missing executable contract marker %q", tc.rel, marker)
				}
			}
		})
	}
}

func TestCatalogIncludesExecutableResultTransitionContract(t *testing.T) {
	t.Parallel()

	target := filepath.Join(".lufy", "contracts", "result-transition.md")
	rootCatalog, err := BuildCatalog(repoRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	embeddedCatalog, err := BuildEmbeddedCatalog()
	if err != nil {
		t.Fatal(err)
	}
	for name, catalog := range map[string]Catalog{"root": rootCatalog, "embedded": embeddedCatalog} {
		found := false
		for _, asset := range catalog.Assets {
			if asset.TargetRel == target {
				found = true
				if asset.Kind != KindFile || asset.Policy != PolicyManaged || asset.SourceSHA256 == "" {
					t.Fatalf("%s transition contract metadata = %#v", name, asset)
				}
			}
		}
		if !found {
			t.Fatalf("%s catalog missing %s", name, target)
		}
	}
}

func readAssetTestFile(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	return string(body)
}
