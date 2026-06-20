package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunContextStatusJSONNotAvailable(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"context", "status", "--target", t.TempDir(), "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"status": "not_available"`)) {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestRunContextBuildAndQueryJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc Hello() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"context", "build", "--target", root, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("build code=%d stderr=%s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"context", "query", "--target", root, "--json", "Hello"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("query code=%d stderr=%s", code, stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte("Hello")) {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}
