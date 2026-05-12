package merger

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

type Options struct {
	Target string
	Path   string
}

type Service struct{}

func NewService() Service { return Service{} }

func (s Service) Run(opts Options, stdout io.Writer) error {
	targetRoot, err := platform.ResolveTargetPath(opts.Target)
	if err != nil {
		return err
	}
	targetRel, err := platform.EnsureRelativeSafe(opts.Path)
	if err != nil {
		return err
	}
	st, err := state.Load(targetRoot)
	if err != nil {
		return err
	}
	if st == nil {
		return fmt.Errorf("merge requiere install-state.json")
	}
	asset, ok := st.AssetMap()[targetRel]
	if !ok {
		return fmt.Errorf("merge requiere asset gestionado en install-state: %s", targetRel)
	}
	targetPath, err := platform.SafeJoin(targetRoot, targetRel)
	if err != nil {
		return err
	}
	ancestorRel := asset.AncestorRel
	if ancestorRel == "" {
		ancestorRel, err = state.AncestorRel(targetRel)
		if err != nil {
			return err
		}
	}
	ancestorPath, err := platform.SafeJoin(targetRoot, ancestorRel)
	if err != nil {
		return err
	}
	lufyNewPath, err := platform.SafeJoin(targetRoot, targetRel+".lufy-new")
	if err != nil {
		return err
	}
	for label, path := range map[string]string{"target": targetPath, "ancestor": ancestorPath, "lufy-new": lufyNewPath} {
		if !regularFile(path) {
			return fmt.Errorf("merge requiere %s existente y seguro: %s", label, path)
		}
	}
	tool := os.Getenv("LUFY_MERGE_TOOL")
	if tool == "" {
		return fmt.Errorf("merge requiere LUFY_MERGE_TOOL configurado")
	}
	parts := strings.Fields(tool)
	if len(parts) == 0 {
		return fmt.Errorf("merge requiere LUFY_MERGE_TOOL configurado")
	}
	cmd := exec.Command(parts[0], append(parts[1:], ancestorPath, targetPath, lufyNewPath)...)
	cmd.Stdout = stdout
	cmd.Stderr = stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("merge tool falló; archivos preservados: %w", err)
	}
	fmt.Fprintf(stdout, "Merge tool completado para %s\n", targetRel)
	return nil
}

func regularFile(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0
}
