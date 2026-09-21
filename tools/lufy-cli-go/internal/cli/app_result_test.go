package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/resultcontract"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

const (
	resultValidationSchema   = "lufy-result-validation/v1"
	transitionRequestSchema  = "lufy-result-transition-request/v1"
	transitionDecisionSchema = "lufy-result-transition-decision/v1"
)

type resultCLIDecision struct {
	SchemaVersion string   `json:"schema_version"`
	Status        string   `json:"status"`
	Fingerprint   string   `json:"fingerprint"`
	Reason        string   `json:"reason"`
	Recovery      string   `json:"recovery"`
	Missing       []string `json:"missing_evidence"`
	NextOwner     string   `json:"next_owner"`
	DecisionID    string   `json:"decision_id"`
	NextVersion   uint64   `json:"next_version"`
}

func TestResultValidateSupportsExplicitStdinAndFileWithHumanOrJSON(t *testing.T) {
	t.Parallel()

	valid := resultFixture(t, "valid-minimal.yaml")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"result", "validate", "--stdin", "--role", "orchestrator"}, Dependencies{
		Stdin: bytes.NewReader(valid), Stdout: &stdout, Stderr: &stderr,
	})
	if code != ExitOK || !strings.Contains(stdout.String(), "result validate: valid") || !strings.Contains(stdout.String(), "fingerprint=") {
		t.Fatalf("human validate code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"result", "validate", "--file", resultFixturePath(t, "valid-minimal.yaml"), "--role", "orchestrator", "--json"}, Dependencies{
		Stdout: &stdout, Stderr: &stderr,
	})
	if code != ExitOK {
		t.Fatalf("file validate code=%d stderr=%s", code, stderr.String())
	}
	decision := decodeResultDecision(t, stdout.Bytes())
	if decision.SchemaVersion != resultValidationSchema || decision.Status != "valid" || len(decision.Fingerprint) != 64 {
		t.Fatalf("validation decision = %#v", decision)
	}
}

func TestResultInputRequiresExactlyOneOfStdinOrFile(t *testing.T) {
	t.Parallel()

	commands := [][]string{
		{"result", "validate"},
		{"result", "normalize"},
		{"result", "transition"},
		{"result", "validate", "--stdin", "--file", resultFixturePath(t, "valid-minimal.yaml")},
		{"result", "normalize", "--stdin", "--file", resultFixturePath(t, "legacy-minimal.yaml")},
		{"result", "transition", "--stdin", "--file", resultFixturePath(t, "valid-minimal.json")},
	}
	for _, args := range commands {
		var stdout, stderr bytes.Buffer
		code := Run(args, Dependencies{Stdin: bytes.NewReader(nil), Stdout: &stdout, Stderr: &stderr})
		if code != ExitUsageErr {
			t.Fatalf("Run(%v) code=%d, want usage=%d; stdout=%s stderr=%s", args, code, ExitUsageErr, stdout.String(), stderr.String())
		}
	}
}

