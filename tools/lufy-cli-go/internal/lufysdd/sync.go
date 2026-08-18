package lufysdd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"gopkg.in/yaml.v3"
)

type changeMetadata struct {
	SchemaVersion int    `yaml:"schemaVersion"`
	ID            string `yaml:"id"`
	Mode          Mode   `yaml:"mode,omitempty"`
	Status        string `yaml:"status"`
	SyncedDigest  string `yaml:"syncedDigest"`
}

type plannedSpec struct {
	Capability string
	Path       string
	Body       []byte
	Actions    []Action
}

type snapshot struct {
	Path   string
	Body   []byte
	Mode   os.FileMode
	Exists bool
}

func (s Service) Sync(target, change string) (Report, error) {
	report, err := inspect(target, change, "sync", true)
	if err != nil {
		return Report{}, err
	}
	if !report.Valid() {
		report.Status = "blocked"
		return report, nil
	}
	if err := s.renderOverview(report.Root, change, report.Mode); err != nil {
		return Report{}, err
	}
	if report.Mode == ModeLite {
		report.Status = "not_applicable"
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
	if metadata.SyncedDigest == report.DeltaDigest {
		report.Status = "synced"
		return report, nil
	}
	deltas, err := loadDeltas(changeRoot, change)
	if err != nil {
		return Report{}, err
	}
	plans, diagnostics, err := planSync(root, deltas)
	if err != nil {
		return Report{}, err
	}
	if len(diagnostics) > 0 {
		report.Diagnostics = append(report.Diagnostics, diagnostics...)
		sortDiagnostics(report.Diagnostics)
		report.Status = "blocked"
		return report, nil
	}
	metadata.Status = "verification"
	metadata.SyncedDigest = report.DeltaDigest
	metadataBody, err := yaml.Marshal(metadata)
	if err != nil {
		return Report{}, err
	}

	snapshots := []snapshot{}
	write := s.write
	if write == nil {
		write = atomicWrite
	}
	backupRoot, err := s.createSyncBackup(root, change, report.DeltaDigest, append(plans, plannedSpec{Path: metadataPath}))
	if err != nil {
		return Report{}, err
	}
	_ = backupRoot
	for _, plan := range plans {
		before, snapErr := takeSnapshot(plan.Path)
		if snapErr != nil {
			rollback(snapshots)
			return Report{}, snapErr
		}
		snapshots = append(snapshots, before)
		if writeErr := write(plan.Path, plan.Body, 0o644); writeErr != nil {
			rollback(snapshots)
			return Report{}, fmt.Errorf("sync falló escribiendo %s: %w", plan.Path, writeErr)
		}
		report.Actions = append(report.Actions, plan.Actions...)
	}
	metadataSnapshot, err := takeSnapshot(metadataPath)
	if err != nil {
		rollback(snapshots)
		return Report{}, err
	}
	snapshots = append(snapshots, metadataSnapshot)
	if err := write(metadataPath, metadataBody, 0o644); err != nil {
		rollback(snapshots)
		return Report{}, fmt.Errorf("sync aplicó rollback porque no pudo persistir metadata: %w", err)
	}
	report.Status = "synced"
	return report, nil
}

func loadDeltas(changeRoot, change string) ([]Delta, error) {
	specRoot, err := platform.SafeJoin(changeRoot, "specs")
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(specRoot)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	deltas := []Delta{}
	for _, entry := range entries {
		if !entry.IsDir() || !safeID.MatchString(entry.Name()) {
			continue
		}
		rel := filepath.Join("specs", entry.Name(), "spec.md")
		body, err := readRegular(changeRoot, rel)
		if err != nil {
			return nil, err
		}
		path := filepath.ToSlash(filepath.Join("changes", change, rel))
		delta, diagnostics := ParseDelta(path, entry.Name(), string(body))
		if len(diagnostics) > 0 {
			return nil, fmt.Errorf("delta inválido después de preflight: %s", path)
		}
		deltas = append(deltas, delta)
	}
	return deltas, nil
}

func planSync(root string, deltas []Delta) ([]plannedSpec, []Diagnostic, error) {
	plans := []plannedSpec{}
	diagnostics := []Diagnostic{}
	for _, delta := range deltas {
		targetRel := filepath.Join("specs", delta.Capability, "spec.md")
		target, err := platform.SafeJoin(root, targetRel)
		if err != nil {
			return nil, nil, err
		}
		content := "# " + delta.Capability + " Specification\n"
		if info, statErr := os.Lstat(target); statErr == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
				diagnostics = append(diagnostics, Diagnostic{Level: LevelError, Code: "unsafe_spec_target", Path: filepath.ToSlash(targetRel), Message: "spec destino no es un archivo regular"})
				continue
			}
			body, readErr := os.ReadFile(target)
			if readErr != nil {
				return nil, nil, readErr
			}
			content = strings.ReplaceAll(string(body), "\r\n", "\n")
		} else if !os.IsNotExist(statErr) {
			return nil, nil, statErr
		}
		actions := []Action{}
		for _, requirement := range delta.Requirements {
			updated, action, diagnostic := applyRequirement(content, filepath.ToSlash(targetRel), requirement)
			if diagnostic != nil {
				diagnostics = append(diagnostics, *diagnostic)
				continue
			}
			content = updated
			actions = append(actions, action)
		}
		plans = append(plans, plannedSpec{Capability: delta.Capability, Path: target, Body: []byte(strings.TrimRight(content, "\n") + "\n"), Actions: actions})
	}
	sortDiagnostics(diagnostics)
	sort.Slice(plans, func(i, j int) bool { return plans[i].Path < plans[j].Path })
	return plans, diagnostics, nil
}

