package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/adaptive/domain"
)

const RecommendationResultSchemaVersion = "lufy-adaptive-recommendation-result/v1"

var (
	ErrRecordingRequired = errors.New("adaptive recording explícito requerido")
	ErrMutationDisabled  = errors.New("mutación adaptativa no permitida en el modo actual")
	ErrLedgerUnavailable = errors.New("adaptive ledger no disponible")
	ErrLeaseOutOfBounds  = errors.New("lease adaptativa fuera del TTL configurado")
)

type RuntimeConfig struct {
	Enabled               bool
	Mode                  domain.Mode
	PolicyVersion         string
	LeaseTTLSeconds       int
	MaxCandidates         int
	MaxWaitingItems       int
	StarvationAfterCycles int
	ParallelEnabled       bool
	MaxParallelAgents     int
	MaxConcurrentSlices   int
}

type LedgerPort struct {
	RecordDemand         func(context.Context, string, domain.DemandSignal, string) (domain.LedgerOperation, error)
	RecordRecommendation func(context.Context, string, domain.Recommendation, string) (domain.LedgerOperation, error)
	Assign               func(context.Context, string, domain.Assignment, string, AssignmentLimits) (domain.LedgerOperation, error)
	Yield                func(context.Context, string, domain.YieldCheckpoint, string) (domain.LedgerOperation, error)
	Status               func(context.Context, string, int, int) (domain.AdaptiveStatus, error)
}

type AssignmentLimits struct {
	MaxActiveAssignments int
	ActorBudgetCeiling   int
}

type CapacitySignal struct {
	Availability      string   `json:"availability"`
	Limit             int      `json:"limit,omitempty"`
	ActiveAssignments int      `json:"active_assignments,omitempty"`
	Exhausted         bool     `json:"exhausted"`
	Sources           []string `json:"sources"`
}

type RecommendRequest struct {
	RunID          string
	Evaluation     domain.EvaluationRequest
	Config         RuntimeConfig
	Record         bool
	IdempotencyKey string
}

type RecommendationResult struct {
	SchemaVersion         string                  `json:"schema_version"`
	Decision              domain.Decision         `json:"decision"`
	Capacity              CapacitySignal          `json:"capacity"`
	Durability            string                  `json:"durability"`
	DemandReceipt         *domain.LedgerOperation `json:"demand_receipt,omitempty"`
	RecommendationReceipt *domain.LedgerOperation `json:"recommendation_receipt,omitempty"`
}

type MutationRequest[T any] struct {
	RunID          string
	Value          T
	Config         RuntimeConfig
	Record         bool
	IdempotencyKey string
}

type Service struct {
	ledger LedgerPort
	now    func() time.Time
}

func NewService(ledger LedgerPort) *Service {
	return NewServiceWithClock(ledger, time.Now)
}

func NewServiceWithClock(ledger LedgerPort, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{ledger: ledger, now: now}
}

