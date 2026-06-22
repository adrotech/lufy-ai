package application

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/adapters"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/extractors"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
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
	Status       string        `json:"status"`
	GraphPath    string        `json:"graph_path,omitempty"`
	SummaryPath  string        `json:"summary_path,omitempty"`
	ReportPath   string        `json:"report_path,omitempty"`
	ManifestPath string        `json:"manifest_path,omitempty"`
	Sources      int           `json:"sources"`
	Nodes        int           `json:"nodes"`
	Edges        int           `json:"edges"`
	Changed      bool          `json:"changed"`
	CacheHits    int           `json:"cache_hits"`
	CacheMisses  int           `json:"cache_misses"`
	Health       domain.Health `json:"health"`
	Errors       []string      `json:"errors,omitempty"`
	Recovery     string        `json:"recovery,omitempty"`
}

type StatusResult struct {
	Status    string        `json:"status"`
	Reason    string        `json:"reason,omitempty"`
	Recovery  string        `json:"recovery,omitempty"`
	Sources   int           `json:"sources,omitempty"`
	Nodes     int           `json:"nodes,omitempty"`
	Edges     int           `json:"edges,omitempty"`
	GraphPath string        `json:"graph_path,omitempty"`
	Health    domain.Health `json:"health,omitempty"`
}

type QueryResult struct {
	Term          string   `json:"term"`
	ExpandedTerms []string `json:"expanded_terms,omitempty"`
	TokenSavings  string   `json:"token_savings"`
	Matches       []Match  `json:"matches"`
	Questions     []string `json:"suggested_questions,omitempty"`
}

type Match struct {
	Node      domain.Node   `json:"node"`
	Score     int           `json:"score"`
	Reason    string        `json:"reason"`
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
	Evidence    []string     `json:"evidence,omitempty"`
}

type DiffResult struct {
	Base         string             `json:"base"`
	ChangedFiles []string           `json:"changed_files"`
	Impact       []Match            `json:"impact"`
	Communities  []domain.Community `json:"communities,omitempty"`
	TokenSavings string             `json:"token_savings"`
	Status       string             `json:"status"`
	Recovery     string             `json:"recovery,omitempty"`
}

type buildMeta struct {
	CacheHits   int
	CacheMisses int
}

type discoveredFile struct {
	Path      string
	Hash      string
	Supported bool
	Skipped   bool
	Reason    string
}

func (s Service) Scan(root string) (BuildResult, error) {
	graph, errs, meta, cfg, err := s.buildGraph(root)
	if err != nil {
		return BuildResult{}, err
	}
	return buildResult("scan", root, graph, false, errs, meta, cfg), nil
}

func (s Service) Build(root string) (BuildResult, error) {
	graph, errs, meta, cfg, err := s.buildGraph(root)
	if err != nil {
		return BuildResult{}, err
	}
	changed := true
	if old, err := s.store.LoadManifest(root, cfg.Root); err == nil && old.SourcesHash == graph.Manifest.SourcesHash && old.ExtractorVersion == graph.Manifest.ExtractorVersion {
		changed = false
	}
	if changed {
		if err := s.store.Save(root, graph, summary(graph), report(graph), cfg.Root, cfg.Report); err != nil {
			return BuildResult{}, err
		}
		if cfg.Cache {
			_ = s.store.SaveCache(root, cfg.Root, cacheFromGraph(graph))
		}
	}
	return buildResult("ready", root, graph, changed, errs, meta, cfg), nil
}

func (s Service) Status(root string) StatusResult {
	cfg := s.config(root)
	graph, err := s.store.LoadGraph(root, cfg.Root)
	if err != nil {
		return StatusResult{Status: "not_available", Reason: err.Error(), Recovery: "lufy-ai context build"}
	}
	currentHash, err := s.currentSourcesHash(root, cfg)
	if err != nil {
		return StatusResult{Status: "stale", Reason: err.Error(), Recovery: "lufy-ai context build", GraphPath: adapters.GraphPathFor(root, cfg.Root)}
	}
	status := "ready"
	reason := ""
	if currentHash != graph.Manifest.SourcesHash || graph.Manifest.ExtractorVersion != extractors.Version {
		status, reason = "stale", "inputs or extractor version changed"
	}
	return StatusResult{Status: status, Reason: reason, Recovery: recoveryIf(status), Sources: len(graph.Sources), Nodes: len(graph.Nodes), Edges: len(graph.Edges), GraphPath: adapters.GraphPathFor(root, cfg.Root), Health: graph.Health}
}

