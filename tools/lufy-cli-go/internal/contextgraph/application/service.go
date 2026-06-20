package application

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/adapters"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/extractors"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/version"
)

type Service struct{ store adapters.Store }

func NewService() Service { return Service{store: adapters.NewStore()} }

type Options struct {
	Target string
	JSON   bool
	Term   string
	From   string
	To     string
	Node   string
	Base   string
}
type BuildResult struct {
	Status       string   `json:"status"`
	GraphPath    string   `json:"graph_path,omitempty"`
	SummaryPath  string   `json:"summary_path,omitempty"`
	ManifestPath string   `json:"manifest_path,omitempty"`
	Sources      int      `json:"sources"`
	Nodes        int      `json:"nodes"`
	Edges        int      `json:"edges"`
	Changed      bool     `json:"changed"`
	Errors       []string `json:"errors,omitempty"`
	Recovery     string   `json:"recovery,omitempty"`
}
type StatusResult struct {
	Status    string `json:"status"`
	Reason    string `json:"reason,omitempty"`
	Recovery  string `json:"recovery,omitempty"`
	Sources   int    `json:"sources,omitempty"`
	Nodes     int    `json:"nodes,omitempty"`
	Edges     int    `json:"edges,omitempty"`
	GraphPath string `json:"graph_path,omitempty"`
}
type QueryResult struct {
	Term    string  `json:"term"`
	Matches []Match `json:"matches"`
}
type Match struct {
	Node      domain.Node   `json:"node"`
	Neighbors []domain.Edge `json:"neighbors,omitempty"`
}
type PathResult struct {
	From  string        `json:"from"`
	To    string        `json:"to"`
	Found bool          `json:"found"`
	Nodes []string      `json:"nodes,omitempty"`
	Edges []domain.Edge `json:"edges,omitempty"`
}
type ExplainResult struct {
	ID          string       `json:"id"`
	Node        *domain.Node `json:"node,omitempty"`
	Edge        *domain.Edge `json:"edge,omitempty"`
	Explanation string       `json:"explanation"`
}
type DiffResult struct {
	Base         string   `json:"base"`
	ChangedFiles []string `json:"changed_files"`
	Impact       []Match  `json:"impact"`
	Status       string   `json:"status"`
	Recovery     string   `json:"recovery,omitempty"`
}

func (s Service) Scan(root string) (BuildResult, error) {
	graph, errs, err := s.buildGraph(root)
	if err != nil {
		return BuildResult{}, err
	}
	return buildResult("scan", root, graph, false, errs), nil
}

func (s Service) Build(root string) (BuildResult, error) {
	graph, errs, err := s.buildGraph(root)
	if err != nil {
		return BuildResult{}, err
	}
	changed := true
	if old, err := s.store.LoadManifest(root); err == nil && old.SourcesHash == graph.Manifest.SourcesHash && old.ExtractorVersion == graph.Manifest.ExtractorVersion {
		changed = false
	}
	if changed {
		if err := s.store.Save(root, graph, summary(graph)); err != nil {
			return BuildResult{}, err
		}
	}
	return buildResult("ready", root, graph, changed, errs), nil
}

func (s Service) Status(root string) StatusResult {
	graph, err := s.store.LoadGraph(root)
	if err != nil {
		return StatusResult{Status: "not_available", Reason: err.Error(), Recovery: "lufy-ai context build"}
	}
	current, _, err := s.buildGraph(root)
	if err != nil {
		return StatusResult{Status: "stale", Reason: err.Error(), Recovery: "lufy-ai context build", GraphPath: adapters.GraphPath(root)}
	}
	status := "ready"
	reason := ""
	if current.Manifest.SourcesHash != graph.Manifest.SourcesHash || current.Manifest.ExtractorVersion != graph.Manifest.ExtractorVersion {
		status, reason = "stale", "inputs or extractor version changed"
	}
	return StatusResult{Status: status, Reason: reason, Recovery: recoveryIf(status), Sources: len(graph.Sources), Nodes: len(graph.Nodes), Edges: len(graph.Edges), GraphPath: adapters.GraphPath(root)}
}

