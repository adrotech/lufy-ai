package state

import (
	"path/filepath"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

const AncestorsDir = "ancestors"

func AncestorRel(targetRel string) (string, error) {
	clean, err := platform.EnsureRelativeSafe(targetRel)
	if err != nil {
		return "", err
	}
	parts := strings.Split(filepath.ToSlash(clean), "/")
	return filepath.ToSlash(filepath.Join(append([]string{".lufy-ai", AncestorsDir}, parts...)...)), nil
}

func AncestorPath(targetRoot, targetRel string) (string, error) {
	rel, err := AncestorRel(targetRel)
	if err != nil {
		return "", err
	}
	return platform.SafeJoin(targetRoot, rel)
}
