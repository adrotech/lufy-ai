package adapters

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"sort"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/runledger"
)

const maxWorkflowLedgerEvents = 4096

type WorkflowEvidence struct {
	Category   string `json:"category"`
	Result     string `json:"result"`
	PathSHA256 string `json:"path_sha256,omitempty"`
}

type WorkflowLedgerEvent struct {
	EventID         string             `json:"event_id"`
	RunID           string             `json:"run_id"`
	LocalSequence   uint64             `json:"local_sequence"`
	ParentRunID     string             `json:"parent_run_id,omitempty"`
	CausedByEventID string             `json:"caused_by_event_id,omitempty"`
	EventName       string             `json:"event_name"`
	TaskRef         string             `json:"task_ref,omitempty"`
	OccurredAt      time.Time          `json:"occurred_at,omitempty"`
	ObservedAt      time.Time          `json:"observed_at"`
	Evidence        []WorkflowEvidence `json:"evidence,omitempty"`
}

type WorkflowProjection struct {
	Availability string                `json:"availability"`
	Digest       string                `json:"digest,omitempty"`
	Events       []WorkflowLedgerEvent `json:"events"`
	Truncated    bool                  `json:"truncated"`
	Recovery     string                `json:"recovery,omitempty"`
}

// LoadWorkflowProjection is the only Context Graph adapter allowed to consult
// Run Ledger storage. It returns a bounded, content-free view and never exposes
// runtime paths, event bodies or human-authored payloads.
func LoadWorkflowProjection(root string) WorkflowProjection {
	config := projectconfig.DefaultRunLedgerConfig()
	path, err := projectconfig.ExistingPath(root)
	if err != nil {
		return unavailableWorkflowProjection()
	}
	if _, statErr := os.Stat(path); statErr == nil {
		loaded, loadErr := projectconfig.Load(path)
		if loadErr != nil {
			return unavailableWorkflowProjection()
		}
		config = loaded.RunLedger
	} else if !os.IsNotExist(statErr) {
		return unavailableWorkflowProjection()
	}
	if !config.IsEnabled() {
		return unavailableWorkflowProjection()
	}
	store, err := runledger.NewFileStore(root, runledger.Options{RuntimeRoot: config.Root})
	if err != nil {
		return unavailableWorkflowProjection()
	}
	runIDs, err := store.ListRuns(context.Background())
	if err != nil || len(runIDs) == 0 {
		return unavailableWorkflowProjection()
	}
	result := WorkflowProjection{Availability: "available", Events: []WorkflowLedgerEvent{}}
	for _, runID := range runIDs {
		events, loadErr := store.LoadRun(context.Background(), runID)
		if loadErr != nil {
			result.Availability = "partial"
			continue
		}
		for _, event := range events {
			if len(result.Events) >= maxWorkflowLedgerEvents {
				result.Availability = "partial"
				result.Truncated = true
				break
			}
			projected := WorkflowLedgerEvent{
				EventID: event.EventID, RunID: event.RunID, LocalSequence: event.LocalSequence, ParentRunID: event.ParentRunID,
				CausedByEventID: event.CausedByEventID, EventName: event.Source.EventName,
				TaskRef: event.TaskRef, OccurredAt: event.OccurredAt, ObservedAt: event.ObservedAt,
			}
			for _, ref := range event.EvidenceRefs {
				projected.Evidence = append(projected.Evidence, WorkflowEvidence{Category: ref.Category, Result: ref.Result, PathSHA256: ref.PathSHA256})
			}
			result.Events = append(result.Events, projected)
		}
		if result.Truncated {
			break
		}
	}
	if len(result.Events) == 0 {
		return unavailableWorkflowProjection()
	}
	sort.Slice(result.Events, func(i, j int) bool {
		if result.Events[i].RunID != result.Events[j].RunID {
			return result.Events[i].RunID < result.Events[j].RunID
		}
		if result.Events[i].LocalSequence != result.Events[j].LocalSequence {
			return result.Events[i].LocalSequence < result.Events[j].LocalSequence
		}
		return result.Events[i].EventID < result.Events[j].EventID
	})
	body, _ := json.Marshal(result.Events)
	sum := sha256.Sum256(body)
	result.Digest = hex.EncodeToString(sum[:])
	if result.Availability == "partial" {
		result.Recovery = "verificar Run Ledger y volver a construir el Context Graph"
	}
	return result
}

func unavailableWorkflowProjection() WorkflowProjection {
	return WorkflowProjection{Availability: "unavailable", Events: []WorkflowLedgerEvent{}, Recovery: "registrar eventos Run Ledger content-free"}
}
