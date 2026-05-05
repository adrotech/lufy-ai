package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunInstallDryRun(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"install", "--target", ".", "--dry-run", "--yes", "--no-engram"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("expected ExitOK, got %d, stderr=%s", code, errOut.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("Modo dry-run")) {
		t.Fatalf("expected dry-run message, got: %s", out.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"nope"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitUsageErr {
		t.Fatalf("expected ExitUsageErr, got %d", code)
	}
}

func TestRunInstallUnknownFlag(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"install", "--unknown"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitUsageErr {
		t.Fatalf("expected ExitUsageErr, got %d", code)
	}
}

func TestRunHelpCommandsAndRestoreRequiresBackup(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	if code := Run([]string{"help"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitOK {
		t.Fatalf("help expected ExitOK, got %d", code)
	}
	if !bytes.Contains(out.Bytes(), []byte("install")) || !bytes.Contains(out.Bytes(), []byte("restore")) {
		t.Fatalf("help output missing commands: %s", out.String())
	}

	out.Reset()
	errOut.Reset()
	if code := Run([]string{"restore", "--target", t.TempDir(), "--dry-run"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitUsageErr {
		t.Fatalf("restore without --backup expected ExitUsageErr, got %d", code)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("restore requiere --backup")) {
		t.Fatalf("restore missing backup output unexpected: %s", errOut.String())
	}
}

func TestRunRestoreDryRunParsesRequiredBackup(t *testing.T) {
	target := t.TempDir()
	backupDir := filepath.Join(t.TempDir(), "backup")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := []byte(`{"schemaVersion":1,"toolVersion":"dev","targetRoot":"` + target + `","files":[]}`)
	if err := os.WriteFile(filepath.Join(backupDir, "manifest.json"), manifest, 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"restore", "--target", target, "--backup", backupDir, "--dry-run", "--yes"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("restore dry-run expected ExitOK, got %d stderr=%s", code, errOut.String())
	}
}