func (s Service) Query(root, term string) (QueryResult, error) {
	cfg := s.config(root)
	graph, err := s.store.LoadGraph(root, cfg.Root)
	if err != nil {
		return QueryResult{}, err
	}
	terms := expandTerms(term, graph)
	matches := rankedMatches(graph, terms, cfg.MaxQueryResults, cfg.MaxNeighborsPerHint)
	return QueryResult{Term: term, ExpandedTerms: terms, TokenSavings: savingsSummary(len(matches), len(graph.Nodes)), Matches: matches, Questions: graph.Questions}, nil
}

func (s Service) Path(root, from, to string) (PathResult, error) {
	cfg := s.config(root)
	graph, err := s.store.LoadGraph(root, cfg.Root)
	if err != nil {
		return PathResult{}, err
	}
	return findPath(graph, from, to), nil
}

func (s Service) Explain(root, id string) (ExplainResult, error) {
	cfg := s.config(root)
	graph, err := s.store.LoadGraph(root, cfg.Root)
	if err != nil {
		return ExplainResult{}, err
	}
	for _, n := range graph.Nodes {
		if n.ID == id {
			node := n
			return ExplainResult{ID: id, Node: &node, Explanation: explanationNode(n), Evidence: evidenceForNode(n)}, nil
		}
	}
	parts := strings.Split(id, "->")
	if len(parts) == 2 {
		for _, e := range graph.Edges {
			if e.From == parts[0] && e.To == parts[1] {
				edge := e
				return ExplainResult{ID: id, Edge: &edge, Explanation: e.Reason, Evidence: []string{e.From, e.Type, e.To}}, nil
			}
		}
	}
	return ExplainResult{ID: id, Explanation: "node_or_edge_not_found"}, nil
}

func (s Service) Diff(root, base string) (DiffResult, error) {
	cfg := s.config(root)
	graph, err := s.store.LoadGraph(root, cfg.Root)
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
			impact = append(impact, Match{Node: node, Score: nodeDegree(graph, node.ID) + 10, Reason: "changed file or direct symbol in diff", Neighbors: neighbors(graph, node.ID, cfg.MaxNeighborsPerHint+3)})
		}
	}
	sortMatches(impact)
	communities := impactedCommunities(graph, fileSet)
	return DiffResult{Base: base, ChangedFiles: files, Impact: limitMatches(impact, cfg.MaxQueryResults), Communities: communities, TokenSavings: savingsSummary(len(impact), len(graph.Nodes)), Status: "ready"}, nil
}

