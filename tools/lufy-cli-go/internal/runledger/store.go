package runledger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
)

var ErrIdempotencyConflict = errors.New("conflicto de idempotencia")

const AppendResultSchemaVersion = "lufy-run-append-result/v1"

type AppendStatus string

const (
	AppendRecorded      AppendStatus = "recorded"
	AppendDuplicateNoop AppendStatus = "duplicate_noop"
	AppendConflict      AppendStatus = "conflict"
)

type AppendRequest struct {
	Draft           EventDraft
	IdempotencyKey  string
	ProposedLamport uint64
}

type AppendResult struct {
	SchemaVersion string       `json:"schema_version"`
	Status        AppendStatus `json:"status"`
	Event         Event        `json:"event"`
}

type VerificationIssue struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Ref      string `json:"ref,omitempty"`
}

type VerificationReport struct {
	RunID      string              `json:"run_id"`
	Healthy    bool                `json:"healthy"`
	EventCount int                 `json:"event_count"`
	Issues     []VerificationIssue `json:"issues,omitempty"`
}

type Store interface {
	Append(context.Context, AppendRequest) (AppendResult, error)
	LoadRun(context.Context, string) ([]Event, error)
	ListRuns(context.Context) ([]string, error)
	VerifyRun(context.Context, string) (VerificationReport, error)
}

type Options struct {
	Now                func() time.Time
	NewID              func() (string, error)
	LockTimeout        time.Duration
	LockPollInterval   time.Duration
	LockLease          time.Duration
	RuntimeRoot        string
	BeforeReceiptWrite func(Event) error
}

type FileStore struct {
	runsRoot string
	options  Options
}

type receipt struct {
	SchemaVersion      string    `json:"schema_version"`
	IdempotencyKeyHash string    `json:"idempotency_key_hash"`
	Fingerprint        string    `json:"fingerprint"`
	EventID            string    `json:"event_id"`
	EventFile          string    `json:"event_file"`
	RecordedAt         time.Time `json:"recorded_at"`
}

type lockOwner struct {
	Token          string    `json:"token"`
	PID            int       `json:"pid"`
	CreatedAt      time.Time `json:"created_at"`
	LeaseExpiresAt time.Time `json:"lease_expires_at"`
}

type runLock struct {
	path  string
	token string
}

