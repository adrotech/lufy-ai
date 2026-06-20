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

func ContextDir(root string) string   { return filepath.Join(root, ".lufy", "context") }
func GraphPath(root string) string    { return filepath.Join(ContextDir(root), "graph.json") }
func SummaryPath(root string) string  { return filepath.Join(ContextDir(root), "graph-summary.md") }
func ManifestPath(root string) string { return filepath.Join(ContextDir(root), "manifest.json") }

func (Store) LoadGraph(root string) (domain.Graph, error) {
	data, err := os.ReadFile(GraphPath(root))
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

func (Store) LoadManifest(root string) (domain.Manifest, error) {
	data, err := os.ReadFile(ManifestPath(root))
	if err != nil {
		return domain.Manifest{}, err
	}
	var manifest domain.Manifest
	return manifest, json.Unmarshal(data, &manifest)
}

func (Store) Save(root string, graph domain.Graph, summary string) error {
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
	if err := platform.WriteFileAtomic(GraphPath(root), graphData, 0o644); err != nil {
		return err
	}
	if err := platform.WriteFileAtomic(SummaryPath(root), []byte(summary), 0o644); err != nil {
		return err
	}
	return platform.WriteFileAtomic(ManifestPath(root), manifestData, 0o644)
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
