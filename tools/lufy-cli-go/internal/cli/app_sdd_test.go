package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/lufypaths"
)

func TestSDDNewStatusValidateLifecycle(t *testing.T) {
	target := t.TempDir()
	for _, rel := range []string{"", "changes", "specs", "decisions", "verification", "archive"} {
		if err := os.MkdirAll(filepath.Join(target, lufypaths.LufySDD, rel), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	var out, errOut bytes.Buffer
	deps := Dependencies{Stdout: &out, Stderr: &errOut}
	code := Run([]string{"sdd", "new", "--target", target, "--change", "add-audit-log", "--capability", "audit-log", "--json"}, deps)
	if code != ExitOK || !strings.Contains(out.String(), `"schema": "lufy-sdd-report/v1"`) {
		t.Fatalf("new code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()
	code = Run([]string{"sdd", "validate", "--target", target, "--change", "add-audit-log", "--strict", "--json"}, deps)
	if code != ExitOK || !strings.Contains(out.String(), `"status": "valid"`) {
		t.Fatalf("validate code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()
	code = Run([]string{"sdd", "sync", "--target", target, "--change", "add-audit-log", "--json"}, deps)
	if code != ExitOK || !strings.Contains(out.String(), `"status": "synced"`) {
		t.Fatalf("sync code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()
	code = Run([]string{"sdd", "status", "--target", target, "--change", "add-audit-log"}, deps)
	if code != ExitOK || !strings.Contains(out.String(), "tasks: 0/20") {
		t.Fatalf("status code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func TestSDDHelpAndUnsafeChange(t *testing.T) {
	var out, errOut bytes.Buffer
	deps := Dependencies{Stdout: &out, Stderr: &errOut}
	if code := Run([]string{"sdd", "--help"}, deps); code != ExitOK || !strings.Contains(out.String(), "lufy-ai sdd") {
		t.Fatalf("help code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()
	if code := Run([]string{"sdd", "new", "--change", "../escape"}, deps); code != ExitRuntimeErr {
		t.Fatalf("unsafe code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func TestSDDNewLiteCreatesOverviewWithoutFullArtifacts(t *testing.T) {
	target := t.TempDir()
	for _, rel := range []string{"", "changes", "specs", "decisions", "verification", "archive"} {
		if err := os.MkdirAll(filepath.Join(target, lufypaths.LufySDD, rel), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	var out, errOut bytes.Buffer
	deps := Dependencies{Stdout: &out, Stderr: &errOut}
	code := Run([]string{"sdd", "new", "--target", target, "--change", "quick-fix", "--mode", "lite", "--json"}, deps)
	if code != ExitOK || !strings.Contains(out.String(), `"mode": "lite"`) {
		t.Fatalf("new lite code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	changeRoot := filepath.Join(target, lufypaths.LufySDD, "changes", "quick-fix")
	if _, err := os.Stat(filepath.Join(changeRoot, "change-overview.html")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(changeRoot, "design.md")); !os.IsNotExist(err) {
		t.Fatalf("lite creó design: %v", err)
	}
}
