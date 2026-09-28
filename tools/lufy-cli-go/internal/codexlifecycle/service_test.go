package codexlifecycle

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
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

func TestLifecycleRecordsPseudonymizedCausalRunIdempotently(t *testing.T) {
	target := t.TempDir()
	writeLifecycleFile(t, filepath.Join(target, ".agents", "skills", "demo", "SKILL.md"), "---\nname: demo\ndescription: demo skill\n---\n")

	fixtures := []string{
		`{"cwd":` + quoteJSON(target) + `,"hook_event_name":"SessionStart","session_id":"session-secret","turn_id":"turn-secret","transcript_path":"TRANSCRIPT-CANARY","unknown_payload":"UNKNOWN-CANARY"}`,
		`{"cwd":` + quoteJSON(target) + `,"hook_event_name":"SubagentStart","session_id":"session-secret","turn_id":"turn-secret","agent_id":"agent-secret","agent_type":"implementer","transcript_path":"TRANSCRIPT-CANARY"}`,
		`{"cwd":` + quoteJSON(target) + `,"hook_event_name":"SubagentStop","session_id":"session-secret","turn_id":"turn-secret","agent_id":"agent-secret","agent_type":"implementer","last_assistant_message":"MESSAGE-CANARY"}`,
		`{"cwd":` + quoteJSON(target) + `,"hook_event_name":"Stop","session_id":"session-secret","turn_id":"turn-secret"}`,
		`{"cwd":` + quoteJSON(target) + `,"hook_event_name":"SessionEnd","session_id":"session-secret","turn_id":"turn-secret"}`,
	}
	for _, payload := range fixtures {
		for attempt := 0; attempt < 2; attempt++ {
			result := runLifecycleFixture(t, payload)
			if !result.Continue {
				t.Fatalf("lifecycle bloqueó el workflow: %#v", result)
			}
		}
	}

	store, err := runledger.NewFileStore(target, runledger.Options{})
	if err != nil {
		t.Fatal(err)
	}
	runIDs, err := store.ListRuns(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(runIDs)
	if len(runIDs) != 2 {
		t.Fatalf("want root + child run, got %v", runIDs)
	}
	var rootID string
	for _, runID := range runIDs {
		events, loadErr := store.LoadRun(t.Context(), runID)
		if loadErr != nil {
			t.Fatal(loadErr)
		}
		if events[0].ParentRunID == "" {
			rootID = runID
			if len(events) != 3 {
				t.Fatalf("root retries must be no-op, events=%d", len(events))
			}
		} else if len(events) != 2 {
			t.Fatalf("child retries must be no-op, events=%d", len(events))
		}
	}
	if rootID == "" {
		t.Fatal("root run not found")
	}
	summary, err := runledger.NewProjector(store).Build(t.Context(), rootID)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Status != "closed" || !summary.Terminal || len(summary.Children) != 1 || summary.Children[0].Status != "closed" {
		t.Fatalf("unexpected causal summary: %#v", summary)
	}

	runtimeBody := readLifecycleTree(t, filepath.Join(target, ".lufy", "runtime"))
	for _, forbidden := range []string{"session-secret", "turn-secret", "agent-secret", "TRANSCRIPT-CANARY", "UNKNOWN-CANARY", "MESSAGE-CANARY"} {
		if strings.Contains(runtimeBody, forbidden) {
			t.Fatalf("runtime leaked %q", forbidden)
		}
	}
	if !strings.Contains(runtimeBody, "lufy-run-binding/v1") {
		t.Fatalf("expected durable pseudonymized bindings, body=%s", runtimeBody)
	}
}

func TestLifecycleLedgerFailureRemainsBestEffort(t *testing.T) {
	target := t.TempDir()
	writeLifecycleFile(t, filepath.Join(target, ".lufy", "config", "project.yaml"), "schema_version: 1\nrun_ledger:\n  enabled: true\n  root: outside-runtime\n")
	result := runLifecycleFixture(t, `{"cwd":`+quoteJSON(target)+`,"hook_event_name":"SessionStart","session_id":"session-secret"}`)
	if !result.Continue || !strings.Contains(result.SystemMessage, "Run Ledger") || !strings.Contains(result.SystemMessage, "continúa") {
		t.Fatalf("ledger failure changed primary hook semantics: %#v", result)
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
