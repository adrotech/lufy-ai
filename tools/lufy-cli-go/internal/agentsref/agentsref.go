package agentsref

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

const (
	AgentsFile  = "AGENTS.md"
	HarnessFile = "lufy-ia.harness.md"
	Reference   = "@lufy-ia.harness.md"
)

func MinimalContent() []byte {
	return []byte("# AGENTS.md\n\n" + Reference + "\n")
}

func RecommendedInstallAction() string {
	return "ejecuta `lufy-ai install --target <target> --yes` para crear/agregar `" + Reference + "`, o edita AGENTS.md manualmente"
}

func ContainsReference(body []byte) bool {
	return bytes.Contains(body, []byte(Reference))
}

func Status(targetRoot string) (exists bool, hasReference bool, err error) {
	path, err := platform.SafeJoin(targetRoot, AgentsFile)
	if err != nil {
		return false, false, err
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return true, false, fmt.Errorf("%s no es archivo regular seguro", AgentsFile)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return true, false, err
	}
	return true, ContainsReference(body), nil
}

func InsertReference(targetRoot string) error {
	path, err := platform.SafeJoin(targetRoot, AgentsFile)
	if err != nil {
		return err
	}
	body, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return writeAgentsFile(path, MinimalContent())
	}
	if err != nil {
		return err
	}
	if ContainsReference(body) {
		return nil
	}
	updated := appendReference(body)
	return writeAgentsFile(path, updated)
}

func appendReference(body []byte) []byte {
	text := string(body)
	if text == "" {
		return MinimalContent()
	}
	var b strings.Builder
	b.WriteString(text)
	if !strings.HasSuffix(text, "\n") {
		b.WriteString("\n")
	}
	if !strings.HasSuffix(b.String(), "\n\n") {
		b.WriteString("\n")
	}
	b.WriteString(Reference)
	b.WriteString("\n")
	return []byte(b.String())
}

func writeAgentsFile(path string, body []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("destino symlink no permitido: %s", path)
	}
	return platform.WriteFileAtomic(path, body, 0o644)
}