func TestResultValidateRolePolicyAndInvalidPayloadUseStableExitCodes(t *testing.T) {
	t.Parallel()

	valid := resultFixture(t, "valid-minimal.yaml")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"result", "validate", "--stdin", "--role", "validator", "--json"}, Dependencies{
		Stdin: bytes.NewReader(valid), Stdout: &stdout, Stderr: &stderr,
	})
	if code != ExitResultRejected {
		t.Fatalf("role rejection code=%d want=%d stdout=%s stderr=%s", code, ExitResultRejected, stdout.String(), stderr.String())
	}
	decision := decodeResultDecision(t, stdout.Bytes())
	if decision.Status != "rejected" || decision.Reason != "role_status_not_allowed" || decision.Recovery == "" {
		t.Fatalf("role decision = %#v", decision)
	}

	const canary = "PRIVATE_VALIDATE_CANARY_do_not_echo"
	invalid := append(append([]byte(nil), valid...), []byte("\nsecret: "+canary+"\n")...)
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"result", "validate", "--stdin", "--json"}, Dependencies{
		Stdin: bytes.NewReader(invalid), Stdout: &stdout, Stderr: &stderr,
	})
	if code != ExitResultInvalid {
		t.Fatalf("invalid code=%d want=%d stdout=%s stderr=%s", code, ExitResultInvalid, stdout.String(), stderr.String())
	}
	decision = decodeResultDecision(t, stdout.Bytes())
	if decision.Status != "invalid" || decision.Reason == "" || decision.Recovery == "" {
		t.Fatalf("invalid decision = %#v", decision)
	}
	if strings.Contains(stdout.String()+stderr.String(), canary) {
		t.Fatalf("validation output leaked canary: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}

func TestResultNormalizeUsesNarrowStructuredLegacyContract(t *testing.T) {
	t.Parallel()

	legacy := resultFixture(t, "legacy-minimal.yaml")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"result", "normalize", "--stdin", "--json"}, Dependencies{
		Stdin: bytes.NewReader(legacy), Stdout: &stdout, Stderr: &stderr,
	})
	if code != ExitOK {
		t.Fatalf("normalize code=%d stderr=%s", code, stderr.String())
	}
	var contract resultcontract.Contract
	if err := json.Unmarshal(stdout.Bytes(), &contract); err != nil {
		t.Fatalf("normalize JSON invalid: %v\n%s", err, stdout.String())
	}
	if contract.SchemaVersion != resultcontract.SchemaVersion || contract.Status != resultcontract.StatusImplemented || !contract.LegacyFallback {
		t.Fatalf("normalized contract identity = %#v", contract)
	}
	if len(contract.Evidence.Commands) != 1 || contract.Evidence.Commands[0].Result != resultcontract.EvidenceNotRun ||
		contract.Evidence.Commands[0].Notes != "not_available" {
		t.Fatalf("normalizer fabricated or lost uncertainty: %#v", contract.Evidence)
	}
	if len(contract.Evidence.Static) != 1 || contract.Evidence.Static[0] != "not_available" {
		t.Fatalf("normalized static evidence = %#v", contract.Evidence.Static)
	}
	if !containsString(contract.Artifacts.Referenced, "legacy:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Fatalf("provenance not preserved opaquely: %#v", contract.Artifacts.Referenced)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"result", "normalize", "--file", resultFixturePath(t, "legacy-minimal.yaml")}, Dependencies{
		Stdout: &stdout, Stderr: &stderr,
	})
	if code != ExitOK || !strings.Contains(stdout.String(), "schema_version: result-contract/v1") ||
		!strings.Contains(stdout.String(), "legacy_fallback: true") {
		t.Fatalf("human normalize code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestResultNormalizeRejectsAmbiguousFreeTextAndUnboundedLegacy(t *testing.T) {
	t.Parallel()

	const canary = "LEGACY_PRIVATE_CANARY_do_not_echo"
	valid := string(resultFixture(t, "legacy-minimal.yaml"))
	cases := []struct {
		name    string
		payload string
	}{
		{name: "multiple candidates", payload: "- " + strings.ReplaceAll(valid, "\n", "\n  ") + "\n- " + strings.ReplaceAll(valid, "\n", "\n  ")},
		{name: "conflicting duplicate status", payload: valid + "status: blocked\n"},
		{name: "advanced claim without evidence", payload: strings.Replace(valid, "status: implemented", "status: closed", 1)},
		{name: "free text heuristic forbidden", payload: "Terminé todo. status: implemented. schema_version: lufy-result-legacy/v1 " + canary},
		{name: "unbounded", payload: strings.Repeat("x", 64*1024+1) + canary},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			code := Run([]string{"result", "normalize", "--stdin", "--json"}, Dependencies{
				Stdin: strings.NewReader(tc.payload), Stdout: &stdout, Stderr: &stderr,
			})
			if code != ExitResultInvalid {
				t.Fatalf("normalize ambiguous code=%d want=%d stdout=%s stderr=%s", code, ExitResultInvalid, stdout.String(), stderr.String())
			}
			decision := decodeResultDecision(t, stdout.Bytes())
			if decision.Status != "invalid" || decision.Recovery == "" {
				t.Fatalf("normalize rejection = %#v", decision)
			}
			if strings.Contains(stdout.String()+stderr.String(), canary) {
				t.Fatalf("normalize diagnostics leaked canary: stdout=%s stderr=%s", stdout.String(), stderr.String())
			}
		})
	}
}

