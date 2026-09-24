package adapters

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	adaptivedomain "github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adaptive/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

const StatusSchemaVersion = adaptivedomain.AdaptiveStatusSchemaVersion

var (
	ErrAssignmentUnavailable = errors.New("assignment adaptativa no disponible")
	ErrRecommendationStale   = errors.New("recommendation adaptativa stale")
	ErrCapacityExhausted     = errors.New("capacidad adaptativa agotada")
	ErrBudgetExhausted       = errors.New("budget adaptativo agotado")
	ErrLeaseMismatch         = errors.New("lease adaptativa no coincide")
	ErrLeaseExpired          = errors.New("lease adaptativa vencida")
)

type LedgerAdapter struct {
	store runledger.Store
	now   func() time.Time
}

type RecordDemandRequest struct {
	RunID          string
	Demand         adaptivedomain.DemandSignal
	IdempotencyKey string
}

type AssignRequest struct {
	RunID                string
	Assignment           adaptivedomain.Assignment
	IdempotencyKey       string
	MaxActiveAssignments int
	ActorBudgetCeiling   int
}

type RecordRecommendationRequest struct {
	RunID          string
	Recommendation adaptivedomain.Recommendation
	IdempotencyKey string
}

type YieldRequest struct {
	RunID          string
	Checkpoint     adaptivedomain.YieldCheckpoint
	IdempotencyKey string
}

type ActiveAssignment = adaptivedomain.ActiveAssignment
type WaitingItem = adaptivedomain.WaitingItem
type ActorBudget = adaptivedomain.ActorBudget
type AdaptiveStatus = adaptivedomain.AdaptiveStatus

type OperationResult struct {
	Status        runledger.AppendStatus `json:"status"`
	EventID       string                 `json:"event_id"`
	EventSequence uint64                 `json:"event_sequence"`
	State         AdaptiveStatus         `json:"state"`
}

func NewLedgerAdapter(store runledger.Store, now func() time.Time) *LedgerAdapter {
	if now == nil {
		now = time.Now
	}
	return &LedgerAdapter{store: store, now: now}
}

func (a *LedgerAdapter) RecordDemand(ctx context.Context, request RecordDemandRequest) (OperationResult, error) {
	if a == nil || a.store == nil {
		return OperationResult{}, errors.New("run ledger no disponible")
	}
	if err := adaptivedomain.ValidateDemandSignal(request.Demand); err != nil {
		return OperationResult{}, err
	}
	result, err := a.store.Append(ctx, runledger.AppendRequest{
		Draft: runledger.EventDraft{
			RunID: request.RunID, Kind: runledger.KindDemand,
			CausedByEventID: request.Demand.CausedByEventID,
			Source:          runledger.Source{Adapter: "lufy", EventName: "adaptive_demand"},
			TaskRef:         request.Demand.TaskRef,
			Adaptive: &runledger.AdaptiveMetadata{
				SchemaVersion: runledger.AdaptiveMetadataSchemaVersion,
				DemandID:      request.Demand.DemandID, Priority: request.Demand.Priority,
				RequiredBudget: request.Demand.RequiredBudget,
			},
		},
		IdempotencyKey: request.IdempotencyKey,
	})
	if err != nil {
		return OperationResult{}, err
	}
	state, err := a.Status(ctx, request.RunID, 128, 5)
	return operationResult(result, state), err
}