func (s Service) buildGraph(root string) (domain.Graph, []string, buildMeta, projectconfig.ContextGraphConfig, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return domain.Graph{}, nil, buildMeta{}, projectconfig.ContextGraphConfig{}, err
	}
	cfg := s.config(root)
	files, err := discover(root, cfg)
	if err != nil {
		return domain.Graph{}, nil, buildMeta{}, cfg, err
	}
	cache := map[string]domain.CacheEntry{}
	if cfg.Cache {
		if cached, err := s.store.LoadCache(root, cfg.Root); err == nil && cached.Schema == domain.SchemaVersion && cached.ExtractorVersion == extractors.Version {
			for _, entry := range cached.Entries {
				cache[entry.Source.Path] = entry
			}
		}
	}
	var sources []domain.Source
	nodeMap := map[string]domain.Node{}
	edgeMap := map[string]domain.Edge{}
	var errs []string
	meta := buildMeta{}
	for _, file := range files {
		if file.Skipped || !file.Supported {
			sources = append(sources, domain.Source{Path: file.Path, Hash: file.Hash, Parser: extractors.ParserName(file.Path), Status: "skipped", Reason: file.Reason})
			continue
		}
		res := domain.ExtractResult{}
		if entry, ok := cache[file.Path]; ok && entry.Source.Hash == file.Hash {
			res = domain.ExtractResult{Source: entry.Source, Nodes: entry.Nodes, Edges: entry.Edges}
			meta.CacheHits++
		} else {
			res = extractors.Extract(root, file.Path)
			meta.CacheMisses++
		}
		sources = append(sources, res.Source)
		if res.Source.Status != "ok" {
			errs = append(errs, file.Path+": "+res.Source.Error)
		}
		for _, node := range res.Nodes {
			nodeMap[node.ID] = node
		}
		for _, edge := range res.Edges {
			edgeMap[edge.From+"\x00"+edge.Type+"\x00"+edge.To] = edge
		}
	}
	nodes, edges := sortedGraphParts(nodeMap, edgeMap)
	sort.Slice(sources, func(i, j int) bool { return sources[i].Path < sources[j].Path })
	health := buildHealth(files, sources)
	communities := buildCommunities(nodes, edges)
	important := importantNodes(nodes, edges, 12)
	questions := suggestedQuestions(communities, important)
	manifest := domain.Manifest{Schema: domain.SchemaVersion, ExtractorVersion: extractors.Version, Options: map[string]string{"formats": "go,markdown,yaml,json", "config_source": projectconfig.ProjectConfigPath, "context_root": cfg.Root, "cache": fmt.Sprint(cfg.Cache), "exclude": strings.Join(cfg.Exclude, ",")}, SourcesHash: sourcesHash(sources)}
	manifest.GeneratedFromHash = manifest.SourcesHash
	graph := domain.Graph{Schema: domain.SchemaVersion, GeneratedAt: time.Now().UTC().Format(time.RFC3339), Root: domain.Root{Name: filepath.Base(root), CLIVersion: version.Current().String()}, Sources: sources, Nodes: nodes, Edges: edges, Health: health, Communities: communities, Important: important, Questions: questions, Manifest: manifest, Extensions: map[string]interface{}{"config_source": projectconfig.ProjectConfigPath}}
	return graph, errs, meta, cfg, nil
}

func (s Service) currentSourcesHash(root string, cfg projectconfig.ContextGraphConfig) (string, error) {
	files, err := discover(root, cfg)
	if err != nil {
		return "", err
	}
	sources := make([]domain.Source, 0, len(files))
	for _, file := range files {
		status := "ok"
		if file.Skipped || !file.Supported {
			status = "skipped"
		}
		sources = append(sources, domain.Source{Path: file.Path, Hash: file.Hash, Parser: extractors.ParserName(file.Path), Status: status, Reason: file.Reason})
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].Path < sources[j].Path })
	return sourcesHash(sources), nil
}

func (s Service) config(root string) projectconfig.ContextGraphConfig {
	cfg := projectconfig.DefaultContextGraphConfig()
	path, err := projectconfig.ExistingPath(root)
	if err != nil {
		return cfg
	}
	loaded, err := projectconfig.Load(path)
	if err != nil {
		return cfg
	}
	if loaded.ContextGraph.Root != "" {
		cfg = loaded.ContextGraph
	}
	if cfg.Root == "" {
		cfg.Root = projectconfig.DefaultContextGraphConfig().Root
	}
	if cfg.Report == "" {
		cfg.Report = projectconfig.DefaultContextGraphConfig().Report
	}
	if cfg.MaxQueryResults == 0 {
		cfg.MaxQueryResults = projectconfig.DefaultContextGraphConfig().MaxQueryResults
	}
	if cfg.MaxNeighborsPerHint == 0 {
		cfg.MaxNeighborsPerHint = projectconfig.DefaultContextGraphConfig().MaxNeighborsPerHint
	}
	if len(cfg.SensitivePatterns) == 0 {
		cfg.SensitivePatterns = projectconfig.DefaultContextGraphConfig().SensitivePatterns
	}
	if len(cfg.Exclude) == 0 {
		cfg.Exclude = projectconfig.DefaultContextGraphConfig().Exclude
	}
	return cfg
}

func discover(root string, cfg projectconfig.ContextGraphConfig) ([]discoveredFile, error) {
	var files []discoveredFile
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
			if skipDir(rel, cfg.Root, cfg) {
				return filepath.SkipDir
			}
			return nil
		}
		if excludedPath(rel, cfg) {
			return nil
		}
		if !extractors.Supported(rel) && !looksSensitive(rel, cfg) {
			return nil
		}
		hash := fileHash(path)
		file := discoveredFile{Path: rel, Hash: hash, Supported: extractors.Supported(rel)}
		if cfg.SkipSensitive && looksSensitive(rel, cfg) {
			file.Skipped, file.Reason = true, "sensitive file pattern"
		}
		files = append(files, file)
		return nil
	})
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, err
}

