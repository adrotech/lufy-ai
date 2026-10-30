package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEvidenceInspectAndRetiredResultCommand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "checks.json")
	if err := os.WriteFile(path, []byte(`{"checks":[{"name":"factual","result":"passed","evidence":"fuente contrastada"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := Run([]string{"evidence", "inspect", "--file", path}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitOK || !strings.Contains(out.String(), `"conclusion":"supported"`) {
		t.Fatalf("inspect: %d %s %s", code, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()
	if code := Run([]string{"result", "validate", "--file", path}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitUsageErr {
		t.Fatalf("retired command still exists: %d", code)
	}
}
