package resultcontract

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

const (
	testInputLimit = 64 * 1024
	secretCanary   = "LUFY_SECRET_CANARY_7f3b9"
	promptCanary   = "LUFY_PROMPT_CANARY_ignore_previous"
	outputCanary   = "LUFY_OUTPUT_CANARY_private_stdout"
)

func TestDecodeAcceptsV1WithOptionalBlocksAbsentOrPresent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		fixture   string
		optionals bool
	}{
		{name: "optional blocks absent", fixture: "testdata/valid-minimal.yaml", optionals: false},
		{name: "optional blocks present", fixture: "testdata/valid-complete.yaml", optionals: true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			contract := decodeFixture(t, tc.fixture)
			if contract.SchemaVersion != SchemaVersion {
				t.Fatalf("SchemaVersion = %q, want %q", contract.SchemaVersion, SchemaVersion)
			}
			if contract.Status != StatusReady {
				t.Fatalf("Status = %q, want %q", contract.Status, StatusReady)
			}
			if got := contract.Ledger != nil; got != tc.optionals {
				t.Fatalf("Ledger present = %t, want %t", got, tc.optionals)
			}
			if got := contract.Overview != nil; got != tc.optionals {
				t.Fatalf("Overview present = %t, want %t", got, tc.optionals)
			}
			if got := contract.Diagnostics != nil; got != tc.optionals {
				t.Fatalf("Diagnostics present = %t, want %t", got, tc.optionals)
			}
			if got := contract.StructuralAcceptance != nil; got != tc.optionals {
				t.Fatalf("StructuralAcceptance present = %t, want %t", got, tc.optionals)
			}
		})
	}
}

func TestDecodeYAMLAndJSONProduceEquivalentCanonicalIdentity(t *testing.T) {
	t.Parallel()

	yamlContract := decodeFixture(t, "testdata/valid-minimal.yaml")
	jsonContract := decodeFixture(t, "testdata/valid-minimal.json")
	yamlCanonical := mustCanonicalize(t, yamlContract)
	jsonCanonical := mustCanonicalize(t, jsonContract)

	if !bytes.Equal(yamlCanonical.JSON, jsonCanonical.JSON) {
		t.Fatalf("canonical JSON differs:\nYAML: %s\nJSON: %s", yamlCanonical.JSON, jsonCanonical.JSON)
	}
	if yamlCanonical.Fingerprint != jsonCanonical.Fingerprint {
		t.Fatalf("fingerprint differs: %q != %q", yamlCanonical.Fingerprint, jsonCanonical.Fingerprint)
	}
	assertSHA256(t, yamlCanonical.Fingerprint)
	if !json.Valid(yamlCanonical.JSON) {
		t.Fatalf("Canonicalize() returned invalid JSON: %q", yamlCanonical.JSON)
	}
}

func TestDecodeRejectsUnknownAndDuplicateKeys(t *testing.T) {
	t.Parallel()

	base := readFixture(t, "testdata/valid-minimal.yaml")
	jsonBase := string(readFixture(t, "testdata/valid-minimal.json"))
	cases := []struct {
		name    string
		payload []byte
		path    string
	}{
		{
			name:    "unknown top-level key",
			payload: append(append([]byte(nil), base...), []byte("\nprivate_prompt: "+promptCanary+"\n")...),
			path:    "private_prompt",
		},
		{
			name: "unknown nested key",
			payload: []byte(strings.Replace(string(base),
				"  changed:\n", "  private_output: "+outputCanary+"\n  changed:\n", 1)),
			path: "artifacts.private_output",
		},
		{
			name:    "duplicate top-level key",
			payload: append(append([]byte(nil), base...), []byte("\nstatus: blocked\n")...),
			path:    "status",
		},
		{
			name: "duplicate nested key",
			payload: []byte(strings.Replace(string(base),
				"  changed:\n", "  changed:\n    - duplicate\n  changed:\n", 1)),
			path: "artifacts.changed",
		},
		{
			name: "unknown JSON key",
			payload: []byte(strings.Replace(jsonBase,
				`"status": "ready",`, `"status": "ready", "private_prompt": "`+promptCanary+`",`, 1)),
			path: "private_prompt",
		},
		{
			name: "duplicate JSON key",
			payload: []byte(strings.Replace(jsonBase,
				`"status": "ready",`, `"status": "ready", "status": "blocked",`, 1)),
			path: "status",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Decode(bytes.NewReader(tc.payload))
			diagnostic := requireDiagnostic(t, err)
			if diagnostic.Path == "" || !strings.Contains(diagnostic.Path, tc.path) {
				t.Fatalf("diagnostic path = %q, want path containing %q", diagnostic.Path, tc.path)
			}
			assertSanitized(t, diagnostic, secretCanary, promptCanary, outputCanary)
		})
	}
}

func TestDecodeAcceptsEveryDocumentedStatus(t *testing.T) {
	t.Parallel()

	base := string(readFixture(t, "testdata/valid-minimal.yaml"))
	for _, status := range []string{
		"ready", "implemented", "validated", "delivery_pending", "sync_pending",
		"blocked", "escalated", "delivered", "closed",
	} {
		status := status
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			payload := strings.Replace(base, "status: ready", "status: "+status, 1)
			if _, err := Decode(strings.NewReader(payload)); err != nil {
				t.Fatalf("Decode(status=%s) error = %v", status, err)
			}
		})
	}
}

