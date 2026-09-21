package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/platform"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/resultcontract"
)

var resultDigestPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

type fileResultReceiptStore struct {
	target string
}

func newFileResultReceiptStore(target string) *fileResultReceiptStore {
	return &fileResultReceiptStore{target: target}
}

func (store *fileResultReceiptStore) Record(candidate resultcontract.TransitionReceipt) (resultcontract.ReceiptOutcome, resultcontract.TransitionReceipt, error) {
	if !validTransitionReceipt(candidate) {
		return "", resultcontract.TransitionReceipt{}, fmt.Errorf("receipt candidato inválido")
	}
	root, err := platform.SafeJoin(store.target, filepath.Join(".lufy", "runtime", "result-contract"))
	if err != nil {
		return "", resultcontract.TransitionReceipt{}, err
	}
	receiptsDir := filepath.Join(root, "receipts")
	locksDir := filepath.Join(root, "locks")
	if err := os.MkdirAll(receiptsDir, 0o700); err != nil {
		return "", resultcontract.TransitionReceipt{}, err
	}
	if err := os.MkdirAll(locksDir, 0o700); err != nil {
		return "", resultcontract.TransitionReceipt{}, err
	}
	lockPath := filepath.Join(locksDir, candidate.KeyHash)
	if err := acquireResultReceiptLock(lockPath); err != nil {
		return "", resultcontract.TransitionReceipt{}, err
	}
	defer releaseResultReceiptLock(lockPath)

	receiptPath := filepath.Join(receiptsDir, candidate.KeyHash+".json")
	existing, found, err := readTransitionReceipt(receiptPath)
	if err != nil {
		return "", resultcontract.TransitionReceipt{}, err
	}
	if found {
		if existing.IntentFingerprint == candidate.IntentFingerprint {
			return resultcontract.ReceiptDuplicateNoop, existing, nil
		}
		return resultcontract.ReceiptConflict, existing, nil
	}
	body, err := json.MarshalIndent(candidate, "", "  ")
	if err != nil {
		return "", resultcontract.TransitionReceipt{}, err
	}
	body = append(body, '\n')
	if err := platform.WriteFileAtomic(receiptPath, body, 0o600); err != nil {
		return "", resultcontract.TransitionReceipt{}, err
	}
	return resultcontract.ReceiptRecorded, candidate, nil
}

func acquireResultReceiptLock(path string) error {
	deadline := time.Now().Add(2 * time.Second)
	for {
		err := os.Mkdir(path, 0o700)
		if err == nil {
			return nil
		}
		if !os.IsExist(err) {
			return err
		}
		if info, statErr := os.Stat(path); statErr == nil && time.Since(info.ModTime()) > 30*time.Second {
			stale := fmt.Sprintf("%s.stale-%d", path, time.Now().UnixNano())
			if renameErr := os.Rename(path, stale); renameErr == nil {
				_ = os.RemoveAll(stale)
				continue
			}
		}
		if !time.Now().Before(deadline) {
			return fmt.Errorf("timeout esperando receipt lock")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func releaseResultReceiptLock(path string) {
	deadline := time.Now().Add(2 * time.Second)
	for {
		if err := os.RemoveAll(path); err == nil || os.IsNotExist(err) {
			return
		}
		if !time.Now().Before(deadline) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func readTransitionReceipt(path string) (resultcontract.TransitionReceipt, bool, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return resultcontract.TransitionReceipt{}, false, nil
	}
	if err != nil {
		return resultcontract.TransitionReceipt{}, false, err
	}
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var receipt resultcontract.TransitionReceipt
	decodeErr := decoder.Decode(&receipt)
	var trailing any
	trailingErr := decoder.Decode(&trailing)
	closeErr := file.Close()
	if decodeErr != nil || trailingErr != io.EOF || closeErr != nil || !validTransitionReceipt(receipt) {
		return resultcontract.TransitionReceipt{}, false, fmt.Errorf("receipt durable inválido")
	}
	return receipt, true, nil
}

func validTransitionReceipt(receipt resultcontract.TransitionReceipt) bool {
	return resultDigestPattern.MatchString(receipt.KeyHash) &&
		resultDigestPattern.MatchString(receipt.IntentFingerprint) && receipt.DecisionID != "" &&
		receipt.DecisionStatus == resultcontract.DecisionAccepted && receipt.NextVersion > 0 &&
		resultDigestPattern.MatchString(receipt.NextFingerprint)
}
