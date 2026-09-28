package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunLifecycleCodexDispatchesHookPayload(t *testing.T) {
	target := t.TempDir()
	payload, err := json.Marshal(map[string]any{"cwd": target, "hook_event_name": "SubagentStop", "last_assistant_message": ""})
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"lifecycle", "codex"}, Dependencies{Stdin: bytes.NewReader(payload), Stdout: &stdout, Stderr: &stderr})
	if code != ExitOK || !strings.Contains(stdout.String(), "resultado vacío") || stderr.Len() != 0 {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestRunLifecycleRejectsUnsupportedAdapter(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"lifecycle", "opencode"}, Dependencies{Stdin: strings.NewReader("{}"), Stdout: &stdout, Stderr: &stderr})
	if code != ExitUsageErr || !strings.Contains(stderr.String(), "lifecycle codex") {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
}
