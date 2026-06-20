package domain

const SchemaVersion = "lufy-context-graph/v1"

type Graph struct {
	Schema      string                 `json:"schema"`
	GeneratedAt string                 `json:"generated_at"`
	Root        Root                   `json:"root"`
	Sources     []Source               `json:"sources"`
	Nodes       []Node                 `json:"nodes"`
	Edges       []Edge                 `json:"edges"`
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
