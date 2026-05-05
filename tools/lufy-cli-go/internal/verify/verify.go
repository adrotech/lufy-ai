package verify

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

type Options struct {
	Target   string
	NoEngram bool
}

type Service struct{}

func NewService() Service {
	return Service{}
}

func (s Service) Run(opts Options, stdout io.Writer) error {
	target, err := platform.ResolveTargetPath(opts.Target)
	if err != nil {
		return err
	}

	statePath := filepath.Join(target, ".lufy-ai", "install-state.json")
	body, err := os.ReadFile(statePath)
	if err != nil {
		return fmt.Errorf("fail: falta estado de instalación (%s)", statePath)
	}

	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return fmt.Errorf("fail: install-state.json inválido: %w", err)
	}

	fmt.Fprintln(stdout, "ok: install-state.json presente y parseable")
	if opts.NoEngram {
		fmt.Fprintln(stdout, "warn: chequeo de Engram omitido por --no-engram")
	} else if path, ok := platform.ResolveEngram(false, platform.OSResolver{}); ok {
		fmt.Fprintf(stdout, "ok: engram detectado en PATH (%s)\n", path)
	} else {
		fmt.Fprintln(stdout, "warn: engram no encontrado en PATH")
	}

	return nil
}
