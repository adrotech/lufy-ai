package resultcontract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

const fuzzPrivateCanary = "PRIVATE_FUZZ_CANARY_do_not_echo"

type ingressFuzzSeed struct {
	name      string
	payload   []byte
	validFor  string
	rejectAll bool
}

func TestResultContractIngressFuzzSeeds(t *testing.T) {
	seeds, err := resultContractIngressFuzzSeeds()
	if err != nil {
		t.Fatal(err)
	}
	for _, seed := range seeds {
		seed := seed
		t.Run(seed.name, func(t *testing.T) {
			errorsByDecoder := exerciseResultContractIngress(t, seed.payload)
			if seed.validFor != "" && errorsByDecoder[seed.validFor] != nil {
				t.Fatalf("%s rejected valid seed: %v", seed.validFor, errorsByDecoder[seed.validFor])
			}
			if seed.rejectAll {
				for decoder, decodeErr := range errorsByDecoder {
					if decodeErr == nil {
						t.Fatalf("%s accepted hostile seed", decoder)
					}
				}
			}
		})
	}
}

func FuzzResultContractIngressBoundedAndSanitized(f *testing.F) {
	seeds, err := resultContractIngressFuzzSeeds()
	if err != nil {
		f.Fatal(err)
	}
	for _, seed := range seeds {
		f.Add(seed.payload)
	}

	f.Fuzz(func(t *testing.T, payload []byte) {
		exerciseResultContractIngress(t, payload)
	})
}

func exerciseResultContractIngress(t *testing.T, payload []byte) map[string]error {
	t.Helper()
	errorsByDecoder := map[string]error{}
	_, errorsByDecoder["Decode"] = Decode(bytes.NewReader(payload))
	_, errorsByDecoder["NormalizeLegacy"] = NormalizeLegacy(bytes.NewReader(payload))
	_, errorsByDecoder["DecodeTransitionRequest"] = DecodeTransitionRequest(bytes.NewReader(payload))

	for decoder, decodeErr := range errorsByDecoder {
		if len(payload) > maxInputBytes && decodeErr == nil {
			t.Fatalf("%s accepted %d bytes above the %d-byte bound", decoder, len(payload), maxInputBytes)
		}
		if decodeErr != nil && bytes.Contains(payload, []byte(fuzzPrivateCanary)) && strings.Contains(decodeErr.Error(), fuzzPrivateCanary) {
			t.Fatalf("%s diagnostic echoed private canary: %v", decoder, decodeErr)
		}
	}
	return errorsByDecoder
}

func resultContractIngressFuzzSeeds() ([]ingressFuzzSeed, error) {
	validYAML, err := os.ReadFile("testdata/valid-minimal.yaml")
	if err != nil {
		return nil, fmt.Errorf("read valid YAML seed: %w", err)
	}
	validJSON, err := os.ReadFile("testdata/valid-minimal.json")
	if err != nil {
		return nil, fmt.Errorf("read valid JSON seed: %w", err)
	}
	legacyYAML, err := os.ReadFile("testdata/legacy-minimal.yaml")
	if err != nil {
		return nil, fmt.Errorf("read legacy seed: %w", err)
	}
	transitionJSON, err := validTransitionRequestFuzzSeed(validYAML)
	if err != nil {
		return nil, err
	}

	unknown := append(append([]byte(nil), validYAML...), []byte("\nprivate_prompt: "+fuzzPrivateCanary+"\n")...)
	duplicate := append(append([]byte(nil), validYAML...), []byte("\nstatus: "+fuzzPrivateCanary+"\n")...)
	invalidUTF8 := append(append(append([]byte(nil), validYAML...), 0xff), []byte(fuzzPrivateCanary)...)
	oversized := append([]byte("private_prompt: "+fuzzPrivateCanary+"\n"), bytes.Repeat([]byte("x"), maxInputBytes+1)...)

	return []ingressFuzzSeed{
		{name: "valid result YAML", payload: validYAML, validFor: "Decode"},
		{name: "valid result JSON", payload: validJSON, validFor: "Decode"},
		{name: "valid legacy YAML", payload: legacyYAML, validFor: "NormalizeLegacy"},
		{name: "valid transition JSON", payload: transitionJSON, validFor: "DecodeTransitionRequest"},
		{name: "malformed", payload: []byte("schema_version: [" + fuzzPrivateCanary), rejectAll: true},
		{name: "oversized", payload: oversized, rejectAll: true},
		{name: "invalid UTF-8", payload: invalidUTF8, rejectAll: true},
		{name: "unknown field", payload: unknown, rejectAll: true},
		{name: "duplicate key", payload: duplicate, rejectAll: true},
	}, nil
}

func validTransitionRequestFuzzSeed(validContract []byte) ([]byte, error) {
	prior, err := Decode(bytes.NewReader(validContract))
	if err != nil {
		return nil, fmt.Errorf("decode transition prior seed: %w", err)
	}
	prior.Status = StatusReady
	canonical, err := Canonicalize(prior)
	if err != nil {
		return nil, fmt.Errorf("canonicalize transition prior seed: %w", err)
	}
	state, err := DeriveState(StatusReady, nil)
	if err != nil {
		return nil, fmt.Errorf("derive transition prior state: %w", err)
	}
	next := prior
	next.Status = StatusImplemented
	next.ExecutiveSummary = "Fuzz transition seed."
	request := TransitionRequest{
		SchemaVersion: TransitionRequestSchemaVersion,
		Prior: TransitionState{
			Version: 1, Fingerprint: canonical.Fingerprint, Contract: prior, State: state, RunID: "run-fuzz-result",
		},
		Intent: TransitionIntent{
			SchemaVersion: TransitionSchemaVersion, TransitionID: "transition-fuzz-result",
			IdempotencyKey: "fuzz-result-idempotency-key", ExpectedVersion: 1,
			PreviousFingerprint: canonical.Fingerprint,
			Actor:               TransitionActor{Role: "implementer", OwnerRef: strings.Repeat("a", 64)},
			NextContract:        next,
		},
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal transition seed: %w", err)
	}
	return body, nil
}