func NewFileStore(targetRoot string, options Options) (*FileStore, error) {
	runtimeRoot := options.RuntimeRoot
	if runtimeRoot == "" {
		runtimeRoot = filepath.Join(".lufy", "runtime")
	}
	cleanRuntimeRoot := filepath.ToSlash(filepath.Clean(runtimeRoot))
	if cleanRuntimeRoot != ".lufy/runtime" && !strings.HasPrefix(cleanRuntimeRoot, ".lufy/runtime/") {
		return nil, fieldError("runtime_root", "debe permanecer bajo .lufy/runtime")
	}
	root, err := platform.SafeJoin(targetRoot, filepath.Join(runtimeRoot, "runs"))
	if err != nil {
		return nil, err
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.NewID == nil {
		options.NewID = randomEventID
	}
	if options.LockTimeout <= 0 {
		options.LockTimeout = 2 * time.Second
	}
	if options.LockPollInterval <= 0 {
		options.LockPollInterval = 10 * time.Millisecond
	}
	if options.LockLease <= 0 {
		options.LockLease = 30 * time.Second
	}
	return &FileStore{runsRoot: root, options: options}, nil
}

func (s *FileStore) Append(ctx context.Context, request AppendRequest) (result AppendResult, err error) {
	if err := ValidateDraft(request.Draft); err != nil {
		return AppendResult{}, err
	}
	if len(request.IdempotencyKey) == 0 || len(request.IdempotencyKey) > 512 || strings.ContainsAny(request.IdempotencyKey, "\r\n\x00") {
		return AppendResult{}, fieldError("idempotency_key", "valor inválido o fuera de límite")
	}
	fingerprint, err := CanonicalFingerprint(request.Draft)
	if err != nil {
		return AppendResult{}, err
	}
	runDir, err := platform.SafeJoin(s.runsRoot, request.Draft.RunID)
	if err != nil {
		return AppendResult{}, err
	}
	if err := os.MkdirAll(filepath.Join(runDir, "events"), 0o700); err != nil {
		return AppendResult{}, err
	}
	if err := os.MkdirAll(filepath.Join(runDir, "receipts"), 0o700); err != nil {
		return AppendResult{}, err
	}
	lock, err := s.acquireRunLock(ctx, runDir)
	if err != nil {
		return AppendResult{}, err
	}
	defer func() {
		if releaseErr := lock.release(); err == nil && releaseErr != nil {
			err = releaseErr
		}
	}()

	events, err := s.loadRunDirectory(runDir)
	if err != nil {
		return AppendResult{}, err
	}
	keyHash := HashReference(request.IdempotencyKey)
	receiptPath := filepath.Join(runDir, "receipts", keyHash+".json")
	if existing, found, readErr := readReceipt(receiptPath); readErr != nil {
		return AppendResult{}, readErr
	} else if found {
		return resolveExistingReceipt(existing, events, fingerprint, keyHash)
	}

	for _, event := range events {
		if event.IdempotencyKeyHash != keyHash {
			continue
		}
		if event.Fingerprint != fingerprint {
			return AppendResult{SchemaVersion: AppendResultSchemaVersion, Status: AppendConflict, Event: event}, ErrIdempotencyConflict
		}
		if err := s.writeReceipt(receiptPath, receiptFor(event, s.options.Now().UTC())); err != nil {
			return AppendResult{}, err
		}
		return AppendResult{SchemaVersion: AppendResultSchemaVersion, Status: AppendDuplicateNoop, Event: event}, nil
	}
	if len(events) > 0 {
		parent, parentErr := consistentParent(events)
		if parentErr != nil {
			return AppendResult{}, parentErr
		}
		if request.Draft.ParentRunID != parent {
			return AppendResult{}, fieldError("parent_run_id", "no coincide con el parent durable del run")
		}
	}

	if err := s.validateCausalRefs(ctx, request.Draft, events); err != nil {
		return AppendResult{}, err
	}
	var maxClock, maxSequence uint64
	for _, event := range events {
		if event.LamportClock > maxClock {
			maxClock = event.LamportClock
		}
		if event.LocalSequence > maxSequence {
			maxSequence = event.LocalSequence
		}
	}
	if request.ProposedLamport > maxClock {
		maxClock = request.ProposedLamport
	}
	eventID := request.Draft.EventID
	if eventID == "" {
		eventID, err = s.options.NewID()
		if err != nil {
			return AppendResult{}, fmt.Errorf("no se pudo generar event_id: %w", err)
		}
	}
	for _, event := range events {
		if event.EventID == eventID {
			return AppendResult{}, fieldError("event_id", "ya existe en el run")
		}
	}
	event := eventFromDraft(request.Draft, eventID, maxClock+1, maxSequence+1, s.options.Now().UTC(), keyHash, fingerprint)
	if err := ValidateEvent(event); err != nil {
		return AppendResult{}, err
	}
	eventPath := filepath.Join(runDir, "events", eventFilename(event))
	if err := writeJSONAtomic(eventPath, event); err != nil {
		return AppendResult{}, err
	}
	if s.options.BeforeReceiptWrite != nil {
		if err := s.options.BeforeReceiptWrite(event); err != nil {
			return AppendResult{}, err
		}
	}
	if err := s.writeReceipt(receiptPath, receiptFor(event, s.options.Now().UTC())); err != nil {
		return AppendResult{}, err
	}
	return AppendResult{SchemaVersion: AppendResultSchemaVersion, Status: AppendRecorded, Event: event}, nil
}

func (s *FileStore) LoadRun(ctx context.Context, runID string) ([]Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validateID("run_id", runID, true); err != nil {
		return nil, err
	}
	runDir, err := platform.SafeJoin(s.runsRoot, runID)
	if err != nil {
		return nil, err
	}
	return s.loadRunDirectory(runDir)
}

func (s *FileStore) ListRuns(ctx context.Context) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.runsRoot)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	runs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if validateID("run_id", entry.Name(), true) != nil {
			continue
		}
		runs = append(runs, entry.Name())
	}
	sort.Strings(runs)
	return runs, nil
}

