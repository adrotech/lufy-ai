package lufysdd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"gopkg.in/yaml.v3"
)

func (s Service) Archive(target, change string) (Report, error) {
	report, err := inspect(target, change, "archive", true)
	if err != nil {
		return Report{}, err
	}
	if !report.Valid() {
		report.Status = "blocked"
		return report, nil
	}
	if report.Progress.Total == 0 || report.Progress.Complete != report.Progress.Total {
		report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "incomplete_tasks", Path: filepath.ToSlash(filepath.Join("changes", change, "tasks.md")), Message: "archive requiere todas las tareas completas"})
		report.Status = "blocked"
		return report, nil
	}
	root := report.Root
	changeRoot, err := platform.SafeJoin(root, filepath.Join("changes", change))
	if err != nil {
		return Report{}, err
	}
	metadataPath, err := platform.SafeJoin(changeRoot, "change.yaml")
	if err != nil {
		return Report{}, err
	}
	metadata, err := readMetadata(metadataPath)
	if err != nil {
		return Report{}, err
	}
	if metadata.ID != change {
		report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "metadata_id_mismatch", Path: filepath.ToSlash(filepath.Join("changes", change, "change.yaml")), Message: "metadata.id no coincide con el change solicitado"})
	}
	if report.Mode == ModeFull && (metadata.SyncedDigest == "" || metadata.SyncedDigest != report.DeltaDigest) {
		report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "sync_pending", Path: filepath.ToSlash(filepath.Join("changes", change, "change.yaml")), Message: "los deltas actuales no están sincronizados"})
	}
	evidence, evidenceDiagnostics, err := verificationEvidence(root, change)
	if err != nil {
		return Report{}, err
	}
	report.Diagnostics = append(report.Diagnostics, evidenceDiagnostics...)
	if !evidence {
		report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "verification_pending", Path: filepath.ToSlash(filepath.Join("verification", change)), Message: "archive requiere al menos un archivo regular de evidencia"})
	}
	sortDiagnostics(report.Diagnostics)
	if !report.Valid() {
		report.Status = "blocked"
		return report, nil
	}
	if err := s.renderOverview(root, change, report.Mode); err != nil {
		return Report{}, err
	}

	archiveRoot, err := platform.SafeJoin(root, "archive")
	if err != nil {
		return Report{}, err
	}
	if err := requireDirectory(archiveRoot); err != nil {
		return Report{}, err
	}
	now := time.Now()
	if s.now != nil {
		now = s.now()
	}
	destinationRel := filepath.Join("archive", now.UTC().Format("2006-01-02")+"-"+change)
	destination, err := platform.SafeJoin(root, destinationRel)
	if err != nil {
		return Report{}, err
	}
	if _, err := os.Lstat(destination); err == nil {
		return Report{}, fmt.Errorf("archive destino ya existe: %s", destination)
	} else if !os.IsNotExist(err) {
		return Report{}, err
	}
	originalMetadata, err := os.ReadFile(metadataPath)
	if err != nil {
		return Report{}, err
	}
	metadata.Status = "archived"
	metadataBody, err := yaml.Marshal(metadata)
	if err != nil {
		return Report{}, err
	}
	if err := atomicWrite(metadataPath, metadataBody, 0o644); err != nil {
		return Report{}, err
	}
	if err := os.Rename(changeRoot, destination); err != nil {
		_ = atomicWrite(metadataPath, originalMetadata, 0o644)
		return Report{}, err
	}
	report.Actions = []Action{{Kind: "archive", Path: filepath.ToSlash(destinationRel)}}
	report.Status = "archived"
	return report, nil
}

func verificationEvidence(root, change string) (bool, []Diagnostic, error) {
	verificationRoot, err := platform.SafeJoin(root, filepath.Join("verification", change))
	if err != nil {
		return false, nil, err
	}
	info, err := os.Lstat(verificationRoot)
	if os.IsNotExist(err) {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return false, []Diagnostic{{Level: LevelError, Code: "unsafe_verification", Path: filepath.ToSlash(filepath.Join("verification", change)), Message: "verification no es un directorio regular"}}, nil
	}
	found := false
	diagnostics := []Diagnostic{}
	err = filepath.WalkDir(verificationRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == verificationRoot {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return infoErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			diagnostics = append(diagnostics, Diagnostic{Level: LevelError, Code: "unsafe_verification", Path: filepath.ToSlash(rel), Message: "symlink no permitido en evidencia"})
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Mode().IsRegular() {
			found = true
		}
		return nil
	})
	return found, diagnostics, err
}
