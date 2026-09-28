package resultcontractledger

import (
	"context"
	"errors"
	"sort"
	"strconv"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/resultcontract"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

type Status string

const (
	StatusRecorded      Status = "recorded"
	StatusDuplicateNoop Status = "duplicate_noop"
	StatusConflict      Status = "conflict"
	StatusUnavailable   Status = "unavailable"
)

type Appender interface {
	Append(context.Context, runledger.AppendRequest) (runledger.AppendResult, error)
}

type RecordRequest struct {
	RunID            string
	IdempotencyKey   string
	Contract         resultcontract.Contract
	Decision         resultcontract.TransitionDecision
	ReferenceDigests []string
}

type RecordResult struct {
	Status  Status
	EventID string
	Warning string
}

type Bridge struct {
	appender Appender
}

func New(appender Appender) Bridge {
	return Bridge{appender: appender}
}

func (bridge Bridge) Record(ctx context.Context, request RecordRequest) RecordResult {
	if bridge.appender == nil {
		return unavailableResult()
	}
	canonical, err := resultcontract.Canonicalize(request.Contract)
	if err != nil || request.RunID == "" || request.IdempotencyKey == "" || request.Decision.DecisionID == "" {
		return unavailableResult()
	}

	refs := append([]string(nil), request.ReferenceDigests...)
	sort.Strings(refs)
	artifacts := make([]runledger.ArtifactRef, 0, len(refs))
	for index, digest := range refs {
		if index > 0 && digest == refs[index-1] {
			continue
		}
		artifacts = append(artifacts, runledger.ArtifactRef{Kind: "result_reference", PathSHA256: digest})
	}
	draft := runledger.EventDraft{
		RunID: request.RunID,
		Kind:  runledger.KindCheckpoint,
		Source: runledger.Source{
			Adapter:   "lufy-result-contract",
			EventName: "transition-decision",
		},
		TaskRef:      resultcontract.SchemaVersion,
		ArtifactRefs: artifacts,
		EvidenceRefs: []runledger.EvidenceRef{
			{Category: "result_contract", Result: string(request.Contract.Status), PathSHA256: canonical.Fingerprint},
			{Category: "transition_decision", Result: string(request.Decision.Status), PathSHA256: request.Decision.DecisionID},
			{Category: "transition_version", Result: strconv.FormatUint(request.Decision.NextVersion, 10)},
		},
		Checkpoint: &runledger.Checkpoint{
			Status: string(request.Contract.Status),
			Gate:   string(request.Decision.Status),
		},
	}
	appendResult, appendErr := bridge.appender.Append(ctx, runledger.AppendRequest{
		Draft:          draft,
		IdempotencyKey: request.IdempotencyKey,
	})
	if errors.Is(appendErr, runledger.ErrIdempotencyConflict) || appendResult.Status == runledger.AppendConflict {
		return RecordResult{Status: StatusConflict, EventID: appendResult.Event.EventID}
	}
	if appendErr != nil {
		return unavailableResult()
	}
	switch appendResult.Status {
	case runledger.AppendRecorded:
		return RecordResult{Status: StatusRecorded, EventID: appendResult.Event.EventID}
	case runledger.AppendDuplicateNoop:
		return RecordResult{Status: StatusDuplicateNoop, EventID: appendResult.Event.EventID}
	default:
		return unavailableResult()
	}
}

func unavailableResult() RecordResult {
	return RecordResult{
		Status:  StatusUnavailable,
		Warning: "Run Ledger no disponible; reintentar la correlación durable sin avanzar el gate.",
	}
}
