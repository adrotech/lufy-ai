package backup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

const SchemaVersion = 1

type Options struct {
	Target string
}

type Manifest struct {
	SchemaVersion int        `json:"schemaVersion"`
	CreatedAt     string     `json:"createdAt"`
	TargetRoot    string     `json:"targetRoot"`
	Cause         string     `json:"cause"`
	Files         []FileItem `json:"files"`
}

type FileItem struct {
	Path       string `json:"path"`
	BackupPath string `json:"backupPath"`
	SHA256     string `json:"sha256"`
	Size       int64  `json:"size"`
	CapturedAt string `json:"capturedAt"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
	Cause      string `json:"cause"`
}

type Service struct{}

func NewService() Service {
	return Service{}
}

func (s Service) Run(opts Options, stdout io.Writer) (string, error) {
	target, err := platform.ResolveTargetPath(opts.Target)
	if err != nil {
		return "", err
	}
	rels, err := managedExistingFiles(target)
	if err != nil {
		return "", err
	}
	return BackupFiles(target, rels, "manual-backup", stdout)
}

func BackupFiles(target string, rels []string, cause string, stdout io.Writer) (string, error) {
	if len(rels) == 0 {
		return "", fmt.Errorf("no hay archivos gestionados existentes para respaldar")
	}
	ts := time.Now().UTC().Format("20060102T150405.000000000Z")
	backupDir := filepath.Join(target, ".lufy-ai", "backups", ts)
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", err
	}

	manifest := Manifest{SchemaVersion: SchemaVersion, CreatedAt: time.Now().UTC().Format(time.RFC3339), TargetRoot: target, Cause: cause}
	for _, rel := range uniqueStrings(rels) {
		clean, err := platform.EnsureRelativeSafe(rel)
		if err != nil {
			return "", err
		}
		src, err := platform.SafeJoin(target, clean)
		if err != nil {
			return "", err
		}
		item := FileItem{Path: clean, BackupPath: clean, CapturedAt: time.Now().UTC().Format(time.RFC3339), Cause: cause}
		info, err := os.Lstat(src)
		if os.IsNotExist(err) {
			item.Status = "missing"
			manifest.Files = append(manifest.Files, item)
			continue
		}
		if err != nil {
			return "", err
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			item.Status = "error"
			item.Error = "no es archivo regular seguro"
			manifest.Files = append(manifest.Files, item)
			continue
		}
		item.Size = info.Size()
		item.SHA256, err = assets.FileSHA256(src)
		if err != nil {
			return "", err
		}
		if err := copyFile(src, backupDir, clean); err != nil {
			return "", err
		}
		item.Status = "captured"
		manifest.Files = append(manifest.Files, item)
	}

	manifestPath := filepath.Join(backupDir, "manifest.json")
	body, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", err
	}
	body = append(body, '\n')
	if err := os.WriteFile(manifestPath, body, 0o644); err != nil {
		return "", err
	}
	fmt.Fprintf(stdout, "Backup creado: %s (%d archivo(s))\n", backupDir, len(manifest.Files))
	return backupDir, nil
}

func (s Service) Restore(target, backupPath string, dryRun bool, stdout io.Writer) error {
	absTarget, err := platform.ResolveTargetPath(target)
	if err != nil {
		return err
	}
	absBackup, err := platform.ResolveTargetPath(backupPath)
	if err != nil {
		return err
	}

	manifestPath := absBackup
	if filepath.Base(absBackup) != "manifest.json" {
		manifestPath = filepath.Join(absBackup, "manifest.json")
	}
	body, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	var manifest Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return err
	}
	if manifest.SchemaVersion != SchemaVersion {
		return fmt.Errorf("schema de backup no soportado: %d", manifest.SchemaVersion)
	}
	if manifest.TargetRoot != absTarget {
		return fmt.Errorf("backup pertenece a otro target: %s", manifest.TargetRoot)
	}

	restoreFiles, err := capturedPaths(manifest.Files)
	if err != nil {
		return err
	}
	recoveryBackup := ""
	if !dryRun {
		existing := existingFiles(absTarget, restoreFiles)
		if len(existing) > 0 {
			recoveryBackup, err = BackupFiles(absTarget, existing, "pre-restore-recovery", stdout)
			if err != nil {
				return err
			}
			fmt.Fprintf(stdout, "Backup previo a restore: %s\n", recoveryBackup)
		}
	}

	for _, file := range manifest.Files {
		if file.Status != "captured" {
			fmt.Fprintf(stdout, "[skip] %s no fue capturado (%s)\n", file.Path, file.Status)
			continue
		}
		path, err := platform.EnsureRelativeSafe(file.Path)
		if err != nil {
			return err
		}
		backupRel, err := platform.EnsureRelativeSafe(file.BackupPath)
		if err != nil {
			return err
		}
		src := filepath.Join(filepath.Dir(manifestPath), backupRel)
		if dryRun {
			fmt.Fprintf(stdout, "[dry-run] restauraría %s sha256=%s\n", path, shortHash(file.SHA256))
			continue
		}
		if err := copyFile(src, absTarget, path); err != nil {
			if recoveryBackup != "" {
				return fmt.Errorf("restore falló después de crear backup de recovery en %s: %w", recoveryBackup, err)
			}
			return err
		}
		fmt.Fprintf(stdout, "Restaurado %s\n", path)
	}

	return nil
}

func capturedPaths(files []FileItem) ([]string, error) {
	var out []string
	for _, file := range files {
		if file.Status != "captured" {
			continue
		}
		path, err := platform.EnsureRelativeSafe(file.Path)
		if err != nil {
			return nil, err
		}
		out = append(out, path)
	}
	return out, nil
}

func existingFiles(target string, rels []string) []string {
	var out []string
	for _, rel := range rels {
		path, err := platform.SafeJoin(target, rel)
		if err != nil {
			continue
		}
		if info, err := os.Lstat(path); err == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
			out = append(out, rel)
		}
	}
	return out
}

func managedExistingFiles(target string) ([]string, error) {
	st, err := state.Load(target)
	if err != nil {
		return nil, err
	}
	if st == nil {
		if _, err := os.Stat(filepath.Join(target, "AGENTS.md")); err == nil {
			return []string{"AGENTS.md"}, nil
		}
		return nil, nil
	}
	var rels []string
	for _, asset := range st.Assets {
		clean, err := platform.EnsureRelativeSafe(asset.TargetRel)
		if err != nil {
			return nil, err
		}
		path, err := platform.SafeJoin(target, clean)
		if err != nil {
			continue
		}
		if info, err := os.Lstat(path); err == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
			rels = append(rels, clean)
		}
	}
	return rels, nil
}

func copyFile(src, targetRoot, targetRel string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("archivo no seguro para copiar: %s", src)
	}
	dst, err := platform.SafeJoin(targetRoot, targetRel)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0o644)
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func shortHash(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}
