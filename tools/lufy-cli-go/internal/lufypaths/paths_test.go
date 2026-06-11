package lufypaths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveExistingPrefersCanonicalWhenBothExist(t *testing.T) {
	target := t.TempDir()
	write(t, filepath.Join(target, ProjectConfig), "new\n")
	write(t, filepath.Join(target, LegacyProjectConfig), "legacy\n")

	resolved, err := ResolveExisting(target, ProjectConfig, LegacyProjectConfig)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Rel != ProjectConfig || resolved.Legacy || !resolved.Exists {
		t.Fatalf("resolved = %#v", resolved)
	}
}

func TestResolveExistingFallsBackToLegacyWhenOnlyLegacyExists(t *testing.T) {
	target := t.TempDir()
	write(t, filepath.Join(target, LegacyProjectConfig), "legacy\n")

	resolved, err := ResolveExisting(target, ProjectConfig, LegacyProjectConfig)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Rel != LegacyProjectConfig || !resolved.Legacy || !resolved.Exists {
		t.Fatalf("resolved = %#v", resolved)
	}
}

func TestResolveExistingReturnsCanonicalWritePathWhenMissing(t *testing.T) {
	target := t.TempDir()

	resolved, err := ResolveExisting(target, InstallState, LegacyInstallState)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Rel != InstallState || resolved.Legacy || resolved.Exists {
		t.Fatalf("resolved = %#v", resolved)
	}
	want := filepath.Join(target, InstallState)
	if resolved.Path != want {
		t.Fatalf("path = %q, want %q", resolved.Path, want)
	}
}

func TestWritePathAlwaysUsesProvidedCanonicalPath(t *testing.T) {
	target := t.TempDir()
	write(t, filepath.Join(target, LegacyInstallState), "{}\n")

	path, err := WritePath(target, InstallState)
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(target, InstallState) {
		t.Fatalf("path = %q", path)
	}
}

func TestExistingPathsReturnsCanonicalAndLegacy(t *testing.T) {
	target := t.TempDir()
	write(t, filepath.Join(target, Backups, "new", "manifest.json"), "{}\n")
	write(t, filepath.Join(target, LegacyBackups, "old", "manifest.json"), "{}\n")

	paths, err := ExistingPaths(target, Backups, LegacyBackups)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 || paths[0].Rel != Backups || paths[0].Legacy || paths[1].Rel != LegacyBackups || !paths[1].Legacy {
		t.Fatalf("paths = %#v", paths)
	}
}

func TestResolveBackupReferenceFindsCanonicalLegacyAndFallback(t *testing.T) {
	target := t.TempDir()
	write(t, filepath.Join(target, Backups, "new", "manifest.json"), "{}\n")
	write(t, filepath.Join(target, LegacyBackups, "old", "manifest.json"), "{}\n")

	canonical, err := ResolveBackupReference(target, "new")
	if err != nil {
		t.Fatal(err)
	}
	if canonical != filepath.Join(target, Backups, "new") {
		t.Fatalf("canonical = %q", canonical)
	}
	legacy, err := ResolveBackupReference(target, "old")
	if err != nil {
		t.Fatal(err)
	}
	if legacy != filepath.Join(target, LegacyBackups, "old") {
		t.Fatalf("legacy = %q", legacy)
	}
	missing, err := ResolveBackupReference(target, "missing")
	if err != nil {
		t.Fatal(err)
	}
	if missing != filepath.Join(target, Backups, "missing") {
		t.Fatalf("missing fallback = %q", missing)
	}
}

func TestPathHelpersRejectUnsafeTargetEscape(t *testing.T) {
	target := t.TempDir()
	if _, err := WritePath(target, "../outside"); err == nil {
		t.Fatal("WritePath expected safe join error")
	}
	if _, err := ResolveExisting(target, "../outside", LegacyInstallState); err == nil {
		t.Fatal("ResolveExisting expected safe join error")
	}
	if _, err := ExistingPaths(target, InstallState, "../outside"); err == nil {
		t.Fatal("ExistingPaths expected safe join error")
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
