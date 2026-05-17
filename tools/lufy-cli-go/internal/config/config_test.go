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
	if _, ok := decoded["x-lufy-ai"]; ok {
		t.Fatalf("opencode.json contiene clave no soportada por OpenCode: %s", string(body))
	}
}

func TestEnsureRemovesLegacyManagedNamespace(t *testing.T) {
	target := t.TempDir()
	writeConfig(t, target, `{"$schema":"https://opencode.ai/config.json","plugin":[],"x-lufy-ai":{"version":"v0.3.0"}}`)
	if _, err := NewService().Ensure(Options{TargetRoot: target, NoEngram: true, Resolver: fakeResolver{}}); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	body := readConfig(t, target)
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, ok := decoded["x-lufy-ai"]; ok {
		t.Fatalf("legacy managed namespace not removed: %s", string(body))
	}
}

func TestEnsurePreservesExistingEngramUserOptions(t *testing.T) {
	target := t.TempDir()
	writeConfig(t, target, `{"mcp":{"engram":{"type":"local","command":["/old/engram","mcp"],"enabled":false,"timeout":10000,"env":{"KEEP":"1"}}}}`)
	if _, err := NewService().Ensure(Options{TargetRoot: target, Resolver: fakeResolver{path: "/usr/local/bin/engram"}}); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	body := readConfig(t, target)
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	engram := decoded["mcp"].(map[string]any)["engram"].(map[string]any)
	cmd := engram["command"].([]any)
	if cmd[0] != "/usr/local/bin/engram" {
		t.Fatalf("unexpected engram command: %#v", cmd)
	}
	if engram["enabled"] != false || engram["timeout"] != float64(10000) || engram["env"] == nil {
		t.Fatalf("existing engram user options not preserved: %#v", engram)
	}
}

func TestEnsureTreatsNullMCPAsAbsent(t *testing.T) {
	target := t.TempDir()
	writeConfig(t, target, `{"mcp":null}`)
	if _, err := NewService().Ensure(Options{TargetRoot: target, Resolver: fakeResolver{path: "/usr/local/bin/engram"}}); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	body := readConfig(t, target)
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, ok := decoded["mcp"].(map[string]any)["engram"]; !ok {
		t.Fatalf("engram not written when mcp was null: %s", string(body))
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

func TestEnsurePreservesExistingMCPKeysWhenAddingEngram(t *testing.T) {
	target := t.TempDir()
	writeConfig(t, target, `{"mcp":{"local-tool":{"enabled":false}},"custom":true}`)
	if _, err := NewService().Ensure(Options{TargetRoot: target, Resolver: fakeResolver{path: "/usr/local/bin/engram"}}); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	body := readConfig(t, target)
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	mcp := decoded["mcp"].(map[string]any)
	if _, ok := mcp["local-tool"]; !ok {
		t.Fatalf("existing mcp key not preserved: %s", string(body))
	}
	if _, ok := mcp["engram"]; !ok {
		t.Fatalf("engram key not added: %s", string(body))
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

func TestPlanAndEnsureRejectNonRegularOpenCode(t *testing.T) {
	target := t.TempDir()
	if err := os.Mkdir(filepath.Join(target, OpenCodeFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := NewService().Plan(Options{TargetRoot: target, Resolver: fakeResolver{}}); err == nil || !strings.Contains(err.Error(), "archivo regular seguro") {
		t.Fatalf("expected non-regular Plan error, got %v", err)
	}
	if _, err := NewService().Ensure(Options{TargetRoot: target, Resolver: fakeResolver{}}); err == nil || !strings.Contains(err.Error(), "archivo regular seguro") {
		t.Fatalf("expected non-regular Ensure error, got %v", err)
	}
	if _, err := NewService().ValidateManagedStructure(target); err == nil || !strings.Contains(err.Error(), "archivo regular seguro") {
		t.Fatalf("expected non-regular ValidateManagedStructure error, got %v", err)
	}
}

func TestPlanRejectsSymlinkOpenCode(t *testing.T) {
	target := t.TempDir()
	outside := filepath.Join(t.TempDir(), OpenCodeFile)
	if err := os.WriteFile(outside, []byte(`{"$schema":"https://opencode.ai/config.json","plugin":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(target, OpenCodeFile)); err != nil {
		t.Skipf("symlink no soportado en este entorno: %v", err)
	}
	if _, err := NewService().Plan(Options{TargetRoot: target, Resolver: fakeResolver{}}); err == nil || !unsafeOpenCodeError(err) {
		t.Fatalf("expected symlink Plan error, got %v", err)
	}
}

func TestPlanRejectsInvalidManagedKeyTypes(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{name: "schema", content: `{"$schema":123,"plugin":[]}`, want: "$schema debe ser string"},
		{name: "plugin", content: `{"$schema":"https://opencode.ai/config.json","plugin":{}}`, want: "plugin debe ser array"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := t.TempDir()
			writeConfig(t, target, tt.content)
			_, err := NewService().Plan(Options{TargetRoot: target, Resolver: fakeResolver{}})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q error, got %v", tt.want, err)
			}
			_, err = NewService().Ensure(Options{TargetRoot: target, Resolver: fakeResolver{}})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected Ensure %q error, got %v", tt.want, err)
			}
			_, err = NewService().ValidateManagedStructure(target)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected ValidateManagedStructure %q error, got %v", tt.want, err)
			}
			if got := string(readConfig(t, target)); got != tt.content {
				t.Fatalf("invalid managed key type was overwritten: %q", got)
			}
		})
	}
}

func TestPlanAndEnsureRejectNonObjectMCPWhenEngramEnabled(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "string", content: `{"mcp":"local"}`},
		{name: "array", content: `{"mcp":[]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := t.TempDir()
			writeConfig(t, target, tt.content)
			_, err := NewService().Plan(Options{TargetRoot: target, Resolver: fakeResolver{path: "/usr/local/bin/engram"}})
			if err == nil || !strings.Contains(err.Error(), "mcp debe ser object") {
				t.Fatalf("expected mcp object Plan error, got %v", err)
			}
			_, err = NewService().Ensure(Options{TargetRoot: target, Resolver: fakeResolver{path: "/usr/local/bin/engram"}})
			if err == nil || !strings.Contains(err.Error(), "mcp debe ser object") {
				t.Fatalf("expected mcp object Ensure error, got %v", err)
			}
			if got := string(readConfig(t, target)); got != tt.content {
				t.Fatalf("non-object mcp was overwritten: %q", got)
			}
		})
	}
}

func unsafeOpenCodeError(err error) bool {
	return strings.Contains(err.Error(), "archivo regular seguro") || strings.Contains(err.Error(), "symlink no permitido")
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
