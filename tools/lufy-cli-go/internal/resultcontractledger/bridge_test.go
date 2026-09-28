package resultcontractledger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/resultcontract"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

func TestBridgeRecordsAndDeduplicatesContentFreeCorrelation(t *testing.T) {
	t.Parallel()

	target := t.TempDir()
	store, err := runledger.NewFileStore(target, runledger.Options{})
	if err != nil {
		t.Fatal(err)
	}
	bridge := New(store)
	request := bridgeRequest(t)
	first := bridge.Record(t.Context(), request)
	if first.Status != StatusRecorded || first.EventID == "" {
		t.Fatalf("first result = %#v", first)
	}
	second := bridge.Record(t.Context(), request)
	if second.Status != StatusDuplicateNoop || second.EventID != first.EventID {
		t.Fatalf("duplicate result = %#v, first=%#v", second, first)
	}

	events, err := store.LoadRun(t.Context(), request.RunID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events = %d, want one durable correlation", len(events))
	}
	event := events[0]
	canonical, err := resultcontract.Canonicalize(request.Contract)
	if err != nil {
		t.Fatal(err)
	}
	if event.TaskRef != resultcontract.SchemaVersion || event.Checkpoint == nil ||
		event.Checkpoint.Status != string(request.Contract.Status) || event.Checkpoint.Gate != string(request.Decision.Status) {
		t.Fatalf("event lost allow-listed decision metadata: %#v", event)
	}
	if !eventHasEvidence(event, "result_contract", string(request.Contract.Status), canonical.Fingerprint) {
		t.Fatalf("event missing contract fingerprint correlation %s: %#v", canonical.Fingerprint, event)
	}
	if !eventHasEvidence(event, "transition_decision", string(request.Decision.Status), request.Decision.DecisionID) {
		t.Fatalf("event missing decision correlation: %#v", event)
	}
	if !eventHasEvidence(event, "transition_version", "2", "") {
		t.Fatalf("event missing transition version: %#v", event)
	}
	for _, digest := range request.ReferenceDigests {
		if !eventContainsDigest(event, digest) {
			t.Fatalf("event missing reference digest %s: %#v", digest, event)
		}
	}
}

func TestBridgeMapsSameKeyDifferentContractToConflictWithoutOverwrite(t *testing.T) {
	t.Parallel()

	target := t.TempDir()
	store, err := runledger.NewFileStore(target, runledger.Options{})
	if err != nil {
		t.Fatal(err)
	}
	bridge := New(store)
	request := bridgeRequest(t)
	first := bridge.Record(t.Context(), request)
	if first.Status != StatusRecorded {
		t.Fatalf("first result = %#v", first)
	}
	request.Contract.ExecutiveSummary = "semantic change for conflict"
	canonical, err := resultcontract.Canonicalize(request.Contract)
	if err != nil {
		t.Fatal(err)
	}
	request.Decision.NextFingerprint = canonical.Fingerprint
	conflict := bridge.Record(t.Context(), request)
	if conflict.Status != StatusConflict || conflict.EventID != first.EventID {
		t.Fatalf("conflict result = %#v, first=%#v", conflict, first)
	}
	events, err := store.LoadRun(t.Context(), request.RunID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("conflict overwrote or appended history: %d events", len(events))
	}
}

func TestBridgeUnavailableRemainsObservationAndDoesNotPromoteDecision(t *testing.T) {
	t.Parallel()

	decision := bridgeRequest(t).Decision
	bridge := New(unavailableAppender{})
	request := bridgeRequest(t)
	result := bridge.Record(t.Context(), request)
	if result.Status != StatusUnavailable || result.Warning == "" {
		t.Fatalf("unavailable result = %#v", result)
	}
	if strings.Contains(result.Warning, "PRIVATE_STORAGE_ERROR_CANARY") {
		t.Fatalf("unavailable warning leaked storage error: %s", result.Warning)
	}
	if !reflect.DeepEqual(request.Decision, decision) || request.Decision.Status != resultcontract.DecisionAccepted {
		t.Fatalf("ledger outcome promoted or mutated workflow truth: before=%#v after=%#v", decision, request.Decision)
	}
	resultType := reflect.TypeOf(result)
	for _, forbidden := range []string{"Contract", "Workflow", "Gate", "Delivered", "Closed"} {
		for index := 0; index < resultType.NumField(); index++ {
			if strings.Contains(resultType.Field(index).Name, forbidden) {
				t.Fatalf("bridge result exposes workflow authority field %q", resultType.Field(index).Name)
			}
		}
	}
}