func TestResultTransitionIsReadOnlyByDefaultAndJSONIsContentFree(t *testing.T) {
	t.Parallel()

	target := t.TempDir()
	const canary = "TRANSITION_PROMPT_CANARY_private"
	request := transitionRequest(t, resultcontract.StatusImplemented, 1, canary)
	var stdout, stderr bytes.Buffer
	code := Run([]string{"result", "transition", "--stdin", "--target", target, "--json"}, Dependencies{
		Stdin: bytes.NewReader(request), Stdout: &stdout, Stderr: &stderr,
	})
	if code != ExitOK {
		t.Fatalf("transition code=%d stderr=%s", code, stderr.String())
	}
	decision := decodeResultDecision(t, stdout.Bytes())
	if decision.SchemaVersion != transitionDecisionSchema || decision.Status != "accepted" || decision.NextVersion != 2 {
		t.Fatalf("transition decision = %#v", decision)
	}
	if strings.Contains(stdout.String()+stderr.String(), canary) {
		t.Fatalf("transition decision leaked contract content: %s %s", stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy", "runtime")); !os.IsNotExist(err) {
		t.Fatalf("read-only transition created durable runtime state: err=%v", err)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"result", "transition", "--stdin", "--target", target}, Dependencies{
		Stdin: bytes.NewReader(request), Stdout: &stdout, Stderr: &stderr,
	})
	if code != ExitOK || !strings.Contains(stdout.String(), "result transition: accepted") {
		t.Fatalf("human transition code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestResultTransitionStrictlyParsesTypedPriorAndIntentWithoutEcho(t *testing.T) {
	t.Parallel()

	const canary = "TRANSITION_PRIVATE_FIELD_CANARY"
	base := transitionRequest(t, resultcontract.StatusImplemented, 1, "typed request")
	var document map[string]any
	if err := json.Unmarshal(base, &document); err != nil {
		t.Fatal(err)
	}
	intent, ok := document["intent"].(map[string]any)
	if !ok {
		t.Fatalf("intent fixture type = %T", document["intent"])
	}
	intent["private_prompt"] = canary
	body, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"result", "transition", "--stdin", "--json"}, Dependencies{
		Stdin: bytes.NewReader(body), Stdout: &stdout, Stderr: &stderr,
	})
	if code != ExitResultInvalid {
		t.Fatalf("typed rejection code=%d want=%d stdout=%s stderr=%s", code, ExitResultInvalid, stdout.String(), stderr.String())
	}
	decision := decodeResultDecision(t, stdout.Bytes())
	if decision.SchemaVersion != transitionDecisionSchema || decision.Status != "invalid" || decision.Recovery == "" {
		t.Fatalf("typed rejection = %#v", decision)
	}
	if strings.Contains(stdout.String()+stderr.String(), canary) {
		t.Fatalf("typed transition diagnostics leaked canary: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}

func TestResultTransitionRecordIsExplicitAndDurablyIdempotent(t *testing.T) {
	t.Parallel()

	target := t.TempDir()
	request := transitionRequest(t, resultcontract.StatusImplemented, 1, "recorded transition")
	for attempt, wantStatus := range []string{"accepted", "duplicate_noop"} {
		var stdout, stderr bytes.Buffer
		code := Run([]string{"result", "transition", "--stdin", "--target", target, "--record", "--json"}, Dependencies{
			Stdin: bytes.NewReader(request), Stdout: &stdout, Stderr: &stderr,
		})
		if code != ExitOK {
			t.Fatalf("record attempt %d code=%d stderr=%s", attempt+1, code, stderr.String())
		}
		decision := decodeResultDecision(t, stdout.Bytes())
		if decision.Status != wantStatus || decision.DecisionID == "" {
			t.Fatalf("record attempt %d decision=%#v want status=%s", attempt+1, decision, wantStatus)
		}
	}
}

func TestResultTransitionRecordCorrelatesContentFreeRunLedgerAndFailsClosedWhenUnavailable(t *testing.T) {
	const (
		summaryCanary = "SUMMARY_CANARY_cli_private"
		promptCanary  = "PROMPT_CANARY_cli_ignore_previous"
		pathCanary    = "/Users/private/PATH_CANARY_cli/project"
		outputCanary  = "OUTPUT_CANARY_cli_command_stdout"
	)
	canaries := []string{summaryCanary, promptCanary, pathCanary, outputCanary}
	request := transitionRequestWithPrivateCanaries(t, canaries)

	t.Run("records once for prior run and retry stays idempotent", func(t *testing.T) {
		target := t.TempDir()
		for attempt, wantStatus := range []string{"accepted", "duplicate_noop"} {
			var stdout, stderr bytes.Buffer
			code := Run([]string{"result", "transition", "--stdin", "--target", target, "--record", "--json"}, Dependencies{
				Stdin: bytes.NewReader(request), Stdout: &stdout, Stderr: &stderr,
			})
			if code != ExitOK {
				t.Fatalf("record attempt %d code=%d stdout=%s stderr=%s", attempt+1, code, stdout.String(), stderr.String())
			}
			decision := decodeResultDecision(t, stdout.Bytes())
			if decision.Status != wantStatus || decision.DecisionID == "" || decision.Fingerprint == "" {
				t.Fatalf("record attempt %d decision=%#v want status=%s", attempt+1, decision, wantStatus)
			}
			for _, canary := range canaries {
				if strings.Contains(stdout.String()+stderr.String(), canary) {
					t.Fatalf("record attempt %d leaked canary %q", attempt+1, canary)
				}
			}
		}

		store, err := runledger.NewFileStore(target, runledger.Options{})
		if err != nil {
			t.Fatal(err)
		}
		events, err := store.LoadRun(t.Context(), "run-cli-result")
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 1 {
			t.Fatalf("Run Ledger correlations=%d, want one durable event after retry", len(events))
		}
		event := events[0]
		if event.RunID != "run-cli-result" || event.TaskRef != resultcontract.SchemaVersion || event.Checkpoint == nil ||
			event.Checkpoint.Status != string(resultcontract.StatusImplemented) || event.Checkpoint.Gate != "accepted" {
			t.Fatalf("Run Ledger correlation lost allow-listed identity: %#v", event)
		}

		durable := readResultRuntimeTree(t, filepath.Join(target, ".lufy", "runtime"))
		for _, canary := range canaries {
			if strings.Contains(durable, canary) {
				t.Fatalf("result transition persisted private canary %q", canary)
			}
		}
	})

	t.Run("ledger unavailable returns unavailable without accepted identity", func(t *testing.T) {
		target := t.TempDir()
		cfg := projectconfig.ProjectConfig{RunLedger: projectconfig.DefaultRunLedgerConfig()}
		cfg.RunLedger.Root = "outside-runtime"
		body, err := projectconfig.Marshal(cfg)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(target, projectconfig.ProjectConfigPath)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, body, 0o600); err != nil {
			t.Fatal(err)
		}

		var stdout, stderr bytes.Buffer
		code := Run([]string{"result", "transition", "--stdin", "--target", target, "--record", "--json"}, Dependencies{
			Stdin: bytes.NewReader(request), Stdout: &stdout, Stderr: &stderr,
		})
		if code != ExitResultUnavailable {
			t.Fatalf("ledger unavailable code=%d want=%d stdout=%s stderr=%s", code, ExitResultUnavailable, stdout.String(), stderr.String())
		}
		decision := decodeResultDecision(t, stdout.Bytes())
		if decision.Status != "unavailable" || decision.Recovery == "" {
			t.Fatalf("ledger unavailable decision=%#v", decision)
		}
		if decision.DecisionID != "" || decision.NextVersion != 0 || decision.Fingerprint != "" {
			t.Fatalf("ledger unavailable promoted accepted identity: %#v", decision)
		}
		for _, canary := range canaries {
			if strings.Contains(stdout.String()+stderr.String(), canary) {
				t.Fatalf("ledger unavailable diagnostics leaked canary %q", canary)
			}
		}
	})
}

func TestResultTransitionExitCodesDifferentiateRejectedConflictAndUnavailable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		request    []byte
		targetFile bool
		wantCode   int
		wantStatus string
	}{
		{name: "rejected edge", request: transitionRequest(t, resultcontract.StatusClosed, 1, "skip gates"), wantCode: ExitResultRejected, wantStatus: "rejected"},
		{name: "stale conflict", request: transitionRequest(t, resultcontract.StatusImplemented, 0, "stale"), wantCode: ExitResultConflict, wantStatus: "conflict"},
		{name: "record unavailable", request: transitionRequest(t, resultcontract.StatusImplemented, 1, "unavailable"), targetFile: true, wantCode: ExitResultUnavailable, wantStatus: "unavailable"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			target := t.TempDir()
			if tc.targetFile {
				target = filepath.Join(target, "not-a-directory")
				if err := os.WriteFile(target, []byte("occupied"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{"result", "transition", "--stdin", "--target", target, "--json"}
			if tc.targetFile {
				args = append(args, "--record")
			}
			var stdout, stderr bytes.Buffer
			code := Run(args, Dependencies{Stdin: bytes.NewReader(tc.request), Stdout: &stdout, Stderr: &stderr})
			if code != tc.wantCode {
				t.Fatalf("code=%d want=%d stdout=%s stderr=%s", code, tc.wantCode, stdout.String(), stderr.String())
			}
			decision := decodeResultDecision(t, stdout.Bytes())
			if decision.SchemaVersion != transitionDecisionSchema || decision.Status != tc.wantStatus || decision.Recovery == "" {
				t.Fatalf("decision = %#v, want status=%s", decision, tc.wantStatus)
			}
		})
	}
}

func TestResultHelpListsCommandsFlagsAndStableExitCategories(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		args []string
		want []string
	}{
		{args: []string{"--help"}, want: []string{"result"}},
		{args: []string{"result", "--help"}, want: []string{"validate", "normalize", "transition"}},
		{args: []string{"result", "validate", "--help"}, want: []string{"--stdin", "--file", "--role", "--json"}},
		{args: []string{"result", "normalize", "--help"}, want: []string{"--stdin", "--file", "--json"}},
		{args: []string{"result", "transition", "--help"}, want: []string{"--stdin", "--file", "--record", "--json"}},
	} {
		var stdout, stderr bytes.Buffer
		code := Run(tc.args, Dependencies{Stdout: &stdout, Stderr: &stderr})
		if code != ExitOK {
			t.Fatalf("Run(%v) code=%d stderr=%s", tc.args, code, stderr.String())
		}
		for _, want := range tc.want {
			if !strings.Contains(stdout.String(), want) {
				t.Fatalf("Run(%v) help missing %q:\n%s", tc.args, want, stdout.String())
			}
		}
	}
	if ExitResultInvalid == ExitResultRejected || ExitResultInvalid == ExitResultConflict || ExitResultInvalid == ExitResultUnavailable ||
		ExitResultRejected == ExitResultConflict || ExitResultRejected == ExitResultUnavailable || ExitResultConflict == ExitResultUnavailable ||
		ExitUsageErr == ExitResultInvalid || ExitUsageErr == ExitResultRejected || ExitUsageErr == ExitResultConflict || ExitUsageErr == ExitResultUnavailable {
		t.Fatal("result exit categories must remain pairwise distinct from usage")
	}
}

func transitionRequest(t *testing.T, nextStatus resultcontract.Status, expectedVersion uint64, summary string) []byte {
	t.Helper()
	prior := decodedResultFixture(t, "valid-minimal.yaml")
	prior.Status = resultcontract.StatusReady
	priorCanonical, err := resultcontract.Canonicalize(prior)
	if err != nil {
		t.Fatal(err)
	}
	next := decodedResultFixture(t, "valid-minimal.yaml")
	next.Status = nextStatus
	next.ExecutiveSummary = summary
	if nextStatus == resultcontract.StatusClosed {
		next.Evidence.Commands[0].Result = resultcontract.EvidencePassed
	}
	request := map[string]any{
		"schema_version": transitionRequestSchema,
		"prior": map[string]any{
			"version": 1, "fingerprint": priorCanonical.Fingerprint, "run_id": "run-cli-result",
			"contract": prior,
			"state":    map[string]any{"work": "ready", "delivery": "not_required", "sync": "not_required", "attention": "none", "terminal": false},
		},
		"intent": map[string]any{
			"schema_version":       resultcontract.TransitionSchemaVersion,
			"transition_id":        "transition-cli-result",
			"idempotency_key":      "cli-result-idempotency-key",
			"expected_version":     expectedVersion,
			"previous_fingerprint": priorCanonical.Fingerprint,
			"actor":                map[string]any{"role": "implementer", "owner_ref": testCLIDigest("cli-result-owner")},
			"next_contract":        next,
		},
	}
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func transitionRequestWithPrivateCanaries(t *testing.T, canaries []string) []byte {
	t.Helper()
	if len(canaries) != 4 {
		t.Fatalf("private canaries=%d, want summary/prompt/path/output", len(canaries))
	}
	request := transitionRequest(t, resultcontract.StatusImplemented, 1, canaries[0]+" "+canaries[1])
	var document map[string]any
	if err := json.Unmarshal(request, &document); err != nil {
		t.Fatal(err)
	}
	intent, ok := document["intent"].(map[string]any)
	if !ok {
		t.Fatalf("intent fixture type=%T", document["intent"])
	}
	contract, ok := intent["next_contract"].(map[string]any)
	if !ok {
		t.Fatalf("next contract fixture type=%T", intent["next_contract"])
	}
	contract["artifacts"] = map[string]any{"changed": []any{canaries[2]}, "referenced": []any{"none"}}
	contract["evidence"] = map[string]any{
		"commands": []any{map[string]any{"command": "none", "result": "not_run", "notes": canaries[3]}},
		"static":   []any{"not_applicable"},
	}
	body, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func readResultRuntimeTree(t *testing.T, root string) string {
	t.Helper()
	var output strings.Builder
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		output.Write(body)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return output.String()
}

func decodeResultDecision(t *testing.T, body []byte) resultCLIDecision {
	t.Helper()
	var decision resultCLIDecision
	if err := json.Unmarshal(body, &decision); err != nil {
		t.Fatalf("decision JSON invalid: %v\n%s", err, body)
	}
	return decision
}

func decodedResultFixture(t *testing.T, name string) resultcontract.Contract {
	t.Helper()
	contract, err := resultcontract.Decode(bytes.NewReader(resultFixture(t, name)))
	if err != nil {
		t.Fatalf("Decode(%s) error = %v", name, err)
	}
	return contract
}

func resultFixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(resultFixturePath(t, name))
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func resultFixturePath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("..", "resultcontract", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func testCLIDigest(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
