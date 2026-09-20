package runledger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

const (
	SummarySchemaVersion      = "lufy-run-summary/v1"
	VerificationSchemaVersion = "lufy-run-verification/v1"
)

type MetricsSummary struct {
	Availability   string   `json:"availability"`
	DurationMillis *int64   `json:"duration_millis,omitempty"`
	Tokens         *int64   `json:"tokens,omitempty"`
	CostUSD        *float64 `json:"cost_usd,omitempty"`
}

type RunSummary struct {
	SchemaVersion   string         `json:"schema_version"`
	RunID           string         `json:"run_id"`
	ParentRunID     string         `json:"parent_run_id,omitempty"`
	Status          string         `json:"status"`
	Terminal        bool           `json:"terminal"`
	EventCount      int            `json:"event_count"`
	LastEventID     string         `json:"last_event_id,omitempty"`
	LastObservedAt  time.Time      `json:"last_observed_at,omitempty"`
	Tasks           []string       `json:"tasks,omitempty"`
	ArtifactRefs    []ArtifactRef  `json:"artifact_refs,omitempty"`
	EvidenceRefs    []EvidenceRef  `json:"evidence_refs,omitempty"`
	BlockerEventIDs []string       `json:"blocker_event_ids,omitempty"`
	Metrics         MetricsSummary `json:"metrics"`
	Children        []RunSummary   `json:"children,omitempty"`
	SourceDigest    string         `json:"source_digest"`
}

type RunVerification struct {
	SchemaVersion     string               `json:"schema_version"`
	RunID             string               `json:"run_id"`
	SourceHealthy     bool                 `json:"source_healthy"`
	ProjectionPresent bool                 `json:"projection_present"`
	ProjectionFresh   bool                 `json:"projection_fresh"`
	Repaired          bool                 `json:"repaired"`
	Reports           []VerificationReport `json:"reports"`
	Summary           *RunSummary          `json:"summary,omitempty"`
}

type Projector struct {
	store *FileStore
}

func NewProjector(store *FileStore) *Projector {
	return &Projector{store: store}
}

func (p *Projector) Build(ctx context.Context, rootRunID string) (RunSummary, error) {
	if p == nil || p.store == nil {
		return RunSummary{}, fmt.Errorf("projector sin store")
	}
	if err := validateID("run_id", rootRunID, true); err != nil {
		return RunSummary{}, err
	}
	runIDs, err := p.store.ListRuns(ctx)
	if err != nil {
		return RunSummary{}, err
	}
	eventsByRun := make(map[string][]Event, len(runIDs))
	parentByRun := make(map[string]string, len(runIDs))
	for _, runID := range runIDs {
		events, loadErr := p.store.LoadRun(ctx, runID)
		if loadErr != nil {
			if runID == rootRunID {
				return RunSummary{}, loadErr
			}
			continue
		}
		if len(events) == 0 {
			continue
		}
		parent, parentErr := consistentParent(events)
		if parentErr != nil {
			if runID == rootRunID {
				return RunSummary{}, parentErr
			}
			continue
		}
		eventsByRun[runID] = events
		parentByRun[runID] = parent
	}
	if len(eventsByRun[rootRunID]) == 0 {
		return RunSummary{}, fmt.Errorf("run no disponible: %s", rootRunID)
	}
	childrenByParent := map[string][]string{}
	for runID, parent := range parentByRun {
		if parent != "" {
			childrenByParent[parent] = append(childrenByParent[parent], runID)
		}
	}
	for parent := range childrenByParent {
		sort.Strings(childrenByParent[parent])
	}
	return buildSummaryTree(rootRunID, eventsByRun, childrenByParent, map[string]bool{})
}

