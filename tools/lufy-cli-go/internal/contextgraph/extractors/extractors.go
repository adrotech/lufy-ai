package extractors

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/domain"
)

const Version = "contextgraph-extractors/v1"

var pathRefRE = regexp.MustCompile(`(?:[A-Za-z0-9_.-]+/)+[A-Za-z0-9_.-]+(?:\.[A-Za-z0-9_.-]+)?`)

func Supported(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".go" || ext == ".md" || ext == ".yaml" || ext == ".yml" || ext == ".json"
}

func ParserName(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go/parser"
	case ".md":
		return "markdown/linear"
	case ".yaml", ".yml":
		return "yaml.v3"
	case ".json":
		return "encoding/json"
	default:
		return "unsupported"
	}
}

func Extract(root, rel string) domain.ExtractResult {
	abs := filepath.Join(root, filepath.FromSlash(rel))
	data, err := os.ReadFile(abs)
	source := domain.Source{Path: rel, Parser: ParserName(rel), Status: "ok"}
	if err != nil {
		source.Status, source.Error = "error", err.Error()
		return domain.ExtractResult{Source: source}
	}
	sum := sha256.Sum256(data)
	source.Hash = hex.EncodeToString(sum[:])
	fileNode := domain.Node{ID: fileID(rel), Type: fileNodeType(rel), Label: filepath.Base(rel), Path: rel, Reason: "workspace source file"}
	res := domain.ExtractResult{Source: source, Nodes: []domain.Node{fileNode}}
	if dir := filepath.ToSlash(filepath.Dir(rel)); dir != "." {
		res.Nodes = append(res.Nodes, domain.Node{ID: "dir:" + dir, Type: "directory", Label: filepath.Base(dir), Path: dir, Reason: "parent directory"})
		res.Edges = append(res.Edges, domain.Edge{From: "dir:" + dir, Type: "contains", To: fileNode.ID, Reason: "directory contains file"})
	}
	switch strings.ToLower(filepath.Ext(rel)) {
	case ".go":
		extractGo(root, rel, data, &res)
	case ".md":
		extractMarkdown(rel, string(data), &res)
	case ".yaml", ".yml":
		extractYAML(rel, data, &res)
	case ".json":
		extractJSON(rel, data, &res)
	}
	normalize(&res)
	return res
}

func fileID(rel string) string { return "file:" + filepath.ToSlash(rel) }

func fileNodeType(rel string) string {
	rel = filepath.ToSlash(rel)
	switch {
	case strings.HasPrefix(rel, ".opencode/agents/"):
		return "opencode_agent"
	case strings.HasPrefix(rel, ".opencode/skills/"):
		return "opencode_skill"
	case strings.HasPrefix(rel, ".opencode/commands/"):
		return "opencode_command"
	case strings.HasPrefix(rel, ".agents/agents/"):
		return "codex_agent"
	case strings.HasPrefix(rel, ".agents/skills/"):
		return "codex_skill"
	}
	return "file"
}

func addNode(res *domain.ExtractResult, node domain.Node) {
	res.Nodes = append(res.Nodes, node)
	res.Edges = append(res.Edges, domain.Edge{From: fileID(res.Source.Path), Type: "defines", To: node.ID, Reason: "extracted from source file"})
}

func extractGo(root, rel string, data []byte, res *domain.ExtractResult) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath.Join(root, filepath.FromSlash(rel)), data, parser.ParseComments)
	if err != nil {
		res.Source.Status, res.Source.Error = "error", err.Error()
		return
	}
	pkgID := fileID(rel) + "#package:" + file.Name.Name
	addNode(res, domain.Node{ID: pkgID, Type: "go_package", Label: file.Name.Name, Path: rel, Span: &domain.Span{Line: fset.Position(file.Name.Pos()).Line}, Reason: "go package declaration"})
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, "\"")
		id := fileID(rel) + "#import:" + path
		addNode(res, domain.Node{ID: id, Type: "go_import", Label: path, Path: rel, Span: &domain.Span{Line: fset.Position(imp.Pos()).Line}, Reason: "go import declaration"})
		res.Edges = append(res.Edges, domain.Edge{From: pkgID, Type: "imports", To: id, Reason: "package import"})
	}
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			for _, spec := range d.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					addNode(res, domain.Node{ID: fileID(rel) + "#type:" + ts.Name.Name, Type: "go_type", Label: ts.Name.Name, Path: rel, Span: &domain.Span{Line: fset.Position(ts.Pos()).Line}, Reason: "go type declaration"})
				}
			}
		case *ast.FuncDecl:
			typ := "go_function"
			name := d.Name.Name
			if d.Recv != nil && len(d.Recv.List) > 0 {
				typ = "go_method"
				name = receiverName(d.Recv.List[0].Type) + "." + d.Name.Name
			}
			attrs := map[string]string{}
			if strings.HasSuffix(rel, "_test.go") || strings.HasPrefix(d.Name.Name, "Test") {
				attrs["test"] = "true"
			}
			addNode(res, domain.Node{ID: fileID(rel) + "#func:" + name, Type: typ, Label: name, Path: rel, Span: &domain.Span{Line: fset.Position(d.Pos()).Line}, Attrs: attrs, Reason: "go function declaration"})
		}
	}
}

func receiverName(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.StarExpr:
		return receiverName(v.X)
	case *ast.IndexExpr:
		return receiverName(v.X)
	default:
		return "receiver"
	}
}

