package runledger

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

const (
	SchemaVersion    = "lufy-run-event/v1"
	maxReferenceLen  = 128
	maxEventNameLen  = 96
	maxReferences    = 32
	maxInputJSONSize = 64 * 1024
)

var (
	idPattern     = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{1,127}$`)
	digestPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
	tokenPattern  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/-]*$`)
)

type EventKind string

const (
	KindStart      EventKind = "start"
	KindLink       EventKind = "link"
	KindCheckpoint EventKind = "checkpoint"
	KindEvidence   EventKind = "evidence"
	KindValidation EventKind = "validation"
	KindBlocked    EventKind = "blocked"
	KindFinish     EventKind = "finish"
	KindWarning    EventKind = "warning"
)

type Source struct {
	Adapter        string `json:"adapter"`
	EventName      string `json:"event_name"`
	SessionRefHash string `json:"session_ref_hash,omitempty"`
	TurnRefHash    string `json:"turn_ref_hash,omitempty"`
	AgentRefHash   string `json:"agent_ref_hash,omitempty"`
	AgentType      string `json:"agent_type,omitempty"`
}

type ArtifactRef struct {
	Kind          string `json:"kind"`
	PathSHA256    string `json:"path_sha256"`
	ContentSHA256 string `json:"content_sha256,omitempty"`
}

type EvidenceRef struct {
	Category   string `json:"category"`
	Result     string `json:"result"`
	PathSHA256 string `json:"path_sha256,omitempty"`
}

type Checkpoint struct {
	Status    string `json:"status"`
	Gate      string `json:"gate,omitempty"`
	NextOwner string `json:"next_owner,omitempty"`
}

type Metrics struct {
	Availability   string   `json:"availability"`
	DurationMillis *int64   `json:"duration_millis,omitempty"`
	Tokens         *int64   `json:"tokens,omitempty"`
	CostUSD        *float64 `json:"cost_usd,omitempty"`
}

type EventDraft struct {
	EventID         string        `json:"event_id,omitempty"`
	RunID           string        `json:"run_id"`
	ParentRunID     string        `json:"parent_run_id,omitempty"`
	CausedByEventID string        `json:"caused_by_event_id,omitempty"`
	OccurredAt      time.Time     `json:"occurred_at,omitempty"`
	Kind            EventKind     `json:"kind"`
	Source          Source        `json:"source"`
	TaskRef         string        `json:"task_ref,omitempty"`
	ArtifactRefs    []ArtifactRef `json:"artifact_refs,omitempty"`
	EvidenceRefs    []EvidenceRef `json:"evidence_refs,omitempty"`
	Checkpoint      *Checkpoint   `json:"checkpoint,omitempty"`
	Metrics         *Metrics      `json:"metrics,omitempty"`
}

type Event struct {
	SchemaVersion      string        `json:"schema_version"`
	EventID            string        `json:"event_id"`
	RunID              string        `json:"run_id"`
	ParentRunID        string        `json:"parent_run_id,omitempty"`
	CausedByEventID    string        `json:"caused_by_event_id,omitempty"`
	LamportClock       uint64        `json:"lamport_clock"`
	LocalSequence      uint64        `json:"local_sequence"`
	OccurredAt         time.Time     `json:"occurred_at,omitempty"`
	ObservedAt         time.Time     `json:"observed_at"`
	Kind               EventKind     `json:"kind"`
	Source             Source        `json:"source"`
	TaskRef            string        `json:"task_ref,omitempty"`
	ArtifactRefs       []ArtifactRef `json:"artifact_refs,omitempty"`
	EvidenceRefs       []EvidenceRef `json:"evidence_refs,omitempty"`
	Checkpoint         *Checkpoint   `json:"checkpoint,omitempty"`
	Metrics            *Metrics      `json:"metrics,omitempty"`
	IdempotencyKeyHash string        `json:"idempotency_key_hash"`
	Fingerprint        string        `json:"fingerprint"`
}

func DecodeDraft(input io.Reader) (EventDraft, error) {
	var draft EventDraft
	body, err := io.ReadAll(io.LimitReader(input, maxInputJSONSize+1))
	if err != nil {
		return EventDraft{}, fmt.Errorf("event payload inválido: lectura fallida")
	}
	if len(body) > maxInputJSONSize {
		return EventDraft{}, fmt.Errorf("event payload inválido: excede el límite")
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&draft); err != nil {
		return EventDraft{}, fmt.Errorf("event payload inválido: %w", sanitizeJSONError(err))
	}
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		return EventDraft{}, fmt.Errorf("event payload inválido: contenido adicional")
	}
	if err := ValidateDraft(draft); err != nil {
		return EventDraft{}, err
	}
	return draft, nil
}

