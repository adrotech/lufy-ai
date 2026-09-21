package resultcontract

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"testing"
)

func TestIdempotencySameKeyAndIntentReturnsDuplicateNoop(t *testing.T) {
	t.Parallel()

	store := newMemoryReceiptStore()
	prior, intent, policy := leasedTransition(t)
	policy.Receipts = store
	intent.IdempotencyKey = "retry-key-slice-c"

	first := EvaluateTransition(prior, intent, policy)
	if first.Status != DecisionAccepted {
		t.Fatalf("first decision = %#v, want accepted", first)
	}
	second := EvaluateTransition(prior, intent, policy)
	if second.Status != DecisionDuplicateNoop {
		t.Fatalf("retry decision = %#v, want duplicate_noop", second)
	}
	if second.DecisionID != first.DecisionID || second.NextVersion != first.NextVersion || second.NextFingerprint != first.NextFingerprint {
		t.Fatalf("duplicate did not preserve original identity: first=%#v second=%#v", first, second)
	}
	if got := store.saveCount(); got != 1 {
		t.Fatalf("receipt saves = %d, want 1", got)
	}
}

func TestConcurrentIdempotencyRecordsOneDecisionAndDeduplicatesTheRest(t *testing.T) {
	t.Parallel()

	store := newMemoryReceiptStore()
	prior, intent, policy := leasedTransition(t)
	policy.Receipts = store
	intent.IdempotencyKey = "concurrent-retry-key-slice-c"
	const workers = 16
	decisions := make(chan TransitionDecision, workers)
	var group sync.WaitGroup
	for index := 0; index < workers; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			decisions <- EvaluateTransition(prior, intent, policy)
		}()
	}
	group.Wait()
	close(decisions)

	accepted, duplicates := 0, 0
	for decision := range decisions {
		switch decision.Status {
		case DecisionAccepted:
			accepted++
		case DecisionDuplicateNoop:
			duplicates++
		default:
			t.Fatalf("concurrent retry returned unexpected decision: %#v", decision)
		}
	}
	if accepted != 1 || duplicates != workers-1 || store.saveCount() != 1 {
		t.Fatalf("accepted=%d duplicates=%d saves=%d, want 1/%d/1", accepted, duplicates, store.saveCount(), workers-1)
	}
}

func TestIdempotencySameKeyWithDifferentIntentConflictsWithoutOverwrite(t *testing.T) {
	t.Parallel()

	store := newMemoryReceiptStore()
	prior, intent, policy := leasedTransition(t)
	policy.Receipts = store
	intent.IdempotencyKey = "conflict-key-slice-c"
	first := EvaluateTransition(prior, intent, policy)
	if first.Status != DecisionAccepted {
		t.Fatalf("first decision = %#v", first)
	}
	original := store.onlyReceipt(t)

	intent.NextContract.ExecutiveSummary = "Cambio semántico para reuso conflictivo."
	conflict := EvaluateTransition(prior, intent, policy)
	assertRejectedWithoutNextState(t, conflict, DecisionConflict, "idempotency_conflict")
	if got := store.saveCount(); got != 1 {
		t.Fatalf("receipt saves = %d, want original receipt only", got)
	}
	if after := store.onlyReceipt(t); !reflect.DeepEqual(after, original) {
		t.Fatalf("original receipt mutated:\nbefore=%#v\nafter=%#v", original, after)
	}
}

func TestTransitionReceiptIsContentFreeAndStoresOnlyDigestsAndDecisionRefs(t *testing.T) {
	t.Parallel()

	const promptCanary = "PROMPT_CANARY_ignore_all_previous"
	const outputCanary = "OUTPUT_CANARY_private_command_stdout"
	const pathCanary = "/Users/private/PATH_CANARY/repository"
	const keyCanary = "IDEMPOTENCY_KEY_CANARY_raw_secret"
	store := newMemoryReceiptStore()
	prior, intent, policy := leasedTransition(t)
	policy.Receipts = store
	intent.IdempotencyKey = keyCanary
	intent.NextContract.ExecutiveSummary = promptCanary
	intent.NextContract.Evidence.Commands[0].Notes = outputCanary
	intent.NextContract.Risks = []string{pathCanary}

	decision := EvaluateTransition(prior, intent, policy)
	if decision.Status != DecisionAccepted {
		t.Fatalf("decision = %#v, want accepted", decision)
	}
	receipt := store.onlyReceipt(t)
	body, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	observed := string(body)
	for _, canary := range []string{promptCanary, outputCanary, pathCanary, keyCanary} {
		if strings.Contains(observed, canary) {
			t.Fatalf("receipt leaked canary %q: %s", canary, observed)
		}
	}
	sha256Pattern := regexp.MustCompile(`^[a-f0-9]{64}$`)
	if !sha256Pattern.MatchString(receipt.KeyHash) || !sha256Pattern.MatchString(receipt.IntentFingerprint) {
		t.Fatalf("receipt refs are not digests: %#v", receipt)
	}
	if receipt.DecisionStatus != DecisionAccepted || receipt.NextFingerprint != decision.NextFingerprint {
		t.Fatalf("receipt lost decision correlation: receipt=%#v decision=%#v", receipt, decision)
	}
	allowedFields := map[string]bool{
		"KeyHash": true, "IntentFingerprint": true, "DecisionID": true,
		"DecisionStatus": true, "NextVersion": true, "NextFingerprint": true,
	}
	receiptType := reflect.TypeOf(receipt)
	for index := 0; index < receiptType.NumField(); index++ {
		field := receiptType.Field(index)
		if !allowedFields[field.Name] {
			t.Fatalf("receipt contains non-allow-listed field %q of type %s", field.Name, field.Type)
		}
	}
}

type memoryReceiptStore struct {
	mu       sync.Mutex
	receipts map[string]TransitionReceipt
	saves    int
}

func newMemoryReceiptStore() *memoryReceiptStore {
	return &memoryReceiptStore{receipts: make(map[string]TransitionReceipt)}
}

func (store *memoryReceiptStore) Record(candidate TransitionReceipt) (ReceiptOutcome, TransitionReceipt, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if receipt, found := store.receipts[candidate.KeyHash]; found {
		if receipt.IntentFingerprint == candidate.IntentFingerprint {
			return ReceiptDuplicateNoop, receipt, nil
		}
		return ReceiptConflict, receipt, nil
	}
	store.receipts[candidate.KeyHash] = candidate
	store.saves++
	return ReceiptRecorded, candidate, nil
}

func (store *memoryReceiptStore) saveCount() int {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.saves
}

func (store *memoryReceiptStore) onlyReceipt(t *testing.T) TransitionReceipt {
	t.Helper()
	store.mu.Lock()
	defer store.mu.Unlock()
	if len(store.receipts) != 1 {
		t.Fatalf("receipts = %d, want 1", len(store.receipts))
	}
	for _, receipt := range store.receipts {
		return receipt
	}
	return TransitionReceipt{}
}
