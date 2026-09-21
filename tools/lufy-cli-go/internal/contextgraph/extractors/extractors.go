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

const (
	Version                 = "contextgraph-extractors/v2"
	maxWorkflowMarkers      = 32
	maxWorkflowTargetLength = 256
)

var (
	pathRefRE      = regexp.MustCompile(`(?:[A-Za-z0-9_.-]+/)+[A-Za-z0-9_.-]+(?:\.[A-Za-z0-9_.-]+)?`)
	workflowTaskRE = regexp.MustCompile(`^- \[[ xX]\] ([0-9]+(?:\.[0-9]+)*)\s+(.+)$`)
	workflowMarkRE = regexp.MustCompile(`lufy:([a-z_]+)[ \t]+([^ \t\r\n<]+)`)
	githubRefRE    = regexp.MustCompile(`https://github\.com/([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+)/(issues|pull)/([1-9][0-9]*)\b`)
	workflowIDRE   = regexp.MustCompile(`^(?:change:[a-z0-9][a-z0-9._-]*|spec:[a-z0-9][a-z0-9._-]*/[a-z0-9][a-z0-9._-]*|requirement:[a-z0-9][a-z0-9._-]*/[a-z0-9][a-z0-9._-]*#[a-z0-9][a-z0-9-]*|scenario:[a-z0-9][a-z0-9._-]*/[a-z0-9][a-z0-9._-]*#[a-z0-9][a-z0-9-]*/[a-z0-9][a-z0-9-]*|task:[a-z0-9][a-z0-9._-]*#[0-9]+(?:\.[0-9]+)*|decision:[a-z0-9][a-z0-9._-]*#[a-z0-9][a-z0-9-]*|test:[A-Za-z0-9_.\-/]+#[A-Za-z][A-Za-z0-9_]*|issue:github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+#[1-9][0-9]*|pr:github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+#[1-9][0-9]*)$`)
)

var workflowRelations = map[string]bool{
	"implements":  true,
	"verifies":    true,
	"depends_on":  true,
	"caused_by":   true,
	"reviewed_by": true,
	"supersedes":  true,
}

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
	markerCount := 0
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
			funcID := fileID(rel) + "#func:" + name
			addNode(res, domain.Node{ID: funcID, Type: typ, Label: name, Path: rel, Span: &domain.Span{Line: fset.Position(d.Pos()).Line}, Attrs: attrs, Reason: "go function declaration"})
			markerSource := funcID
			if strings.HasSuffix(filepath.ToSlash(rel), "_test.go") && strings.HasPrefix(d.Name.Name, "Test") {
				testID := "test:" + filepath.ToSlash(rel) + "#" + d.Name.Name
				addNode(res, workflowNode(testID, "workflow_test", d.Name.Name, rel, fset.Position(d.Pos()).Line, "recognized Go test function"))
				markerSource = testID
			}
			if d.Doc != nil {
				for _, comment := range d.Doc.List {
					line := fset.Position(comment.Pos()).Line
					extractWorkflowMarkers(rel, line, comment.Text, markerSource, res, &markerCount)
				}
			}
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
	extractWorkflowMarkdown(rel, text, res)
}

type workflowArtifact struct {
	change string
	spec   string
	kind   string
}

func workflowArtifactForPath(rel string) (workflowArtifact, bool) {
	const prefix = ".lufy/workflows/sdd/changes/"
	rel = filepath.ToSlash(rel)
	if !strings.HasPrefix(rel, prefix) {
		return workflowArtifact{}, false
	}
	parts := strings.Split(strings.TrimPrefix(rel, prefix), "/")
	if len(parts) < 2 || slug(parts[0]) != parts[0] {
		return workflowArtifact{}, false
	}
	artifact := workflowArtifact{change: parts[0]}
	switch {
	case len(parts) == 2 && parts[1] == "proposal.md":
		artifact.kind = "proposal"
	case len(parts) == 2 && parts[1] == "tasks.md":
		artifact.kind = "tasks"
	case len(parts) == 2 && parts[1] == "design.md":
		artifact.kind = "design"
	case len(parts) == 4 && parts[1] == "specs" && parts[3] == "spec.md" && slug(parts[2]) == parts[2]:
		artifact.kind, artifact.spec = "spec", parts[2]
	default:
		return workflowArtifact{}, false
	}
	return artifact, true
}

