package runledger

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestCanonicalFingerprintIsStableAndExcludesGeneratedFields(t *testing.T) {
	draft := validDraft("run-root", KindCheckpoint)
	draft.EventID = "evt_first"

	first, err := CanonicalFingerprint(draft)
	if err != nil {
		t.Fatalf("CanonicalFingerprint() error = %v", err)
	}
	draft.EventID = "evt_retry"
	second, err := CanonicalFingerprint(draft)
	if err != nil {
		t.Fatalf("CanonicalFingerprint() retry error = %v", err)
	}
	if first != second {
		t.Fatalf("generated event id changed fingerprint: %q != %q", first, second)
	}

	draft.Checkpoint.Status = "blocked"
	changed, err := CanonicalFingerprint(draft)
	if err != nil {
		t.Fatalf("CanonicalFingerprint() changed error = %v", err)
	}
	if changed == first {
		t.Fatal("semantic checkpoint change did not change fingerprint")
	}
}

func TestDecodeDraftRejectsUnknownSensitiveFieldsWithoutEchoingValues(t *testing.T) {
	const canary = "SUPER_SECRET_PROMPT_VALUE"
	for _, field := range []string{"prompt", "response", "message", "transcript_path", "tool_arguments", "diff", "file_content", "secret"} {
		t.Run(field, func(t *testing.T) {
			payload := `{"run_id":"run-root","kind":"start","` + field + `":"` + canary + `"}`
			_, err := DecodeDraft(strings.NewReader(payload))
			if err == nil {
				t.Fatal("DecodeDraft() expected unknown-field error")
			}
			if strings.Contains(err.Error(), canary) {
				t.Fatalf("error leaked rejected value: %v", err)
			}
		})
	}
}

func TestValidateDraftRejectsInvalidReferencesWithSanitizedErrors(t *testing.T) {
	draft := validDraft("run-root", KindEvidence)
	draft.ArtifactRefs = []ArtifactRef{{
		Kind:          "source",
		PathSHA256:    strings.Repeat("a", 63),
		ContentSHA256: strings.Repeat("b", 64),
	}}

	err := ValidateDraft(draft)
	if err == nil {
		t.Fatal("ValidateDraft() expected digest validation error")
	}
	if strings.Contains(err.Error(), strings.Repeat("a", 63)) {
		t.Fatalf("error leaked invalid digest: %v", err)
	}
}

func TestDecodeDraftAcceptsOnlyOneBoundedObject(t *testing.T) {
	draft := validDraft("run-decode", KindStart)
	body, err := json.Marshal(draft)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeDraft(strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("DecodeDraft() valid error = %v", err)
	}
	if decoded.RunID != draft.RunID || decoded.Kind != draft.Kind {
		t.Fatalf("decoded = %#v", decoded)
	}

	for name, payload := range map[string]string{
		"malformed": `{"run_id":`,
		"trailing":  string(body) + ` {}`,
		"oversized": strings.Repeat("x", maxInputJSONSize+1),
		"empty":     "",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeDraft(strings.NewReader(payload)); err == nil {
				t.Fatal("DecodeDraft() expected error")
			}
		})
	}
}