func skipDir(rel, contextRoot string, cfg projectconfig.ContextGraphConfig) bool {
	return rel == ".git" || rel == "node_modules" || rel == filepath.ToSlash(contextRoot) || strings.HasPrefix(rel, filepath.ToSlash(contextRoot)+"/") || excludedPath(rel, cfg)
}

func excludedPath(rel string, cfg projectconfig.ContextGraphConfig) bool {
	rel = filepath.ToSlash(rel)
	for _, pattern := range cfg.Exclude {
		pattern = filepath.ToSlash(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}
		if strings.HasSuffix(pattern, "/**") {
			prefix := strings.TrimSuffix(pattern, "/**")
			if rel == prefix || strings.HasPrefix(rel, prefix+"/") {
				return true
			}
			continue
		}
		if ok, _ := filepath.Match(pattern, rel); ok {
			return true
		}
		if rel == pattern {
			return true
		}
	}
	return false
}

func looksSensitive(rel string, cfg projectconfig.ContextGraphConfig) bool {
	base := strings.ToLower(filepath.Base(rel))
	path := strings.ToLower(rel)
	for _, pattern := range cfg.SensitivePatterns {
		p := strings.ToLower(pattern)
		if ok, _ := filepath.Match(p, base); ok {
			return true
		}
		if strings.Contains(path, strings.Trim(p, "*")) && strings.Contains(p, "*") {
			return true
		}
	}
	return false
}

