package managedcontent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
)

func TestRenderInjectsImplementerValidationPermissionsFromProjectConfig(t *testing.T) {
	source := t.TempDir()
	target := t.TempDir()
	writeManagedContentTestFile(t, source, implementerAgentRel, `---
permission:
  edit:
    "go test*": allow
  task:
    "*": deny
---

body
`)
	writeManagedContentTestFile(t, target, projectconfig.ProjectConfigPath, `schema_version: 1
validation:
  allowed_commands:
    implementer:
      - pnpm typecheck*
      - pnpm lint*
`)

	body, err := Render(source, implementerAgentRel, target, implementerAgentRel)
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	for _, want := range []string{`"pnpm typecheck*": allow`, `"pnpm lint*": allow`} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing rendered permission %q in:\n%s", want, got)
		}
	}
	if strings.Index(got, `"pnpm typecheck*": allow`) > strings.Index(got, `  task:`) {
		t.Fatalf("rendered permissions must stay under permission.edit:\n%s", got)
	}
}

func TestRenderIgnoresDuplicateAndUnsafeValidationPermissions(t *testing.T) {
	source := t.TempDir()
	target := t.TempDir()
	writeManagedContentTestFile(t, source, implementerAgentRel, `---
permission:
  edit:
    "pnpm test*": allow
  task:
    "*": deny
---
`)
	writeManagedContentTestFile(t, target, projectconfig.ProjectConfigPath, "schema_version: 1\nvalidation:\n  allowed_commands:\n    implementer:\n      - pnpm test*\n      - \"bad\\ncommand\"\n")

	body, err := Render(source, implementerAgentRel, target, implementerAgentRel)
	if err != nil {
		t.Fatal(err)
	}
	if count := strings.Count(string(body), `"pnpm test*": allow`); count != 1 {
		t.Fatalf("expected exactly one pnpm test permission, got %d in:\n%s", count, body)
	}
	if strings.Contains(string(body), "bad") {
		t.Fatalf("unsafe command was rendered:\n%s", body)
	}
}

func TestRenderReturnsOriginalContentWhenConfigOrAgentShapeDoesNotApply(t *testing.T) {
	source := t.TempDir()
	target := t.TempDir()
	writeManagedContentTestFile(t, source, "README.md", "readme\n")
	writeManagedContentTestFile(t, source, implementerAgentRel, "no frontmatter\n")

	body, err := Render(source, "README.md", target, "README.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "readme\n" {
		t.Fatalf("non implementer content changed: %q", body)
	}

	body, err = Render(source, implementerAgentRel, target, implementerAgentRel)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "no frontmatter\n" {
		t.Fatalf("implementer without project config changed: %q", body)
	}
}

func TestRenderReturnsProjectConfigErrors(t *testing.T) {
	source := t.TempDir()
	target := t.TempDir()
	writeManagedContentTestFile(t, source, implementerAgentRel, `---
permission:
  edit:
    "go test*": allow
---
`)
	writeManagedContentTestFile(t, target, projectconfig.ProjectConfigPath, "schema_version: [bad\n")

	if _, err := Render(source, implementerAgentRel, target, implementerAgentRel); err == nil {
		t.Fatal("expected invalid project config to fail")
	}
}

func TestCatalogWithRenderedHashesUsesRenderedImplementerContent(t *testing.T) {
	source := t.TempDir()
	target := t.TempDir()
	writeManagedContentTestFile(t, source, implementerAgentRel, `---
permission:
  edit:
    "go test*": allow
  task:
    "*": deny
---
`)
	writeManagedContentTestFile(t, source, "README.md", "readme\n")
	writeManagedContentTestFile(t, target, projectconfig.ProjectConfigPath, `schema_version: 1
validation:
  allowed_commands:
    implementer:
      - pnpm build*
`)
	implementer, err := assets.FileSHA256(filepath.Join(source, implementerAgentRel))
	if err != nil {
		t.Fatal(err)
	}
	readme, err := assets.FileSHA256(filepath.Join(source, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	catalog := assets.Catalog{
		SourceRoot: source,
		Assets: []assets.Asset{
			{Kind: assets.KindFile, SourceRel: implementerAgentRel, TargetRel: implementerAgentRel, SourceSHA256: implementer},
			{Kind: assets.KindFile, SourceRel: "README.md", TargetRel: "README.md", SourceSHA256: readme},
		},
	}

	rendered, err := CatalogWithRenderedHashes(catalog, target)
	if err != nil {
		t.Fatal(err)
	}
	if rendered.Assets[0].SourceSHA256 == catalog.Assets[0].SourceSHA256 {
		t.Fatalf("implementer hash should reflect rendered permissions")
	}
	if rendered.Assets[1].SourceSHA256 != catalog.Assets[1].SourceSHA256 {
		t.Fatalf("unrelated asset hash changed")
	}
}

func writeManagedContentTestFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
