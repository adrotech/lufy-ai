package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRetiredWorkflowCommandsDoNotWrite(t *testing.T) {
	for _, command := range [][]string{{"sdd", "new", "--change", "example"}, {"opsx", "render", "--change", "example"}, {"context", "coverage"}, {"context", "trace", "example"}} {
		target := t.TempDir()
		var out, errOut bytes.Buffer
		args := append(append([]string{}, command...), "--target", target)
		if code := Run(args, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitUsageErr {
			t.Fatalf("comando retirado aceptado: %v: %d %s", command, code, errOut.String())
		}
		entries, err := os.ReadDir(target)
		if err != nil || len(entries) != 0 {
			t.Fatalf("comando retirado escribió: %v %v", entries, err)
		}
	}
}

func TestFreshInstallationHasNoMethodologyMetadata(t *testing.T) {
	target := t.TempDir()
	var out, errOut bytes.Buffer
	if code := Run([]string{"install", "--target", target, "--yes"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitOK {
		t.Fatalf("install: %d %s", code, errOut.String())
	}
	body, err := os.ReadFile(filepath.Join(target, ".lufy/managed-state/install-state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(body, []byte("methodology")) {
		t.Fatal("manifest conserva metadatos de metodología")
	}
}

func TestVersionReportsCurrentRuntimeContract(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := Run([]string{"version", "--runtime-contract"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitOK {
		t.Fatalf("runtime contract: %d %s", code, errOut.String())
	}
	if out.String() != "# lufy-runtime: execution=inline config=2 state=4 backup=2\n" {
		t.Fatalf("contrato inesperado: %q", out.String())
	}
}