func fileHash(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func buildResult(status, root string, graph domain.Graph, changed bool, errs []string, meta buildMeta, cfg projectconfig.ContextGraphConfig) BuildResult {
	return BuildResult{Status: status, GraphPath: adapters.GraphPathFor(root, cfg.Root), SummaryPath: adapters.SummaryPathFor(root, cfg.Root), ReportPath: adapters.ReportPathFor(root, cfg.Root, cfg.Report), ManifestPath: adapters.ManifestPathFor(root, cfg.Root), Sources: len(graph.Sources), Nodes: len(graph.Nodes), Edges: len(graph.Edges), Changed: changed, CacheHits: meta.CacheHits, CacheMisses: meta.CacheMisses, Health: graph.Health, Errors: errs}
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

func sortedGraphParts(nodeMap map[string]domain.Node, edgeMap map[string]domain.Edge) ([]domain.Node, []domain.Edge) {
	nodes := make([]domain.Node, 0, len(nodeMap))
	for _, node := range nodeMap {
		nodes = append(nodes, node)
	}
	edges := make([]domain.Edge, 0, len(edgeMap))
	for _, edge := range edgeMap {
		edges = append(edges, edge)
	}
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
	return nodes, edges
}

func buildHealth(files []discoveredFile, sources []domain.Source) domain.Health {
	health := domain.Health{TotalFiles: len(files), SupportedFormats: []string{".go", ".md", ".yaml", ".yml", ".json"}}
	for _, source := range sources {
		switch source.Status {
		case "ok":
			health.IndexedFiles++
		case "skipped":
			health.SkippedFiles++
		case "error":
			health.ErroredFiles++
		}
	}
	if health.SkippedFiles > 0 {
		health.Warnings = append(health.Warnings, fmt.Sprintf("%d archivos omitidos por patrones sensibles", health.SkippedFiles))
	}
	if health.ErroredFiles > 0 {
		health.Warnings = append(health.Warnings, fmt.Sprintf("%d archivos no pudieron parsearse", health.ErroredFiles))
	}
	return health
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
	fmt.Fprintf(&b, "# LUFY Context Graph\n\n- schema: `%s`\n- sources: %d\n- nodes: %d\n- edges: %d\n- indexed_files: %d\n- skipped_files: %d\n\n## Node counts\n", graph.Schema, len(graph.Sources), len(graph.Nodes), len(graph.Edges), graph.Health.IndexedFiles, graph.Health.SkippedFiles)
	for _, k := range keys {
		fmt.Fprintf(&b, "- %s: %d\n", k, counts[k])
	}
	b.WriteString("\n## Suggested commands\n- `lufy-ai context status`\n- `lufy-ai context query <term>`\n- `lufy-ai context diff --base origin/develop`\n")
	return b.String()
}

func report(graph domain.Graph) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# LUFY Context Report\n\nEste reporte es derivado y regenerable desde `.lufy/config/project.yaml` y el workspace actual. Su objetivo es ahorrar tokens orientando lecturas iniciales.\n\n")
	fmt.Fprintf(&b, "## Health\n\n- indexed files: %d\n- skipped files: %d\n- parse errors: %d\n\n", graph.Health.IndexedFiles, graph.Health.SkippedFiles, graph.Health.ErroredFiles)
	b.WriteString("## Important Nodes\n\n")
	for _, n := range graph.Important {
		fmt.Fprintf(&b, "- `%s` [%s] degree=%d: %s\n", n.ID, n.Type, n.Degree, n.Reason)
	}
	b.WriteString("\n## Communities\n\n")
	for _, c := range graph.Communities {
		fmt.Fprintf(&b, "- `%s` %s: %d nodes, %d edges, paths=%s\n", c.ID, c.Label, c.Nodes, c.Edges, strings.Join(c.Paths, ", "))
	}
	b.WriteString("\n## Suggested Questions\n\n")
	for _, q := range graph.Questions {
		fmt.Fprintf(&b, "- %s\n", q)
	}
	b.WriteString("\n## Audit Trail\n\n- Todos los nodos y edges actuales son `EXTRACTED` por parsers determinísticos o referencias explícitas.\n- No hay llamadas LLM, embeddings ni servicios remotos en el build default.\n- Archivos sensibles se omiten cuando `context_graph.skip_sensitive` está activo.\n")
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

func evidenceForNode(n domain.Node) []string {
	var evidence []string
	if n.Path != "" {
		evidence = append(evidence, n.Path)
	}
	if n.Span != nil {
		evidence = append(evidence, fmt.Sprintf("line:%d", n.Span.Line))
	}
	if n.Reason != "" {
		evidence = append(evidence, "EXTRACTED: "+n.Reason)
	}
	return evidence
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

func expandTerms(term string, graph domain.Graph) []string {
	seen := map[string]bool{}
	add := func(value string, out *[]string) {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" && !seen[value] {
			seen[value] = true
			*out = append(*out, value)
		}
	}
	var terms []string
	for _, part := range strings.FieldsFunc(term, func(r rune) bool { return r == ' ' || r == '/' || r == '-' || r == '_' || r == '.' }) {
		add(part, &terms)
	}
	needle := strings.ToLower(term)
	for _, n := range graph.Nodes {
		if len(terms) >= 8 {
			break
		}
		text := strings.ToLower(n.ID + " " + n.Label + " " + n.Type + " " + n.Path)
		if strings.Contains(text, needle) || anyTerm(text, terms) {
			add(n.Label, &terms)
			add(n.Type, &terms)
		}
	}
	return terms
}

func rankedMatches(graph domain.Graph, terms []string, limit, neighborLimit int) []Match {
	var matches []Match
	for _, node := range graph.Nodes {
		text := strings.ToLower(node.ID + " " + node.Label + " " + node.Type + " " + node.Path)
		score := 0
		for _, term := range terms {
			if strings.Contains(text, term) {
				score += 10
			}
		}
		if score == 0 {
			continue
		}
		score += nodeDegree(graph, node.ID)
		matches = append(matches, Match{Node: node, Score: score, Reason: "lexical match plus graph degree", Neighbors: neighbors(graph, node.ID, neighborLimit)})
	}
	sortMatches(matches)
	return limitMatches(matches, limit)
}

func anyTerm(text string, terms []string) bool {
	for _, term := range terms {
		if strings.Contains(text, term) {
			return true
		}
	}
	return false
}

func sortMatches(matches []Match) {
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}
		return matches[i].Node.ID < matches[j].Node.ID
	})
}

func limitMatches(matches []Match, limit int) []Match {
	if limit <= 0 || len(matches) <= limit {
		return matches
	}
	return matches[:limit]
}

func nodeDegree(graph domain.Graph, id string) int {
	degree := 0
	for _, edge := range graph.Edges {
		if edge.From == id || edge.To == id {
			degree++
		}
	}
	return degree
}