func TestDecodeRejectsAnchorsAliasesAndCustomTags(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		payload string
	}{
		{
			name: "anchor",
			payload: strings.Replace(string(readFixture(t, "testdata/valid-minimal.yaml")),
				"schema_version: result-contract/v1", "schema_version: &schema result-contract/v1", 1),
		},
		{
			name: "alias",
			payload: strings.Replace(string(readFixture(t, "testdata/valid-minimal.yaml")),
				"schema_version: result-contract/v1\nstatus: ready", "schema_version: &schema result-contract/v1\nstatus: *schema", 1),
		},
		{
			name: "custom tag",
			payload: strings.Replace(string(readFixture(t, "testdata/valid-minimal.yaml")),
				"executive_summary: Slice A listo para validar.", "executive_summary: !private "+secretCanary, 1),
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Decode(strings.NewReader(tc.payload))
			diagnostic := requireDiagnostic(t, err)
			assertSanitized(t, diagnostic, secretCanary)
		})
	}
}

func TestDecodeRejectsInvalidUTF8AndOversizedInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		payload []byte
	}{
		{name: "invalid UTF-8", payload: append(readFixture(t, "testdata/valid-minimal.yaml"), 0xff)},
		{name: "oversized", payload: bytes.Repeat([]byte("x"), testInputLimit+1)},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Decode(bytes.NewReader(tc.payload))
			diagnostic := requireDiagnostic(t, err)
			if diagnostic.Code == "" || diagnostic.Recovery == "" {
				t.Fatalf("diagnostic must include code and recovery: %#v", diagnostic)
			}
			assertSanitized(t, diagnostic, string(tc.payload))
		})
	}
}

func TestDecodeRejectsUnsupportedSchemaAndEnums(t *testing.T) {
	t.Parallel()

	base := string(readFixture(t, "testdata/valid-complete.yaml"))
	cases := []struct {
		name string
		old  string
		new  string
		path string
	}{
		{name: "schema", old: "schema_version: result-contract/v1", new: "schema_version: result-contract/v2", path: "schema_version"},
		{name: "status", old: "status: ready", new: "status: pretending", path: "status"},
		{name: "ledger status", old: "  status: not_applicable", new: "  status: uploaded", path: "ledger.status"},
		{name: "evidence result", old: "      result: not_run", new: "      result: probably", path: "evidence.commands.result"},
		{name: "next owner", old: "  owner: implementer", new: "  owner: stranger", path: "next_recommended.owner"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			payload := strings.Replace(base, tc.old, tc.new, 1)
			if payload == base {
				t.Fatalf("invalid test fixture: %q not found", tc.old)
			}
			_, err := Decode(strings.NewReader(payload))
			diagnostic := requireDiagnostic(t, err)
			if !strings.Contains(diagnostic.Path, tc.path) {
				t.Fatalf("diagnostic path = %q, want containing %q", diagnostic.Path, tc.path)
			}
			assertSanitized(t, diagnostic, tc.new)
		})
	}
}

func TestDiagnosticNeverLeaksRejectedCanaries(t *testing.T) {
	t.Parallel()

	payloads := [][]byte{
		[]byte("schema_version: result-contract/v1\nstatus: ready\nsecret: " + secretCanary + "\n"),
		[]byte("schema_version: result-contract/v1\nstatus: " + promptCanary + "\n"),
		[]byte("schema_version: result-contract/v1\nstatus: ready\nexecutive_summary: !private " + outputCanary + "\n"),
	}
	for i, payload := range payloads {
		_, err := Decode(bytes.NewReader(payload))
		diagnostic := requireDiagnostic(t, err)
		assertSanitized(t, diagnostic, secretCanary, promptCanary, outputCanary)
		if diagnostic.Code == "" || diagnostic.Recovery == "" {
			t.Fatalf("case %d: expected structured code and recovery: %#v", i, diagnostic)
		}
	}
}

func decodeFixture(t *testing.T, path string) Contract {
	t.Helper()
	contract, err := Decode(bytes.NewReader(readFixture(t, path)))
	if err != nil {
		t.Fatalf("Decode(%s) error = %v", path, err)
	}
	return contract
}

func readFixture(t *testing.T, path string) []byte {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%s) error = %v", path, err)
	}
	return body
}

func mustCanonicalize(t *testing.T, contract Contract) Canonical {
	t.Helper()
	canonical, err := Canonicalize(contract)
	if err != nil {
		t.Fatalf("Canonicalize() error = %v", err)
	}
	return canonical
}

func requireDiagnostic(t *testing.T, err error) *DiagnosticError {
	t.Helper()
	if err == nil {
		t.Fatal("Decode() expected error")
	}
	var diagnostic *DiagnosticError
	if !errors.As(err, &diagnostic) {
		t.Fatalf("error type = %T, want *DiagnosticError: %v", err, err)
	}
	return diagnostic
}

func assertSanitized(t *testing.T, diagnostic *DiagnosticError, canaries ...string) {
	t.Helper()
	body, err := json.Marshal(diagnostic)
	if err != nil {
		t.Fatalf("json.Marshal(diagnostic) error = %v", err)
	}
	observed := diagnostic.Error() + "\n" + string(body)
	for _, canary := range canaries {
		if canary != "" && strings.Contains(observed, canary) {
			t.Fatalf("diagnostic leaked rejected input %q: %s", canary, observed)
		}
	}
}

func assertSHA256(t *testing.T, value string) {
	t.Helper()
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 {
		t.Fatalf("fingerprint = %q, want 64-char SHA-256 hex", value)
	}
}
