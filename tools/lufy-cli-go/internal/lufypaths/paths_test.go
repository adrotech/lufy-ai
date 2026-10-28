package lufypaths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupIDsNeverResolveRetiredRoots(t *testing.T) {
	root := t.TempDir()
	old := filepath.Join(root, ".lufy-ai/backups/old")
	if err := os.MkdirAll(old, 0755); err != nil {
		t.Fatal(err)
	}
	path, err := ResolveBackupReference(root, "old")
	if err != nil || path != filepath.Join(root, Backups, "old") {
		t.Fatalf("ruta retirada seleccionada: %s %v", path, err)
	}
}
func TestCanonicalPathsRejectEscape(t *testing.T) {
	root := t.TempDir()
	if _, err := WritePath(root, "../outside"); err == nil {
		t.Fatal("escape aceptado")
	}
	if _, err := ResolveBackupReference(root, "../../../../outside"); err == nil {
		t.Fatal("escape de backup aceptado")
	}
}
