package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

type Kind string

const (
	KindFile Kind = "file"
	KindDir  Kind = "dir"
)

type Policy string

const (
	PolicyManaged  Policy = "managed"
	PolicyMetadata Policy = "metadata"
)

type Asset struct {
	ID           string `json:"id"`
	SourceRel    string `json:"sourceRel"`
	TargetRel    string `json:"targetRel"`
	Kind         Kind   `json:"kind"`
	Policy       Policy `json:"policy"`
	SourceSHA256 string `json:"sourceSHA256,omitempty"`
}

type Catalog struct {
	SourceRoot string
	Assets     []Asset
}

type entry struct {
	sourceRel string
	targetRel string
	kind      Kind
	policy    Policy
}

var allowedEntries = []entry{
	{sourceRel: ".opencode/agents", targetRel: ".opencode/agents", kind: KindDir, policy: PolicyManaged},
	{sourceRel: ".opencode/commands", targetRel: ".opencode/commands", kind: KindDir, policy: PolicyManaged},
	{sourceRel: ".opencode/skills", targetRel: ".opencode/skills", kind: KindDir, policy: PolicyManaged},
	{sourceRel: ".opencode/policies", targetRel: ".opencode/policies", kind: KindDir, policy: PolicyManaged},
	{sourceRel: ".opencode/plugins", targetRel: ".opencode/plugins", kind: KindDir, policy: PolicyManaged},
	{sourceRel: ".opencode/agent-observatory", targetRel: ".opencode/agent-observatory", kind: KindDir, policy: PolicyManaged},
	{sourceRel: ".opencode/README.md", targetRel: ".opencode/README.md", kind: KindFile, policy: PolicyManaged},
	{sourceRel: ".opencode/package.json", targetRel: ".opencode/package.json", kind: KindFile, policy: PolicyManaged},
	{sourceRel: ".opencode/package-lock.json", targetRel: ".opencode/package-lock.json", kind: KindFile, policy: PolicyManaged},
	{sourceRel: ".opencode/.gitignore", targetRel: ".opencode/.gitignore", kind: KindFile, policy: PolicyManaged},
	{sourceRel: "AGENTS.md.template", targetRel: "AGENTS.md", kind: KindFile, policy: PolicyManaged},
	{sourceRel: "tui.json", targetRel: "tui.json", kind: KindFile, policy: PolicyManaged},
	{sourceRel: "openspec", targetRel: "openspec", kind: KindDir, policy: PolicyManaged},
}

func BuildCatalog(sourceRoot string) (Catalog, error) {
	var out []Asset
	for _, ent := range allowedEntries {
		sourceRel, err := platform.EnsureRelativeSafe(ent.sourceRel)
		if err != nil {
			return Catalog{}, err
		}
		targetRel, err := platform.EnsureRelativeSafe(ent.targetRel)
		if err != nil {
			return Catalog{}, err
		}
		if ent.kind == KindDir {
			out = append(out, Asset{ID: targetRel, SourceRel: sourceRel, TargetRel: targetRel, Kind: KindDir, Policy: ent.policy})
			files, err := expandDir(sourceRoot, sourceRel, targetRel, ent.policy)
			if err != nil {
				return Catalog{}, err
			}
			out = append(out, files...)
			continue
		}
		asset, err := fileAsset(sourceRoot, sourceRel, targetRel, ent.policy)
		if err != nil {
			return Catalog{}, err
		}
		out = append(out, asset)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TargetRel < out[j].TargetRel })
	return Catalog{SourceRoot: sourceRoot, Assets: out}, nil
}

func expandDir(sourceRoot, sourceRel, targetRel string, policy Policy) ([]Asset, error) {
	root := filepath.Join(sourceRoot, sourceRel)
	if info, err := os.Lstat(root); err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("directorio fuente inválido: %s", sourceRel)
	}
	var out []Asset
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink fuente no soportado: %s", path)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if sourceRel == "openspec" && rel == "changes" && d.IsDir() {
			return filepath.SkipDir
		}
		if sourceRel == "openspec" && strings.HasPrefix(rel, "changes"+string(filepath.Separator)) {
			return nil
		}
		src := filepath.Join(sourceRel, rel)
		dst := filepath.Join(targetRel, rel)
		src, err = platform.EnsureRelativeSafe(src)
		if err != nil {
			return err
		}
		dst, err = platform.EnsureRelativeSafe(dst)
		if err != nil {
			return err
		}
		if d.IsDir() {
			out = append(out, Asset{ID: dst, SourceRel: src, TargetRel: dst, Kind: KindDir, Policy: policy})
			return nil
		}
		asset, err := fileAsset(sourceRoot, src, dst, policy)
		if err != nil {
			return err
		}
		out = append(out, asset)
		return nil
	})
	return out, err
}

func fileAsset(sourceRoot, sourceRel, targetRel string, policy Policy) (Asset, error) {
	hash, err := FileSHA256(filepath.Join(sourceRoot, sourceRel))
	if err != nil {
		return Asset{}, err
	}
	return Asset{ID: targetRel, SourceRel: sourceRel, TargetRel: targetRel, Kind: KindFile, Policy: policy, SourceSHA256: hash}, nil
}

func FileSHA256(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("archivo fuente inválido: %s", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
