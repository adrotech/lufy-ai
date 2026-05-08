package verify

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/config"
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

var requiredDirs = []string{
	filepath.Join(".opencode", "agents"),
	filepath.Join(".opencode", "commands"),
	filepath.Join(".opencode", "skills"),
	filepath.Join(".opencode", "plugins"),
	filepath.Join(".opencode", "policies"),
}

var requiredManagedFiles = []string{
	"AGENTS.md",
	filepath.Join(".opencode", "plugins", "agent-observatory.tsx"),
	"tui.json",
	filepath.Join("openspec", "config.yaml"),
}

var requiredStateFiles = []string{filepath.Join(".lufy-ai", "install-state.json")}

var jsonValidationFiles = []string{
	config.OpenCodeFile,
	"tui.json",
	filepath.Join(".opencode", "package.json"),
	filepath.Join(".opencode", "package-lock.json"),
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
	failures := 0
	for _, rel := range requiredStateFiles {
		path, err := platform.SafeJoin(target, rel)
		if err != nil {
			return err
		}
		if !regularFile(path) {
			fmt.Fprintf(stdout, "fail: falta archivo crítico: %s\n", rel)
			failures++
			continue
		}
		fmt.Fprintf(stdout, "ok: archivo crítico %s\n", rel)
	}
	if st.SourceChangeID == "" || st.SourceRootFingerprint == "" {
		fmt.Fprintln(stdout, "fail: install-state.json no contiene fingerprint de catálogo")
		failures++
	}
	for _, rel := range jsonValidationFiles {
		status, err := validateJSONFile(target, rel)
		if err != nil {
			fmt.Fprintf(stdout, "fail: JSON inválido en %s: %s\n", rel, err.Error())
			failures++
			continue
		}
		if status != "" {
			fmt.Fprintf(stdout, "ok: JSON parseable %s\n", rel)
		}
	}
	if status, err := config.NewService().ValidateManagedStructure(target); err != nil {
		fmt.Fprintf(stdout, "fail: estructura gestionada inválida en %s: %s\n", config.OpenCodeFile, err.Error())
		failures++
	} else if status.Exists {
		fmt.Fprintf(stdout, "ok: estructura merge-managed %s\n", config.OpenCodeFile)
	}
	manifestTarget := st.TargetRoot
	if manifestTarget != "" {
		if resolved, err := platform.ResolveTargetPath(manifestTarget); err == nil {
			manifestTarget = resolved
		}
	}
	if manifestTarget != "" && manifestTarget != target {
		fmt.Fprintf(stdout, "fail: targetRoot del manifest no coincide: manifest=%s actual=%s\n", st.TargetRoot, target)
		failures++
	}
	fmt.Fprintf(stdout, "ok: install-state.json schema=%d assets=%d\n", st.SchemaVersion, len(st.Assets))

	assetMap := st.AssetMap()
	for _, required := range requiredDirs {
		clean, err := platform.EnsureRelativeSafe(required)
		if err != nil {
			return err
		}
		path, err := platform.SafeJoin(target, clean)
		if err != nil {
			fmt.Fprintf(stdout, "fail: directorio crítico inseguro: %s (%s)\n", clean, err.Error())
			failures++
			continue
		}
		if !directory(path) {
			fmt.Fprintf(stdout, "fail: falta directorio crítico: %s\n", clean)
			failures++
			continue
		}
		fmt.Fprintf(stdout, "ok: directorio crítico %s\n", clean)
	}

	for _, required := range requiredManagedFiles {
		clean, err := platform.EnsureRelativeSafe(required)
		if err != nil {
			return err
		}
		if _, ok := assetMap[clean]; !ok {
			fmt.Fprintf(stdout, "fail: asset clave no está en manifest: %s\n", clean)
			failures++
		}
		path, err := platform.SafeJoin(target, clean)
		if err != nil {
			fmt.Fprintf(stdout, "fail: asset clave inseguro: %s (%s)\n", clean, err.Error())
			failures++
			continue
		}
		if !regularFile(path) {
			fmt.Fprintf(stdout, "fail: falta archivo crítico: %s\n", clean)
			failures++
			continue
		}
		fmt.Fprintf(stdout, "ok: archivo crítico %s\n", clean)
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

func directory(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0
}

func validateJSONFile(target, rel string) (string, error) {
	path, err := platform.SafeJoin(target, rel)
	if err != nil {
		return "", err
	}
	body, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return "", err
	}
	return "ok", nil
}

func shortHash(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}
