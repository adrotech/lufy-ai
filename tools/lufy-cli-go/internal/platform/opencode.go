package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

func ResolveOpenCodeConfigRoot() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return ResolveTargetPath(filepath.Join(xdg, "opencode"))
	}
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("no se pudo resolver config global OpenCode: HOME no definido")
	}
	return ResolveTargetPath(filepath.Join(home, ".config", "opencode"))
}
