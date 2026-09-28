package resultcontract

import (
	"regexp"
	"strings"
	"time"
)

const TransitionSchemaVersion = "result-transition/v1"

type DecisionStatus string

const (
	DecisionAccepted      DecisionStatus = "accepted"
	DecisionDuplicateNoop DecisionStatus = "duplicate_noop"
	DecisionRejected      DecisionStatus = "rejected"
	DecisionConflict      DecisionStatus = "conflict"
)

type TransitionActor struct {
	Role     string `json:"role" yaml:"role"`
	OwnerRef string `json:"owner_ref" yaml:"owner_ref"`
}

type TransitionIntent struct {
	SchemaVersion       string           `json:"schema_version" yaml:"schema_version"`
	TransitionID        string           `json:"transition_id" yaml:"transition_id"`
	IdempotencyKey      string           `json:"idempotency_key" yaml:"idempotency_key"`
	ExpectedVersion     uint64           `json:"expected_version" yaml:"expected_version"`
	PreviousFingerprint string           `json:"previous_fingerprint" yaml:"previous_fingerprint"`
	Actor               TransitionActor  `json:"actor" yaml:"actor"`
	Lease               *TransitionLease `json:"lease,omitempty" yaml:"lease,omitempty"`
	Join                *TransitionJoin  `json:"join,omitempty" yaml:"join,omitempty"`
	NextContract        Contract         `json:"next_contract" yaml:"next_contract"`
}

type TransitionState struct {
	Version            uint64               `json:"version" yaml:"version"`
	Fingerprint        string               `json:"fingerprint" yaml:"fingerprint"`
	Contract           Contract             `json:"contract" yaml:"contract"`
	State              StateVector          `json:"state" yaml:"state"`
	RunID              string               `json:"run_id" yaml:"run_id"`
	Ownership          *TransitionOwnership `json:"ownership,omitempty" yaml:"ownership,omitempty"`
	RequiredJoinRunIDs []string             `json:"required_join_run_ids,omitempty" yaml:"required_join_run_ids,omitempty"`
}

type TransitionPolicy struct {
	Roles    RolePolicy
	Evidence EvidencePolicy
	Now      func() time.Time
	Receipts ReceiptStore
	Joins    JoinResolver
}

type TransitionDecision struct {
	Status          DecisionStatus
	DecisionID      string
	Reason          string
	Recovery        string
	NextVersion     uint64
	NextFingerprint string
	State           StateVector
	MissingEvidence []EvidenceCategory
	NextOwner       string
}

var (
	transitionIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{1,127}$`)
	fingerprintPattern  = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

func EvaluateTransition(previous TransitionState, intent TransitionIntent, policy TransitionPolicy) TransitionDecision {
	if intent.SchemaVersion != TransitionSchemaVersion {
		return rejectedTransition("unsupported_transition_schema", "usar result-transition/v1")
	}
	if !isValidTransitionIntent(intent) {
		return rejectedTransition("invalid_transition", "completar identidad, actor y referencias requeridas")
	}
	if intent.ExpectedVersion != previous.Version {
		return conflictTransition("stale_version", "recargar el estado y reintentar sobre la versión actual")
	}
	if intent.PreviousFingerprint != previous.Fingerprint {
		return conflictTransition("stale_fingerprint", "recargar el estado y reintentar sobre el fingerprint actual")
	}
	if previous.Contract.Status == StatusClosed || previous.State.Terminal {
		return rejectedTransition("terminal_state", "crear un nuevo workflow; closed no admite transiciones")
	}
	canonical, err := Canonicalize(intent.NextContract)
	if err != nil {
		return rejectedTransition("invalid_contract", "corregir el next contract antes de reintentar")
	}
	if !IsAllowedTransition(previous.Contract.Status, intent.NextContract.Status) {
		return rejectedTransition("transition_not_allowed", "usar una transición declarada en la tabla de estados")
	}
	now := time.Now().UTC()
	if policy.Now != nil {
		now = policy.Now().UTC()
	}
	if ownershipDecision := evaluateOwnership(previous, intent, now); ownershipDecision != nil {
		return *ownershipDecision
	}

	claim := EvaluateClaim(intent.NextContract, intent.Actor.Role, policy.Roles, policy.Evidence)
	if !claim.Accepted {
		return TransitionDecision{
			Status: DecisionRejected, Reason: claim.Reason, Recovery: claim.Recovery,
			MissingEvidence: claim.MissingEvidence, NextOwner: claim.NextOwner,
		}
	}
	if recoveryTransition(previous.Contract.Status, intent.NextContract.Status) {
		if len(policy.Evidence.Recovery.Categories) == 0 {
			return rejectedTransition("recovery_policy_unavailable", "configurar evidencia de recuperación")
		}
		missing := missingEvidence(intent.NextContract, policy.Evidence.Recovery)
		if len(missing) > 0 {
			nextOwner := policy.Evidence.Recovery.NextOwner
			if nextOwner == "" {
				nextOwner = intent.Actor.Role
			}
			return TransitionDecision{
				Status: DecisionRejected, Reason: "recovery_evidence_missing",
				Recovery:        "agregar evidencia observable de recuperación",
				MissingEvidence: missing, NextOwner: nextOwner,
			}
		}
	}
	if joinDecision := evaluateJoin(previous, intent, policy.Joins); joinDecision != nil {
		return *joinDecision
	}

	nextState, err := DeriveState(intent.NextContract.Status, &previous.State)
	if err != nil {
		return rejectedTransition("invalid_state", "corregir el status y el estado previo")
	}
	intentFingerprint, err := fingerprintIntent(intent, canonical.Fingerprint)
	if err != nil {
		return rejectedTransition("intent_fingerprint_failed", "reconstruir el intent tipado")
	}
	decision := TransitionDecision{
		Status:          DecisionAccepted,
		DecisionID:      digestString("decision:" + intentFingerprint),
		NextVersion:     previous.Version + 1,
		NextFingerprint: canonical.Fingerprint,
		State:           nextState,
	}
	if policy.Receipts != nil {
		return recordDecision(policy.Receipts, intent, canonical.Fingerprint, decision)
	}
	return decision
}

func isValidTransitionIntent(intent TransitionIntent) bool {
	return transitionIDPattern.MatchString(intent.TransitionID) &&
		validIdempotencyKey(intent.IdempotencyKey) &&
		intent.Actor.Role != "" && intent.Actor.OwnerRef != "" &&
		fingerprintPattern.MatchString(intent.PreviousFingerprint)
}

func validIdempotencyKey(value string) bool {
	return value != "" && len(value) <= 512 && !strings.ContainsAny(value, "\r\n\x00")
}

func recoveryTransition(previous, next Status) bool {
	wasInterrupted := previous == StatusBlocked || previous == StatusEscalated
	remainsInterrupted := next == StatusBlocked || next == StatusEscalated
	return wasInterrupted && !remainsInterrupted
}

func rejectedTransition(reason, recovery string) TransitionDecision {
	return TransitionDecision{Status: DecisionRejected, Reason: reason, Recovery: recovery}
}

func conflictTransition(reason, recovery string) TransitionDecision {
	return TransitionDecision{Status: DecisionConflict, Reason: reason, Recovery: recovery}
}