func (s Service) Query(root, term string) (QueryResult, error) {
	graph, err := s.store.LoadGraph(root)
	if err != nil {
		return QueryResult{}, err
	}
	termLower := strings.ToLower(term)
	var matches []Match
	for _, node := range graph.Nodes {
		if strings.Contains(strings.ToLower(node.ID+" "+node.Label+" "+node.Type+" "+node.Path), termLower) {
			matches = append(matches, Match{Node: node, Neighbors: neighbors(graph, node.ID, 5)})
			if len(matches) >= 20 {
				break
			}
		}
	}
	return QueryResult{Term: term, Matches: matches}, nil
}

func (s Service) Path(root, from, to string) (PathResult, error) {
	graph, err := s.store.LoadGraph(root)
	if err != nil {
		return PathResult{}, err
	}
	return findPath(graph, from, to), nil
}
func (s Service) Explain(root, id string) (ExplainResult, error) {
	graph, err := s.store.LoadGraph(root)
	if err != nil {
		return ExplainResult{}, err
	}
	for _, n := range graph.Nodes {
		if n.ID == id {
			node := n
			return ExplainResult{ID: id, Node: &node, Explanation: explanationNode(n)}, nil
		}
	}
	parts := strings.Split(id, "->")
	if len(parts) == 2 {
		for _, e := range graph.Edges {
			if e.From == parts[0] && e.To == parts[1] {
				edge := e
				return ExplainResult{ID: id, Edge: &edge, Explanation: e.Reason}, nil
			}
		}
	}
	return ExplainResult{ID: id, Explanation: "node_or_edge_not_found"}, nil
}

func (s Service) Diff(root, base string) (DiffResult, error) {
	graph, err := s.store.LoadGraph(root)
	if err != nil {
		return DiffResult{Base: base, Status: "not_available", Recovery: "lufy-ai context build"}, err
	}
	files, err := adapters.ChangedFiles(root, base)
	if err != nil {
		return DiffResult{}, err
	}
	fileSet := map[string]bool{}
	for _, f := range files {
		fileSet[f] = true
	}
	var impact []Match
	for _, node := range graph.Nodes {
		if fileSet[node.Path] || fileSet[strings.TrimPrefix(node.ID, "file:")] {
			impact = append(impact, Match{Node: node, Neighbors: neighbors(graph, node.ID, 8)})
		}
	}
	return DiffResult{Base: base, ChangedFiles: files, Impact: impact, Status: "ready"}, nil
}

func (s Service) buildGraph(root string) (domain.Graph, []string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return domain.Graph{}, nil, err
	}
	files, err := discover(root)
	if err != nil {
		return domain.Graph{}, nil, err
	}
	var sources []domain.Source
	nodeMap := map[string]domain.Node{}
	edgeMap := map[string]domain.Edge{}
	var errs []string
	for _, rel := range files {
		res := extractors.Extract(root, rel)
		sources = append(sources, res.Source)
		if res.Source.Status != "ok" {
			errs = append(errs, rel+": "+res.Source.Error)
		}
		for _, node := range res.Nodes {
			nodeMap[node.ID] = node
		}
		for _, edge := range res.Edges {
			edgeMap[edge.From+"\x00"+edge.Type+"\x00"+edge.To] = edge
		}
	}
	nodes := make([]domain.Node, 0, len(nodeMap))
	for _, node := range nodeMap {
		nodes = append(nodes, node)
	}
	edges := make([]domain.Edge, 0, len(edgeMap))
	for _, edge := range edgeMap {
		edges = append(edges, edge)
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].Path < sources[j].Path })
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		if edges[i].Type != edges[j].Type {
			return edges[i].Type < edges[j].Type
		}
		return edges[i].To < edges[j].To
	})
	manifest := domain.Manifest{Schema: domain.SchemaVersion, ExtractorVersion: extractors.Version, Options: map[string]string{"formats": "go,markdown,yaml,json"}, SourcesHash: sourcesHash(sources)}
	manifest.GeneratedFromHash = manifest.SourcesHash
	return domain.Graph{Schema: domain.SchemaVersion, GeneratedAt: time.Now().UTC().Format(time.RFC3339), Root: domain.Root{Name: filepath.Base(root), CLIVersion: version.Current().String()}, Sources: sources, Nodes: nodes, Edges: edges, Manifest: manifest, Extensions: map[string]interface{}{}}, errs, nil
}