func ValidateDraft(draft EventDraft) error {
	if err := validateID("run_id", draft.RunID, true); err != nil {
		return err
	}
	if err := validateID("event_id", draft.EventID, false); err != nil {
		return err
	}
	if err := validateID("parent_run_id", draft.ParentRunID, false); err != nil {
		return err
	}
	if draft.ParentRunID != "" && draft.ParentRunID == draft.RunID {
		return fieldError("parent_run_id", "no puede ser igual a run_id")
	}
	if err := validateID("caused_by_event_id", draft.CausedByEventID, false); err != nil {
		return err
	}
	if !validKind(draft.Kind) {
		return fieldError("kind", "valor no soportado")
	}
	if err := validateToken("source.adapter", draft.Source.Adapter, true, maxReferenceLen); err != nil {
		return err
	}
	if err := validateToken("source.event_name", draft.Source.EventName, true, maxEventNameLen); err != nil {
		return err
	}
	if draft.Source.SessionRefHash != "" && !digestPattern.MatchString(draft.Source.SessionRefHash) {
		return fieldError("source.session_ref_hash", "debe ser sha256 hexadecimal")
	}
	if draft.Source.TurnRefHash != "" && !digestPattern.MatchString(draft.Source.TurnRefHash) {
		return fieldError("source.turn_ref_hash", "debe ser sha256 hexadecimal")
	}
	if draft.Source.AgentRefHash != "" && !digestPattern.MatchString(draft.Source.AgentRefHash) {
		return fieldError("source.agent_ref_hash", "debe ser sha256 hexadecimal")
	}
	if err := validateToken("source.agent_type", draft.Source.AgentType, false, maxReferenceLen); err != nil {
		return err
	}
	if err := validateToken("task_ref", draft.TaskRef, false, maxReferenceLen); err != nil {
		return err
	}
	if len(draft.ArtifactRefs) > maxReferences {
		return fieldError("artifact_refs", "excede el límite")
	}
	for _, ref := range draft.ArtifactRefs {
		if err := validateToken("artifact_refs.kind", ref.Kind, true, maxReferenceLen); err != nil {
			return err
		}
		if !digestPattern.MatchString(ref.PathSHA256) {
			return fieldError("artifact_refs.path_sha256", "debe ser sha256 hexadecimal")
		}
		if ref.ContentSHA256 != "" && !digestPattern.MatchString(ref.ContentSHA256) {
			return fieldError("artifact_refs.content_sha256", "debe ser sha256 hexadecimal")
		}
	}
	if len(draft.EvidenceRefs) > maxReferences {
		return fieldError("evidence_refs", "excede el límite")
	}
	for _, ref := range draft.EvidenceRefs {
		if err := validateToken("evidence_refs.category", ref.Category, true, maxReferenceLen); err != nil {
			return err
		}
		if err := validateToken("evidence_refs.result", ref.Result, true, maxReferenceLen); err != nil {
			return err
		}
		if ref.PathSHA256 != "" && !digestPattern.MatchString(ref.PathSHA256) {
			return fieldError("evidence_refs.path_sha256", "debe ser sha256 hexadecimal")
		}
	}
	if draft.Checkpoint != nil {
		if !allowedCheckpointStatus(draft.Checkpoint.Status) {
			return fieldError("checkpoint.status", "valor no soportado")
		}
		if err := validateToken("checkpoint.gate", draft.Checkpoint.Gate, false, maxReferenceLen); err != nil {
			return err
		}
		if err := validateToken("checkpoint.next_owner", draft.Checkpoint.NextOwner, false, maxReferenceLen); err != nil {
			return err
		}
	}
	if draft.Metrics != nil {
		if draft.Metrics.Availability != "available" && draft.Metrics.Availability != "unavailable" && draft.Metrics.Availability != "partial" {
			return fieldError("metrics.availability", "valor no soportado")
		}
		if draft.Metrics.DurationMillis != nil && *draft.Metrics.DurationMillis < 0 {
			return fieldError("metrics.duration_millis", "no puede ser negativo")
		}
		if draft.Metrics.Tokens != nil && *draft.Metrics.Tokens < 0 {
			return fieldError("metrics.tokens", "no puede ser negativo")
		}
		if draft.Metrics.CostUSD != nil && *draft.Metrics.CostUSD < 0 {
			return fieldError("metrics.cost_usd", "no puede ser negativo")
		}
	}
	return nil
}