func (s *FileStore) VerifyRun(ctx context.Context, runID string) (VerificationReport, error) {
	report := VerificationReport{RunID: runID, Healthy: true}
	if err := ctx.Err(); err != nil {
		return report, err
	}
	if err := validateID("run_id", runID, true); err != nil {
		return report, err
	}
	runDir, err := platform.SafeJoin(s.runsRoot, runID)
	if err != nil {
		return report, err
	}
	events, eventIssues, err := loadEventsForVerify(filepath.Join(runDir, "events"))
	if err != nil {
		return report, err
	}
	report.EventCount = len(events)
	report.Issues = append(report.Issues, eventIssues...)
	eventByID := make(map[string]Event, len(events))
	sequenceSeen := map[uint64]bool{}
	clockSeen := map[uint64]bool{}
	receiptByKey := map[string]receipt{}
	for _, event := range events {
		if _, exists := eventByID[event.EventID]; exists {
			report.Issues = append(report.Issues, issue("duplicate_event_id", HashReference(event.EventID)))
		}
		eventByID[event.EventID] = event
		if sequenceSeen[event.LocalSequence] {
			report.Issues = append(report.Issues, issue("duplicate_local_sequence", fmt.Sprint(event.LocalSequence)))
		}
		sequenceSeen[event.LocalSequence] = true
		if clockSeen[event.LamportClock] {
			report.Issues = append(report.Issues, issue("duplicate_lamport_clock", fmt.Sprint(event.LamportClock)))
		}
		clockSeen[event.LamportClock] = true
		if event.CausedByEventID != "" {
			causalEvents := events
			if event.ParentRunID != "" {
				parent, parentErr := s.LoadRun(ctx, event.ParentRunID)
				if parentErr != nil || len(parent) == 0 {
					report.Issues = append(report.Issues, issue("missing_parent_run", HashReference(event.ParentRunID)))
					continue
				}
				causalEvents = parent
			}
			foundCause := false
			for _, cause := range causalEvents {
				if cause.EventID == event.CausedByEventID {
					foundCause = true
					break
				}
			}
			if !foundCause {
				report.Issues = append(report.Issues, issue("missing_causal_event", HashReference(event.CausedByEventID)))
			}
		}
	}
	receiptsDir := filepath.Join(runDir, "receipts")
	entries, readErr := os.ReadDir(receiptsDir)
	if readErr != nil && !os.IsNotExist(readErr) {
		return report, readErr
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".write-") {
			report.Issues = append(report.Issues, issue("partial_file", HashReference(entry.Name())))
			continue
		}
		path := filepath.Join(receiptsDir, entry.Name())
		value, found, readErr := readReceipt(path)
		if readErr != nil || !found {
			report.Issues = append(report.Issues, issue("invalid_receipt", HashReference(entry.Name())))
			continue
		}
		receiptByKey[value.IdempotencyKeyHash] = value
		event, ok := eventByID[value.EventID]
		if !ok {
			report.Issues = append(report.Issues, issue("orphan_receipt", value.IdempotencyKeyHash))
			continue
		}
		if event.Fingerprint != value.Fingerprint || event.IdempotencyKeyHash != value.IdempotencyKeyHash || value.EventFile != eventFilename(event) {
			report.Issues = append(report.Issues, issue("receipt_mismatch", value.IdempotencyKeyHash))
		}
	}
	for _, event := range events {
		if _, ok := receiptByKey[event.IdempotencyKeyHash]; !ok {
			report.Issues = append(report.Issues, issue("missing_receipt", event.IdempotencyKeyHash))
		}
	}
	report.Healthy = len(report.Issues) == 0
	return report, nil
}

func (s *FileStore) validateCausalRefs(ctx context.Context, draft EventDraft, current []Event) error {
	if draft.ParentRunID == "" && draft.CausedByEventID == "" {
		return nil
	}
	search := current
	if draft.ParentRunID != "" {
		parent, err := s.LoadRun(ctx, draft.ParentRunID)
		if err != nil || len(parent) == 0 {
			return fieldError("parent_run_id", "run padre no disponible")
		}
		search = parent
	}
	if draft.CausedByEventID != "" {
		for _, event := range search {
			if event.EventID == draft.CausedByEventID {
				return nil
			}
		}
		return fieldError("caused_by_event_id", "evento causal no disponible")
	}
	return nil
}

func (s *FileStore) loadRunDirectory(runDir string) ([]Event, error) {
	eventsDir := filepath.Join(runDir, "events")
	entries, err := os.ReadDir(eventsDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	events := make([]Event, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".write-") {
			continue
		}
		var event Event
		if err := readJSONStrict(filepath.Join(eventsDir, entry.Name()), &event); err != nil {
			return nil, fmt.Errorf("evento durable inválido %s", HashReference(entry.Name()))
		}
		if err := ValidateEvent(event); err != nil {
			return nil, fmt.Errorf("evento durable inválido %s: %w", HashReference(entry.Name()), err)
		}
		events = append(events, event)
	}
	sortEvents(events)
	return events, nil
}