func extractWorkflowMarkdown(rel, text string, res *domain.ExtractResult) {
	artifact, ok := workflowArtifactForPath(rel)
	if !ok {
		return
	}
	lines := strings.Split(text, "\n")
	changeID := "change:" + artifact.change
	defaultSource := ""
	if artifact.kind == "proposal" {
		defaultSource = changeID
		addNode(res, workflowNode(changeID, "workflow_change", artifact.change, rel, 1, "recognized SDD proposal"))
	}
	specID := ""
	if artifact.kind == "spec" {
		specID = "spec:" + artifact.change + "/" + artifact.spec
		defaultSource = specID
		addNode(res, workflowNode(specID, "workflow_spec", artifact.spec, rel, 1, "recognized SDD specification"))
	}

	currentSource := defaultSource
	currentRequirement := ""
	markerCount := 0
	for index, rawLine := range lines {
		lineNumber := index + 1
		line := strings.TrimSpace(strings.TrimSuffix(rawLine, "\r"))
		lineSource := currentSource

		if artifact.kind == "spec" && strings.HasPrefix(line, "### Requirement:") {
			label := strings.TrimSpace(strings.TrimPrefix(line, "### Requirement:"))
			if value := slug(label); value != "" {
				currentRequirement = value
				lineSource = "requirement:" + artifact.change + "/" + artifact.spec + "#" + value
				currentSource = lineSource
				addNode(res, workflowNode(lineSource, "workflow_requirement", label, rel, lineNumber, "recognized SDD requirement heading"))
				res.Edges = append(res.Edges, domain.Edge{From: specID, Type: "contains", To: lineSource, Reason: "specification contains requirement", Attrs: workflowProvenance("recognized_structure")})
			}
		} else if artifact.kind == "spec" && strings.HasPrefix(line, "#### Scenario:") && currentRequirement != "" {
			label := strings.TrimSpace(strings.TrimPrefix(line, "#### Scenario:"))
			if value := slug(label); value != "" {
				lineSource = "scenario:" + artifact.change + "/" + artifact.spec + "#" + currentRequirement + "/" + value
				currentSource = lineSource
				addNode(res, workflowNode(lineSource, "workflow_scenario", label, rel, lineNumber, "recognized SDD scenario heading"))
				parent := "requirement:" + artifact.change + "/" + artifact.spec + "#" + currentRequirement
				res.Edges = append(res.Edges, domain.Edge{From: parent, Type: "contains", To: lineSource, Reason: "requirement contains scenario", Attrs: workflowProvenance("recognized_structure")})
			}
		} else if artifact.kind == "tasks" {
			if match := workflowTaskRE.FindStringSubmatch(line); len(match) == 3 {
				label := cleanWorkflowLabel(match[2])
				lineSource = "task:" + artifact.change + "#" + match[1]
				currentSource = lineSource
				addNode(res, workflowNode(lineSource, "workflow_task", label, rel, lineNumber, "recognized SDD task item"))
			}
		} else if artifact.kind == "design" && strings.HasPrefix(line, "### Decision:") {
			label := strings.TrimSpace(strings.TrimPrefix(line, "### Decision:"))
			if value := slug(label); value != "" {
				lineSource = "decision:" + artifact.change + "#" + value
				currentSource = lineSource
				addNode(res, workflowNode(lineSource, "workflow_decision", label, rel, lineNumber, "recognized SDD decision heading"))
			}
		}

		githubSources := extractGitHubReferences(rel, lineNumber, line, res)
		if workflowMarkRE.MatchString(line) && len(githubSources.pullRequests) > 0 && markerRelation(line) == "reviewed_by" {
			lineSource = githubSources.pullRequests[0]
		}
		extractWorkflowMarkers(rel, lineNumber, line, lineSource, res, &markerCount)
	}
}