func discover(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == ".git" || rel == ".lufy/context" || strings.HasPrefix(rel, ".lufy/context/") || rel == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if extractors.Supported(rel) {
			files = append(files, rel)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func buildResult(status, root string, graph domain.Graph, changed bool, errs []string) BuildResult {
	return BuildResult{Status: status, GraphPath: adapters.GraphPath(root), SummaryPath: adapters.SummaryPath(root), ManifestPath: adapters.ManifestPath(root), Sources: len(graph.Sources), Nodes: len(graph.Nodes), Edges: len(graph.Edges), Changed: changed, Errors: errs}
}
func sourcesHash(sources []domain.Source) string {
	b, _ := json.Marshal(sources)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
func recoveryIf(status string) string {
	if status == "ready" {
		return ""
	}
	return "lufy-ai context build"
}

func summary(graph domain.Graph) string {
	counts := map[string]int{}
	for _, n := range graph.Nodes {
		counts[n.Type]++
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	fmt.Fprintf(&b, "# LUFY Context Graph\n\n- schema: `%s`\n- sources: %d\n- nodes: %d\n- edges: %d\n\n## Node counts\n", graph.Schema, len(graph.Sources), len(graph.Nodes), len(graph.Edges))
	for _, k := range keys {
		fmt.Fprintf(&b, "- %s: %d\n", k, counts[k])
	}
	b.WriteString("\n## Suggested commands\n- `lufy-ai context status`\n- `lufy-ai context query <term>`\n- `lufy-ai context diff --base origin/develop`\n")
	return b.String()
}

func neighbors(graph domain.Graph, id string, limit int) []domain.Edge {
	var out []domain.Edge
	for _, e := range graph.Edges {
		if e.From == id || e.To == id {
			out = append(out, e)
			if len(out) >= limit {
				break
			}
		}
	}
	return out
}
func explanationNode(n domain.Node) string {
	if n.Reason != "" {
		return n.Reason
	}
	return "node extracted deterministically from " + n.Path
}

func findPath(graph domain.Graph, from, to string) PathResult {
	adj := map[string][]domain.Edge{}
	for _, e := range graph.Edges {
		adj[e.From] = append(adj[e.From], e)
		adj[e.To] = append(adj[e.To], domain.Edge{From: e.To, Type: e.Type, To: e.From, Reason: e.Reason})
	}
	queue := []string{from}
	prev := map[string]domain.Edge{}
	seen := map[string]bool{from: true}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur == to {
			break
		}
		for _, e := range adj[cur] {
			if !seen[e.To] {
				seen[e.To] = true
				prev[e.To] = e
				queue = append(queue, e.To)
			}
		}
	}
	if !seen[to] {
		return PathResult{From: from, To: to, Found: false}
	}
	var nodes []string
	var edges []domain.Edge
	for cur := to; cur != from; {
		e := prev[cur]
		edges = append([]domain.Edge{e}, edges...)
		nodes = append([]string{cur}, nodes...)
		cur = e.From
	}
	nodes = append([]string{from}, nodes...)
	return PathResult{From: from, To: to, Found: true, Nodes: nodes, Edges: edges}
}