func (s *Service) Recommend(ctx context.Context, request RecommendRequest) (RecommendationResult, error) {
	if err := validateRuntimeConfig(request.Config); err != nil {
		return RecommendationResult{}, err
	}
	capacity, status, statusAvailable := s.capacitySignal(ctx, request.RunID, request.Config)
	evaluation := request.Evaluation
	if statusAvailable {
		evaluation.Profiles = applyProjectedBudget(evaluation.Profiles, status, 100)
	}
	if capacity.Exhausted {
		evaluation.Profiles = append([]domain.CapabilityProfile(nil), evaluation.Profiles...)
		for index := range evaluation.Profiles {
			evaluation.Profiles[index].AvailableBudget = 0
		}
	}
	decision, err := domain.Evaluate(evaluation, domain.Policy{
		Enabled: request.Config.Enabled, Mode: request.Config.Mode,
		Version: request.Config.PolicyVersion, MaxCandidates: request.Config.MaxCandidates,
	})
	if err != nil {
		return RecommendationResult{}, err
	}
	result := RecommendationResult{
		SchemaVersion: RecommendationResultSchemaVersion, Decision: decision,
		Capacity: capacity, Durability: "not_recorded",
	}
	if !request.Record {
		return result, nil
	}
	if !request.Config.Enabled || request.Config.Mode != domain.ModeAdvisory || decision.Action != domain.ActionRecommend {
		return RecommendationResult{}, ErrMutationDisabled
	}
	if strings.TrimSpace(request.RunID) == "" || strings.TrimSpace(request.IdempotencyKey) == "" {
		return RecommendationResult{}, ErrRecordingRequired
	}
	if s == nil || s.ledger.RecordDemand == nil || s.ledger.RecordRecommendation == nil {
		return RecommendationResult{}, ErrLedgerUnavailable
	}
	demandReceipt, err := s.ledger.RecordDemand(ctx, request.RunID, evaluation.Demand, request.IdempotencyKey+":demand")
	if err != nil {
		return RecommendationResult{}, fmt.Errorf("record demand: %w", err)
	}
	candidate, ok := selectedCandidate(decision)
	if !ok {
		return RecommendationResult{}, ErrMutationDisabled
	}
	recommendation := domain.Recommendation{
		SchemaVersion: domain.RecommendationSchemaVersion,
		DemandID:      evaluation.Demand.DemandID, TaskRef: evaluation.Demand.TaskRef,
		ActorRef: decision.SelectedActorRef, RoleHint: decision.SelectedRoleHint,
		PolicyVersion: decision.PolicyVersion, Fingerprint: decision.Fingerprint,
		Priority: evaluation.Demand.Priority, RequiredBudget: evaluation.Demand.RequiredBudget,
		Score: candidate.Breakdown.Total, ExpectedProjectionVersion: demandReceipt.EventSequence,
	}
	recommendationReceipt, err := s.ledger.RecordRecommendation(
		ctx, request.RunID, recommendation, request.IdempotencyKey+":recommendation",
	)
	if err != nil {
		return RecommendationResult{}, fmt.Errorf("record recommendation: %w", err)
	}
	result.Durability = "recorded"
	result.DemandReceipt = &demandReceipt
	result.RecommendationReceipt = &recommendationReceipt
	return result, nil
}

func (s *Service) Assign(ctx context.Context, request MutationRequest[domain.Assignment]) (domain.LedgerOperation, error) {
	if err := validateRuntimeConfig(request.Config); err != nil {
		return domain.LedgerOperation{}, err
	}
	if !request.Record {
		return domain.LedgerOperation{}, ErrRecordingRequired
	}
	if !request.Config.Enabled || request.Config.Mode != domain.ModeAdvisory {
		return domain.LedgerOperation{}, ErrMutationDisabled
	}
	if err := domain.ValidateAssignment(request.Value); err != nil {
		return domain.LedgerOperation{}, err
	}
	now := s.now().UTC()
	expiresAt := request.Value.Lease.ExpiresAt.UTC()
	if !expiresAt.After(now) || expiresAt.After(now.Add(time.Duration(request.Config.LeaseTTLSeconds)*time.Second)) {
		return domain.LedgerOperation{}, ErrLeaseOutOfBounds
	}
	if s == nil || s.ledger.Assign == nil {
		return domain.LedgerOperation{}, ErrLedgerUnavailable
	}
	maxActive, _ := effectiveCapacity(request.Config.MaxParallelAgentsIfEnabled(), request.Config.MaxConcurrentSlices)
	return s.ledger.Assign(ctx, request.RunID, request.Value, request.IdempotencyKey, AssignmentLimits{
		MaxActiveAssignments: maxActive,
		ActorBudgetCeiling:   100,
	})
}

func (s *Service) Yield(ctx context.Context, request MutationRequest[domain.YieldCheckpoint]) (domain.LedgerOperation, error) {
	if err := validateRuntimeConfig(request.Config); err != nil {
		return domain.LedgerOperation{}, err
	}
	if !request.Record {
		return domain.LedgerOperation{}, ErrRecordingRequired
	}
	if request.Config.Enabled && request.Config.Mode != domain.ModeAdvisory {
		return domain.LedgerOperation{}, ErrMutationDisabled
	}
	if err := domain.ValidateYieldCheckpoint(request.Value); err != nil {
		return domain.LedgerOperation{}, err
	}
	if s == nil || s.ledger.Yield == nil {
		return domain.LedgerOperation{}, ErrLedgerUnavailable
	}
	return s.ledger.Yield(ctx, request.RunID, request.Value, request.IdempotencyKey)
}