func applyRequirement(content, path string, requirement Requirement) (string, Action, *Diagnostic) {
	blocks := requirementBlocks(content)
	matches := []requirementBlock{}
	for _, block := range blocks {
		if strings.EqualFold(block.Title, requirement.Title) {
			matches = append(matches, block)
		}
	}
	want := 0
	if requirement.Marker == "MODIFIED" || requirement.Marker == "REMOVED" {
		want = 1
	}
	if len(matches) != want {
		message := fmt.Sprintf("%s requiere %d coincidencia(s), encontró %d", requirement.Marker, want, len(matches))
		return content, Action{}, &Diagnostic{Level: LevelError, Code: "ambiguous_requirement", Path: path, Line: requirement.Line, Message: message}
	}
	action := Action{Path: path, Requirement: requirement.Title}
	switch requirement.Marker {
	case "ADDED":
		action.Kind = "create"
		separator := "\n\n"
		if strings.TrimSpace(content) == "" {
			separator = ""
		}
		return strings.TrimRight(content, "\n") + separator + strings.TrimSpace(requirement.Body) + "\n", action, nil
	case "MODIFIED":
		action.Kind = "modify"
		block := matches[0]
		return content[:block.Start] + strings.TrimRight(requirement.Body, "\n") + "\n" + content[block.End:], action, nil
	case "REMOVED":
		action.Kind = "remove"
		block := matches[0]
		return content[:block.Start] + content[block.End:], action, nil
	default:
		return content, Action{}, &Diagnostic{Level: LevelError, Code: "unsupported_marker", Path: path, Line: requirement.Line, Message: "marker delta no soportado"}
	}
}

func readMetadata(path string) (changeMetadata, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return changeMetadata{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return changeMetadata{}, fmt.Errorf("metadata no regular: %s", path)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return changeMetadata{}, err
	}
	metadata := changeMetadata{}
	if err := yaml.Unmarshal(body, &metadata); err != nil {
		return changeMetadata{}, fmt.Errorf("metadata inválida: %w", err)
	}
	if metadata.SchemaVersion != 1 || metadata.ID == "" {
		return changeMetadata{}, fmt.Errorf("metadata Lufy SDD incompleta en %s", path)
	}
	if metadata.Mode == "" {
		metadata.Mode = ModeFull
	}
	if metadata.Mode != ModeFull && metadata.Mode != ModeLite {
		return changeMetadata{}, fmt.Errorf("metadata mode inválido %q en %s", metadata.Mode, path)
	}
	return metadata, nil
}

func (s Service) createSyncBackup(root, change, digest string, plans []plannedSpec) (string, error) {
	now := time.Now()
	if s.now != nil {
		now = s.now()
	}
	id := fmt.Sprintf("%s-%s-%d", change, shortDigest(digest), now.UTC().UnixNano())
	backupRoot, err := platform.SafeJoin(root, filepath.Join(".backups", id))
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(backupRoot, 0o700); err != nil {
		return "", err
	}
	for _, plan := range plans {
		info, statErr := os.Lstat(plan.Path)
		if os.IsNotExist(statErr) {
			continue
		}
		if statErr != nil {
			return "", statErr
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return "", fmt.Errorf("no se puede respaldar path no regular: %s", plan.Path)
		}
		body, readErr := os.ReadFile(plan.Path)
		if readErr != nil {
			return "", readErr
		}
		rel := filepath.Base(plan.Path)
		if strings.Contains(filepath.ToSlash(plan.Path), "/specs/") {
			rel = filepath.Join("specs", filepath.Base(filepath.Dir(plan.Path)), filepath.Base(plan.Path))
		}
		destination, joinErr := platform.SafeJoin(backupRoot, rel)
		if joinErr != nil {
			return "", joinErr
		}
		if mkdirErr := os.MkdirAll(filepath.Dir(destination), 0o700); mkdirErr != nil {
			return "", mkdirErr
		}
		if writeErr := writeExclusive(destination, body); writeErr != nil {
			return "", writeErr
		}
	}
	return backupRoot, nil
}

func shortDigest(digest string) string {
	if len(digest) > 12 {
		return digest[:12]
	}
	return digest
}

func takeSnapshot(path string) (snapshot, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return snapshot{Path: path}, nil
	}
	if err != nil {
		return snapshot{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return snapshot{}, fmt.Errorf("path no regular: %s", path)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return snapshot{}, err
	}
	return snapshot{Path: path, Body: body, Mode: info.Mode().Perm(), Exists: true}, nil
}

func rollback(snapshots []snapshot) {
	for i := len(snapshots) - 1; i >= 0; i-- {
		snapshot := snapshots[i]
		if snapshot.Exists {
			_ = atomicWrite(snapshot.Path, snapshot.Body, snapshot.Mode)
			continue
		}
		_ = os.Remove(snapshot.Path)
	}
}

func atomicWrite(path string, body []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".lufy-sdd-*")
	if err != nil {
		return err
	}
	temp := file.Name()
	defer os.Remove(temp)
	if err := file.Chmod(mode); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(body); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(temp, path)
}