func ValidateEvent(event Event) error {
	if event.SchemaVersion != SchemaVersion {
		return fieldError("schema_version", "versión no soportada")
	}
	if event.LamportClock == 0 || event.LocalSequence == 0 {
		return fieldError("logical_order", "debe ser mayor que cero")
	}
	if event.ObservedAt.IsZero() {
		return fieldError("observed_at", "es obligatorio")
	}
	if !digestPattern.MatchString(event.IdempotencyKeyHash) {
		return fieldError("idempotency_key_hash", "debe ser sha256 hexadecimal")
	}
	if !digestPattern.MatchString(event.Fingerprint) {
		return fieldError("fingerprint", "debe ser sha256 hexadecimal")
	}
	return ValidateDraft(EventDraft{
		EventID:         event.EventID,
		RunID:           event.RunID,
		ParentRunID:     event.ParentRunID,
		CausedByEventID: event.CausedByEventID,
		OccurredAt:      event.OccurredAt,
		Kind:            event.Kind,
		Source:          event.Source,
		TaskRef:         event.TaskRef,
		ArtifactRefs:    event.ArtifactRefs,
		EvidenceRefs:    event.EvidenceRefs,
		Checkpoint:      event.Checkpoint,
		Metrics:         event.Metrics,
	})
}

func CanonicalFingerprint(draft EventDraft) (string, error) {
	if err := ValidateDraft(draft); err != nil {
		return "", err
	}
	payload := struct {
		RunID           string        `json:"run_id"`
		ParentRunID     string        `json:"parent_run_id,omitempty"`
		CausedByEventID string        `json:"caused_by_event_id,omitempty"`
		OccurredAt      time.Time     `json:"occurred_at,omitempty"`
		Kind            EventKind     `json:"kind"`
		Source          Source        `json:"source"`
		TaskRef         string        `json:"task_ref,omitempty"`
		ArtifactRefs    []ArtifactRef `json:"artifact_refs,omitempty"`
		EvidenceRefs    []EvidenceRef `json:"evidence_refs,omitempty"`
		Checkpoint      *Checkpoint   `json:"checkpoint,omitempty"`
		Metrics         *Metrics      `json:"metrics,omitempty"`
	}{
		RunID: draft.RunID, ParentRunID: draft.ParentRunID, CausedByEventID: draft.CausedByEventID,
		OccurredAt: draft.OccurredAt.UTC(), Kind: draft.Kind, Source: draft.Source, TaskRef: draft.TaskRef,
		ArtifactRefs: draft.ArtifactRefs, EvidenceRefs: draft.EvidenceRefs, Checkpoint: draft.Checkpoint, Metrics: draft.Metrics,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("no se pudo canonicalizar evento: %w", err)
	}
	return HashReference(string(body)), nil
}

func HashReference(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func validKind(kind EventKind) bool {
	switch kind {
	case KindStart, KindLink, KindCheckpoint, KindEvidence, KindValidation, KindBlocked, KindFinish, KindWarning:
		return true
	default:
		return false
	}
}

func allowedCheckpointStatus(status string) bool {
	switch status {
	case "ready", "implemented", "validated", "delivery_pending", "sync_pending", "blocked", "escalated", "delivered", "closed":
		return true
	default:
		return false
	}
}

func validateID(field, value string, required bool) error {
	if value == "" && !required {
		return nil
	}
	if !idPattern.MatchString(value) {
		return fieldError(field, "identificador inválido")
	}
	return nil
}

func validateToken(field, value string, required bool, limit int) error {
	if value == "" {
		if required {
			return fieldError(field, "es obligatorio")
		}
		return nil
	}
	if len(value) > limit || strings.ContainsAny(value, "\r\n\x00") || !tokenPattern.MatchString(value) {
		return fieldError(field, "valor inválido o fuera de límite")
	}
	return nil
}

func fieldError(field, reason string) error {
	return fmt.Errorf("campo %s: %s", field, reason)
}

func sanitizeJSONError(err error) error {
	message := err.Error()
	if strings.HasPrefix(message, "json: unknown field ") {
		return fmt.Errorf("campo no permitido %s", strings.TrimPrefix(message, "json: unknown field "))
	}
	return fmt.Errorf("JSON malformado")
}
