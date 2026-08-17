package skillregistry

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

func TestBuildPrefersProjectSkillAndPreservesExactEvidence(t *testing.T) {
	target := t.TempDir()
	projectRoot := filepath.Join(target, "project-skills")
	globalRoot := filepath.Join(target, "global-skills")
	globalPath := writeSkill(t, globalRoot, "reviewer", "reviewer", "global description")
	projectPath := writeSkill(t, projectRoot, "reviewer", "reviewer", "project\nmultiline description")
	writeSkill(t, globalRoot, "delivery", "delivery", "deliver safely")

	index, err := Build(domain.ToolInitialDefault, []ports.SkillRoot{{Path: globalRoot, Scope: "global"}, {Path: projectRoot, Scope: "project"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(index.Skills) != 2 {
		t.Fatalf("skills = %+v", index.Skills)
	}
	var reviewer Skill
	for _, skill := range index.Skills {
		if skill.Name == "reviewer" {
			reviewer = skill
		}
	}
	if reviewer.Path != projectPath || reviewer.Scope != "project" || reviewer.Description != "project\nmultiline description" {
		t.Fatalf("reviewer = %+v", reviewer)
	}
	if len(reviewer.ShadowedPaths) != 1 || reviewer.ShadowedPaths[0] != globalPath {
		t.Fatalf("shadowed evidence = %+v", reviewer.ShadowedPaths)
	}
}

func TestBuildIsDeterministicAndDoesNotModifySkills(t *testing.T) {
	root := filepath.Join(t.TempDir(), "skills")
	path := writeSkill(t, root, "safe", "safe", "read only")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	first, err := Build(domain.ToolCodex, []ports.SkillRoot{{Path: root, Scope: "project"}})
	if err != nil {
		t.Fatal(err)
	}
	second, err := Build(domain.ToolCodex, []ports.SkillRoot{{Path: root, Scope: "project"}})
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, _ := Marshal(first)
	secondJSON, _ := Marshal(second)
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("registry output is not deterministic\n%s\n%s", firstJSON, secondJSON)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("Build modified a source SKILL.md")
	}
}

func TestBuildUsesAdapterPriorityWithinProjectScope(t *testing.T) {
	target := t.TempDir()
	nativeRoot := filepath.Join(target, ".opencode", "skills")
	compatibleRoot := filepath.Join(target, ".agents", "skills")
	nativePath := writeSkill(t, nativeRoot, "reviewer", "reviewer", "native")
	compatiblePath := writeSkill(t, compatibleRoot, "reviewer", "reviewer", "compatible")
	index, err := Build(domain.ToolInitialDefault, []ports.SkillRoot{
		{Path: compatibleRoot, Scope: "project", Priority: 10},
		{Path: nativeRoot, Scope: "project", Priority: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(index.Skills) != 1 || index.Skills[0].Path != nativePath || len(index.Skills[0].ShadowedPaths) != 1 || index.Skills[0].ShadowedPaths[0] != compatiblePath {
		t.Fatalf("priority selection = %+v", index.Skills)
	}
}

func TestBuildSkipsSymlinkedSkills(t *testing.T) {
	target := t.TempDir()
	external := filepath.Join(t.TempDir(), "external")
	writeSkill(t, external, "escaped", "escaped", "outside")
	root := filepath.Join(target, "skills")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(root, "linked")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	index, err := Build(domain.ToolInitialDefault, []ports.SkillRoot{{Path: root, Scope: "project"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(index.Skills) != 0 {
		t.Fatalf("symlinked skill was indexed: %+v", index.Skills)
	}
}

func TestBuildWarnsAndSkipsInvalidMetadata(t *testing.T) {
	root := filepath.Join(t.TempDir(), "skills")
	path := filepath.Join(root, "broken", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("# missing frontmatter\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	index, err := Build(domain.ToolInitialDefault, []ports.SkillRoot{{Path: root, Scope: "project"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(index.Skills) != 0 || len(index.Warnings) != 1 || index.Warnings[0].Path != path {
		t.Fatalf("index = %+v", index)
	}
}

func writeSkill(t *testing.T, root, dir, name, description string) string {
	t.Helper()
	path := filepath.Join(root, dir, "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: " + name + "\ndescription: |-\n  " + bytes.NewBufferString(description).String() + "\n---\n\n# " + name + "\n"
	content = bytes.NewBufferString(content).String()
	content = string(bytes.ReplaceAll([]byte(content), []byte("\nmultiline"), []byte("\n  multiline")))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(abs)
}
