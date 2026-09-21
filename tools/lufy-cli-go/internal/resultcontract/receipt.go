package resultcontract

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

type ReceiptOutcome string

const (
	ReceiptRecorded      ReceiptOutcome = "recorded"
	ReceiptDuplicateNoop ReceiptOutcome = "duplicate_noop"
	ReceiptConflict      ReceiptOutcome = "conflict"
)

type TransitionReceipt struct {
	KeyHash           string
	IntentFingerprint string
	DecisionID        string
	DecisionStatus    DecisionStatus
	NextVersion       uint64
	NextFingerprint   string
}

type ReceiptStore interface {
	Record(TransitionReceipt) (ReceiptOutcome, TransitionReceipt, error)
}

func recordDecision(store ReceiptStore, intent TransitionIntent, nextFingerprint string, decision TransitionDecision) TransitionDecision {
	intentFingerprint, err := fingerprintIntent(intent, nextFingerprint)
	if err != nil {
		return rejectedTransition("intent_fingerprint_failed", "reconstruir el intent tipado")
	}
	candidate := TransitionReceipt{
		KeyHash:           digestString(intent.IdempotencyKey),
		IntentFingerprint: intentFingerprint,
		DecisionID:        digestString("decision:" + intentFingerprint),
		DecisionStatus:    DecisionAccepted,
		NextVersion:       decision.NextVersion,
		NextFingerprint:   decision.NextFingerprint,
	}
	outcome, stored, err := store.Record(candidate)
	if err != nil {
		return rejectedTransition("receipt_unavailable", "reintentar cuando el receipt store esté disponible")
	}
	switch outcome {
	case ReceiptRecorded:
		decision.DecisionID = candidate.DecisionID
		return decision
	case ReceiptDuplicateNoop:
		if stored.KeyHash != candidate.KeyHash || stored.IntentFingerprint != candidate.IntentFingerprint ||
			stored.DecisionStatus != DecisionAccepted || stored.DecisionID == "" ||
			stored.NextVersion == 0 || !fingerprintPattern.MatchString(stored.NextFingerprint) {
			return conflictTransition("receipt_invalid", "verificar o reconstruir el receipt store")
		}
		decision.Status = DecisionDuplicateNoop
		decision.DecisionID = stored.DecisionID
		decision.NextVersion = stored.NextVersion
		decision.NextFingerprint = stored.NextFingerprint
		return decision
	case ReceiptConflict:
		return conflictTransition("idempotency_conflict", "usar una idempotency key nueva o reintentar el intent original")
	default:
		return rejectedTransition("receipt_invalid", "verificar o reconstruir el receipt store")
	}
}

func fingerprintIntent(intent TransitionIntent, nextFingerprint string) (string, error) {
	type joinPair struct {
		RunID       string `json:"run_id"`
		Fingerprint string `json:"fingerprint"`
	}
	type safeJoin struct {
		Children                   []joinPair `json:"children,omitempty"`
		GroupedEvidenceFingerprint string     `json:"grouped_evidence_fingerprint,omitempty"`
	}
	type safeLease struct {
		TokenDigest string `json:"token_digest,omitempty"`
		ExpiresAt   string `json:"expires_at,omitempty"`
	}
	payload := struct {
		SchemaVersion       string    `json:"schema_version"`
		TransitionID        string    `json:"transition_id"`
		ExpectedVersion     uint64    `json:"expected_version"`
		PreviousFingerprint string    `json:"previous_fingerprint"`
		Role                string    `json:"role"`
		OwnerRef            string    `json:"owner_ref"`
		Lease               safeLease `json:"lease"`
		Join                safeJoin  `json:"join"`
		NextFingerprint     string    `json:"next_fingerprint"`
	}{
		SchemaVersion: intent.SchemaVersion, TransitionID: intent.TransitionID,
		ExpectedVersion: intent.ExpectedVersion, PreviousFingerprint: intent.PreviousFingerprint,
		Role: intent.Actor.Role, OwnerRef: intent.Actor.OwnerRef, NextFingerprint: nextFingerprint,
	}
	if intent.Lease != nil {
		payload.Lease.TokenDigest = intent.Lease.TokenDigest
		payload.Lease.ExpiresAt = intent.Lease.ExpiresAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00")
	}
	if intent.Join != nil {
		for index, runID := range intent.Join.RequiredRunIDs {
			fingerprint := ""
			if index < len(intent.Join.TerminalContractFingerprints) {
				fingerprint = intent.Join.TerminalContractFingerprints[index]
			}
			payload.Join.Children = append(payload.Join.Children, joinPair{RunID: runID, Fingerprint: fingerprint})
		}
		sort.Slice(payload.Join.Children, func(i, j int) bool { return payload.Join.Children[i].RunID < payload.Join.Children[j].RunID })
		payload.Join.GroupedEvidenceFingerprint = intent.Join.GroupedEvidenceFingerprint
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return digestBytes(body), nil
}

func digestString(value string) string {
	return digestBytes([]byte(value))
}

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}