func TestValidateDraftCoversContractBoundaries(t *testing.T) {
	negative := int64(-1)
	negativeCost := -0.01
	cases := []struct {
		name   string
		mutate func(*EventDraft)
	}{
		{"missing run", func(d *EventDraft) { d.RunID = "" }},
		{"invalid event id", func(d *EventDraft) { d.EventID = "bad id" }},
		{"same parent", func(d *EventDraft) { d.ParentRunID = d.RunID }},
		{"invalid cause", func(d *EventDraft) { d.CausedByEventID = "?" }},
		{"invalid kind", func(d *EventDraft) { d.Kind = "message" }},
		{"missing adapter", func(d *EventDraft) { d.Source.Adapter = "" }},
		{"invalid event name", func(d *EventDraft) { d.Source.EventName = "Stop event" }},
		{"invalid session hash", func(d *EventDraft) { d.Source.SessionRefHash = "not-a-hash" }},
		{"invalid turn hash", func(d *EventDraft) { d.Source.TurnRefHash = "not-a-hash" }},
		{"invalid agent hash", func(d *EventDraft) { d.Source.AgentRefHash = "not-a-hash" }},
		{"invalid agent type", func(d *EventDraft) { d.Source.AgentType = "test writer" }},
		{"invalid task ref", func(d *EventDraft) { d.TaskRef = "private task text" }},
		{"too many artifacts", func(d *EventDraft) { d.ArtifactRefs = make([]ArtifactRef, maxReferences+1) }},
		{"missing artifact kind", func(d *EventDraft) { d.ArtifactRefs = []ArtifactRef{{PathSHA256: strings.Repeat("a", 64)}} }},
		{"invalid artifact content hash", func(d *EventDraft) {
			d.ArtifactRefs = []ArtifactRef{{Kind: "source", PathSHA256: strings.Repeat("a", 64), ContentSHA256: "bad"}}
		}},
		{"too many evidence refs", func(d *EventDraft) { d.EvidenceRefs = make([]EvidenceRef, maxReferences+1) }},
		{"missing evidence category", func(d *EventDraft) { d.EvidenceRefs = []EvidenceRef{{Result: "passed"}} }},
		{"missing evidence result", func(d *EventDraft) { d.EvidenceRefs = []EvidenceRef{{Category: "test"}} }},
		{"invalid evidence hash", func(d *EventDraft) {
			d.EvidenceRefs = []EvidenceRef{{Category: "test", Result: "passed", PathSHA256: "bad"}}
		}},
		{"invalid checkpoint status", func(d *EventDraft) { d.Checkpoint = &Checkpoint{Status: "unknown"} }},
		{"invalid checkpoint gate", func(d *EventDraft) { d.Checkpoint = &Checkpoint{Status: "ready", Gate: "not valid"} }},
		{"invalid checkpoint owner", func(d *EventDraft) { d.Checkpoint = &Checkpoint{Status: "ready", NextOwner: "not valid"} }},
		{"invalid metrics availability", func(d *EventDraft) { d.Metrics = &Metrics{Availability: "maybe"} }},
		{"negative duration", func(d *EventDraft) { d.Metrics = &Metrics{Availability: "available", DurationMillis: &negative} }},
		{"negative tokens", func(d *EventDraft) { d.Metrics = &Metrics{Availability: "available", Tokens: &negative} }},
		{"negative cost", func(d *EventDraft) { d.Metrics = &Metrics{Availability: "available", CostUSD: &negativeCost} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			draft := validDraft("run-boundary", KindCheckpoint)
			tc.mutate(&draft)
			if err := ValidateDraft(draft); err == nil {
				t.Fatal("ValidateDraft() expected error")
			}
		})
	}
}

func TestValidateEventRejectsInvalidEnvelopeFields(t *testing.T) {
	draft := validDraft("run-event", KindStart)
	fingerprint, err := CanonicalFingerprint(draft)
	if err != nil {
		t.Fatal(err)
	}
	valid := eventFromDraft(draft, "evt_valid_event", 1, 1, time.Now().UTC(), HashReference("key"), fingerprint)
	if err := ValidateEvent(valid); err != nil {
		t.Fatalf("ValidateEvent() valid error = %v", err)
	}
	cases := []struct {
		name   string
		mutate func(*Event)
	}{
		{"schema", func(e *Event) { e.SchemaVersion = "v0" }},
		{"clock", func(e *Event) { e.LamportClock = 0 }},
		{"observed", func(e *Event) { e.ObservedAt = time.Time{} }},
		{"key hash", func(e *Event) { e.IdempotencyKeyHash = "bad" }},
		{"fingerprint", func(e *Event) { e.Fingerprint = "bad" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			event := valid
			tc.mutate(&event)
			if err := ValidateEvent(event); err == nil {
				t.Fatal("ValidateEvent() expected error")
			}
		})
	}
}

func validDraft(runID string, kind EventKind) EventDraft {
	return EventDraft{
		RunID:      runID,
		Kind:       kind,
		OccurredAt: time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC),
		Source: Source{
			Adapter:   "codex",
			EventName: "Stop",
		},
		TaskRef: "task-2.1",
		Checkpoint: &Checkpoint{
			Status:    "implemented",
			Gate:      "implementation",
			NextOwner: "validator",
		},
	}
}
