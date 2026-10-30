package backup

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufypaths"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

func TestPublicRestoreRejectsPreviousGenerationBeforeWriting(t *testing.T) {
	for _, field := range []string{"schemaVersion", "runtimeStateSchema"} {
		t.Run(field, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, filepath.Join(root, "note.md"), "before")
			dir, err := BackupFiles(root, []string{"note.md"}, "test", io.Discard)
			if err != nil {
				t.Fatal(err)
			}
			manifestPath := filepath.Join(dir, "manifest.json")
			data, err := os.ReadFile(manifestPath)
			if err != nil {
				t.Fatal(err)
			}
			var manifest map[string]any
			if err := json.Unmarshal(data, &manifest); err != nil {
				t.Fatal(err)
			}
			manifest[field] = float64(1)
			data, _ = json.Marshal(manifest)
			if err := os.WriteFile(manifestPath, data, 0600); err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(root, "note.md"), "keep")
			if err := NewService().Restore(root, dir, false, true, io.Discard); err == nil {
				t.Fatal("backup de generación retirada aceptado")
			}
			if _, err := RestoreCapturedFiles(root, dir, io.Discard); err == nil {
				t.Fatal("rollback aceptó generación retirada")
			}
			if got := readFile(t, filepath.Join(root, "note.md")); got != "keep" {
				t.Fatal("restore incompatible modificó el target")
			}
			var out strings.Builder
			if err := NewService().List(root, &out); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(out.String(), filepath.Base(dir)) {
				t.Fatal("backup incompatible listado como disponible")
			}
		})
	}
}

func TestRestoreValidatesPayloadAndCanRepairDamagedCurrentState(t *testing.T) {
	root := t.TempDir()
	stateWithFiles(t, root, nil)
	good, err := BackupFiles(root, []string{lufypaths.InstallState}, "good", io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, state.Path(root), "damaged")
	damaged, err := BackupFiles(root, []string{lufypaths.InstallState}, "preimage", io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if err := NewService().Restore(root, good, false, true, io.Discard); err != nil {
		t.Fatal(err)
	}
	if _, err := state.Load(root); err != nil {
		t.Fatalf("no reparó estado: %v", err)
	}
	if err := NewService().Restore(root, damaged, false, true, io.Discard); err == nil {
		t.Fatal("restore público aceptó payload dañado")
	}
	if _, err := state.Load(root); err != nil {
		t.Fatalf("rechazo alteró estado: %v", err)
	}
	// Un rollback transaccional restaura exactamente el preimage, incluso si estaba dañado.
	if _, err := RestoreCapturedFiles(root, damaged, io.Discard); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, state.Path(root)); got != "damaged" {
		t.Fatal("rollback no preservó preimage")
	}
}

func TestRestoreRejectsRetiredConfigPayloadBeforeWriting(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, lufypaths.ProjectConfig), "schema_version: 1\n")
	dir, err := BackupFiles(root, []string{lufypaths.ProjectConfig}, "test", io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, lufypaths.ProjectConfig), "keep")
	if err := NewService().Restore(root, dir, false, true, io.Discard); err == nil {
		t.Fatal("config retirada aceptada")
	}
	if got := readFile(t, filepath.Join(root, lufypaths.ProjectConfig)); got != "keep" {
		t.Fatal("config alterada")
	}
}

func TestRetentionPreservesUnknownAndPreviousGenerationDirectories(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "unknown", "keep"), "custom")
	writeManifest(t, filepath.Join(root, "previous", "manifest.json"), Manifest{SchemaVersion: 1})
	for _, name := range []string{"001", "002", "003"} {
		writeManifest(t, filepath.Join(root, name, "manifest.json"), Manifest{SchemaVersion: SchemaVersion, RuntimeStateSchema: state.SchemaVersion})
	}
	if err := pruneBackups(root, filepath.Join(root, "003"), 2); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "001")); !os.IsNotExist(err) {
		t.Fatal("no aplicó retención actual")
	}
	for _, name := range []string{"unknown/keep", "previous/manifest.json", "002/manifest.json", "003/manifest.json"} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Fatal(err)
		}
	}
}
