package domain

const SchemaVersion = "lufy-context-graph"

type Graph struct {
	Schema      string                 `json:"schema"`
	GeneratedAt string                 `json:"generated_at"`
	Root        Root                   `json:"root"`
	Sources     []Source               `json:"sources"`
	Nodes       []Node                 `json:"nodes"`
	Edges       []Edge                 `json:"edges"`
	Health      Health                 `json:"health"`
	Communities []Community            `json:"communities,omitempty"`
	Important   []ImportantNode        `json:"important_nodes,omitempty"`
	Questions   []string               `json:"suggested_questions,omitempty"`
	Manifest    Manifest               `json:"manifest"`
	Extensions  map[string]interface{} `json:"extensions"`
}

type Root struct {
	Path       string `json:"path,omitempty"`
	Name       string `json:"name"`
	CLIVersion string `json:"cli_version"`
}

type Source struct {
	Path   string `json:"path"`
	Hash   string `json:"hash"`
	Parser string `json:"parser"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type Health struct {
	TotalFiles       int      `json:"total_files"`
	IndexedFiles     int      `json:"indexed_files"`
	SkippedFiles     int      `json:"skipped_files"`
	ErroredFiles     int      `json:"errored_files"`
	SupportedFormats []string `json:"supported_formats"`
	Warnings         []string `json:"warnings,omitempty"`
}

type Community struct {
	ID    string   `json:"id"`
	Label string   `json:"label"`
	Paths []string `json:"paths"`
	Nodes int      `json:"nodes"`
	Edges int      `json:"edges"`
}

type ImportantNode struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Type   string `json:"type"`
	Path   string `json:"path,omitempty"`
	Degree int    `json:"degree"`
	Reason string `json:"reason"`
}

type Node struct {
	ID     string            `json:"id"`
	Type   string            `json:"type"`
	Label  string            `json:"label"`
	Path   string            `json:"path,omitempty"`
	Span   *Span             `json:"span,omitempty"`
	Attrs  map[string]string `json:"attrs,omitempty"`
	Reason string            `json:"reason,omitempty"`
}

type Span struct {
	Line int `json:"line"`
}

type Edge struct {
	From   string            `json:"from"`
	Type   string            `json:"type"`
	To     string            `json:"to"`
	Reason string            `json:"reason,omitempty"`
	Attrs  map[string]string `json:"attrs,omitempty"`
}

type Manifest struct {
	Schema            string            `json:"schema"`
	ExtractorVersion  string            `json:"extractor_version"`
	Options           map[string]string `json:"options"`
	SourcesHash       string            `json:"sources_hash"`
	GeneratedFromHash string            `json:"generated_from_hash"`
}

type ExtractResult struct {
	Source Source
	Nodes  []Node
	Edges  []Edge
}

type Cache struct {
	Schema           string       `json:"schema"`
	ExtractorVersion string       `json:"extractor_version"`
	Entries          []CacheEntry `json:"entries"`
}

type CacheEntry struct {
	Source Source `json:"source"`
	Nodes  []Node `json:"nodes"`
	Edges  []Edge `json:"edges"`
}
