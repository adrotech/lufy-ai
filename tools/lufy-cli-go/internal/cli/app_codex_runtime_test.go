package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDoctorCodexRuntimeOptInJSONOmitsRuntimeWithoutFlag(t *testing.T) {
	target := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := Run([]string{"doctor", "--target", target, "--json"}, Dependencies{Stdout: &stdout, Stderr: &stderr})
	// El target deliberadamente no tiene manifest; eso no debe activar probes.
	if code != ExitRuntimeErr {
		t.Fatalf("expected missing-manifest runtime error, got %d", code)
	}
	var report map[string]json.RawMessage
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("doctor output is not one JSON object: %v", err)
	}
	if _, exists := report["codex_runtime"]; exists {
		t.Fatal("doctor without explicit opt-in must preserve JSON shape")
	}
	for _, key := range []string{"ok", "targetRoot", "checks"} {
		if _, exists := report[key]; !exists {
			t.Fatalf("legacy doctor field %s missing", key)
		}
	}
}

func TestDoctorCodexRuntimeMissingLiteralExecutableDoesNotInvokeShell(t *testing.T) {
	target := t.TempDir()
	sentinel := filepath.Join(target, "must-not-exist")
	// Una cadena con sintaxis shell debe permanecer un nombre literal inexistente.
	selected := filepath.Join(target, "missing codex") + "; touch " + sentinel
	var stdout, stderr bytes.Buffer
	code := Run([]string{"doctor", "--target", target, "--codex", selected, "--json"}, Dependencies{Stdout: &stdout, Stderr: &stderr})
	if code != ExitRuntimeErr {
		t.Fatalf("expected incomplete target error, got %d", code)
	}
	var report struct {
		Runtime struct {
			Schema    string   `json:"schema_version"`
			Status    string   `json:"status"`
			Selection string   `json:"selection"`
			Native    string   `json:"native_delegation"`
			Reasons   []string `json:"reasons"`
		} `json:"codex_runtime"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Runtime.Schema != "codex-runtime-report/v1" || report.Runtime.Status != "unavailable" || report.Runtime.Selection != "explicit" || report.Runtime.Native != "unknown" || len(report.Runtime.Reasons) == 0 {
		t.Fatalf("missing executable must return structured evidence: %#v", report.Runtime)
	}
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Fatal("literal executable selection must not run shell syntax")
	}
}
