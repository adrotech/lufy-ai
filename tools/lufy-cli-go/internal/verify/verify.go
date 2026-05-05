package verify

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

type Options struct {
	Target   string
	NoEngram bool
}

type Service struct{}

func NewService() Service {
	return Service{}
}

var requiredAssets = []string{
	"AGENTS.md",
	filepath.Join(".opencode", "agents", "orchestrator.md"),
	filepath.Join(".opencode", "commands", "opsx-apply.md"),
	filepath.Join(".opencode", "package.json"),
	"tui.json",
	filepath.Join("openspec", "config.yaml"),
}

func (s Service) Run(opts Options, stdout io.Writer) error {
	target, err := platform.ResolveTargetPath(opts.Target)
	if err != nil {
		return err
	}

	st, err := state.Load(target)
	if err != nil {
		return fmt.Errorf("fail: install-state.json inválido: %w", err)
	}
	if st == nil {
		return fmt.Errorf("fail: falta estado de instalación (%s)", state.Path(target))
	}
	if st.SourceChangeID != state.SourceChangeID {
		fmt.Fprintf(stdout, "warn: sourceChangeID inesperado: %s\n", st.SourceChangeID)
	}
	failures := 0
	if st.TargetRoot != "" && st.TargetRoot != target {
		fmt.Fprintf(stdout, "fail: targetRoot del manifest no coincide: manifest=%s actual=%s\n", st.TargetRoot, target)
		failures++
	}
	fmt.Fprintf(stdout, "ok: install-state.json schema=%d assets=%d\n", st.SchemaVersion, len(st.Assets))

	assetMap := st.AssetMap()
	for _, required := range requiredAssets {
		clean, err := platform.EnsureRelativeSafe(required)
		if err != nil {
			return err
		}
		if _, ok := assetMap[clean]; !ok {
			fmt.Fprintf(stdout, "fail: asset clave no está en manifest: %s\n", clean)
			failures++
			continue
		}
		path, err := platform.SafeJoin(target, clean)
		if err != nil {
			fmt.Fprintf(stdout, "fail: asset clave inseguro: %s (%s)\n", clean, err.Error())
			failures++
			continue
		}
		if !regularFile(path) {
			fmt.Fprintf(stdout, "fail: falta asset clave: %s\n", clean)
			failures++
		}
	}

	for _, asset := range st.Assets {
		clean, err := platform.EnsureRelativeSafe(asset.TargetRel)
		if err != nil {
			return err
		}
		path, err := platform.SafeJoin(target, clean)
		if err != nil {
			fmt.Fprintf(stdout, "fail: asset inseguro: %s (%s)\n", clean, err.Error())
			failures++
			continue
		}
		if !regularFile(path) {
			fmt.Fprintf(stdout, "fail: falta asset gestionado: %s\n", clean)
			failures++
			continue
		}
		actual, err := assets.FileSHA256(path)
		if err != nil {
			return err
		}
		if actual != asset.TargetSHA256 {
			fmt.Fprintf(stdout, "fail: drift en %s expected=%s actual=%s\n", clean, shortHash(asset.TargetSHA256), shortHash(actual))
			failures++
			continue
		}
		fmt.Fprintf(stdout, "ok: %s sha256=%s\n", clean, shortHash(actual))
	}

	if opts.NoEngram {
		fmt.Fprintln(stdout, "warn: chequeo de Engram omitido por --no-engram")
	} else if path, ok := platform.ResolveEngram(false, platform.OSResolver{}); ok {
		fmt.Fprintf(stdout, "ok: engram detectado en PATH (%s)\n", path)
	} else {
		fmt.Fprintln(stdout, "warn: engram no encontrado en PATH")
	}

	if failures > 0 {
		return fmt.Errorf("verify falló con %d problema(s) crítico(s)", failures)
	}
	fmt.Fprintln(stdout, "ok: verify estructural completo")
	return nil
}

func regularFile(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0
}

func shortHash(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}
