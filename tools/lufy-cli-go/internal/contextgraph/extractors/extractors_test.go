package extractors

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/domain"
)

func TestSupportedAndParserName(t *testing.T) {
	cases := map[string]string{
		"main.go":     "go/parser",
		"README.md":   "markdown/linear",
		"config.yaml": "yaml.v3",
		"config.yml":  "yaml.v3",
		"data.json":   "encoding/json",
		"notes.txt":   "unsupported",
	}
	for path, parser := range cases {
		t.Run(path, func(t *testing.T) {
			if got := ParserName(path); got != parser {
				t.Fatalf("ParserName(%q) = %q, want %q", path, got, parser)
			}
			if got := Supported(path); got != (parser != "unsupported") {
				t.Fatalf("Supported(%q) = %t", path, got)
			}
		})
	}
}

func TestExtractGoMarkdownYAMLAndJSON(t *testing.T) {
	root := t.TempDir()
	writeExtractorFixture(t, root, "pkg/demo.go", `package demo

import "fmt"

type User struct{}

func Run() {}
func (u *User) Name() string { return fmt.Sprint("user") }
`)
	writeExtractorFixture(t, root, "pkg/demo_test.go", `package demo

func TestRun() {}
`)
	writeExtractorFixture(t, root, "README.md", "# Demo Title\n\nSee tools/lufy-cli-go/main.go and https://example.com/skip.\n")
	writeExtractorFixture(t, root, "config.yaml", "agent:\n  path: .opencode/agents/explorer.md\n")
	writeExtractorFixture(t, root, "data.json", `{"skill":{"path":".opencode/skills/demo/SKILL.md"},"codex_skill":{"path":".agents/skills/demo/SKILL.md"},"alpha":true}`)
	writeExtractorFixture(t, root, ".opencode/agents/explorer.md", "# Explorer\n")
	writeExtractorFixture(t, root, ".opencode/skills/demo/SKILL.md", "# Demo skill\n")
	writeExtractorFixture(t, root, ".opencode/commands/demo.md", "# Demo command\n")
	writeExtractorFixture(t, root, ".agents/skills/demo/SKILL.md", "# Codex skill\n")

	goRes := Extract(root, "pkg/demo.go")
	assertNode(t, goRes.Nodes, "file:pkg/demo.go#package:demo", "go_package")
	assertNode(t, goRes.Nodes, "file:pkg/demo.go#import:fmt", "go_import")
	assertNode(t, goRes.Nodes, "file:pkg/demo.go#type:User", "go_type")
	assertNode(t, goRes.Nodes, "file:pkg/demo.go#func:Run", "go_function")
	assertNode(t, goRes.Nodes, "file:pkg/demo.go#func:User.Name", "go_method")
	assertEdge(t, goRes.Edges, "dir:pkg", "contains", "file:pkg/demo.go")

	testRes := Extract(root, "pkg/demo_test.go")
	node := findNode(testRes.Nodes, "file:pkg/demo_test.go#func:TestRun")
	if node == nil || node.Attrs["test"] != "true" {
		t.Fatalf("expected test function attr, got %#v", node)
	}

	mdRes := Extract(root, "README.md")
	assertNode(t, mdRes.Nodes, "file:README.md#markdown", "markdown_document")
	assertNode(t, mdRes.Nodes, "file:README.md#heading:1:demo-title", "markdown_heading")
	assertEdge(t, mdRes.Edges, "file:README.md#markdown", "references", "file:tools/lufy-cli-go/main.go")

	yamlRes := Extract(root, "config.yaml")
	assertNode(t, yamlRes.Nodes, "file:config.yaml#yaml_key:agent.path", "yaml_key")
	assertEdge(t, yamlRes.Edges, "file:config.yaml#yaml_key:agent.path", "references", "file:.opencode/agents/explorer.md")

	jsonRes := Extract(root, "data.json")
	assertNode(t, jsonRes.Nodes, "file:data.json#json_key:alpha", "json_key")
	assertNode(t, jsonRes.Nodes, "file:data.json#json_key:skill.path", "json_key")
	assertEdge(t, jsonRes.Edges, "file:data.json#json_key:skill.path", "references", "file:.opencode/skills/demo/SKILL.md")
	assertEdge(t, jsonRes.Edges, "file:data.json#json_key:codex_skill.path", "references", "file:.agents/skills/demo/SKILL.md")

	assertNode(t, Extract(root, ".opencode/agents/explorer.md").Nodes, "file:.opencode/agents/explorer.md", "opencode_agent")
	assertNode(t, Extract(root, ".opencode/skills/demo/SKILL.md").Nodes, "file:.opencode/skills/demo/SKILL.md", "opencode_skill")
	assertNode(t, Extract(root, ".opencode/commands/demo.md").Nodes, "file:.opencode/commands/demo.md", "opencode_command")
	assertNode(t, Extract(root, ".agents/skills/demo/SKILL.md").Nodes, "file:.agents/skills/demo/SKILL.md", "codex_skill")
}

func TestExtractReportsReadAndParseErrors(t *testing.T) {
	root := t.TempDir()
	missing := Extract(root, "missing.go")
	if missing.Source.Status != "error" || missing.Source.Error == "" {
		t.Fatalf("missing file should report source error: %+v", missing.Source)
	}

	writeExtractorFixture(t, root, "bad.json", `{`)
	badJSON := Extract(root, "bad.json")
	if badJSON.Source.Status != "error" || badJSON.Source.Hash == "" {
		t.Fatalf("bad json should retain hash and report error: %+v", badJSON.Source)
	}

	writeExtractorFixture(t, root, "bad.yaml", "root: [unterminated")
	badYAML := Extract(root, "bad.yaml")
	if badYAML.Source.Status != "error" || badYAML.Source.Hash == "" {
		t.Fatalf("bad yaml should retain hash and report error: %+v", badYAML.Source)
	}

	writeExtractorFixture(t, root, "bad.go", "package main\nfunc")
	badGo := Extract(root, "bad.go")
	if badGo.Source.Status != "error" || badGo.Source.Hash == "" {
		t.Fatalf("bad go should retain hash and report error: %+v", badGo.Source)
	}
}

func writeExtractorFixture(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertNode(t *testing.T, nodes []domain.Node, id, typ string) {
	t.Helper()
	node := findNode(nodes, id)
	if node == nil {
		t.Fatalf("missing node %s in %#v", id, nodes)
	}
	if node.Type != typ {
		t.Fatalf("node %s type = %s, want %s", id, node.Type, typ)
	}
}

func findNode(nodes []domain.Node, id string) *domain.Node {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
	}
	return nil
}

func assertEdge(t *testing.T, edges []domain.Edge, from, typ, to string) {
	t.Helper()
	for _, edge := range edges {
		if edge.From == from && edge.Type == typ && edge.To == to {
			return
		}
	}
	t.Fatalf("missing edge %s --%s--> %s in %#v", from, typ, to, edges)
}