func (a *LedgerAdapter) Assign(ctx context.Context, request AssignRequest) (OperationResult, error) {
	if a == nil || a.store == nil {
		return OperationResult{}, errors.New("run ledger no disponible")
	}
	if err := adaptivedomain.ValidateAssignment(request.Assignment); err != nil {
		return OperationResult{}, err
	}
	events, err := a.store.LoadRun(ctx, request.RunID)
	if err != nil {
		return OperationResult{}, err
	}
	state := Project(request.RunID, events, 128, 5)
	_, demandExists := findHistoricalDemand(events, request.Assignment.DemandID)
	recommendationEvent, recommendationExists := findCurrentRecommendation(events, request.Assignment.DemandID)
	var waitingCycle int
	if state.Version == request.Assignment.ExpectedProjectionVersion {
		if _, exists := findHistoricalAssignment(events, request.Assignment.AssignmentID); exists {
			return OperationResult{}, ErrAssignmentUnavailable
		}
		waiting, ok := findWaiting(state, request.Assignment.DemandID)
		if !ok || !demandExists || !recommendationExists || waiting.State != "waiting" || activeDemand(state, request.Assignment.DemandID) {
			return OperationResult{}, ErrAssignmentUnavailable
		}
		recommendation := recommendationEvent.Adaptive
		if recommendation.RecommendationFingerprint != request.Assignment.RecommendationFingerprint ||
			recommendation.ActorRef != request.Assignment.ActorRef || recommendation.RoleHint != request.Assignment.RoleHint ||
			recommendation.PolicyVersion != request.Assignment.PolicyVersion || recommendation.Priority != request.Assignment.Priority ||
			recommendation.RequiredBudget != request.Assignment.RequiredBudget || recommendation.Score != request.Assignment.Score {
			return OperationResult{}, ErrAssignmentUnavailable
		}
		if recommendationEvent.LocalSequence != request.Assignment.ExpectedProjectionVersion {
			return OperationResult{}, ErrRecommendationStale
		}
		if request.MaxActiveAssignments > 0 && len(state.ActiveAssignments) >= request.MaxActiveAssignments {
			return OperationResult{}, ErrCapacityExhausted
		}
		if request.ActorBudgetCeiling > 0 && consumedBudget(state, request.Assignment.ActorRef)+request.Assignment.RequiredBudget > request.ActorBudgetCeiling {
			return OperationResult{}, ErrBudgetExhausted
		}
		if !a.now().UTC().Before(request.Assignment.Lease.ExpiresAt.UTC()) {
			return OperationResult{}, ErrLeaseExpired
		}
		waitingCycle = waiting.WaitingCycle
	} else if historical, ok := findHistoricalAssignment(events, request.Assignment.AssignmentID); ok {
		waitingCycle = historical.WaitingCycle
	}
	expected := request.Assignment.ExpectedProjectionVersion
	result, err := a.store.Append(ctx, runledger.AppendRequest{
		Draft: runledger.EventDraft{
			RunID: request.RunID, Kind: runledger.KindAssignment,
			CausedByEventID: recommendationEvent.EventID,
			Source:          runledger.Source{Adapter: "lufy", EventName: "adaptive_assignment"},
			TaskRef:         request.Assignment.TaskRef,
			Adaptive: &runledger.AdaptiveMetadata{
				SchemaVersion: runledger.AdaptiveMetadataSchemaVersion,
				DemandID:      request.Assignment.DemandID, AssignmentID: request.Assignment.AssignmentID,
				ActorRef: request.Assignment.ActorRef, RoleHint: request.Assignment.RoleHint,
				PolicyVersion:             request.Assignment.PolicyVersion,
				RecommendationFingerprint: request.Assignment.RecommendationFingerprint,
				LeaseTokenDigest:          request.Assignment.Lease.TokenDigest,
				LeaseExpiresAt:            request.Assignment.Lease.ExpiresAt.UTC(),
				ExpectedVersion:           request.Assignment.ExpectedProjectionVersion,
				Priority:                  request.Assignment.Priority, RequiredBudget: request.Assignment.RequiredBudget,
				Score: request.Assignment.Score, WaitingCycle: waitingCycle,
			},
		},
		IdempotencyKey: request.IdempotencyKey, ExpectedLocalSequence: &expected,
	})
	if err != nil {
		return OperationResult{}, err
	}
	next, err := a.Status(ctx, request.RunID, 128, 5)
	return operationResult(result, next), err
}

