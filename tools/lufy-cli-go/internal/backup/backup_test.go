package backup

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

func TestBackupAndRestoreMultiassetCreatesPreRestoreRecovery(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, "AGENTS.md"), "agents original\n")
	writeFile(t, filepath.Join(target, ".opencode", "package.json"), "{\"name\":\"original\"}\n")
	stateWithFiles(t, target, []string{"AGENTS.md", filepath.Join(".opencode", "package.json")})

	var out bytes.Buffer
	backupDir, err := NewService().Run(Options{Target: target}, &out)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	manifestPath := filepath.Join(backupDir, "manifest.json")
	body, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Files) != 2 || manifest.Files[0].SHA256 == "" || manifest.Files[0].Size == 0 {
		t.Fatalf("manifest incompleto: %#v", manifest.Files)
	}
	if manifest.ToolVersion == "" || manifest.ToolCommit == "" || manifest.ToolBuildDate == "" {
		t.Fatalf("manifest missing runtime tool metadata: %#v", manifest)
	}

	writeFile(t, filepath.Join(target, "AGENTS.md"), "agents broken\n")
	out.Reset()
	if err := NewService().Restore(target, backupDir, false, true, &out); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}
	if got := readFile(t, filepath.Join(target, "AGENTS.md")); got != "agents original\n" {
		t.Fatalf("restore did not restore AGENTS.md: %q", got)
	}
	if !strings.Contains(out.String(), "Backup previo a restore:") {
		t.Fatalf("restore output missing recovery backup: %s", out.String())
	}
}

func TestRestoreDryRunDoesNotWrite(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, "AGENTS.md"), "original\n")
	stateWithFiles(t, target, []string{"AGENTS.md"})
	backupDir, err := NewService().Run(Options{Target: target}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(target, "AGENTS.md"), "changed\n")
	var out bytes.Buffer
	if err := NewService().Restore(target, backupDir, true, false, &out); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(target, "AGENTS.md")); got != "changed\n" {
		t.Fatalf("dry-run mutated file: %q", got)
	}
	if !strings.Contains(out.String(), "[dry-run] restauraría AGENTS.md") {
		t.Fatalf("dry-run output unexpected: %s", out.String())
	}
}

func TestRestoreRejectsTargetMismatchAndPathEscape(t *testing.T) {
	target := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "backup")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := Manifest{SchemaVersion: SchemaVersion, TargetRoot: t.TempDir(), Files: []FileItem{{Path: "AGENTS.md", BackupPath: "AGENTS.md", Status: "captured"}}}
	writeManifest(t, filepath.Join(backupDir, "manifest.json"), manifest)
	if err := NewService().Restore(target, backupDir, true, false, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "otro target") {
		t.Fatalf("expected target mismatch error, got %v", err)
	}

	manifest.TargetRoot = target
	manifest.Files[0].Path = "../evil"
	writeManifest(t, filepath.Join(backupDir, "manifest.json"), manifest)
	if err := NewService().Restore(target, backupDir, true, false, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "escapa") {
		t.Fatalf("expected path escape error, got %v", err)
	}

	manifest.Files[0].Path = "AGENTS.md"
	manifest.Files[0].BackupPath = "../evil"
	writeManifest(t, filepath.Join(backupDir, "manifest.json"), manifest)
	if err := NewService().Restore(target, backupDir, true, false, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "escapa") {
		t.Fatalf("expected backup path escape error, got %v", err)
	}
}

func TestRestorePartialFailureReportsRecoveryBackup(t *testing.T) {
	target := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "backup")
	writeFile(t, filepath.Join(target, "AGENTS.md"), "current agents\n")
	writeFile(t, filepath.Join(target, "blocked"), "not a directory\n")
	writeFile(t, filepath.Join(backupDir, "AGENTS.md"), "restored agents\n")
	writeFile(t, filepath.Join(backupDir, "blocked", "child.txt"), "restored child\n")
	agentsHash := hashFile(t, filepath.Join(backupDir, "AGENTS.md"))
	childHash := hashFile(t, filepath.Join(backupDir, "blocked", "child.txt"))
	manifest := Manifest{SchemaVersion: SchemaVersion, TargetRoot: target, Files: []FileItem{
		{Path: "AGENTS.md", BackupPath: "AGENTS.md", SHA256: agentsHash, Status: "captured"},
		{Path: filepath.Join("blocked", "child.txt"), BackupPath: filepath.Join("blocked", "child.txt"), SHA256: childHash, Status: "captured"},
	}}
	writeManifest(t, filepath.Join(backupDir, "manifest.json"), manifest)
	err := NewService().Restore(target, backupDir, false, true, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "backup de recovery") {
		t.Fatalf("expected recovery backup error, got %v", err)
	}
}