func TestBridgeNeverPersistsContractPromptPathOrOutputCanaries(t *testing.T) {
	t.Parallel()

	const summaryCanary = "SUMMARY_CANARY_private"
	const promptCanary = "PROMPT_CANARY_ignore_previous"
	const pathCanary = "/Users/private/PATH_CANARY/project"
	const outputCanary = "OUTPUT_CANARY_command_stdout"
	target := t.TempDir()
	store, err := runledger.NewFileStore(target, runledger.Options{})
	if err != nil {
		t.Fatal(err)
	}
	request := bridgeRequest(t)
	request.Contract.ExecutiveSummary = summaryCanary + " " + promptCanary
	request.Contract.Artifacts.Changed = []string{pathCanary}
	request.Contract.Evidence.Commands[0].Notes = outputCanary
	canonical, err := resultcontract.Canonicalize(request.Contract)
	if err != nil {
		t.Fatal(err)
	}
	request.Decision.NextFingerprint = canonical.Fingerprint
	result := New(store).Record(t.Context(), request)
	if result.Status != StatusRecorded {
		t.Fatalf("record result = %#v", result)
	}
	body := readBridgeTree(t, filepath.Join(target, ".lufy", "runtime"))
	for _, canary := range []string{summaryCanary, promptCanary, pathCanary, outputCanary} {
		if strings.Contains(body, canary) {
			t.Fatalf("ledger bridge persisted canary %q", canary)
		}
	}
	if !strings.Contains(body, canonical.Fingerprint) {
		t.Fatalf("ledger bridge omitted safe fingerprint: %s", body)
	}
}

func bridgeRequest(t *testing.T) RecordRequest {
	t.Helper()
	fixture := filepath.Join("..", "resultcontract", "testdata", "valid-minimal.yaml")
	body, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	contract, err := resultcontract.Decode(bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	contract.Status = resultcontract.StatusValidated
	contract.Evidence.Commands[0].Result = resultcontract.EvidencePassed
	canonical, err := resultcontract.Canonicalize(contract)
	if err != nil {
		t.Fatal(err)
	}
	return RecordRequest{
		RunID: "run-result-bridge", IdempotencyKey: "result-bridge-key",
		Contract: contract,
		Decision: resultcontract.TransitionDecision{
			Status: resultcontract.DecisionAccepted, DecisionID: testBridgeDigest("decision"),
			NextVersion: 2, NextFingerprint: canonical.Fingerprint,
		},
		ReferenceDigests: []string{testBridgeDigest("artifact-ref"), testBridgeDigest("evidence-ref")},
	}
}

func eventContainsDigest(event runledger.Event, digest string) bool {
	body, err := json.Marshal(event)
	return err == nil && strings.Contains(string(body), digest)
}

func eventHasEvidence(event runledger.Event, category, result, digest string) bool {
	for _, evidence := range event.EvidenceRefs {
		if evidence.Category == category && evidence.Result == result && evidence.PathSHA256 == digest {
			return true
		}
	}
	return false
}

func testBridgeDigest(value string) string { return runledger.HashReference(value) }

func readBridgeTree(t *testing.T, root string) string {
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

type unavailableAppender struct{}

func (unavailableAppender) Append(context.Context, runledger.AppendRequest) (runledger.AppendResult, error) {
	return runledger.AppendResult{}, errors.New("PRIVATE_STORAGE_ERROR_CANARY")
}