func (a *LedgerAdapter) RecordRecommendation(ctx context.Context, request RecordRecommendationRequest) (OperationResult, error) {
	if a == nil || a.store == nil {
		return OperationResult{}, errors.New("run ledger no disponible")
	}
	if err := adaptivedomain.ValidateRecommendation(request.Recommendation); err != nil {
		return OperationResult{}, err
	}
	events, err := a.store.LoadRun(ctx, request.RunID)
	if err != nil {
		return OperationResult{}, err
	}
	demandEvent, ok := findHistoricalDemand(events, request.Recommendation.DemandID)
	if !ok {
		return OperationResult{}, ErrAssignmentUnavailable
	}
	expected := request.Recommendation.ExpectedProjectionVersion
	result, err := a.store.Append(ctx, runledger.AppendRequest{
		Draft: runledger.EventDraft{
			RunID: request.RunID, Kind: runledger.KindRecommendation,
			CausedByEventID: demandEvent.EventID,
			Source:          runledger.Source{Adapter: "lufy", EventName: "adaptive_recommendation"},
			TaskRef:         request.Recommendation.TaskRef,
			Adaptive: &runledger.AdaptiveMetadata{
				SchemaVersion: runledger.AdaptiveMetadataSchemaVersion,
				DemandID:      request.Recommendation.DemandID, ActorRef: request.Recommendation.ActorRef,
				RoleHint: request.Recommendation.RoleHint, PolicyVersion: request.Recommendation.PolicyVersion,
				RecommendationFingerprint: request.Recommendation.Fingerprint,
				ExpectedVersion:           request.Recommendation.ExpectedProjectionVersion,
				Priority:                  request.Recommendation.Priority,
				RequiredBudget:            request.Recommendation.RequiredBudget,
				Score:                     request.Recommendation.Score,
			},
		},
		IdempotencyKey: request.IdempotencyKey, ExpectedLocalSequence: &expected,
	})
	if err != nil {
		return OperationResult{}, err
	}
	next, err := a.Status(ctx, request.RunID, 128, 5)
	return operationResult(result, next), err
}

func (a *LedgerAdapter) Yield(ctx context.Context, request YieldRequest) (OperationResult, error) {
	if a == nil || a.store == nil {
		return OperationResult{}, errors.New("run ledger no disponible")
	}
	if err := adaptivedomain.ValidateYieldCheckpoint(request.Checkpoint); err != nil {
		return OperationResult{}, err
	}
	events, err := a.store.LoadRun(ctx, request.RunID)
	if err != nil {
		return OperationResult{}, err
	}
	historicalEvent, ok := findHistoricalAssignmentEvent(events, request.Checkpoint.AssignmentID)
	if !ok {
		return OperationResult{}, ErrAssignmentUnavailable
	}
	historical := activeFromEvent(historicalEvent)
	state := Project(request.RunID, events, 128, 5)
	if state.Version == request.Checkpoint.ExpectedProjectionVersion {
		active, ok := findActive(state, request.Checkpoint.AssignmentID)
		if !ok || active.DemandID != request.Checkpoint.DemandID || active.ActorRef != request.Checkpoint.ActorRef {
			return OperationResult{}, ErrAssignmentUnavailable
		}
		if active.LeaseTokenDigest != request.Checkpoint.Lease.TokenDigest ||
			!active.LeaseExpiresAt.Equal(request.Checkpoint.Lease.ExpiresAt.UTC()) {
			return OperationResult{}, ErrLeaseMismatch
		}
		if !a.now().UTC().Before(active.LeaseExpiresAt) && !isExpiredLeaseRecovery(request.Checkpoint) {
			return OperationResult{}, ErrLeaseExpired
		}
	}
	expected := request.Checkpoint.ExpectedProjectionVersion
	artifactRefs := make([]runledger.ArtifactRef, len(request.Checkpoint.ArtifactRefs))
	for index, ref := range request.Checkpoint.ArtifactRefs {
		artifactRefs[index] = runledger.ArtifactRef{Kind: ref.Kind, PathSHA256: ref.PathSHA256, ContentSHA256: ref.ContentSHA256}
	}
	evidenceRefs := make([]runledger.EvidenceRef, len(request.Checkpoint.EvidenceRefs))
	for index, ref := range request.Checkpoint.EvidenceRefs {
		evidenceRefs[index] = runledger.EvidenceRef{Category: ref.Category, Result: ref.Result, PathSHA256: ref.PathSHA256}
	}
	result, err := a.store.Append(ctx, runledger.AppendRequest{
		Draft: runledger.EventDraft{
			RunID: request.RunID, Kind: runledger.KindYield,
			CausedByEventID: historicalEvent.EventID,
			Source:          runledger.Source{Adapter: "lufy", EventName: "adaptive_yield"},
			TaskRef:         historical.TaskRef, ArtifactRefs: artifactRefs, EvidenceRefs: evidenceRefs,
			Adaptive: &runledger.AdaptiveMetadata{
				SchemaVersion: runledger.AdaptiveMetadataSchemaVersion,
				DemandID:      historical.DemandID, AssignmentID: historical.AssignmentID,
				ActorRef: historical.ActorRef, RoleHint: historical.RoleHint,
				PolicyVersion:             historical.PolicyVersion,
				RecommendationFingerprint: historical.RecommendationFingerprint,
				LeaseTokenDigest:          historical.LeaseTokenDigest, LeaseExpiresAt: historical.LeaseExpiresAt,
				ExpectedVersion: request.Checkpoint.ExpectedProjectionVersion,
				Priority:        historical.Priority, RequiredBudget: historical.RequiredBudget,
				Score: historical.Score, WaitingCycle: historical.WaitingCycle + 1,
				Reason: request.Checkpoint.Reason, NextStatus: request.Checkpoint.NextStatus,
				HypothesisRefs:        append([]string(nil), request.Checkpoint.HypothesisRefs...),
				FailedAttemptRefs:     append([]string(nil), request.Checkpoint.FailedAttemptRefs...),
				SuccessorCapabilities: append([]string(nil), request.Checkpoint.SuccessorCapabilities...),
			},
		},
		IdempotencyKey: request.IdempotencyKey, ExpectedLocalSequence: &expected,
	})
	if err != nil {
		return OperationResult{}, err
	}
	next, err := a.Status(ctx, request.RunID, 128, 5)
	return operationResult(result, next), err
}