func resolveExistingReceipt(existing receipt, events []Event, fingerprint, keyHash string) (AppendResult, error) {
	for _, event := range events {
		if event.EventID != existing.EventID {
			continue
		}
		if existing.IdempotencyKeyHash != keyHash || event.IdempotencyKeyHash != keyHash || existing.Fingerprint != fingerprint || event.Fingerprint != fingerprint || existing.EventFile != eventFilename(event) {
			return AppendResult{SchemaVersion: AppendResultSchemaVersion, Status: AppendConflict, Event: event}, ErrIdempotencyConflict
		}
		return AppendResult{SchemaVersion: AppendResultSchemaVersion, Status: AppendDuplicateNoop, Event: event}, nil
	}
	return AppendResult{}, fmt.Errorf("receipt sin evento: %s", existing.IdempotencyKeyHash)
}

func (s *FileStore) writeReceipt(path string, value receipt) error {
	return writeJSONAtomic(path, value)
}

func receiptFor(event Event, recordedAt time.Time) receipt {
	return receipt{
		SchemaVersion: "lufy-run-receipt/v1", IdempotencyKeyHash: event.IdempotencyKeyHash,
		Fingerprint: event.Fingerprint, EventID: event.EventID, EventFile: eventFilename(event), RecordedAt: recordedAt,
	}
}

func readReceipt(path string) (receipt, bool, error) {
	var value receipt
	err := readJSONStrict(path, &value)
	if os.IsNotExist(err) {
		return receipt{}, false, nil
	}
	if err != nil {
		return receipt{}, false, fmt.Errorf("receipt inválido: %s", HashReference(filepath.Base(path)))
	}
	if value.SchemaVersion != "lufy-run-receipt/v1" || !digestPattern.MatchString(value.IdempotencyKeyHash) || !digestPattern.MatchString(value.Fingerprint) {
		return receipt{}, false, fmt.Errorf("receipt inválido: %s", HashReference(filepath.Base(path)))
	}
	return value, true, nil
}

func eventFromDraft(draft EventDraft, eventID string, clock, sequence uint64, observedAt time.Time, keyHash, fingerprint string) Event {
	occurredAt := draft.OccurredAt.UTC()
	if draft.OccurredAt.IsZero() {
		occurredAt = observedAt
	}
	return Event{
		SchemaVersion: SchemaVersion, EventID: eventID, RunID: draft.RunID, ParentRunID: draft.ParentRunID,
		CausedByEventID: draft.CausedByEventID, LamportClock: clock, LocalSequence: sequence,
		OccurredAt: occurredAt, ObservedAt: observedAt, Kind: draft.Kind, Source: draft.Source,
		TaskRef: draft.TaskRef, ArtifactRefs: draft.ArtifactRefs, EvidenceRefs: draft.EvidenceRefs,
		Checkpoint: draft.Checkpoint, Metrics: draft.Metrics, IdempotencyKeyHash: keyHash, Fingerprint: fingerprint,
	}
}

func eventFilename(event Event) string {
	return fmt.Sprintf("%020d-%020d-%s.json", event.LamportClock, event.LocalSequence, event.EventID)
}

func writeJSONAtomic(path string, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return platform.WriteFileAtomic(path, body, 0o600)
}

func readJSONStrict(path string, target any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("contenido adicional")
	}
	return nil
}

func sortEvents(events []Event) {
	sort.Slice(events, func(i, j int) bool {
		if events[i].LamportClock != events[j].LamportClock {
			return events[i].LamportClock < events[j].LamportClock
		}
		if events[i].LocalSequence != events[j].LocalSequence {
			return events[i].LocalSequence < events[j].LocalSequence
		}
		return events[i].EventID < events[j].EventID
	})
}

func loadEventsForVerify(eventsDir string) ([]Event, []VerificationIssue, error) {
	entries, err := os.ReadDir(eventsDir)
	if os.IsNotExist(err) {
		return nil, []VerificationIssue{issue("run_missing", "")}, nil
	}
	if err != nil {
		return nil, nil, err
	}
	events := make([]Event, 0, len(entries))
	issues := []VerificationIssue{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".write-") {
			issues = append(issues, issue("partial_file", HashReference(entry.Name())))
			continue
		}
		var event Event
		if err := readJSONStrict(filepath.Join(eventsDir, entry.Name()), &event); err != nil {
			issues = append(issues, issue("invalid_event", HashReference(entry.Name())))
			continue
		}
		if err := ValidateEvent(event); err != nil || entry.Name() != eventFilename(event) {
			issues = append(issues, issue("invalid_event", HashReference(entry.Name())))
			continue
		}
		events = append(events, event)
	}
	sortEvents(events)
	return events, issues, nil
}

func issue(code, ref string) VerificationIssue {
	return VerificationIssue{Code: code, Severity: "error", Ref: ref}
}

func randomEventID() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return "evt_" + hex.EncodeToString(buffer), nil
}
