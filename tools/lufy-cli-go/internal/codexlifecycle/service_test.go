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

func TestOtherEventsAreIgnoredWithoutReadingOrRecordingReports(t *testing.T) {
	for _, event := range []string{"SubagentStart", "SubagentStop", "Stop", "SessionEnd", "unknown"} {
		target := t.TempDir()
		out := runLifecycleFixture(t, `{"cwd":`+quoteJSON(target)+`,"hook_event_name":"`+event+`","last_assistant_message":"PRIVATE-CANARY"}`)
		if !out.Continue || !out.SuppressOutput || out.SystemMessage != "" {
			t.Fatalf("unexpected report processing: %#v", out)
		}
		files, err := os.ReadDir(target)
		if err != nil || len(files) != 0 {
			t.Fatal("ignored event wrote metadata")
		}
	}
}
func TestSessionStartNeverCreatesAutomaticRunLedger(t *testing.T) {
	target := t.TempDir()
	writeLifecycleFile(t, filepath.Join(target, ".lufy", "config", "project.yaml"), "schema_version: 2\nrun_ledger:\n  enabled: true\n")
	out := runLifecycleFixture(t, `{"cwd":`+quoteJSON(target)+`,"hook_event_name":"SessionStart","session_id":"PRIVATE-SESSION","last_assistant_message":"PRIVATE-MESSAGE"}`)
	if !out.Continue {
		t.Fatal("orientation blocked")
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy", "runtime")); !os.IsNotExist(err) {
		t.Fatal("automatic ledger was created")
	}
	body, _ := json.Marshal(out)
	if strings.Contains(string(body), "PRIVATE-") {
		t.Fatal("private identifiers or messages leaked")
	}
}
func TestInvalidLifecycleInputRemainsNonBlocking(t *testing.T) {
	var out bytes.Buffer
	if err := NewService().Run(strings.NewReader("broken"), &out); err != nil {
		t.Fatal(err)
	}
	var parsed Output
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil || !parsed.Continue {
		t.Fatal("invalid payload blocked orientation")
	}
}
func runLifecycleFixture(t *testing.T, payload string) Output {
	t.Helper()
	var out bytes.Buffer
	if err := NewService().Run(strings.NewReader(payload), &out); err != nil {
		t.Fatal(err)
	}
	var result Output
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("invalid hook output: %v body=%s", err, out.String())
	}
	return result
}

func readLifecycleTree(t *testing.T, root string) string {
	t.Helper()
	var body strings.Builder
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		body.Write(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return body.String()
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