func (a *LedgerAdapter) Status(ctx context.Context, runID string, maxWaiting, starvationAfter int) (AdaptiveStatus, error) {
	if a == nil || a.store == nil {
		return AdaptiveStatus{}, errors.New("run ledger no disponible")
	}
	events, err := a.store.LoadRun(ctx, runID)
	if err != nil {
		return AdaptiveStatus{}, err
	}
	return Project(runID, events, maxWaiting, starvationAfter), nil
}

func Project(runID string, events []runledger.Event, maxWaiting, starvationAfter int) AdaptiveStatus {
	if maxWaiting <= 0 {
		maxWaiting = 128
	}
	if starvationAfter <= 0 {
		starvationAfter = 5
	}
	status := AdaptiveStatus{
		SchemaVersion: StatusSchemaVersion, RunID: runID,
		ActiveAssignments: []ActiveAssignment{}, Waiting: []WaitingItem{}, ConsumedBudget: []ActorBudget{},
	}
	demands := map[string]WaitingItem{}
	waiting := map[string]WaitingItem{}
	active := map[string]ActiveAssignment{}
	digestParts := []string{}
	for _, event := range events {
		if event.LocalSequence > status.Version {
			status.Version = event.LocalSequence
		}
		digestParts = append(digestParts, event.EventID+":"+event.Fingerprint)
		if event.Adaptive == nil {
			continue
		}
		value := event.Adaptive
		switch event.Kind {
		case runledger.KindDemand:
			item := WaitingItem{
				DemandID: value.DemandID, TaskRef: event.TaskRef, Priority: value.Priority,
				RequiredBudget: value.RequiredBudget, WaitingCycle: value.WaitingCycle, State: "waiting",
			}
			demands[value.DemandID] = item
			if !activeDemandMap(active, value.DemandID) {
				waiting[value.DemandID] = item
			}
		case runledger.KindAssignment:
			delete(waiting, value.DemandID)
			active[value.AssignmentID] = activeFromEvent(event)
		case runledger.KindYield:
			delete(active, value.AssignmentID)
			delete(waiting, value.DemandID)
			if value.NextStatus != "completed" {
				item := demands[value.DemandID]
				item.DemandID = value.DemandID
				item.TaskRef = event.TaskRef
				item.Priority = value.Priority
				item.RequiredBudget = value.RequiredBudget
				item.WaitingCycle = value.WaitingCycle
				item.State = value.NextStatus
				waiting[value.DemandID] = item
				demands[value.DemandID] = item
			}
		}
	}
	for _, assignment := range active {
		status.ActiveAssignments = append(status.ActiveAssignments, assignment)
	}
	sort.Slice(status.ActiveAssignments, func(i, j int) bool {
		return status.ActiveAssignments[i].AssignmentID < status.ActiveAssignments[j].AssignmentID
	})
	budget := map[string]int{}
	for _, assignment := range status.ActiveAssignments {
		budget[assignment.ActorRef] += assignment.RequiredBudget
	}
	for actor, consumed := range budget {
		status.ConsumedBudget = append(status.ConsumedBudget, ActorBudget{ActorRef: actor, Consumed: consumed})
	}
	sort.Slice(status.ConsumedBudget, func(i, j int) bool { return status.ConsumedBudget[i].ActorRef < status.ConsumedBudget[j].ActorRef })
	for _, item := range waiting {
		item.StarvationRisk = item.WaitingCycle >= starvationAfter
		status.Waiting = append(status.Waiting, item)
	}
	sort.Slice(status.Waiting, func(i, j int) bool {
		if status.Waiting[i].Priority != status.Waiting[j].Priority {
			return status.Waiting[i].Priority > status.Waiting[j].Priority
		}
		if status.Waiting[i].WaitingCycle != status.Waiting[j].WaitingCycle {
			return status.Waiting[i].WaitingCycle > status.Waiting[j].WaitingCycle
		}
		return status.Waiting[i].DemandID < status.Waiting[j].DemandID
	})
	if len(status.Waiting) > maxWaiting {
		status.Waiting = status.Waiting[:maxWaiting]
		status.Truncated = true
	}
	status.SourceDigest = runledger.HashReference(strings.Join(digestParts, "|"))
	return status
}