func (p *Projector) Verify(ctx context.Context, rootRunID string, repair bool) (RunVerification, error) {
	result := RunVerification{SchemaVersion: VerificationSchemaVersion, RunID: rootRunID, SourceHealthy: true, Reports: []VerificationReport{}}
	rootReport, err := p.store.VerifyRun(ctx, rootRunID)
	if err != nil {
		return result, err
	}
	result.Reports = append(result.Reports, rootReport)
	if !rootReport.Healthy {
		result.SourceHealthy = false
		return result, nil
	}
	summary, err := p.Build(ctx, rootRunID)
	if err != nil {
		return result, err
	}
	for _, childID := range flattenChildRunIDs(summary) {
		report, verifyErr := p.store.VerifyRun(ctx, childID)
		if verifyErr != nil {
			return result, verifyErr
		}
		result.Reports = append(result.Reports, report)
		if !report.Healthy {
			result.SourceHealthy = false
		}
	}
	result.Summary = &summary
	projection, present, readErr := p.readProjection(rootRunID)
	result.ProjectionPresent = present
	result.ProjectionFresh = readErr == nil && present && projection.SourceDigest == summary.SourceDigest
	if repair && result.SourceHealthy && !result.ProjectionFresh {
		if err := p.writeProjection(rootRunID, summary); err != nil {
			return result, err
		}
		result.ProjectionPresent = true
		result.ProjectionFresh = true
		result.Repaired = true
	}
	return result, nil
}

func (p *Projector) readProjection(runID string) (RunSummary, bool, error) {
	runDir, err := platformSafeRunDir(p.store.runsRoot, runID)
	if err != nil {
		return RunSummary{}, false, err
	}
	path := filepath.Join(runDir, "projections", "summary.json")
	var summary RunSummary
	if err := readJSONStrict(path, &summary); err != nil {
		if os.IsNotExist(err) {
			return RunSummary{}, false, nil
		}
		return RunSummary{}, true, err
	}
	if summary.SchemaVersion != SummarySchemaVersion || summary.RunID != runID {
		return RunSummary{}, true, fmt.Errorf("projection inválida")
	}
	return summary, true, nil
}

func (p *Projector) writeProjection(runID string, summary RunSummary) error {
	runDir, err := platformSafeRunDir(p.store.runsRoot, runID)
	if err != nil {
		return err
	}
	return writeJSONAtomic(filepath.Join(runDir, "projections", "summary.json"), summary)
}

func buildSummaryTree(runID string, eventsByRun map[string][]Event, childrenByParent map[string][]string, visiting map[string]bool) (RunSummary, error) {
	if visiting[runID] {
		return RunSummary{}, fmt.Errorf("ciclo causal detectado en run_id")
	}
	visiting[runID] = true
	defer delete(visiting, runID)
	summary := summarizeEvents(eventsByRun[runID])
	for _, childID := range childrenByParent[runID] {
		child, err := buildSummaryTree(childID, eventsByRun, childrenByParent, visiting)
		if err != nil {
			return RunSummary{}, err
		}
		summary.Children = append(summary.Children, child)
		if !child.Terminal {
			summary.Terminal = false
		}
	}
	summary.SourceDigest = summaryDigest(summary, eventsByRun[runID])
	return summary, nil
}