func TestRestoreRequiresYesForRealMutation(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, "AGENTS.md"), "original\n")
	stateWithFiles(t, target, []string{"AGENTS.md"})
	backupDir, err := NewService().Run(Options{Target: target}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(target, "AGENTS.md"), "changed\n")
	err = NewService().Restore(target, backupDir, false, false, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "requiere --yes") {
		t.Fatalf("expected --yes error, got %v", err)
	}
}

func TestRestoreRejectsCorruptBackupHash(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, "AGENTS.md"), "original\n")
	stateWithFiles(t, target, []string{"AGENTS.md"})
	backupDir, err := NewService().Run(Options{Target: target}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(backupDir, "AGENTS.md"), "tampered\n")
	err = NewService().Restore(target, backupDir, true, false, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "hash de backup no coincide") {
		t.Fatalf("expected hash mismatch error, got %v", err)
	}
}

func TestRestoreRejectsCorruptBackupBeforeRecoveryBackup(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, "AGENTS.md"), "original\n")
	stateWithFiles(t, target, []string{"AGENTS.md"})
	backupDir, err := NewService().Run(Options{Target: target}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(target, "AGENTS.md"), "changed\n")
	writeFile(t, filepath.Join(backupDir, "AGENTS.md"), "tampered\n")
	err = NewService().Restore(target, backupDir, false, true, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "hash de backup no coincide") {
		t.Fatalf("expected hash mismatch error, got %v", err)
	}
	backupsRoot := filepath.Join(target, ".lufy-ai", "backups")
	entries, err := os.ReadDir(backupsRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected no recovery backup after corrupt source, entries=%d", len(entries))
	}
}

func TestRestoreCapturedFilesRestoresWithoutRecoveryBackup(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, "AGENTS.md"), "original\n")
	stateWithFiles(t, target, []string{"AGENTS.md"})
	backupDir, err := NewService().Run(Options{Target: target}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(target, "AGENTS.md"), "changed\n")
	restored, err := RestoreCapturedFiles(target, backupDir, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("RestoreCapturedFiles() error = %v", err)
	}
	if restored != 1 {
		t.Fatalf("RestoreCapturedFiles() restored %d, want 1", restored)
	}
	if got := readFile(t, filepath.Join(target, "AGENTS.md")); got != "original\n" {
		t.Fatalf("RestoreCapturedFiles() did not restore file: %q", got)
	}
}

func TestListAndRestoreByBackupID(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, "AGENTS.md"), "original\n")
	stateWithFiles(t, target, []string{"AGENTS.md"})
	backupDir, err := NewService().Run(Options{Target: target}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	backupID := filepath.Base(backupDir)
	var out bytes.Buffer
	if err := NewService().List(target, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), backupID) || !strings.Contains(out.String(), "manifest.json") {
		t.Fatalf("backup list output unexpected: %s", out.String())
	}
	writeFile(t, filepath.Join(target, "AGENTS.md"), "changed\n")
	if err := NewService().Restore(target, backupID, false, true, &bytes.Buffer{}); err != nil {
		t.Fatalf("Restore(id) error = %v", err)
	}
	if got := readFile(t, filepath.Join(target, "AGENTS.md")); got != "original\n" {
		t.Fatalf("restore by id did not restore file: %q", got)
	}
}

func TestBackupPrunesOldBackups(t *testing.T) {
	target := t.TempDir()
	writeFile(t, filepath.Join(target, "AGENTS.md"), "original\n")
	stateWithFiles(t, target, []string{"AGENTS.md"})
	backupsRoot := filepath.Join(target, ".lufy-ai", "backups")
	for i := 0; i < defaultBackupRetention; i++ {
		path := filepath.Join(backupsRoot, "20000101T000000.00000000"+string(rune('0'+i))+"Z")
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	backupDir, err := NewService().Run(Options{Target: target}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	entries, err := os.ReadDir(backupsRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != defaultBackupRetention {
		t.Fatalf("backup retention kept %d entries, want %d", len(entries), defaultBackupRetention)
	}
	if _, err := os.Stat(backupDir); err != nil {
		t.Fatalf("current backup pruned: %v", err)
	}
}

func stateWithFiles(t *testing.T, target string, rels []string) {
	t.Helper()
	var states []state.AssetState
	for _, rel := range rels {
		hash, err := assets.FileSHA256(filepath.Join(target, rel))
		if err != nil {
			t.Fatal(err)
		}
		states = append(states, state.AssetState{ID: rel, SourceRel: rel, TargetRel: rel, SourceSHA256: hash, TargetSHA256: hash, LastAction: "copy"})
	}
	if err := state.WriteAtomic(target, state.New(target, nil, states, "test-fingerprint")); err != nil {
		t.Fatal(err)
	}
}

func writeManifest(t *testing.T, path string, manifest Manifest) {
	t.Helper()
	body, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func hashFile(t *testing.T, path string) string {
	t.Helper()
	hash, err := assets.FileSHA256(path)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}
