package opsx

import "time"

type Layer string

const (
	LayerPath     Layer = "PATH"
	LayerCache    Layer = "cache"
	LayerEmbedded Layer = "embedded"
)

type Upstream struct {
	SchemaVersion                int      `json:"schemaVersion"`
	Workflow                     string   `json:"workflow"`
	EffectiveOpenSpecVersion     string   `json:"effectiveOpenSpecVersion"`
	MinimumCompatibleSpecVersion string   `json:"minimumCompatibleOpenSpecVersion"`
	Profile                      string   `json:"profile"`
	Source                       Source   `json:"source"`
	Resolver                     Resolver `json:"resolver"`
	Capabilities                 []string `json:"capabilities"`
}

type Source struct {
	Type        string `json:"type"`
	Repository  string `json:"repository,omitempty"`
	AssetRoot   string `json:"assetRoot,omitempty"`
	Description string `json:"description,omitempty"`
}

type Resolver struct {
	Layers            []string `json:"layers"`
	CacheRoot         string   `json:"cacheRoot"`
	ManifestName      string   `json:"manifestName"`
	OfflineFallback   string   `json:"offlineFallback"`
	PathCommand       string   `json:"pathCommand"`
	MutationByDefault bool     `json:"mutationByDefault"`
}

type CacheManifest struct {
	SchemaVersion int          `json:"schemaVersion"`
	Version       string       `json:"version"`
	Source        Source       `json:"source"`
	CreatedAt     time.Time    `json:"createdAt"`
	Assets        []CacheAsset `json:"assets"`
}

type CacheAsset struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256,omitempty"`
}

type Resolution struct {
	Layer       Layer
	Version     string
	Path        string
	Source      Source
	Diagnostics []Diagnostic
}

type Diagnostic struct {
	Layer   Layer
	Message string
}
