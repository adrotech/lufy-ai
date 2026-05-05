package backup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

type Options struct {
	Target string
}

type Manifest struct {
	SchemaVersion int      `json:"schemaVersion"`
	CreatedAt     string   `json:"createdAt"`
	TargetRoot    string   `json:"targetRoot"`
	Files         []string `json:"files"`
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

	ts := time.Now().UTC().Format("20060102T150405Z")
	backupDir := filepath.Join(target, ".lufy-ai", "backups", ts)
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", err
	}

	files := []string{}
	agentsPath := filepath.Join(target, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err == nil {
		backupAgents := filepath.Join(backupDir, "AGENTS.md")
		content, readErr := os.ReadFile(agentsPath)
		if readErr != nil {
			return "", readErr
		}
		if writeErr := os.WriteFile(backupAgents, content, 0o644); writeErr != nil {
			return "", writeErr
		}
		files = append(files, "AGENTS.md")
	}

	manifest := Manifest{
		SchemaVersion: 1,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		TargetRoot:    target,
		Files:         files,
	}
	manifestPath := filepath.Join(backupDir, "manifest.json")
	body, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(manifestPath, body, 0o644); err != nil {
		return "", err
	}

	fmt.Fprintf(stdout, "Backup creado: %s\n", backupDir)
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
	if manifest.TargetRoot != absTarget {
		return fmt.Errorf("backup pertenece a otro target: %s", manifest.TargetRoot)
	}

	for _, file := range manifest.Files {
		src := filepath.Join(filepath.Dir(manifestPath), file)
		dst := filepath.Join(absTarget, file)
		if dryRun {
			fmt.Fprintf(stdout, "[dry-run] restauraría %s\n", file)
			continue
		}
		content, readErr := os.ReadFile(src)
		if readErr != nil {
			return readErr
		}
		if writeErr := os.WriteFile(dst, content, 0o644); writeErr != nil {
			return writeErr
		}
		fmt.Fprintf(stdout, "Restaurado %s\n", file)
	}

	return nil
}
