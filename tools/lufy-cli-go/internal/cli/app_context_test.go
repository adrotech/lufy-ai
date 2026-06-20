package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunContextLifecycleCommands(t *testing.T) {
	target := t.TempDir()
	writeContextCLITestFile(t, filepath.Join(target, "main.go"), "package main\ntype User struct{}\nfunc Run() {}\n")

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"context", "scan", "--target", target}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK || !strings.Contains(out.String(), "context scan:") {
		t.Fatalf("context scan code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy", "context", "graph.json")); !os.IsNotExist(err) {
		t.Fatalf("scan should not write graph: %v", err)
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"context", "build", "--target", target, "--json"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK {
		t.Fatalf("context build code=%d stderr=%s", code, errOut.String())
	}
	var build map[string]any
	if err := json.Unmarshal(out.Bytes(), &build); err != nil || build["status"] != "ready" {
		t.Fatalf("build json=%s err=%v", out.String(), err)
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"context", "status", "--target", target}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK || !strings.Contains(out.String(), "context graph: ready") {
		t.Fatalf("context status code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"context", "query", "--target", target, "User"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK || !strings.Contains(out.String(), "file:main.go#type:User") {
		t.Fatalf("context query code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"context", "path", "--target", target, "file:main.go", "file:main.go#type:User"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK || !strings.Contains(out.String(), "path: file:main.go") {
		t.Fatalf("context path code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"context", "explain", "--target", target, "file:main.go#type:User"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK || !strings.Contains(out.String(), "go type declaration") {
		t.Fatalf("context explain code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func TestRunContextNotAvailableAndUsageErrors(t *testing.T) {
	target := t.TempDir()
	for _, args := range [][]string{
		{"context"},
		{"context", "unknown"},
		{"context", "scan", "extra"},
		{"context", "build", "extra"},
		{"context", "status", "extra"},
		{"context", "query"},
		{"context", "path", "only-one"},
		{"context", "explain"},
		{"context", "diff"},
		{"context", "diff", "--base", "HEAD", "extra"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var out bytes.Buffer
			var errOut bytes.Buffer
			if code := Run(args, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitUsageErr {
				t.Fatalf("Run(%v) code=%d stdout=%s stderr=%s", args, code, out.String(), errOut.String())
			}
		})
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run([]string{"context", "query", "--target", target, "--json", "missing"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK || !strings.Contains(out.String(), `"status": "not_available"`) {
		t.Fatalf("query missing graph code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}

	out.Reset()
	errOut.Reset()
	code = Run([]string{"context", "diff", "--target", target, "--base", "HEAD"}, Dependencies{Stdout: &out, Stderr: &errOut})
	if code != ExitOK || !strings.Contains(errOut.String(), "context graph not_available") {
		t.Fatalf("diff missing graph code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}

	out.Reset()
	errOut.Reset()
	if code = Run([]string{"context", "--help"}, Dependencies{Stdout: &out, Stderr: &errOut}); code != ExitOK || !strings.Contains(out.String(), "lufy-ai context") {
		t.Fatalf("context help code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func writeContextCLITestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
