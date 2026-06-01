package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadUpstreamReturnsEffectiveVersion(t *testing.T) {
	path := writeJSON(t, `{"effectiveOpenSpecVersion":"1.2.3","profile":"stable"}`)

	data, err := readUpstream(path)
	if err != nil {
		t.Fatalf("readUpstream() error = %v", err)
	}
	if got := data["effectiveOpenSpecVersion"]; got != "1.2.3" {
		t.Fatalf("effectiveOpenSpecVersion = %v", got)
	}
}

func TestWriteUpstreamPreservesKeysAndUpdatesVersion(t *testing.T) {
	path := writeJSON(t, `{"effectiveOpenSpecVersion":"1.2.3","profile":"stable"}`)
	data, err := readUpstream(path)
	if err != nil {
		t.Fatalf("readUpstream() error = %v", err)
	}
	data["effectiveOpenSpecVersion"] = "1.2.4"

	if err := writeUpstream(path, data); err != nil {
		t.Fatalf("writeUpstream() error = %v", err)
	}
	updated, err := readUpstream(path)
	if err != nil {
		t.Fatalf("readUpstream(updated) error = %v", err)
	}
	if got := updated["effectiveOpenSpecVersion"]; got != "1.2.4" {
		t.Fatalf("effectiveOpenSpecVersion = %v", got)
	}
	if got := updated["profile"]; got != "stable" {
		t.Fatalf("profile = %v", got)
	}
}

func TestRunGetVersionWritesVersion(t *testing.T) {
	path := writeJSON(t, `{"effectiveOpenSpecVersion":"1.2.3"}`)
	var out bytes.Buffer

	if err := run([]string{"get-version", path}, &out); err != nil {
		t.Fatalf("run(get-version) error = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "1.2.3" {
		t.Fatalf("output = %q", got)
	}
}

func TestRunSetVersionUpdatesMultipleFiles(t *testing.T) {
	first := writeJSON(t, `{"effectiveOpenSpecVersion":"1.2.3"}`)
	second := writeJSON(t, `{"effectiveOpenSpecVersion":"1.2.3"}`)

	if err := run([]string{"set-version", "1.2.4", first, second}, &bytes.Buffer{}); err != nil {
		t.Fatalf("run(set-version) error = %v", err)
	}
	for _, path := range []string{first, second} {
		data, err := readUpstream(path)
		if err != nil {
			t.Fatalf("readUpstream(%s) error = %v", path, err)
		}
		if got := data["effectiveOpenSpecVersion"]; got != "1.2.4" {
			t.Fatalf("%s effectiveOpenSpecVersion = %v", path, got)
		}
	}
}

func TestRunRejectsInvalidArgs(t *testing.T) {
	err := run([]string{"unknown"}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "uso:") {
		t.Fatalf("expected usage error, got %v", err)
	}
}

func TestRunRejectsMissingEffectiveVersion(t *testing.T) {
	path := writeJSON(t, `{"profile":"stable"}`)
	err := run([]string{"get-version", path}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "effectiveOpenSpecVersion") {
		t.Fatalf("expected missing version error, got %v", err)
	}
}

func TestRunSetVersionRejectsInvalidJSON(t *testing.T) {
	err := run([]string{"set-version", "1.2.4", writeJSON(t, `{bad-json`)}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "JSON inválido") {
		t.Fatalf("expected invalid JSON error, got %v", err)
	}
}

func TestReadUpstreamRejectsInvalidJSON(t *testing.T) {
	_, err := readUpstream(writeJSON(t, `{bad-json`))
	if err == nil || !strings.Contains(err.Error(), "JSON inválido") {
		t.Fatalf("expected invalid JSON error, got %v", err)
	}
}

func TestWriteUpstreamRejectsDirectoryPath(t *testing.T) {
	err := writeUpstream(t.TempDir(), map[string]any{"effectiveOpenSpecVersion": "1.2.3"})
	if err == nil || !strings.Contains(err.Error(), "escribir") {
		t.Fatalf("expected write error, got %v", err)
	}
}

func writeJSON(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "UPSTREAM.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
