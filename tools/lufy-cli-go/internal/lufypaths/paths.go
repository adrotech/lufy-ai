package lufypaths

import (
	"os"
	"path/filepath"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

const (
	Root = ".lufy"

	Readme = ".lufy/README.md"

	ProjectConfig       = ".lufy/config/project.yaml"
	LegacyProjectConfig = ".lufy/project.yaml"

	InstallState       = ".lufy/managed-state/install-state.json"
	LegacyInstallState = ".lufy-ai/install-state.json"

	Backups       = ".lufy/managed-state/backups"
	LegacyBackups = ".lufy-ai/backups"

	Ancestors       = ".lufy/managed-state/ancestors"
	LegacyAncestors = ".lufy-ai/ancestors"

	OpenSpecCache       = ".lufy/cache/openspec"
	LegacyOpenSpecCache = ".lufy-ai/openspec-cache"

	LufySDD       = ".lufy/workflows/sdd"
	LegacyLufySDD = ".lufy/sdd"
)

type Resolved struct {
	Rel    string
	Path   string
	Legacy bool
	Exists bool
}

func WritePath(targetRoot, rel string) (string, error) {
	return platform.SafeJoin(targetRoot, rel)
}

func ResolveExisting(targetRoot, canonicalRel, legacyRel string) (Resolved, error) {
	canonical, err := platform.SafeJoin(targetRoot, canonicalRel)
	if err != nil {
		return Resolved{}, err
	}
	if exists(canonical) {
		return Resolved{Rel: canonicalRel, Path: canonical, Exists: true}, nil
	}
	legacy, err := platform.SafeJoin(targetRoot, legacyRel)
	if err != nil {
		return Resolved{}, err
	}
	if exists(legacy) {
		return Resolved{Rel: legacyRel, Path: legacy, Legacy: true, Exists: true}, nil
	}
	return Resolved{Rel: canonicalRel, Path: canonical}, nil
}

func ExistingPaths(targetRoot, canonicalRel, legacyRel string) ([]Resolved, error) {
	var out []Resolved
	canonical, err := platform.SafeJoin(targetRoot, canonicalRel)
	if err != nil {
		return nil, err
	}
	if exists(canonical) {
		out = append(out, Resolved{Rel: canonicalRel, Path: canonical, Exists: true})
	}
	legacy, err := platform.SafeJoin(targetRoot, legacyRel)
	if err != nil {
		return nil, err
	}
	if exists(legacy) {
		out = append(out, Resolved{Rel: legacyRel, Path: legacy, Legacy: true, Exists: true})
	}
	return out, nil
}

func ResolveBackupReference(targetRoot, backupID string) (string, error) {
	for _, rootRel := range []string{Backups, LegacyBackups} {
		path, err := platform.SafeJoin(targetRoot, filepath.Join(rootRel, backupID))
		if err != nil {
			return "", err
		}
		if exists(path) {
			return path, nil
		}
	}
	return platform.SafeJoin(targetRoot, filepath.Join(Backups, backupID))
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}