func (s *Service) Status(ctx context.Context, runID string, config RuntimeConfig) (domain.AdaptiveStatus, error) {
	if err := validateRuntimeConfig(config); err != nil {
		return domain.AdaptiveStatus{}, err
	}
	if s == nil || s.ledger.Status == nil {
		return domain.AdaptiveStatus{}, ErrLedgerUnavailable
	}
	return s.ledger.Status(ctx, runID, config.MaxWaitingItems, config.StarvationAfterCycles)
}

func (s *Service) capacitySignal(ctx context.Context, runID string, config RuntimeConfig) (CapacitySignal, domain.AdaptiveStatus, bool) {
	parallelLimit := 0
	if config.ParallelEnabled {
		parallelLimit = config.MaxParallelAgents
	}
	limit, sources := effectiveCapacity(parallelLimit, config.MaxConcurrentSlices)
	signal := CapacitySignal{Availability: "not_available", Limit: limit, Sources: sources}
	if strings.TrimSpace(runID) == "" || s == nil || s.ledger.Status == nil {
		return signal, domain.AdaptiveStatus{}, false
	}
	status, err := s.ledger.Status(ctx, runID, config.MaxWaitingItems, config.StarvationAfterCycles)
	if err != nil {
		return signal, domain.AdaptiveStatus{}, false
	}
	signal.Availability = "available"
	signal.ActiveAssignments = len(status.ActiveAssignments)
	signal.Exhausted = limit > 0 && signal.ActiveAssignments >= limit
	return signal, status, true
}

func (config RuntimeConfig) MaxParallelAgentsIfEnabled() int {
	if !config.ParallelEnabled {
		return 0
	}
	return config.MaxParallelAgents
}

func applyProjectedBudget(profiles []domain.CapabilityProfile, status domain.AdaptiveStatus, ceiling int) []domain.CapabilityProfile {
	adjusted := append([]domain.CapabilityProfile(nil), profiles...)
	consumedByActor := make(map[string]int, len(status.ConsumedBudget))
	for _, budget := range status.ConsumedBudget {
		consumedByActor[budget.ActorRef] = budget.Consumed
	}
	for index := range adjusted {
		remaining := ceiling - consumedByActor[adjusted[index].ActorRef]
		if remaining < 0 {
			remaining = 0
		}
		if adjusted[index].AvailableBudget > remaining {
			adjusted[index].AvailableBudget = remaining
		}
	}
	return adjusted
}

func effectiveCapacity(parallel, review int) (int, []string) {
	limit := 0
	sources := []string{}
	if parallel > 0 {
		limit = parallel
		sources = append(sources, "parallel_execution.max_parallel_agents")
	}
	if review > 0 {
		if limit == 0 || review < limit {
			limit = review
		}
		sources = append(sources, "workflow_limits.review.max_concurrent_slices")
	}
	return limit, sources
}

func selectedCandidate(decision domain.Decision) (domain.CandidateDecision, bool) {
	for _, candidate := range decision.Ranking {
		if candidate.ActorRef == decision.SelectedActorRef && candidate.Eligible {
			return candidate, true
		}
	}
	return domain.CandidateDecision{}, false
}

func validateRuntimeConfig(config RuntimeConfig) error {
	if config.Mode != domain.ModeShadow && config.Mode != domain.ModeAdvisory {
		return errors.New("adaptive mode no soportado")
	}
	if config.PolicyVersion != domain.PolicyDeterministicV1 {
		return errors.New("adaptive policy no soportada")
	}
	if config.MaxCandidates < 1 || config.MaxCandidates > 32 || config.MaxWaitingItems < 1 ||
		config.StarvationAfterCycles < 1 || config.LeaseTTLSeconds < 1 {
		return errors.New("adaptive config fuera de límites")
	}
	return nil
}
