package runledger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type RetentionPolicy struct {
	MaxAge          time.Duration
	MaxTerminalRuns int
	MaxBytes        int64
}

type PruneCandidate struct {
	RunID          string    `json:"run_id"`
	RunIDs         []string  `json:"run_ids"`
	LastObservedAt time.Time `json:"last_observed_at"`
	Bytes          int64     `json:"bytes"`
	Reasons        []string  `json:"reasons"`
}

type PruneReport struct {
	SchemaVersion   string           `json:"schema_version"`
	DryRun          bool             `json:"dry_run"`
	Candidates      []PruneCandidate `json:"candidates"`
	Deleted         []string         `json:"deleted,omitempty"`
	ProtectedActive []string         `json:"protected_active,omitempty"`
	ReclaimedBytes  int64            `json:"reclaimed_bytes"`
}

type retentionGroup struct {
	rootID         string
	runIDs         []string
	lastObservedAt time.Time
	bytes          int64
	terminal       bool
}

func (s *FileStore) Prune(ctx context.Context, policy RetentionPolicy, now time.Time, apply bool) (PruneReport, error) {
	report := PruneReport{SchemaVersion: "lufy-run-prune/v1", DryRun: !apply}
	if policy.MaxAge <= 0 || policy.MaxTerminalRuns < 0 || policy.MaxBytes <= 0 {
		return report, fmt.Errorf("política de retención inválida")
	}
	runIDs, err := s.ListRuns(ctx)
	if err != nil {
		return report, err
	}
	eventsByRun := map[string][]Event{}
	parentByRun := map[string]string{}
	childrenByParent := map[string][]string{}
	for _, runID := range runIDs {
		events, loadErr := s.LoadRun(ctx, runID)
		if loadErr != nil {
			return report, fmt.Errorf("prune bloqueado por run inválido: %s", HashReference(runID))
		}
		if len(events) == 0 {
			continue
		}
		parent, parentErr := consistentParent(events)
		if parentErr != nil {
			return report, parentErr
		}
		eventsByRun[runID] = events
		parentByRun[runID] = parent
		if parent != "" {
			childrenByParent[parent] = append(childrenByParent[parent], runID)
		}
	}
	for runID, parent := range parentByRun {
		if parent != "" {
			if _, ok := eventsByRun[parent]; !ok {
				return report, fmt.Errorf("prune bloqueado por parent faltante: %s", HashReference(runID))
			}
		}
	}
	var groups []retentionGroup
	for _, runID := range runIDs {
		if _, ok := eventsByRun[runID]; !ok || parentByRun[runID] != "" {
			continue
		}
		group, groupErr := s.buildRetentionGroup(runID, eventsByRun, childrenByParent, map[string]bool{})
		if groupErr != nil {
			return report, groupErr
		}
		groups = append(groups, group)
		if !group.terminal {
			report.ProtectedActive = append(report.ProtectedActive, runID)
		}
	}
	sort.Strings(report.ProtectedActive)
	sort.Slice(groups, func(i, j int) bool {
		if !groups[i].lastObservedAt.Equal(groups[j].lastObservedAt) {
			return groups[i].lastObservedAt.Before(groups[j].lastObservedAt)
		}
		return groups[i].rootID < groups[j].rootID
	})

	selected := map[string]map[string]bool{}
	terminalGroups := make([]retentionGroup, 0, len(groups))
	var totalBytes int64
	for _, group := range groups {
		totalBytes += group.bytes
		if !group.terminal {
			continue
		}
		terminalGroups = append(terminalGroups, group)
		if now.Sub(group.lastObservedAt) > policy.MaxAge {
			addPruneReason(selected, group.rootID, "age")
		}
	}
	if len(terminalGroups) > policy.MaxTerminalRuns {
		for _, group := range terminalGroups[:len(terminalGroups)-policy.MaxTerminalRuns] {
			addPruneReason(selected, group.rootID, "count")
		}
	}
	remainingBytes := totalBytes
	for _, group := range terminalGroups {
		if remainingBytes <= policy.MaxBytes {
			break
		}
		addPruneReason(selected, group.rootID, "bytes")
		remainingBytes -= group.bytes
	}
	for _, group := range groups {
		reasons, ok := selected[group.rootID]
		if !ok {
			continue
		}
		reasonList := sortedKeys(reasons)
		report.Candidates = append(report.Candidates, PruneCandidate{
			RunID: group.rootID, RunIDs: group.runIDs, LastObservedAt: group.lastObservedAt,
			Bytes: group.bytes, Reasons: reasonList,
		})
	}
	if !apply {
		return report, nil
	}
	for _, candidate := range report.Candidates {
		for index := len(candidate.RunIDs) - 1; index >= 0; index-- {
			runDir, pathErr := platformSafeRunDir(s.runsRoot, candidate.RunIDs[index])
			if pathErr != nil {
				return report, pathErr
			}
			if err := os.RemoveAll(runDir); err != nil {
				return report, err
			}
		}
		report.Deleted = append(report.Deleted, candidate.RunID)
		report.ReclaimedBytes += candidate.Bytes
	}
	return report, nil
}

func (s *FileStore) buildRetentionGroup(runID string, eventsByRun map[string][]Event, childrenByParent map[string][]string, visiting map[string]bool) (retentionGroup, error) {
	if visiting[runID] {
		return retentionGroup{}, fmt.Errorf("ciclo causal detectado durante prune")
	}
	visiting[runID] = true
	defer delete(visiting, runID)
	summary := summarizeEvents(eventsByRun[runID])
	group := retentionGroup{rootID: runID, runIDs: []string{runID}, lastObservedAt: summary.LastObservedAt, terminal: summary.Terminal}
	runDir, err := platformSafeRunDir(s.runsRoot, runID)
	if err != nil {
		return retentionGroup{}, err
	}
	group.bytes, err = directorySize(runDir)
	if err != nil {
		return retentionGroup{}, err
	}
	sort.Strings(childrenByParent[runID])
	for _, childID := range childrenByParent[runID] {
		child, childErr := s.buildRetentionGroup(childID, eventsByRun, childrenByParent, visiting)
		if childErr != nil {
			return retentionGroup{}, childErr
		}
		group.runIDs = append(group.runIDs, child.runIDs...)
		group.bytes += child.bytes
		if child.lastObservedAt.After(group.lastObservedAt) {
			group.lastObservedAt = child.lastObservedAt
		}
		group.terminal = group.terminal && child.terminal
	}
	return group, nil
}

func directorySize(root string) (int64, error) {
	var size int64
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		size += info.Size()
		return nil
	})
	return size, err
}

func addPruneReason(selected map[string]map[string]bool, runID, reason string) {
	if selected[runID] == nil {
		selected[runID] = map[string]bool{}
	}
	selected[runID][reason] = true
}
