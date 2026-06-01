package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckWorkflowsAcceptsValidYAML(t *testing.T) {
	root := t.TempDir()
	writeWorkflow(t, root, "ci.yml", "name: ci\non: push\njobs: {}\n")

	if err := checkWorkflows(root); err != nil {
		t.Fatalf("checkWorkflows() error = %v", err)
	}
}

func TestRunAcceptsRootFlag(t *testing.T) {
	root := t.TempDir()
	writeWorkflow(t, root, "ci.yml", "name: ci\non: push\njobs: {}\n")
	var out bytes.Buffer

	if err := run([]string{"--root", root}, &out, &bytes.Buffer{}); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "yaml ok" {
		t.Fatalf("output = %q", got)
	}
}

func TestCheckWorkflowsRejectsInvalidYAML(t *testing.T) {
	root := t.TempDir()
	writeWorkflow(t, root, "ci.yml", "name: [invalid\n")

	err := checkWorkflows(root)
	if err == nil || !strings.Contains(err.Error(), "YAML inválido") {
		t.Fatalf("expected invalid YAML error, got %v", err)
	}
}

func TestCheckWorkflowsRejectsMissingDirectory(t *testing.T) {
	err := checkWorkflows(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "no se encontraron workflows") {
		t.Fatalf("expected missing workflows error, got %v", err)
	}
}

func TestCheckWorkflowsRejectsUnreadableWorkflowPath(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".github", "workflows", "ci.yml")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}

	err := checkWorkflows(root)
	if err == nil || !strings.Contains(err.Error(), "leer") {
		t.Fatalf("expected read error, got %v", err)
	}
}

func TestRunRejectsInvalidFlag(t *testing.T) {
	var stderr bytes.Buffer
	err := run([]string{"--unknown"}, &bytes.Buffer{}, &stderr)
	if err == nil || !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Fatalf("expected flag error, err=%v stderr=%q", err, stderr.String())
	}
}

func writeWorkflow(t *testing.T, root, name, body string) {
	t.Helper()
	path := filepath.Join(root, ".github", "workflows", name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
