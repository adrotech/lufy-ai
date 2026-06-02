package managedio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/managedcontent"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/mergeblock"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

func CopyRenderedFile(sourceRoot, sourceRel, targetRoot, targetRel string) error {
	content, err := managedcontent.Render(sourceRoot, sourceRel, targetRoot, targetRel)
	if err != nil {
		return err
	}
	return WriteTargetFile(targetRoot, targetRel, content)
}

func WriteAncestor(sourceRoot, sourceRel, targetRoot, targetRel string) error {
	content, err := managedcontent.Render(sourceRoot, sourceRel, targetRoot, targetRel)
	if err != nil {
		return err
	}
	ancestorPath, err := state.AncestorPath(targetRoot, targetRel)
	if err != nil {
		return err
	}
	return platform.WriteFileAtomic(ancestorPath, content, 0o644)
}

func RenderMergeBlock(sourceRoot, sourceRel, targetRoot, targetRel string) ([]byte, error) {
	sourceContent, err := ReadSourceContent(sourceRoot, sourceRel)
	if err != nil {
		return nil, err
	}
	targetPath, err := platform.SafeJoin(targetRoot, targetRel)
	if err != nil {
		return nil, err
	}
	targetContent, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, err
	}
	return mergeblock.Render(targetContent, sourceContent)
}

func ReadSourceContent(sourceRoot, sourceRel string) ([]byte, error) {
	if sourceRoot == assets.EmbeddedSourceRoot {
		return assets.ReadSourceFile(sourceRoot, sourceRel)
	}
	src := filepath.Join(sourceRoot, sourceRel)
	info, err := os.Lstat(src)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("source no es archivo regular seguro: %s", src)
	}
	return os.ReadFile(src)
}

func WriteTargetFile(targetRoot, targetRel string, content []byte) error {
	dst, err := platform.SafeJoin(targetRoot, targetRel)
	if err != nil {
		return err
	}
	if info, err := os.Lstat(dst); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("destino symlink no permitido: %s", dst)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return platform.WriteFileAtomic(dst, content, 0o644)
}

func TrimLufyNewSuffix(path string) string {
	return strings.TrimSuffix(path, ".lufy-new")
}

func UniqueTargets(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, target := range in {
		if seen[target] {
			continue
		}
		seen[target] = true
		out = append(out, target)
	}
	return out
}

func FileExists(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0
}