func extractMarkdown(rel, text string, res *domain.ExtractResult) {
	docID := fileID(rel) + "#markdown"
	addNode(res, domain.Node{ID: docID, Type: "markdown_document", Label: filepath.Base(rel), Path: rel, Reason: "markdown document"})
	for i, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			level := len(trimmed) - len(strings.TrimLeft(trimmed, "#"))
			label := strings.TrimSpace(trimmed[level:])
			if label != "" {
				addNode(res, domain.Node{ID: fmt.Sprintf("%s#heading:%d:%s", fileID(rel), i+1, slug(label)), Type: "markdown_heading", Label: label, Path: rel, Span: &domain.Span{Line: i + 1}, Attrs: map[string]string{"level": fmt.Sprint(level)}, Reason: "markdown heading"})
			}
		}
		for _, ref := range explicitRefs(trimmed) {
			res.Edges = append(res.Edges, domain.Edge{From: docID, Type: "references", To: fileID(ref), Reason: "explicit relative path reference"})
		}
	}
}

func extractYAML(rel string, data []byte, res *domain.ExtractResult) {
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		res.Source.Status, res.Source.Error = "error", err.Error()
		return
	}
	docID := fileID(rel) + "#yaml"
	addNode(res, domain.Node{ID: docID, Type: "yaml_document", Label: filepath.Base(rel), Path: rel, Reason: "yaml document"})
	walkYAML(rel, docID, "", &node, res)
}

func walkYAML(rel, parent, prefix string, node *yaml.Node, res *domain.ExtractResult) {
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		walkYAML(rel, parent, prefix, node.Content[0], res)
		return
	}
	if node.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(node.Content); i += 2 {
			key := node.Content[i].Value
			path := key
			if prefix != "" {
				path = prefix + "." + key
			}
			id := fileID(rel) + "#yaml_key:" + path
			addNode(res, domain.Node{ID: id, Type: "yaml_key", Label: path, Path: rel, Span: &domain.Span{Line: node.Content[i].Line}, Reason: "yaml mapping key"})
			res.Edges = append(res.Edges, domain.Edge{From: parent, Type: "contains", To: id, Reason: "yaml hierarchy"})
			if node.Content[i+1].Kind == yaml.ScalarNode {
				for _, ref := range explicitRefs(node.Content[i+1].Value) {
					res.Edges = append(res.Edges, domain.Edge{From: id, Type: "references", To: fileID(ref), Reason: "yaml scalar path reference"})
				}
			}
			walkYAML(rel, id, path, node.Content[i+1], res)
		}
	}
}

func extractJSON(rel string, data []byte, res *domain.ExtractResult) {
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		res.Source.Status, res.Source.Error = "error", err.Error()
		return
	}
	docID := fileID(rel) + "#json"
	addNode(res, domain.Node{ID: docID, Type: "json_document", Label: filepath.Base(rel), Path: rel, Reason: "json document"})
	walkJSON(rel, docID, "", value, res)
}

func walkJSON(rel, parent, prefix string, value interface{}, res *domain.ExtractResult) {
	m, ok := value.(map[string]interface{})
	if !ok {
		return
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		id := fileID(rel) + "#json_key:" + path
		addNode(res, domain.Node{ID: id, Type: "json_key", Label: path, Path: rel, Reason: "json object key"})
		res.Edges = append(res.Edges, domain.Edge{From: parent, Type: "contains", To: id, Reason: "json hierarchy"})
		if s, ok := m[key].(string); ok {
			for _, ref := range explicitRefs(s) {
				res.Edges = append(res.Edges, domain.Edge{From: id, Type: "references", To: fileID(ref), Reason: "json string path reference"})
			}
		}
		walkJSON(rel, id, path, m[key], res)
	}
}

func explicitRefs(text string) []string {
	seen := map[string]bool{}
	var out []string
	for _, ref := range pathRefRE.FindAllString(text, -1) {
		ref = strings.Trim(ref, "'\"`),;")
		ref = strings.TrimSuffix(ref, ".")
		if strings.HasPrefix(ref, "http") || strings.HasPrefix(ref, "/") || strings.Contains(ref, "..") || seen[ref] {
			continue
		}
		seen[ref] = true
		out = append(out, filepath.ToSlash(ref))
	}
	sort.Strings(out)
	return out
}

func slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func normalize(res *domain.ExtractResult) {
	nodes := map[string]domain.Node{}
	for _, node := range res.Nodes {
		nodes[node.ID] = node
	}
	res.Nodes = res.Nodes[:0]
	for _, node := range nodes {
		res.Nodes = append(res.Nodes, node)
	}
	sort.Slice(res.Nodes, func(i, j int) bool { return res.Nodes[i].ID < res.Nodes[j].ID })
	edges := map[string]domain.Edge{}
	for _, edge := range res.Edges {
		edges[edge.From+"\x00"+edge.Type+"\x00"+edge.To] = edge
	}
	res.Edges = res.Edges[:0]
	for _, edge := range edges {
		res.Edges = append(res.Edges, edge)
	}
	sort.Slice(res.Edges, func(i, j int) bool {
		if res.Edges[i].From != res.Edges[j].From {
			return res.Edges[i].From < res.Edges[j].From
		}
		if res.Edges[i].Type != res.Edges[j].Type {
			return res.Edges[i].Type < res.Edges[j].Type
		}
		return res.Edges[i].To < res.Edges[j].To
	})
}