func importantNodes(nodes []domain.Node, edges []domain.Edge, limit int) []domain.ImportantNode {
	degree := map[string]int{}
	for _, edge := range edges {
		degree[edge.From]++
		degree[edge.To]++
	}
	var out []domain.ImportantNode
	for _, node := range nodes {
		if degree[node.ID] == 0 {
			continue
		}
		out = append(out, domain.ImportantNode{ID: node.ID, Label: node.Label, Type: node.Type, Path: node.Path, Degree: degree[node.ID], Reason: "high structural connectivity"})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Degree != out[j].Degree {
			return out[i].Degree > out[j].Degree
		}
		return out[i].ID < out[j].ID
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func buildCommunities(nodes []domain.Node, edges []domain.Edge) []domain.Community {
	byArea := map[string]*domain.Community{}
	for _, node := range nodes {
		area := firstPathPart(node.Path)
		if area == "" {
			area = "root"
		}
		community := byArea[area]
		if community == nil {
			community = &domain.Community{ID: "area:" + area, Label: area, Paths: []string{area}}
			byArea[area] = community
		}
		community.Nodes++
	}
	for _, edge := range edges {
		area := firstPathPart(pathFromID(edge.From))
		if area == "" {
			area = "root"
		}
		if community := byArea[area]; community != nil {
			community.Edges++
		}
	}
	var out []domain.Community
	for _, community := range byArea {
		out = append(out, *community)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Nodes != out[j].Nodes {
			return out[i].Nodes > out[j].Nodes
		}
		return out[i].ID < out[j].ID
	})
	if len(out) > 12 {
		out = out[:12]
	}
	return out
}

func impactedCommunities(graph domain.Graph, files map[string]bool) []domain.Community {
	areas := map[string]bool{}
	for file := range files {
		areas[firstPathPart(file)] = true
	}
	var out []domain.Community
	for _, community := range graph.Communities {
		if areas[strings.TrimPrefix(community.ID, "area:")] {
			out = append(out, community)
		}
	}
	return out
}

func firstPathPart(path string) string {
	path = strings.Trim(filepath.ToSlash(path), "/")
	if path == "" || path == "." {
		return ""
	}
	return strings.Split(path, "/")[0]
}

func pathFromID(id string) string {
	id = strings.TrimPrefix(id, "file:")
	if idx := strings.Index(id, "#"); idx >= 0 {
		return id[:idx]
	}
	return id
}

func suggestedQuestions(communities []domain.Community, important []domain.ImportantNode) []string {
	var questions []string
	if len(important) > 0 {
		questions = append(questions, fmt.Sprintf("Que depende de %s y que deberia revisar antes de cambiarlo?", important[0].Label))
	}
	if len(communities) > 1 {
		questions = append(questions, fmt.Sprintf("Que conexiones existen entre %s y %s?", communities[0].Label, communities[1].Label))
	}
	questions = append(questions, "Que archivos concentran mas relaciones y pueden ahorrar lectura inicial?")
	return questions
}

func savingsSummary(matches, total int) string {
	if total == 0 {
		return "no graph nodes available"
	}
	return fmt.Sprintf("bounded hints: %d of %d nodes surfaced before broad file reads", matches, total)
}

func cacheFromGraph(graph domain.Graph) domain.Cache {
	entryMap := map[string]*domain.CacheEntry{}
	for _, source := range graph.Sources {
		if source.Status != "ok" {
			continue
		}
		entryMap[source.Path] = &domain.CacheEntry{Source: source}
	}
	for _, node := range graph.Nodes {
		if entry := entryMap[node.Path]; entry != nil {
			entry.Nodes = append(entry.Nodes, node)
		}
	}
	for _, edge := range graph.Edges {
		path := pathFromID(edge.From)
		if entry := entryMap[path]; entry != nil {
			entry.Edges = append(entry.Edges, edge)
			continue
		}
		path = pathFromID(edge.To)
		if entry := entryMap[path]; entry != nil {
			entry.Edges = append(entry.Edges, edge)
		}
	}
	var entries []domain.CacheEntry
	for _, entry := range entryMap {
		entries = append(entries, *entry)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Source.Path < entries[j].Source.Path })
	return domain.Cache{Schema: domain.SchemaVersion, ExtractorVersion: extractors.Version, Entries: entries}
}
