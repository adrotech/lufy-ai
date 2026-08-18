package lufysdd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufypaths"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

var safeID = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Service struct {
	now   func() time.Time
	write func(string, []byte, os.FileMode) error
}

func NewService() Service {
	return Service{now: time.Now, write: atomicWrite}
}

func (s Service) New(target, change, capability string) (Report, error) {
	return s.NewWithMode(target, change, capability, ModeFull)
}

func (s Service) NewWithMode(target, change, capability string, mode Mode) (Report, error) {
	root, err := workflowRoot(target)
	if err != nil {
		return Report{}, err
	}
	if err := validateID("change", change); err != nil {
		return Report{}, err
	}
	if mode == "" {
		mode = ModeFull
	}
	if mode != ModeFull && mode != ModeLite {
		return Report{}, fmt.Errorf("mode inválido %q; usa full o lite", mode)
	}
	if capability == "" {
		capability = change
	}
	if err := validateID("capability", capability); err != nil {
		return Report{}, err
	}
	if err := requireDirectory(root); err != nil {
		return Report{}, fmt.Errorf("workflow Lufy SDD no disponible: %w", err)
	}
	changesRoot, err := platform.SafeJoin(root, "changes")
	if err != nil {
		return Report{}, err
	}
	if err := requireDirectory(changesRoot); err != nil {
		return Report{}, fmt.Errorf("directorio changes inválido: %w", err)
	}
	changeRoot, err := platform.SafeJoin(changesRoot, change)
	if err != nil {
		return Report{}, err
	}
	if err := os.Mkdir(changeRoot, 0o755); err != nil {
		if os.IsExist(err) {
			return Report{}, fmt.Errorf("change %q ya existe; no se sobrescribió ningún artifact", change)
		}
		return Report{}, err
	}
	created := true
	defer func() {
		if !created {
			_ = os.RemoveAll(changeRoot)
		}
	}()

	type scaffoldFile struct {
		rel  string
		body string
	}
	files := []scaffoldFile{
		{rel: "change.yaml", body: fmt.Sprintf("schemaVersion: 1\nid: %s\nmode: %s\nstatus: proposed\nsyncedDigest: \"\"\n", change, mode)},
		{rel: "proposal.md", body: fmt.Sprintf("# Proposal: %s\n\n## Why\n\nDescribe el problema y el outcome esperado.\n\n## What Changes\n\n- Describe el alcance.\n\n## Non-Goals\n\n- Describe lo que queda fuera.\n", change)},
	}
	if mode == ModeFull {
		files = append(files, scaffoldFile{rel: "design.md", body: fmt.Sprintf("# Design: %s\n\n## Context\n\nDescribe el contexto técnico.\n\n## Decisions\n\nDescribe las decisiones y tradeoffs.\n\n## Risks\n\nDescribe riesgos y mitigaciones.\n", change)})
	}
	files = append(files, scaffoldFile{rel: "tasks.md", body: fmt.Sprintf("# Tasks: %s\n\n- [ ] Implementar el cambio.\n- [ ] Ejecutar validación proporcional.\n- [ ] Registrar evidencia en `.lufy/workflows/sdd/verification/%s/`.\n", change, change)})
	if mode == ModeFull {
		files = append(files, scaffoldFile{rel: filepath.Join("specs", capability, "spec.md"), body: fmt.Sprintf("# %s Specification Delta\n\n## ADDED Requirements\n\n### Requirement: Describe the required behavior\n\nEl sistema SHALL describir un comportamiento verificable.\n\n#### Scenario: Describe the expected outcome\n\n- **WHEN** ocurre una condición concreta\n- **THEN** ocurre un resultado observable\n", capability)})
	}
	for _, file := range files {
		rel, body := file.rel, file.body
		path, joinErr := platform.SafeJoin(changeRoot, rel)
		if joinErr != nil {
			created = false
			return Report{}, joinErr
		}
		if mkdirErr := os.MkdirAll(filepath.Dir(path), 0o755); mkdirErr != nil {
			created = false
			return Report{}, mkdirErr
		}
		if writeErr := writeExclusive(path, []byte(body)); writeErr != nil {
			created = false
			return Report{}, writeErr
		}
	}
	if err := s.renderOverview(root, change, mode); err != nil {
		created = false
		return Report{}, err
	}
	return Report{Schema: ReportSchema, Action: "new", Change: change, Mode: mode, Status: "proposed", Root: root, Progress: Progress{}, Diagnostics: []Diagnostic{}}, nil
}

