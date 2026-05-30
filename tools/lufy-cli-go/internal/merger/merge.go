package merger

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

type Options struct {
	Target       string
	Path         string
	AcceptTheirs bool
	AcceptOurs   bool
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
	if opts.AcceptTheirs && opts.AcceptOurs {
		return fmt.Errorf("merge no permite combinar --accept-theirs y --accept-ours")
	}
	if opts.AcceptTheirs || opts.AcceptOurs {
		return acceptResolution(targetRoot, targetRel, targetPath, ancestorPath, lufyNewPath, opts.AcceptTheirs, st, stdout)
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

func acceptResolution(targetRoot, targetRel, targetPath, ancestorPath, lufyNewPath string, acceptTheirs bool, st *state.InstallState, stdout io.Writer) error {
	lufyNewHash, err := assets.FileSHA256(lufyNewPath)
	if err != nil {
		return err
	}
	resolvedPath := targetPath
	action := "merge-accept-ours"
	if acceptTheirs {
		resolvedPath = lufyNewPath
		action = "merge-accept-theirs"
	}
	body, err := os.ReadFile(resolvedPath)
	if err != nil {
		return err
	}
	perm := filePerm(targetPath, 0o644)
	if acceptTheirs {
		if err := platform.WriteFileAtomic(targetPath, body, perm); err != nil {
			return err
		}
	}
	if err := platform.WriteFileAtomic(ancestorPath, body, filePerm(ancestorPath, 0o644)); err != nil {
		return err
	}
	resolvedHash, err := assets.FileSHA256(targetPath)
	if err != nil {
		return err
	}
	updated := false
	for i := range st.Assets {
		if st.Assets[i].TargetRel != targetRel {
			continue
		}
		st.Assets[i].SourceSHA256 = lufyNewHash
		st.Assets[i].TargetSHA256 = resolvedHash
		st.Assets[i].AncestorHash = resolvedHash
		st.Assets[i].LastAction = action
		updated = true
		break
	}
	if !updated {
		return fmt.Errorf("merge requiere asset gestionado en install-state: %s", targetRel)
	}
	st.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := state.WriteAtomic(targetRoot, *st); err != nil {
		return err
	}
	if err := os.Remove(lufyNewPath); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Merge %s aplicado para %s\n", strings.TrimPrefix(action, "merge-"), targetRel)
	return nil
}

func filePerm(path string, fallback os.FileMode) os.FileMode {
	info, err := os.Stat(path)
	if err != nil {
		return fallback
	}
	return info.Mode().Perm()
}

func regularFile(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0
}
