package assets

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// Comprueba toda la distribución Codex, no sólo fragmentos conocidos del texto.
func TestCodexInstructionDistributionParity(t *testing.T) {
	root := repoRoot(t)
	source := filepath.Join(root, ".codex")
	mirror := filepath.Join(root, "tools", "lufy-cli-go", "internal", "assets", "embedded", ".codex")
	read := func(base string) map[string]string {
		files := map[string]string{}
		err := filepath.WalkDir(base, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			if entry.Type()&os.ModeSymlink != 0 {
				t.Errorf("symlink Codex no permitido: %s", path)
				return nil
			}
			rel, err := filepath.Rel(base, path)
			if err != nil {
				return err
			}
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files[rel] = string(body)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		return files
	}
	left, right := read(source), read(mirror)
	if len(left) != len(right) {
		t.Fatalf("superficies Codex con distintos archivos: %d/%d", len(left), len(right))
	}
	for path, body := range left {
		if other, ok := right[path]; !ok || body != other {
			t.Errorf("Codex source/mirror divergen: %s", path)
		}
	}
}