type githubLineSources struct {
	pullRequests []string
}

func extractGitHubReferences(rel string, line int, text string, res *domain.ExtractResult) githubLineSources {
	var sources githubLineSources
	for _, match := range githubRefRE.FindAllStringSubmatch(text, -1) {
		repository := strings.ToLower(match[1])
		prefix, typ, label := "issue:", "github_issue", "Issue #"+match[3]
		if match[2] == "pull" {
			prefix, typ, label = "pr:", "github_pull_request", "Pull request #"+match[3]
		}
		id := prefix + "github.com/" + repository + "#" + match[3]
		addNode(res, workflowNode(id, typ, label, rel, line, "recognized GitHub reference"))
		if typ == "github_pull_request" {
			sources.pullRequests = append(sources.pullRequests, id)
		}
	}
	return sources
}

func markerRelation(text string) string {
	match := workflowMarkRE.FindStringSubmatch(text)
	if len(match) != 3 {
		return ""
	}
	return match[1]
}

func extractWorkflowMarkers(rel string, line int, text, sourceID string, res *domain.ExtractResult, markerCount *int) {
	for _, match := range workflowMarkRE.FindAllStringSubmatch(text, -1) {
		if *markerCount >= maxWorkflowMarkers {
			if *markerCount == maxWorkflowMarkers {
				addWorkflowDiagnostic(res, domain.Diagnostic{Code: "workflow_marker_limit_exceeded", Path: filepath.ToSlash(rel), Line: line, Recovery: "reduce explicit workflow markers in this source"})
			}
			(*markerCount)++
			continue
		}
		(*markerCount)++
		relation := match[1]
		target := strings.Trim(match[2], "'\"`),;")
		if !workflowRelations[relation] || sourceID == "" || !validWorkflowTarget(relation, target) {
			addWorkflowDiagnostic(res, domain.Diagnostic{Code: "invalid_workflow_reference", Path: filepath.ToSlash(rel), Line: line, Relation: safeWorkflowRelation(relation), Recovery: "use an allow-listed relation and a canonical workflow node ID"})
			continue
		}
		res.Edges = append(res.Edges, domain.Edge{From: sourceID, Type: relation, To: target, Reason: "explicit workflow trace marker", Attrs: workflowProvenance("explicit_marker")})
	}
}

func validWorkflowTarget(relation, target string) bool {
	if len(target) == 0 || len(target) > maxWorkflowTargetLength || !workflowIDRE.MatchString(target) {
		return false
	}
	switch relation {
	case "implements":
		return strings.HasPrefix(target, "scenario:") || strings.HasPrefix(target, "task:")
	case "verifies":
		return strings.HasPrefix(target, "scenario:") || strings.HasPrefix(target, "task:")
	case "depends_on":
		return strings.HasPrefix(target, "task:") || strings.HasPrefix(target, "change:")
	case "caused_by":
		return strings.HasPrefix(target, "issue:") || strings.HasPrefix(target, "change:")
	case "reviewed_by":
		return strings.HasPrefix(target, "issue:")
	case "supersedes":
		return strings.HasPrefix(target, "change:") || strings.HasPrefix(target, "spec:") || strings.HasPrefix(target, "decision:")
	default:
		return false
	}
}

func safeWorkflowRelation(relation string) string {
	if workflowRelations[relation] {
		return relation
	}
	return "invalid"
}

func workflowNode(id, typ, label, path string, line int, reason string) domain.Node {
	return domain.Node{ID: id, Type: typ, Label: label, Path: filepath.ToSlash(path), Span: &domain.Span{Line: line}, Attrs: workflowProvenance("recognized_structure"), Reason: reason}
}

func workflowProvenance(value string) map[string]string {
	return map[string]string{"provenance": value}
}

