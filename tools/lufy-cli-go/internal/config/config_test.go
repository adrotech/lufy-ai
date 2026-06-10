package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureCreatesOpenCodeJSONWithManagedMinimum(t *testing.T) {
	target := t.TempDir()
	res, err := NewService().Ensure(Options{TargetRoot: target})
	if err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	if res.Action != "merge-json" {
		t.Fatalf("unexpected result: %#v", res)
	}
	body := readConfig(t, target)
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["$schema"] == nil || decoded["plugin"] == nil {
		t.Fatalf("managed minimum missing: %s", string(body))
	}
	if _, ok := decoded["x-lufy-ai"]; ok {
		t.Fatalf("opencode.json contiene clave no soportada por OpenCode: %s", string(body))
	}
	if _, ok := decoded["mcp"]; ok {
		t.Fatalf("new config should not create mcp: %s", string(body))
	}
}

func TestEnsureRemovesLegacyManagedNamespace(t *testing.T) {
	target := t.TempDir()
	writeConfig(t, target, `{"$schema":"https://opencode.ai/config.json","plugin":[],"x-lufy-ai":{"version":"v0.3.0"}}`)
	if _, err := NewService().Ensure(Options{TargetRoot: target}); err != nil {
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

func TestEnsureRemovesLegacyMemoryMCP(t *testing.T) {
	target := t.TempDir()
	writeConfig(t, target, `{"mcp":{"`+legacyMemoryMCPKey+`":{"type":"local","command":["/old/legacy-memory","mcp"]},"pencil":{"enabled":false}}}`)
	if _, err := NewService().Ensure(Options{TargetRoot: target}); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	body := readConfig(t, target)
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	mcp := decoded["mcp"].(map[string]any)
	if _, ok := mcp[legacyMemoryMCPKey]; ok {
		t.Fatalf("legacy mcp removed incompletely: %s", string(body))
	}
	if _, ok := mcp["pencil"]; !ok {
		t.Fatalf("existing mcp key not preserved: %s", string(body))
	}
}

func TestEnsureRemovesEmptyLegacyMCP(t *testing.T) {
	target := t.TempDir()
	writeConfig(t, target, `{"mcp":{"`+legacyMemoryMCPKey+`":{"type":"local","command":["/old/legacy-memory","mcp"]}}}`)
	if _, err := NewService().Ensure(Options{TargetRoot: target}); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	body := readConfig(t, target)
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, ok := decoded["mcp"]; ok {
		t.Fatalf("empty legacy mcp should be removed: %s", string(body))
	}
}

func TestEnsurePreservesUnknownKeysAndMCP(t *testing.T) {
	target := t.TempDir()
	writeConfig(t, target, `{"custom": {"keep": true}, "mcp": {"pencil": {"enabled": false}}}`)
	if _, err := NewService().Ensure(Options{TargetRoot: target}); err != nil {
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
	if _, ok := mcp[legacyMemoryMCPKey]; ok {
		t.Fatalf("legacy memory mcp unexpectedly present: %s", string(body))
	}
	if _, ok := mcp["pencil"]; !ok {
		t.Fatalf("existing mcp key not preserved: %s", string(body))
	}
}

func TestPlanSkipsWhenConfigStable(t *testing.T) {
	target := t.TempDir()
	if _, err := NewService().Ensure(Options{TargetRoot: target}); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	res, err := NewService().Plan(Options{TargetRoot: target})
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	if res.Action != "skip" {
		t.Fatalf("unexpected plan: %#v", res)
	}
}

func TestValidateManagedStructureValidAndMissing(t *testing.T) {
	target := t.TempDir()
	if _, err := NewService().Ensure(Options{TargetRoot: target}); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	valid, err := NewService().ValidateManagedStructure(target)
	if err != nil {
		t.Fatalf("ValidateManagedStructure() valid error = %v", err)
	}
	if !valid.Exists {
		t.Fatal("ValidateManagedStructure() valid Exists = false")
	}

	missing := t.TempDir()
	if _, err := NewService().ValidateManagedStructure(missing); err == nil || !strings.Contains(err.Error(), "falta opencode.json") {
		t.Fatalf("expected missing opencode.json error, got %v", err)
	}
}

func TestEnsureRejectsInvalidJSON(t *testing.T) {
	target := t.TempDir()
	writeConfig(t, target, `{bad-json`)
	_, err := NewService().Ensure(Options{TargetRoot: target})
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
	if _, err := NewService().Plan(Options{TargetRoot: target}); err == nil || !strings.Contains(err.Error(), "archivo regular seguro") {
		t.Fatalf("expected non-regular Plan error, got %v", err)
	}
	if _, err := NewService().Ensure(Options{TargetRoot: target}); err == nil || !strings.Contains(err.Error(), "archivo regular seguro") {
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
	if _, err := NewService().Plan(Options{TargetRoot: target}); err == nil || !unsafeOpenCodeError(err) {
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
			_, err := NewService().Plan(Options{TargetRoot: target})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q error, got %v", tt.want, err)
			}
			_, err = NewService().Ensure(Options{TargetRoot: target})
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

func TestPlanAndEnsurePreserveNonObjectMCP(t *testing.T) {
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
			res, err := NewService().Plan(Options{TargetRoot: target})
			if err != nil {
				t.Fatalf("Plan() error = %v", err)
			}
			if res.Action != "merge-json" {
				t.Fatalf("expected managed minimum merge, got %#v", res)
			}
			_, err = NewService().Ensure(Options{TargetRoot: target})
			if err != nil {
				t.Fatalf("Ensure() error = %v", err)
			}
			body := readConfig(t, target)
			var decoded map[string]any
			if err := json.Unmarshal(body, &decoded); err != nil {
				t.Fatal(err)
			}
			switch tt.name {
			case "string":
				if decoded["mcp"] != "local" {
					t.Fatalf("non-object mcp not preserved: %s", string(body))
				}
			case "array":
				if _, ok := decoded["mcp"].([]any); !ok {
					t.Fatalf("non-object mcp not preserved: %s", string(body))
				}
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
