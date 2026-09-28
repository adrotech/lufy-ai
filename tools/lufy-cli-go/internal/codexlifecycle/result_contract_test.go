package codexlifecycle

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/resultcontract"
)

func TestExtractResultContractAcceptsPureDocumentOrSingleCompatibleFence(t *testing.T) {
	t.Parallel()

	yamlDocument := lifecycleResultFixture(t, "valid-minimal.yaml")
	jsonDocument := lifecycleResultFixture(t, "valid-minimal.json")
	cases := []struct {
		name    string
		message string
	}{
		{name: "pure YAML", message: yamlDocument},
		{name: "pure JSON", message: jsonDocument},
		{name: "single YAML fence", message: "Resultado:\n```yaml\n" + yamlDocument + "\n```\n"},
		{name: "single JSON fence", message: "Resultado:\n```json\n" + jsonDocument + "\n```\n"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			contract, err := extractResultContract(tc.message)
			if err != nil {
				t.Fatalf("extractResultContract() error = %v", err)
			}
			if contract.SchemaVersion != resultcontract.SchemaVersion || contract.Status != resultcontract.StatusReady {
				t.Fatalf("contract = %#v", contract)
			}
		})
	}
}

func TestExtractResultContractAcceptsValidEnvelopeWithoutOptionalLedger(t *testing.T) {
	t.Parallel()

	contract, err := extractResultContract(lifecycleResultFixture(t, "valid-minimal.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if contract.Ledger != nil {
		t.Fatalf("fixture unexpectedly has ledger: %#v", contract.Ledger)
	}
	result := subagentStop(Input{LastAssistantMessage: stringPointer(lifecycleResultFixture(t, "valid-minimal.yaml"))})
	if !result.Continue || result.SystemMessage != "" || !result.SuppressOutput {
		t.Fatalf("valid ledger-less contract not accepted structurally: %#v", result)
	}
}

func TestExtractResultContractRejectsSubstringInvalidAmbiguousAndOversizedInputs(t *testing.T) {
	t.Parallel()

	const canary = "PRIVATE_RESULT_CANARY_do_not_echo"
	valid := lifecycleResultFixture(t, "valid-minimal.yaml")
	invalid := valid + "\nprivate_prompt: " + canary + "\n"
	cases := []struct {
		name    string
		message string
	}{
		{name: "schema substring", message: "Terminé. schema_version: result-contract/v1. Todo está validado."},
		{name: "invalid envelope", message: invalid},
		{name: "two valid candidates", message: "```yaml\n" + valid + "\n```\n```yaml\n" + valid + "\n```"},
		{name: "mixed compatible candidates", message: "```yaml\n" + invalid + "\n```\n```json\n" + lifecycleResultFixture(t, "valid-minimal.json") + "\n```"},
		{name: "unclosed compatible fence", message: "```yaml\n" + valid},
		{name: "oversized", message: strings.Repeat("x", 64*1024+1) + canary},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := extractResultContract(tc.message); err == nil {
				t.Fatal("extractResultContract() expected rejection")
			} else if strings.Contains(err.Error(), canary) {
				t.Fatalf("extractor leaked canary: %v", err)
			}
			result := subagentStop(Input{LastAssistantMessage: &tc.message})
			if !result.Continue || result.SystemMessage == "" || !strings.Contains(result.SystemMessage, "ningún gate") {
				t.Fatalf("unsafe hook result = %#v", result)
			}
			body, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(body), canary) {
				t.Fatalf("hook warning leaked canary: %s", body)
			}
		})
	}
}

func TestSubagentStopAlwaysContinuesAndNeverMutatesWorkflowGate(t *testing.T) {
	t.Parallel()

	target := t.TempDir()
	tasksPath := filepath.Join(target, ".lufy", "workflows", "sdd", "changes", "demo", "tasks.md")
	writeLifecycleFile(t, tasksPath, "- [ ] remains-pending\n")
	message := lifecycleResultFixture(t, "valid-minimal.yaml")
	payload, err := json.Marshal(Input{
		CWD: target, HookEventName: "SubagentStop", SessionID: "session-result-contract",
		AgentID: "agent-result-contract", AgentType: "validator", LastAssistantMessage: &message,
	})
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := NewService().Run(bytes.NewReader(payload), &output); err != nil {
		t.Fatal(err)
	}
	var result Output
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.Continue {
		t.Fatalf("SubagentStop blocked primary workflow: %#v", result)
	}
	body, err := os.ReadFile(tasksPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "- [ ] remains-pending\n" {
		t.Fatalf("SubagentStop advanced or mutated gate: %q", body)
	}
}

func lifecycleResultFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "resultcontract", "testdata", name)
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func stringPointer(value string) *string { return &value }