func cleanWorkflowLabel(value string) string {
	if index := strings.Index(value, "<!--"); index >= 0 {
		value = value[:index]
	}
	return strings.TrimSpace(value)
}

func addWorkflowDiagnostic(res *domain.ExtractResult, diagnostic domain.Diagnostic) {
	res.Diagnostics = append(res.Diagnostics, diagnostic)
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
		if strings.HasPrefix(ref, "http") || strings.HasPrefix(ref, "/") || strings.HasPrefix(filepath.ToSlash(ref), ".lufy/runtime/") || strings.Contains(ref, "..") || seen[ref] {
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

// ResolveWorkflowReferences resolves only explicit workflow evidence edges
// against a combined extraction result. Extract intentionally remains scoped to
// one source and never scans the workspace to resolve cross-file references.
func ResolveWorkflowReferences(input domain.ExtractResult) domain.ExtractResult {
	resolved := domain.ExtractResult{
		Source:      input.Source,
		Nodes:       append([]domain.Node(nil), input.Nodes...),
		Edges:       append([]domain.Edge(nil), input.Edges...),
		Diagnostics: append([]domain.Diagnostic(nil), input.Diagnostics...),
	}
	nodeByID := make(map[string]domain.Node, len(resolved.Nodes))
	for _, node := range resolved.Nodes {
		nodeByID[node.ID] = node
	}
	edges := resolved.Edges[:0]
	for _, edge := range resolved.Edges {
		if !workflowRelations[edge.Type] {
			edges = append(edges, edge)
			continue
		}
		_, fromExists := nodeByID[edge.From]
		_, toExists := nodeByID[edge.To]
		if fromExists && toExists {
			edges = append(edges, edge)
			continue
		}
		source := nodeByID[edge.From]
		line := 0
		if source.Span != nil {
			line = source.Span.Line
		}
		addWorkflowDiagnostic(&resolved, domain.Diagnostic{
			Code:     "unresolved_workflow_reference",
			Path:     filepath.ToSlash(source.Path),
			Line:     line,
			Relation: edge.Type,
			TargetID: edge.To,
			Recovery: "add the referenced workflow node or correct the explicit marker",
		})
	}
	resolved.Edges = edges
	normalize(&resolved)
	return resolved
}

func normalize(res *domain.ExtractResult) {
	nodes := map[string]domain.Node{}
	for _, node := range res.Nodes {
		if current, exists := nodes[node.ID]; !exists || workflowNodeLess(node, current) {
			nodes[node.ID] = node
		}
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
	diagnostics := map[string]domain.Diagnostic{}
	for _, diagnostic := range res.Diagnostics {
		key := diagnostic.Code + "\x00" + diagnostic.Path + "\x00" + fmt.Sprint(diagnostic.Line) + "\x00" + diagnostic.Relation + "\x00" + diagnostic.TargetID
		diagnostics[key] = diagnostic
	}
	res.Diagnostics = res.Diagnostics[:0]
	for _, diagnostic := range diagnostics {
		res.Diagnostics = append(res.Diagnostics, diagnostic)
	}
	sort.Slice(res.Diagnostics, func(i, j int) bool {
		left := res.Diagnostics[i]
		right := res.Diagnostics[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		if left.Code != right.Code {
			return left.Code < right.Code
		}
		if left.Relation != right.Relation {
			return left.Relation < right.Relation
		}
		return left.TargetID < right.TargetID
	})
}

func workflowNodeLess(left, right domain.Node) bool {
	if left.Path != right.Path {
		return left.Path < right.Path
	}
	leftLine, rightLine := 0, 0
	if left.Span != nil {
		leftLine = left.Span.Line
	}
	if right.Span != nil {
		rightLine = right.Span.Line
	}
	if leftLine != rightLine {
		return leftLine < rightLine
	}
	if left.Type != right.Type {
		return left.Type < right.Type
	}
	return left.Label < right.Label
}