func activeFromEvent(event runledger.Event) ActiveAssignment {
	value := event.Adaptive
	return ActiveAssignment{
		AssignmentID: value.AssignmentID, DemandID: value.DemandID, TaskRef: event.TaskRef,
		ActorRef: value.ActorRef, RoleHint: value.RoleHint, PolicyVersion: value.PolicyVersion,
		RecommendationFingerprint: value.RecommendationFingerprint,
		Priority:                  value.Priority, RequiredBudget: value.RequiredBudget, Score: value.Score,
		WaitingCycle: value.WaitingCycle, LeaseTokenDigest: value.LeaseTokenDigest,
		LeaseExpiresAt: value.LeaseExpiresAt.UTC(),
	}
}

func operationResult(result runledger.AppendResult, state AdaptiveStatus) OperationResult {
	return OperationResult{
		Status: result.Status, EventID: result.Event.EventID,
		EventSequence: result.Event.LocalSequence, State: state,
	}
}

func findWaiting(state AdaptiveStatus, demandID string) (WaitingItem, bool) {
	for _, item := range state.Waiting {
		if item.DemandID == demandID {
			return item, true
		}
	}
	return WaitingItem{}, false
}

func findActive(state AdaptiveStatus, assignmentID string) (ActiveAssignment, bool) {
	for _, item := range state.ActiveAssignments {
		if item.AssignmentID == assignmentID {
			return item, true
		}
	}
	return ActiveAssignment{}, false
}

func activeDemand(state AdaptiveStatus, demandID string) bool {
	for _, item := range state.ActiveAssignments {
		if item.DemandID == demandID {
			return true
		}
	}
	return false
}

func activeDemandMap(values map[string]ActiveAssignment, demandID string) bool {
	for _, item := range values {
		if item.DemandID == demandID {
			return true
		}
	}
	return false
}

func consumedBudget(state AdaptiveStatus, actorRef string) int {
	for _, budget := range state.ConsumedBudget {
		if budget.ActorRef == actorRef {
			return budget.Consumed
		}
	}
	return 0
}

func isExpiredLeaseRecovery(checkpoint adaptivedomain.YieldCheckpoint) bool {
	return checkpoint.Reason == "lease_expiring" && checkpoint.NextStatus == "waiting"
}

func findHistoricalAssignment(events []runledger.Event, assignmentID string) (ActiveAssignment, bool) {
	event, ok := findHistoricalAssignmentEvent(events, assignmentID)
	if !ok {
		return ActiveAssignment{}, false
	}
	return activeFromEvent(event), true
}

func findHistoricalAssignmentEvent(events []runledger.Event, assignmentID string) (runledger.Event, bool) {
	for _, event := range events {
		if event.Kind == runledger.KindAssignment && event.Adaptive != nil && event.Adaptive.AssignmentID == assignmentID {
			return event, true
		}
	}
	return runledger.Event{}, false
}

func findHistoricalDemand(events []runledger.Event, demandID string) (runledger.Event, bool) {
	for _, event := range events {
		if event.Kind == runledger.KindDemand && event.Adaptive != nil && event.Adaptive.DemandID == demandID {
			return event, true
		}
	}
	return runledger.Event{}, false
}

func findCurrentRecommendation(events []runledger.Event, demandID string) (runledger.Event, bool) {
	for index := len(events) - 1; index >= 0; index-- {
		event := events[index]
		if event.Kind == runledger.KindRecommendation && event.Adaptive != nil && event.Adaptive.DemandID == demandID {
			return event, true
		}
	}
	return runledger.Event{}, false
}
