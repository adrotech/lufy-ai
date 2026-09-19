package codexlifecycle

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionStartProducesBoundedMetadataOnly(t *testing.T) {
	target := t.TempDir()
	writeLifecycleFile(t, filepath.Join(target, ".agents", "skills", "demo", "SKILL.md"), "---\nname: demo\ndescription: demo skill\n---\n")
	payload := `{"cwd":` + quoteJSON(target) + `,"hook_event_name":"SessionStart","transcript_path":"SECRET_TRANSCRIPT"}`

	var out bytes.Buffer
	if err := NewService().Run(strings.NewReader(payload), &out); err != nil {
		t.Fatal(err)
	}
	var result Output
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("invalid hook output: %v body=%s", err, out.String())
	}
	if !result.Continue || result.HookSpecificOutput == nil || result.HookSpecificOutput.HookEventName != "SessionStart" {
		t.Fatalf("unexpected output: %#v", result)
	}
	context := result.HookSpecificOutput.AdditionalContext
	for _, want := range []string{"skills=ready", "memory=not_initialized", "context=not_available"} {
		if !strings.Contains(context, want) {
			t.Fatalf("context missing %q: %s", want, context)
		}
	}
	if strings.Contains(out.String(), "SECRET_TRANSCRIPT") || len(context) > maxContextBytes+len("…") {
		t.Fatalf("hook leaked private input or exceeded bound: %s", out.String())
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy", "skill-registry.json")); err != nil {
		t.Fatalf("skill registry was not ensured: %v", err)
	}
}

func TestSubagentStopDiagnosesEmptyPayloadWithoutAdvancing(t *testing.T) {
	target := t.TempDir()
	payload := `{"cwd":` + quoteJSON(target) + `,"hook_event_name":"SubagentStop","last_assistant_message":""}`
	var out bytes.Buffer
	if err := NewService().Run(strings.NewReader(payload), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "resultado vacío") || !strings.Contains(out.String(), "ningún gate fue avanzado") {
		t.Fatalf("missing diagnostic: %s", out.String())
	}
}

func TestStopReportsActiveChangeReadOnly(t *testing.T) {
	target := t.TempDir()
	change := filepath.Join(target, ".lufy", "workflows", "sdd", "changes", "demo-change")
	writeLifecycleFile(t, filepath.Join(change, "change.yaml"), "schemaVersion: 1\nid: demo-change\nmode: lite\nstatus: proposed\nsyncedDigest: \"\"\n")
	writeLifecycleFile(t, filepath.Join(change, "proposal.md"), "# Proposal\n")
	writeLifecycleFile(t, filepath.Join(change, "tasks.md"), "- [ ] implementar\n")
	payload := `{"cwd":` + quoteJSON(target) + `,"hook_event_name":"Stop"}`

	var out bytes.Buffer
	if err := NewService().Run(strings.NewReader(payload), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "change=demo-change") || !strings.Contains(out.String(), "tasks=0/1") || !strings.Contains(out.String(), "read-only") {
		t.Fatalf("unexpected status diagnostic: %s", out.String())
	}
	body, err := os.ReadFile(filepath.Join(change, "tasks.md"))
	if err != nil || string(body) != "- [ ] implementar\n" {
		t.Fatalf("stop hook mutated tasks: body=%q err=%v", body, err)
	}
}

func TestInvalidInputReturnsSafeJSON(t *testing.T) {
	var out bytes.Buffer
	if err := NewService().Run(strings.NewReader("not-json"), &out); err != nil {
		t.Fatal(err)
	}
	var result Output
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || !result.Continue || !strings.Contains(result.SystemMessage, "payload JSON inválido") {
		t.Fatalf("unexpected safe fallback: result=%#v err=%v body=%s", result, err, out.String())
	}
}

func writeLifecycleFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func quoteJSON(value string) string {
	body, _ := json.Marshal(value)
	return string(body)
}
