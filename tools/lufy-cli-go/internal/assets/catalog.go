package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	PolicyManaged    Policy = "managed"
	PolicyNoReplace  Policy = "no-replace"
	PolicyMergeBlock Policy = "merge-block"
	PolicyMergeJSON  Policy = "merge-json"
	PolicyMetadata   Policy = "metadata"
)

func (p Policy) Valid() bool {
	switch p {
	case PolicyManaged, PolicyNoReplace, PolicyMergeBlock, PolicyMergeJSON, PolicyMetadata:
		return true
	default:
		return false
	}
}

func (p Policy) SupportsAncestor() bool {
	switch p {
	case PolicyManaged, PolicyNoReplace, PolicyMergeBlock:
		return true
	default:
		return false
	}
}

type Scope string

const (
	ScopeProject Scope = "project"
	ScopeGlobal  Scope = "global"
	ScopeBoth    Scope = "both"
)

func (s Scope) Valid() bool {
	switch s {
	case ScopeProject, ScopeGlobal, ScopeBoth:
		return true
	default:
		return false
	}
}

func ParseScope(value string) (Scope, error) {
	if value == "" {
		value = string(ScopeProject)
	}
	scope := Scope(value)
	if !scope.Valid() {
		return "", fmt.Errorf("scope no soportado %q; valores permitidos: project, global, both", value)
	}
	return scope, nil
}

type Asset struct {
	ID           string `json:"id"`
	SourceRel    string `json:"sourceRel"`
	TargetRel    string `json:"targetRel"`
	Kind         Kind   `json:"kind"`
	Policy       Policy `json:"policy"`
	Scope        Scope  `json:"scope"`
	SourceSHA256 string `json:"sourceSHA256,omitempty"`
}

type Catalog struct {
	SourceRoot string
	Assets     []Asset
}

func (c Catalog) Fingerprint() (string, error) {
	type item struct {
		TargetRel    string `json:"targetRel"`
		SourceSHA256 string `json:"sourceSHA256"`
	}
	var items []item
	for _, asset := range c.Assets {
		if asset.Kind != KindFile {
			continue
		}
		items = append(items, item{TargetRel: filepath.ToSlash(asset.TargetRel), SourceSHA256: asset.SourceSHA256})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].TargetRel < items[j].TargetRel })
	body, err := json.Marshal(items)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:]), nil
}

type entry struct {
	sourceRel string
	targetRel string
	kind      Kind
	policy    Policy
	scope     Scope
}

var allowedEntries = []entry{
	{sourceRel: ".opencode/agents", targetRel: ".opencode/agents", kind: KindDir, policy: PolicyManaged, scope: ScopeProject},
	{sourceRel: ".opencode/commands", targetRel: ".opencode/commands", kind: KindDir, policy: PolicyManaged, scope: ScopeProject},
	{sourceRel: ".opencode/hooks", targetRel: ".opencode/hooks", kind: KindDir, policy: PolicyManaged, scope: ScopeProject},
	{sourceRel: ".opencode/skills", targetRel: ".opencode/skills", kind: KindDir, policy: PolicyManaged, scope: ScopeProject},
	{sourceRel: ".opencode/templates", targetRel: ".opencode/templates", kind: KindDir, policy: PolicyManaged, scope: ScopeProject},
	{sourceRel: ".opencode/policies", targetRel: ".opencode/policies", kind: KindDir, policy: PolicyManaged, scope: ScopeProject},
	{sourceRel: ".opencode/plugins", targetRel: ".opencode/plugins", kind: KindDir, policy: PolicyManaged, scope: ScopeProject},
	{sourceRel: ".opencode/agent-observatory", targetRel: ".opencode/agent-observatory", kind: KindDir, policy: PolicyManaged, scope: ScopeProject},
	{sourceRel: ".opencode/README.md", targetRel: ".opencode/README.md", kind: KindFile, policy: PolicyManaged, scope: ScopeProject},
	{sourceRel: ".opencode/package.json", targetRel: ".opencode/package.json", kind: KindFile, policy: PolicyManaged, scope: ScopeProject},
	{sourceRel: ".opencode/package-lock.json", targetRel: ".opencode/package-lock.json", kind: KindFile, policy: PolicyManaged, scope: ScopeProject},
	{sourceRel: ".opencode/.gitignore", targetRel: ".opencode/.gitignore", kind: KindFile, policy: PolicyManaged, scope: ScopeProject},
	{sourceRel: "lufy-ia.harness.md", targetRel: "lufy-ia.harness.md", kind: KindFile, policy: PolicyManaged, scope: ScopeProject},
	{sourceRel: "tui.json", targetRel: "tui.json", kind: KindFile, policy: PolicyNoReplace, scope: ScopeProject},
	{sourceRel: "openspec", targetRel: "openspec", kind: KindDir, policy: PolicyManaged, scope: ScopeProject},
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
		if !ent.policy.Valid() {
			return Catalog{}, fmt.Errorf("policy de asset no soportada: %s", ent.policy)
		}
		if !ent.scope.Valid() {
			return Catalog{}, fmt.Errorf("scope de asset no soportado: %s", ent.scope)
		}
		if ent.kind == KindDir {
			out = append(out, Asset{ID: targetRel, SourceRel: sourceRel, TargetRel: targetRel, Kind: KindDir, Policy: ent.policy, Scope: ent.scope})
			files, err := expandDir(sourceRoot, sourceRel, targetRel, ent.policy, ent.scope)
			if err != nil {
				return Catalog{}, err
			}
			out = append(out, files...)
			continue
		}
		asset, err := fileAsset(sourceRoot, sourceRel, targetRel, ent.policy, ent.scope)
		if err != nil {
			return Catalog{}, err
		}
		out = append(out, asset)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TargetRel < out[j].TargetRel })
	return Catalog{SourceRoot: sourceRoot, Assets: out}, nil
}

func expandDir(sourceRoot, sourceRel, targetRel string, policy Policy, scope Scope) ([]Asset, error) {
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
			out = append(out, Asset{ID: dst, SourceRel: src, TargetRel: dst, Kind: KindDir, Policy: policy, Scope: scope})
			return nil
		}
		asset, err := fileAsset(sourceRoot, src, dst, policy, scope)
		if err != nil {
			return err
		}
		out = append(out, asset)
		return nil
	})
	return out, err
}

func fileAsset(sourceRoot, sourceRel, targetRel string, policy Policy, scope Scope) (Asset, error) {
	hash, err := FileSHA256(filepath.Join(sourceRoot, sourceRel))
	if err != nil {
		return Asset{}, err
	}
	return Asset{ID: targetRel, SourceRel: sourceRel, TargetRel: targetRel, Kind: KindFile, Policy: policy, Scope: scope, SourceSHA256: hash}, nil
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