func summarizeEvents(events []Event) RunSummary {
	summary := RunSummary{SchemaVersion: SummarySchemaVersion, Status: "unknown", Metrics: MetricsSummary{Availability: "unavailable"}}
	if len(events) == 0 {
		return summary
	}
	summary.RunID = events[0].RunID
	summary.ParentRunID = events[0].ParentRunID
	summary.EventCount = len(events)
	tasks := map[string]bool{}
	artifacts := map[string]ArtifactRef{}
	evidence := map[string]EvidenceRef{}
	availableMetrics := 0
	unavailableMetrics := 0
	var duration, tokens int64
	var cost float64
	var hasDuration, hasTokens, hasCost bool
	for _, event := range events {
		summary.LastEventID = event.EventID
		if event.ObservedAt.After(summary.LastObservedAt) {
			summary.LastObservedAt = event.ObservedAt
		}
		if event.TaskRef != "" {
			tasks[event.TaskRef] = true
		}
		for _, ref := range event.ArtifactRefs {
			key := ref.Kind + ":" + ref.PathSHA256 + ":" + ref.ContentSHA256
			artifacts[key] = ref
		}
		for _, ref := range event.EvidenceRefs {
			key := ref.Category + ":" + ref.Result + ":" + ref.PathSHA256
			evidence[key] = ref
		}
		if event.Kind == KindBlocked || event.Checkpoint != nil && event.Checkpoint.Status == "blocked" {
			summary.BlockerEventIDs = append(summary.BlockerEventIDs, event.EventID)
		}
		if event.Checkpoint != nil {
			summary.Status = event.Checkpoint.Status
		} else {
			switch event.Kind {
			case KindStart:
				summary.Status = "running"
			case KindBlocked:
				summary.Status = "blocked"
			case KindFinish:
				summary.Status = "closed"
			}
		}
		if event.Metrics == nil || event.Metrics.Availability == "unavailable" {
			unavailableMetrics++
			continue
		}
		availableMetrics++
		if event.Metrics.Availability == "partial" {
			unavailableMetrics++
		}
		if event.Metrics.DurationMillis != nil {
			duration += *event.Metrics.DurationMillis
			hasDuration = true
		}
		if event.Metrics.Tokens != nil {
			tokens += *event.Metrics.Tokens
			hasTokens = true
		}
		if event.Metrics.CostUSD != nil {
			cost += *event.Metrics.CostUSD
			hasCost = true
		}
	}
	summary.Terminal = terminalStatus(summary.Status)
	summary.Tasks = sortedKeys(tasks)
	for _, key := range sortedArtifactKeys(artifacts) {
		summary.ArtifactRefs = append(summary.ArtifactRefs, artifacts[key])
	}
	for _, key := range sortedEvidenceKeys(evidence) {
		summary.EvidenceRefs = append(summary.EvidenceRefs, evidence[key])
	}
	if availableMetrics > 0 && unavailableMetrics == 0 {
		summary.Metrics.Availability = "available"
	} else if availableMetrics > 0 {
		summary.Metrics.Availability = "partial"
	}
	if hasDuration {
		summary.Metrics.DurationMillis = &duration
	}
	if hasTokens {
		summary.Metrics.Tokens = &tokens
	}
	if hasCost {
		summary.Metrics.CostUSD = &cost
	}
	return summary
}

func consistentParent(events []Event) (string, error) {
	if len(events) == 0 {
		return "", nil
	}
	parent := events[0].ParentRunID
	for _, event := range events[1:] {
		if event.ParentRunID != parent {
			return "", fmt.Errorf("parent_run_id inconsistente en run")
		}
	}
	return parent, nil
}

func summaryDigest(summary RunSummary, events []Event) string {
	payload := struct {
		RunID    string   `json:"run_id"`
		Events   []string `json:"events"`
		Children []string `json:"children"`
	}{RunID: summary.RunID}
	for _, event := range events {
		payload.Events = append(payload.Events, event.EventID+":"+event.Fingerprint)
	}
	for _, child := range summary.Children {
		payload.Children = append(payload.Children, child.RunID+":"+child.SourceDigest)
	}
	body, _ := json.Marshal(payload)
	return HashReference(string(body))
}

func flattenChildRunIDs(summary RunSummary) []string {
	var ids []string
	for _, child := range summary.Children {
		ids = append(ids, child.RunID)
		ids = append(ids, flattenChildRunIDs(child)...)
	}
	return ids
}

func terminalStatus(status string) bool {
	switch status {
	case "delivered", "closed":
		return true
	default:
		return false
	}
}

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedArtifactKeys(values map[string]ArtifactRef) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedEvidenceKeys(values map[string]EvidenceRef) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func platformSafeRunDir(runsRoot, runID string) (string, error) {
	if err := validateID("run_id", runID, true); err != nil {
		return "", err
	}
	return platform.SafeJoin(runsRoot, runID)
}
