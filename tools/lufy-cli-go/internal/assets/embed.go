package assets

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

const EmbeddedSourceRoot = "embedded://lufy-ai-managed-assets"

//go:embed all:embedded
var embeddedFS embed.FS

func BuildEmbeddedCatalog() (Catalog, error) {
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
			out = append(out, withOwnership(Asset{ID: targetRel, SourceRel: sourceRel, TargetRel: targetRel, Kind: KindDir, Policy: ent.policy, Scope: ent.scope}))
			files, err := expandEmbeddedDir(sourceRel, targetRel, ent.policy, ent.scope)
			if err != nil {
				return Catalog{}, err
			}
			out = append(out, files...)
			continue
		}
		asset, err := embeddedFileAsset(sourceRel, targetRel, ent.policy, ent.scope)
		if err != nil {
			return Catalog{}, err
		}
		out = append(out, asset)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TargetRel < out[j].TargetRel })
	return Catalog{SourceRoot: EmbeddedSourceRoot, Assets: out}, nil
}

func ReadSourceFile(sourceRoot, sourceRel string) ([]byte, error) {
	if sourceRoot != EmbeddedSourceRoot {
		return nil, fmt.Errorf("source root no embebido: %s", sourceRoot)
	}
	clean, err := platform.EnsureRelativeSafe(sourceRel)
	if err != nil {
		return nil, err
	}
	body, err := embeddedFS.ReadFile(filepath.ToSlash(filepath.Join("embedded", clean)))
	if err != nil {
		return nil, err
	}
	return body, nil
}

func expandEmbeddedDir(sourceRel, targetRel string, policy Policy, scope Scope) ([]Asset, error) {
	root := filepath.ToSlash(filepath.Join("embedded", sourceRel))
	info, err := fs.Stat(embeddedFS, root)
	if err != nil || !info.IsDir() {
		if err == nil {
			err = fmt.Errorf("no es directorio")
		}
		return nil, fmt.Errorf("directorio embebido inválido: %s: %w", sourceRel, err)
	}
	var out []Asset
	err = fs.WalkDir(embeddedFS, root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if sourceRel == "openspec" && (rel == "changes" || strings.HasPrefix(rel, "changes/")) {
			if d.IsDir() {
				return fs.SkipDir
			}
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
			out = append(out, withOwnership(Asset{ID: dst, SourceRel: src, TargetRel: dst, Kind: KindDir, Policy: policy, Scope: scope}))
			return nil
		}
		asset, err := embeddedFileAsset(src, dst, policy, scope)
		if err != nil {
			return err
		}
		out = append(out, asset)
		return nil
	})
	return out, err
}

func embeddedFileAsset(sourceRel, targetRel string, policy Policy, scope Scope) (Asset, error) {
	body, err := ReadSourceFile(EmbeddedSourceRoot, sourceRel)
	if err != nil {
		return Asset{}, err
	}
	h := sha256.Sum256(body)
	return withOwnership(Asset{ID: targetRel, SourceRel: sourceRel, TargetRel: targetRel, Kind: KindFile, Policy: policy, Scope: scope, SourceSHA256: hex.EncodeToString(h[:])}), nil
}
