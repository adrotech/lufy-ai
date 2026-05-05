package cli

import (
	"bytes"
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
