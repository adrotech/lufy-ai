package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeResolver struct{ path string }

func (f fakeResolver) LookPath(file string) (string, error) {
	if f.path == "" {
		return "", errors.New("missing")
	}
	return f.path, nil
}

func TestEnsureCreatesOpenCodeJSONWithPortableEngram(t *testing.T) {
	target := t.TempDir()
	res, err := NewService().Ensure(Options{TargetRoot: target, Resolver: fakeResolver{path: "/usr/local/bin/engram"}})
	if err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	if !res.EngramFound || res.Action != "merge-json" {
		t.Fatalf("unexpected result: %#v", res)
	}
	body := readConfig(t, target)
	if strings.Contains(string(body), "/opt/homebrew/bin/engram") {
		t.Fatal("config hardcodea /opt/homebrew/bin/engram")
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	mcp := decoded["mcp"].(map[string]any)
	engram := mcp["engram"].(map[string]any)
	cmd := engram["command"].([]any)
	if cmd[0] != "/usr/local/bin/engram" {
		t.Fatalf("unexpected engram command: %#v", cmd)
	}
}

func TestEnsurePreservesUnknownKeysAndNoEngram(t *testing.T) {
	target := t.TempDir()
	writeConfig(t, target, `{"custom": {"keep": true}, "mcp": {"pencil": {"enabled": false}}}`)
	if _, err := NewService().Ensure(Options{TargetRoot: target, NoEngram: true, Resolver: fakeResolver{path: "/usr/local/bin/engram"}}); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	body := readConfig(t, target)
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["custom"] == nil {
		t.Fatalf("unknown keys not preserved: %s", string(body))
	}
	mcp := decoded["mcp"].(map[string]any)
	if _, ok := mcp["engram"]; ok {
		t.Fatalf("engram configured despite --no-engram: %s", string(body))
	}
	if _, ok := mcp["pencil"]; !ok {
		t.Fatalf("existing mcp key not preserved: %s", string(body))
	}
}

func TestPlanSkipsWhenEngramAbsentAndConfigStable(t *testing.T) {
	target := t.TempDir()
	if _, err := NewService().Ensure(Options{TargetRoot: target, Resolver: fakeResolver{}}); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	res, err := NewService().Plan(Options{TargetRoot: target, Resolver: fakeResolver{}})
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	if res.Action != "skip" || res.EngramFound {
		t.Fatalf("unexpected plan: %#v", res)
	}
}

func TestEnsureRejectsInvalidJSON(t *testing.T) {
	target := t.TempDir()
	writeConfig(t, target, `{bad-json`)
	_, err := NewService().Ensure(Options{TargetRoot: target, Resolver: fakeResolver{}})
	if err == nil || !strings.Contains(err.Error(), "opencode.json inválido") {
		t.Fatalf("expected invalid JSON error, got %v", err)
	}
	if got := string(readConfig(t, target)); got != `{bad-json` {
		t.Fatalf("invalid JSON was overwritten: %q", got)
	}
}

func writeConfig(t *testing.T, target, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(target, OpenCodeFile), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readConfig(t *testing.T, target string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(target, OpenCodeFile))
	if err != nil {
		t.Fatal(err)
	}
	return body
}
