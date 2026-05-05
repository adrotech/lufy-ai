package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ResolveSourceRoot(start string) (string, error) {
	if start == "" {
		var err error
		start, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	abs = filepath.Clean(abs)

	info, err := os.Stat(abs)
	if err == nil && !info.IsDir() {
		abs = filepath.Dir(abs)
	}

	for {
		if hasSourceMarkers(abs) {
			return abs, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", fmt.Errorf("no se pudo resolver source root desde %s", start)
		}
		abs = parent
	}
}

func hasSourceMarkers(dir string) bool {
	if !isFile(filepath.Join(dir, "AGENTS.md")) {
		return false
	}
	if !isDir(filepath.Join(dir, ".opencode")) {
		return false
	}
	return isFile(filepath.Join(dir, "openspec", "config.yaml"))
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func EnsureRelativeSafe(path string) (string, error) {
	if path == "" || filepath.IsAbs(path) {
		return "", fmt.Errorf("path no permitido en catálogo: %q", path)
	}
	clean := filepath.Clean(path)
	if clean == "." || clean == ".." || len(clean) >= 3 && clean[:3] == "../" {
		return "", fmt.Errorf("path escapa del root permitido: %q", path)
	}
	return clean, nil
}

func SafeJoin(root, rel string) (string, error) {
	cleanRel, err := EnsureRelativeSafe(rel)
	if err != nil {
		return "", err
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absRoot = filepath.Clean(absRoot)
	full := filepath.Join(absRoot, cleanRel)
	back, err := filepath.Rel(absRoot, full)
	if err != nil {
		return "", err
	}
	if back == ".." || strings.HasPrefix(back, ".."+string(filepath.Separator)) || filepath.IsAbs(back) {
		return "", fmt.Errorf("path escapa del root permitido: %q", rel)
	}

	current := absRoot
	for _, part := range strings.Split(cleanRel, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			break
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("symlink no permitido dentro del root: %s", current)
		}
	}
	return full, nil
}