func (s Service) Validate(target, change string, strict bool) (Report, error) {
	report, err := s.Check(target, change, strict)
	if err != nil {
		return Report{}, err
	}
	if err := s.renderOverview(report.Root, change, report.Mode); err != nil {
		return Report{}, err
	}
	return report, nil
}

// Check validates a change without materializing derived artifacts. It is used
// by read-only installation verification; the sdd validate command uses
// Validate so normal authoring still refreshes the integrated overview.
func (Service) Check(target, change string, strict bool) (Report, error) {
	report, err := inspect(target, change, "validate", strict)
	if err != nil {
		return Report{}, err
	}
	if report.Valid() {
		report.Status = "valid"
	} else {
		report.Status = "invalid"
	}
	return report, nil
}

func (Service) Status(target, change string) (Report, error) {
	report, err := inspect(target, change, "status", false)
	if err != nil {
		return Report{}, err
	}
	switch {
	case !report.Valid():
		report.Status = "blocked"
	case report.Progress.Total == 0:
		report.Status = "proposed"
	case report.Progress.Complete < report.Progress.Total:
		report.Status = "in_progress"
	default:
		changeRoot, joinErr := platform.SafeJoin(report.Root, filepath.Join("changes", change))
		if joinErr != nil {
			return Report{}, joinErr
		}
		if report.Mode == ModeFull {
			metadataPath, joinErr := platform.SafeJoin(changeRoot, "change.yaml")
			if joinErr != nil {
				return Report{}, joinErr
			}
			metadata, metadataErr := readMetadata(metadataPath)
			if metadataErr != nil || metadata.SyncedDigest == "" || metadata.SyncedDigest != report.DeltaDigest {
				report.Status = "sync_pending"
				break
			}
		}
		evidence, diagnostics, evidenceErr := verificationEvidence(report.Root, change)
		if evidenceErr != nil {
			return Report{}, evidenceErr
		}
		report.Diagnostics = append(report.Diagnostics, diagnostics...)
		if evidence && report.Valid() {
			report.Status = "completed"
		} else {
			report.Status = "verification"
		}
	}
	return report, nil
}

