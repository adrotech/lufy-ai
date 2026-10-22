// Package reviewevents registra operaciones explícitas sin avanzar gates.
package reviewevents

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

type Options struct {
	RunID, ReviewID, Action, OperationID, Result, EvidenceSHA256 string
}

type Service struct {
	Store runledger.Store
	Now   func() time.Time
}

func NewService(store runledger.Store) Service { return Service{Store: store, Now: time.Now} }

var identifierPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{1,127}$`)
var fingerprintPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

func (s Service) Record(ctx context.Context, opts Options) (runledger.AppendResult, error) {
	if s.Store == nil {
		return runledger.AppendResult{}, fmt.Errorf("review store no disponible")
	}
	if opts.OperationID == "" {
		opts.OperationID = opts.ReviewID
	}
	if !identifierPattern.MatchString(opts.RunID) || !identifierPattern.MatchString(opts.ReviewID) || !identifierPattern.MatchString(opts.OperationID) {
		return runledger.AppendResult{}, fmt.Errorf("review requiere identificadores opacos válidos")
	}
	names := map[string]string{"start": "review.started", "complete": "review.completed", "rework": "review.rework", "reopen": "defect.reopened"}
	eventName, ok := names[opts.Action]
	if !ok {
		return runledger.AppendResult{}, fmt.Errorf("operación review no soportada")
	}
	result := opts.Result
	if opts.Action == "start" {
		if result != "" || opts.EvidenceSHA256 != "" {
			return runledger.AppendResult{}, fmt.Errorf("start no acepta resultado ni evidencia")
		}
		result = "started"
	} else {
		if !fingerprintPattern.MatchString(opts.EvidenceSHA256) {
			return runledger.AppendResult{}, fmt.Errorf("review requiere fingerprint de evidencia válido")
		}
		switch opts.Action {
		case "complete":
			if result != "passed" && result != "failed" {
				return runledger.AppendResult{}, fmt.Errorf("complete requiere resultado passed o failed")
			}
		case "rework", "reopen":
			if result != "" {
				return runledger.AppendResult{}, fmt.Errorf("operación review no acepta resultado adicional")
			}
			if opts.Action == "rework" {
				result = "required"
			} else {
				result = "reopened"
			}
		}
	}
	now := s.Now
	if now == nil {
		now = time.Now
	}
	task := runledger.HashReference(opts.ReviewID)
	key := "codex-review/v1:" + runledger.HashReference(opts.RunID) + ":" + task + ":" + opts.Action + ":" + runledger.HashReference(opts.OperationID)
	keyHash := runledger.HashReference(key)
	for attempt := 0; attempt < 3; attempt++ {
		events, err := s.Store.LoadRun(ctx, opts.RunID)
		if err != nil {
			return runledger.AppendResult{}, err
		}
		var sequence uint64
		var start, existing *runledger.Event
		var parent string
		for index := range events {
			event := &events[index]
			if index == 0 {
				parent = event.ParentRunID
			}
			if event.LocalSequence > sequence {
				sequence = event.LocalSequence
			}
			if event.IdempotencyKeyHash == keyHash {
				existing = event
			}
			if event.Source.Adapter != "codex-review" || event.TaskRef != task {
				continue
			}
			if event.Source.EventName == "review.started" {
				if start != nil {
					return runledger.AppendResult{}, fmt.Errorf("inicio review ambiguo")
				}
				start = event
			}
			if event.Source.EventName == eventName && (opts.Action == "start" || opts.Action == "complete") && event.IdempotencyKeyHash != keyHash {
				return runledger.AppendResult{}, runledger.ErrIdempotencyConflict
			}
		}
		if opts.Action != "start" && (start == nil || start.OccurredAt.IsZero()) {
			return runledger.AppendResult{}, fmt.Errorf("review requiere inicio conocido")
		}
		occurred := now().UTC()
		if existing != nil {
			occurred = existing.OccurredAt
		}
		if occurred.IsZero() || start != nil && occurred.Before(start.OccurredAt) {
			return runledger.AppendResult{}, fmt.Errorf("timestamp review inválido")
		}
		draft := runledger.EventDraft{
			RunID: opts.RunID, ParentRunID: parent, OccurredAt: occurred, Kind: runledger.KindEvidence,
			Source: runledger.Source{Adapter: "codex-review", EventName: eventName}, TaskRef: task,
			EvidenceRefs: []runledger.EvidenceRef{{Category: "review", Result: result}},
		}
		if opts.Action != "start" {
			draft.CausedByEventID = start.EventID
			draft.ArtifactRefs = []runledger.ArtifactRef{{Kind: "result-contract", PathSHA256: runledger.HashReference("result-contract:" + opts.EvidenceSHA256), ContentSHA256: opts.EvidenceSHA256}}
		}
		appended, err := s.Store.Append(ctx, runledger.AppendRequest{Draft: draft, IdempotencyKey: key, ExpectedLocalSequence: &sequence})
		if errors.Is(err, runledger.ErrVersionConflict) || errors.Is(err, runledger.ErrIdempotencyConflict) && existing == nil {
			continue
		}
		return appended, err
	}
	return runledger.AppendResult{}, runledger.ErrVersionConflict
}
