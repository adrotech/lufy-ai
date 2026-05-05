package platform

import "path/filepath"

func ResolveTargetPath(raw string) (string, error) {
	if raw == "" {
		raw = "."
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}
