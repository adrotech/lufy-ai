package runledger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func (s *FileStore) acquireRunLock(ctx context.Context, runDir string) (*runLock, error) {
	lockPath := filepath.Join(runDir, "lock")
	deadline := time.Now().Add(s.options.LockTimeout)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		token, err := randomEventID()
		if err != nil {
			return nil, err
		}
		err = os.Mkdir(lockPath, 0o700)
		if err == nil {
			now := s.options.Now().UTC()
			owner := lockOwner{Token: token, PID: os.Getpid(), CreatedAt: now, LeaseExpiresAt: now.Add(s.options.LockLease)}
			if err := writeJSONAtomic(filepath.Join(lockPath, "owner.json"), owner); err != nil {
				_ = os.RemoveAll(lockPath)
				return nil, err
			}
			return &runLock{
				path:             lockPath,
				token:            token,
				cleanupTimeout:   s.options.LockTimeout,
				cleanupPollDelay: s.options.LockPollInterval,
			}, nil
		}
		// Windows puede devolver access denied mientras otro writer elimina el
		// directorio del lock. Es contención transitoria y debe respetar el mismo
		// timeout acotado que un lock todavía existente.
		if !os.IsExist(err) && !os.IsPermission(err) {
			return nil, err
		}
		recovered, recoverErr := s.recoverExpiredLock(lockPath)
		if recoverErr != nil {
			return nil, recoverErr
		}
		if recovered {
			continue
		}
		if !time.Now().Before(deadline) {
			return nil, fmt.Errorf("timeout esperando lock del run")
		}
		timer := time.NewTimer(s.options.LockPollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func (s *FileStore) recoverExpiredLock(lockPath string) (bool, error) {
	var owner lockOwner
	if err := readJSONStrict(filepath.Join(lockPath, "owner.json"), &owner); err != nil {
		return false, nil
	}
	if owner.Token == "" || owner.LeaseExpiresAt.IsZero() || s.options.Now().Before(owner.LeaseExpiresAt) {
		return false, nil
	}
	var confirmed lockOwner
	if err := readJSONStrict(filepath.Join(lockPath, "owner.json"), &confirmed); err != nil {
		return false, nil
	}
	if confirmed.Token != owner.Token || !confirmed.LeaseExpiresAt.Equal(owner.LeaseExpiresAt) {
		return false, nil
	}
	staleToken, err := randomEventID()
	if err != nil {
		return false, err
	}
	stalePath := lockPath + ".stale-" + HashReference(owner.Token + staleToken)[:12]
	if err := os.Rename(lockPath, stalePath); err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, nil
	}
	if err := os.RemoveAll(stalePath); err != nil {
		return false, err
	}
	return true, nil
}

func (l *runLock) release() error {
	if l == nil {
		return nil
	}
	var owner lockOwner
	if err := readJSONStrict(filepath.Join(l.path, "owner.json"), &owner); err != nil {
		return err
	}
	if owner.Token != l.token {
		return fmt.Errorf("lock del run cambió de owner")
	}
	timeout := l.cleanupTimeout
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	pollDelay := l.cleanupPollDelay
	if pollDelay <= 0 {
		pollDelay = 10 * time.Millisecond
	}
	deadline := time.Now().Add(timeout)
	for {
		err := os.RemoveAll(l.path)
		if err == nil || os.IsNotExist(err) {
			return nil
		}
		if !time.Now().Before(deadline) {
			return err
		}
		time.Sleep(pollDelay)
	}
}