func inspect(target, change, action string, strict bool) (Report, error) {
	root, err := workflowRoot(target)
	if err != nil {
		return Report{}, err
	}
	if err := validateID("change", change); err != nil {
		return Report{}, err
	}
	changeRoot, err := platform.SafeJoin(root, filepath.Join("changes", change))
	if err != nil {
		return Report{}, err
	}
	if err := requireDirectory(changeRoot); err != nil {
		return Report{}, fmt.Errorf("change %q no disponible: %w", change, err)
	}
	report := Report{Schema: ReportSchema, Action: action, Change: change, Mode: ModeFull, Status: "unknown", Root: root, Diagnostics: []Diagnostic{}}
	metadataBody, metadataReadErr := readRegular(changeRoot, "change.yaml")
	if metadataReadErr != nil {
		report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "missing_artifact", Path: filepath.ToSlash(filepath.Join("changes", change, "change.yaml")), Message: metadataReadErr.Error()})
	} else if strings.TrimSpace(string(metadataBody)) == "" {
		report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "empty_artifact", Path: filepath.ToSlash(filepath.Join("changes", change, "change.yaml")), Message: "artifact vacío"})
	} else {
		metadataPath, metadataPathErr := platform.SafeJoin(changeRoot, "change.yaml")
		if metadataPathErr != nil {
			return Report{}, metadataPathErr
		}
		metadata, metadataErr := readMetadata(metadataPath)
		if metadataErr != nil {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "invalid_metadata", Path: filepath.ToSlash(filepath.Join("changes", change, "change.yaml")), Message: metadataErr.Error()})
		} else {
			report.Mode = metadata.Mode
			if metadata.ID != change {
				report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "metadata_id_mismatch", Path: filepath.ToSlash(filepath.Join("changes", change, "change.yaml")), Message: "metadata.id no coincide con el directorio del change"})
			}
		}
	}
	required := []string{"proposal.md", "tasks.md"}
	if report.Mode == ModeFull {
		required = []string{"proposal.md", "design.md", "tasks.md"}
	}
	for _, rel := range required {
		body, readErr := readRegular(changeRoot, rel)
		if readErr != nil {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "missing_artifact", Path: filepath.ToSlash(filepath.Join("changes", change, rel)), Message: readErr.Error()})
			continue
		}
		if strings.TrimSpace(string(body)) == "" {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "empty_artifact", Path: filepath.ToSlash(filepath.Join("changes", change, rel)), Message: "artifact vacío"})
		}
		if rel == "tasks.md" {
			report.Progress = TaskProgress(string(body))
			if strict && report.Progress.Total == 0 {
				report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "missing_tasks", Path: filepath.ToSlash(filepath.Join("changes", change, rel)), Message: "tasks.md no contiene checkboxes"})
			}
		}
	}
	if report.Mode == ModeLite {
		sortDiagnostics(report.Diagnostics)
		return report, nil
	}

	specRoot, joinErr := platform.SafeJoin(changeRoot, "specs")
	if joinErr != nil {
		return Report{}, joinErr
	}
	if directoryErr := requireDirectory(specRoot); directoryErr != nil {
		report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "unsafe_specs", Path: filepath.ToSlash(filepath.Join("changes", change, "specs")), Message: directoryErr.Error()})
		sortDiagnostics(report.Diagnostics)
		return report, nil
	}
	entries, readErr := os.ReadDir(specRoot)
	if readErr != nil {
		report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "missing_specs", Path: filepath.ToSlash(filepath.Join("changes", change, "specs")), Message: readErr.Error()})
	} else {
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
		found := 0
		digest := sha256.New()
		for _, entry := range entries {
			capability := entry.Name()
			if !entry.IsDir() {
				report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "invalid_spec_entry", Path: filepath.ToSlash(filepath.Join("changes", change, "specs", capability)), Message: "se esperaba un directorio de capability"})
				continue
			}
			if err := validateID("capability", capability); err != nil {
				report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "unsafe_capability", Path: filepath.ToSlash(filepath.Join("changes", change, "specs", capability)), Message: err.Error()})
				continue
			}
			rel := filepath.Join("specs", capability, "spec.md")
			body, specErr := readRegular(changeRoot, rel)
			path := filepath.ToSlash(filepath.Join("changes", change, rel))
			if specErr != nil {
				report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "missing_spec", Path: path, Message: specErr.Error()})
				continue
			}
			found++
			_, _ = digest.Write([]byte(capability))
			_, _ = digest.Write([]byte{0})
			_, _ = digest.Write(body)
			_, _ = digest.Write([]byte{0})
			_, diagnostics := ParseDelta(path, capability, string(body))
			report.Diagnostics = append(report.Diagnostics, diagnostics...)
		}
		if found > 0 {
			report.DeltaDigest = hex.EncodeToString(digest.Sum(nil))
		}
		if strict && found == 0 {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{Level: LevelError, Code: "missing_specs", Path: filepath.ToSlash(filepath.Join("changes", change, "specs")), Message: "el mode full requiere al menos una spec delta"})
		}
	}
	sortDiagnostics(report.Diagnostics)
	return report, nil
}

func workflowRoot(target string) (string, error) {
	resolved, err := platform.ResolveTargetPath(target)
	if err != nil {
		return "", err
	}
	return platform.SafeJoin(resolved, lufypaths.LufySDD)
}

func validateID(kind, value string) error {
	if !safeID.MatchString(value) {
		return fmt.Errorf("%s ID inválido %q; usa kebab-case seguro", kind, value)
	}
	return nil
}

func requireDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("%s no es un directorio regular", path)
	}
	return nil
}

func readRegular(root, rel string) ([]byte, error) {
	path, err := platform.SafeJoin(root, rel)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s no es un archivo regular", path)
	}
	return os.ReadFile(path)
}

func writeExclusive(path string, body []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err = file.Write(body); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
