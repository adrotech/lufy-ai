package adapters

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/contextgraph/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

var ErrGraphNotAvailable = errors.New("context graph not_available; ejecuta lufy-ai context build")

type Store struct{}

func NewStore() Store { return Store{} }

func ContextDir(root string) string  { return ContextDirFor(root, ".lufy/context") }
func GraphPath(root string) string   { return GraphPathFor(root, ".lufy/context") }
func SummaryPath(root string) string { return SummaryPathFor(root, ".lufy/context") }
func ReportPath(root string) string {
	return ReportPathFor(root, ".lufy/context", ".lufy/context/GRAPH_REPORT.md")
}
func ManifestPath(root string) string { return ManifestPathFor(root, ".lufy/context") }
func CachePath(root string) string    { return CachePathFor(root, ".lufy/context") }

func ContextDirFor(root, contextRoot string) string {
	return filepath.Join(root, filepath.FromSlash(contextRoot))
}
func GraphPathFor(root, contextRoot string) string {
	return filepath.Join(ContextDirFor(root, contextRoot), "graph.json")
}
func SummaryPathFor(root, contextRoot string) string {
	return filepath.Join(ContextDirFor(root, contextRoot), "graph-summary.md")
}
func ReportPathFor(root, contextRoot, reportPath string) string {
	if reportPath == "" {
		return filepath.Join(ContextDirFor(root, contextRoot), "GRAPH_REPORT.md")
	}
	return filepath.Join(root, filepath.FromSlash(reportPath))
}
func ManifestPathFor(root, contextRoot string) string {
	return filepath.Join(ContextDirFor(root, contextRoot), "manifest.json")
}
func CachePathFor(root, contextRoot string) string {
	return filepath.Join(ContextDirFor(root, contextRoot), "cache", "extract.json")
}

func (Store) LoadGraph(root string, contextRoot ...string) (domain.Graph, error) {
	data, err := os.ReadFile(GraphPathFor(root, firstContextRoot(contextRoot)))
	if err != nil {
		return domain.Graph{}, ErrGraphNotAvailable
	}
	var graph domain.Graph
	if err := json.Unmarshal(data, &graph); err != nil {
		return domain.Graph{}, ErrGraphNotAvailable
	}
	if graph.Schema != domain.SchemaVersion {
		return domain.Graph{}, ErrGraphNotAvailable
	}
	return graph, nil
}

func (Store) LoadManifest(root string, contextRoot ...string) (domain.Manifest, error) {
	data, err := os.ReadFile(ManifestPathFor(root, firstContextRoot(contextRoot)))
	if err != nil {
		return domain.Manifest{}, err
	}
	var manifest domain.Manifest
	return manifest, json.Unmarshal(data, &manifest)
}

func (Store) LoadCache(root string, contextRoot string) (domain.Cache, error) {
	data, err := os.ReadFile(CachePathFor(root, contextRoot))
	if err != nil {
		return domain.Cache{}, err
	}
	var cache domain.Cache
	return cache, json.Unmarshal(data, &cache)
}

func (Store) Save(root string, graph domain.Graph, summary, report string, contextRoot string, reportPath ...string) error {
	rootDir := contextRoot
	if rootDir == "" {
		rootDir = ".lufy/context"
	}
	reportRel := ""
	if len(reportPath) > 0 {
		reportRel = reportPath[0]
	}
	graphData, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return err
	}
	graphData = append(graphData, '\n')
	manifestData, err := json.MarshalIndent(graph.Manifest, "", "  ")
	if err != nil {
		return err
	}
	manifestData = append(manifestData, '\n')
	if err := platform.WriteFileAtomic(GraphPathFor(root, rootDir), graphData, 0o644); err != nil {
		return err
	}
	if err := platform.WriteFileAtomic(SummaryPathFor(root, rootDir), []byte(summary), 0o644); err != nil {
		return err
	}
	if err := platform.WriteFileAtomic(ReportPathFor(root, rootDir, reportRel), []byte(report), 0o644); err != nil {
		return err
	}
	return platform.WriteFileAtomic(ManifestPathFor(root, rootDir), manifestData, 0o644)
}

func (Store) SaveCache(root, contextRoot string, cache domain.Cache) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return platform.WriteFileAtomic(CachePathFor(root, contextRoot), append(data, '\n'), 0o644)
}

func firstContextRoot(values []string) string {
	if len(values) > 0 && values[0] != "" {
		return values[0]
	}
	return ".lufy/context"
}

func ChangedFiles(root, base string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", base, "--")
	cmd.Dir = root
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git diff --name-only %s -- fallo: %s", base, strings.TrimSpace(stderr.String()))
	}
	var files []string
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(filepath.ToSlash(line))
		if line != "" {
			files = append(files, line)
		}
	}
	sort.Strings(files)
	return files, nil
}

const maxReviewStatFiles = 128

type ReviewFileStat struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Binary    bool   `json:"binary"`
}

type ReviewStats struct {
	TotalFiles int              `json:"total_files"`
	Additions  int              `json:"additions"`
	Deletions  int              `json:"deletions"`
	Churn      int              `json:"churn"`
	Files      []ReviewFileStat `json:"files"`
	Truncated  bool             `json:"truncated"`
}

func CollectReviewStats(root, base string) (ReviewStats, error) {
	cmd := exec.Command("git", "diff", "--numstat", "--no-renames", "-z", "--end-of-options", base, "--")
	cmd.Dir = root
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return ReviewStats{}, fmt.Errorf("git diff --numstat base %q failed", sanitizedGitBase(base))
	}
	all := make([]ReviewFileStat, 0)
	for _, record := range bytes.Split(stdout.Bytes(), []byte{0}) {
		if len(record) == 0 {
			continue
		}
		parts := bytes.SplitN(record, []byte{'\t'}, 3)
		if len(parts) != 3 {
			continue
		}
		stat := ReviewFileStat{Path: normalizeGitPath(string(parts[2]))}
		if string(parts[0]) == "-" || string(parts[1]) == "-" {
			stat.Binary = true
		} else {
			if _, err := fmt.Sscan(string(parts[0]), &stat.Additions); err != nil {
				continue
			}
			if _, err := fmt.Sscan(string(parts[1]), &stat.Deletions); err != nil {
				continue
			}
		}
		all = append(all, stat)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Path < all[j].Path })
	result := ReviewStats{TotalFiles: len(all)}
	for _, stat := range all {
		result.Additions += stat.Additions
		result.Deletions += stat.Deletions
	}
	result.Churn = result.Additions + result.Deletions
	limit := len(all)
	if limit > maxReviewStatFiles {
		limit = maxReviewStatFiles
		result.Truncated = true
	}
	result.Files = append([]ReviewFileStat(nil), all[:limit]...)
	return result, nil
}

func normalizeGitPath(path string) string {
	path = filepath.ToSlash(filepath.Clean(path))
	return strings.TrimPrefix(path, "./")
}

func sanitizedGitBase(base string) string {
	if len(base) > 128 {
		base = base[:128]
	}
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return '?'
		}
		return r
	}, base)
}
