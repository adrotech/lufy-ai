package skillregistry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
	"gopkg.in/yaml.v3"
)

const SchemaVersion = 1

type Index struct {
	SchemaVersion int           `json:"schemaVersion"`
	Tool          domain.ToolID `json:"tool"`
	Roots         []Root        `json:"roots"`
	Skills        []Skill       `json:"skills"`
	Warnings      []Warning     `json:"warnings,omitempty"`
}

type Root struct {
	Path     string `json:"path"`
	Scope    string `json:"scope"`
	Priority int    `json:"priority"`
	Exists   bool   `json:"exists"`
}

type Skill struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Scope         string   `json:"scope"`
	Path          string   `json:"path"`
	ShadowedPaths []string `json:"shadowedPaths,omitempty"`
}

type Warning struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type candidate struct {
	Skill
	priority int
}

func Build(tool domain.ToolID, declared []ports.SkillRoot) (Index, error) {
	index := Index{SchemaVersion: SchemaVersion, Tool: tool, Roots: []Root{}, Skills: []Skill{}}
	seenRoots := map[string]bool{}
	candidates := []candidate{}
	for _, declaredRoot := range declared {
		rootPath, err := filepath.Abs(declaredRoot.Path)
		if err != nil {
			return Index{}, fmt.Errorf("resolver raíz de skills %q: %w", declaredRoot.Path, err)
		}
		rootPath = filepath.Clean(rootPath)
		key := declaredRoot.Scope + "\x00" + rootPath
		if seenRoots[key] {
			continue
		}
		seenRoots[key] = true
		root := Root{Path: rootPath, Scope: declaredRoot.Scope, Priority: declaredRoot.Priority}
		info, err := os.Lstat(rootPath)
		if os.IsNotExist(err) {
			index.Roots = append(index.Roots, root)
			continue
		}
		if err != nil {
			index.Roots = append(index.Roots, root)
			index.Warnings = append(index.Warnings, Warning{Path: rootPath, Message: err.Error()})
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			index.Roots = append(index.Roots, root)
			index.Warnings = append(index.Warnings, Warning{Path: rootPath, Message: "la raíz debe ser un directorio real, no un symlink"})
			continue
		}
		root.Exists = true
		index.Roots = append(index.Roots, root)
		walkRoot(root, &candidates, &index.Warnings)
	}
	sort.Slice(index.Roots, func(i, j int) bool {
		if effectivePriority(index.Roots[i].Scope, index.Roots[i].Priority) != effectivePriority(index.Roots[j].Scope, index.Roots[j].Priority) {
			return effectivePriority(index.Roots[i].Scope, index.Roots[i].Priority) < effectivePriority(index.Roots[j].Scope, index.Roots[j].Priority)
		}
		return index.Roots[i].Path < index.Roots[j].Path
	})
	index.Skills = selectSkills(candidates)
	sort.Slice(index.Warnings, func(i, j int) bool {
		if index.Warnings[i].Path != index.Warnings[j].Path {
			return index.Warnings[i].Path < index.Warnings[j].Path
		}
		return index.Warnings[i].Message < index.Warnings[j].Message
	})
	return index, nil
}

func Marshal(index Index) ([]byte, error) {
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func walkRoot(root Root, candidates *[]candidate, warnings *[]Warning) {
	err := filepath.WalkDir(root.Path, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			*warnings = append(*warnings, Warning{Path: path, Message: walkErr.Error()})
			if entry != nil && entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() || entry.Name() != "SKILL.md" {
			return nil
		}
		metadata, err := readMetadata(path)
		if err != nil {
			*warnings = append(*warnings, Warning{Path: path, Message: err.Error()})
			return nil
		}
		*candidates = append(*candidates, candidate{
			Skill:    Skill{Name: metadata.Name, Description: metadata.Description, Scope: root.Scope, Path: filepath.Clean(path)},
			priority: effectivePriority(root.Scope, root.Priority),
		})
		return nil
	})
	if err != nil {
		*warnings = append(*warnings, Warning{Path: root.Path, Message: err.Error()})
	}
}

func readMetadata(path string) (struct{ Name, Description string }, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return struct{ Name, Description string }{}, err
	}
	normalized := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	if !bytes.HasPrefix(normalized, []byte("---\n")) {
		return struct{ Name, Description string }{}, fmt.Errorf("frontmatter YAML ausente")
	}
	end := bytes.Index(normalized[4:], []byte("\n---\n"))
	if end < 0 {
		return struct{ Name, Description string }{}, fmt.Errorf("frontmatter YAML sin cierre")
	}
	var metadata struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal(normalized[4:4+end], &metadata); err != nil {
		return struct{ Name, Description string }{}, fmt.Errorf("frontmatter YAML inválido: %w", err)
	}
	metadata.Name = strings.TrimSpace(metadata.Name)
	metadata.Description = strings.TrimSpace(metadata.Description)
	if metadata.Name == "" {
		return struct{ Name, Description string }{}, fmt.Errorf("frontmatter sin name")
	}
	if metadata.Description == "" {
		return struct{ Name, Description string }{}, fmt.Errorf("frontmatter sin description")
	}
	return struct{ Name, Description string }{Name: metadata.Name, Description: metadata.Description}, nil
}

func selectSkills(candidates []candidate) []Skill {
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Name != candidates[j].Name {
			return candidates[i].Name < candidates[j].Name
		}
		if candidates[i].priority != candidates[j].priority {
			return candidates[i].priority < candidates[j].priority
		}
		return candidates[i].Path < candidates[j].Path
	})
	selected := []Skill{}
	for i := 0; i < len(candidates); {
		winner := candidates[i].Skill
		j := i + 1
		for j < len(candidates) && candidates[j].Name == candidates[i].Name {
			winner.ShadowedPaths = append(winner.ShadowedPaths, candidates[j].Path)
			j++
		}
		selected = append(selected, winner)
		i = j
	}
	return selected
}

func scopePriority(scope string) int {
	if scope == "project" {
		return 0
	}
	return 1
}

func effectivePriority(scope string, declared int) int {
	return scopePriority(scope)*1000 + declared
}
