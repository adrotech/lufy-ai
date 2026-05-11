package platform

import (
	"os"
	"path/filepath"
)

func ResolveTargetPath(raw string) (string, error) {
	if raw == "" {
		raw = "."
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", err
	}
	abs = filepath.Clean(abs)
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return filepath.Clean(resolved), nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	parent := filepath.Dir(abs)
	resolvedParent, parentErr := filepath.EvalSymlinks(parent)
	if parentErr == nil {
		return filepath.Join(filepath.Clean(resolvedParent), filepath.Base(abs)), nil
	}
	if os.IsNotExist(parentErr) {
		return abs, nil
	}
	return "", parentErr
}
