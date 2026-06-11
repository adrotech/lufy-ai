package platform

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

type Lock struct {
	path       string
	file       *os.File
	createdDir bool
}

func AcquireLock(targetRoot string) (*Lock, error) {
	lockDir := filepath.Join(targetRoot, ".lufy", "managed-state")
	createdDir := false
	if _, err := os.Stat(lockDir); os.IsNotExist(err) {
		createdDir = true
	} else if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		return nil, err
	}
	lockPath := filepath.Join(lockDir, ".lock")
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if os.IsExist(err) {
		return nil, fmt.Errorf("otra operación lufy-ai está en curso para %s; si no hay procesos activos, elimina %s", targetRoot, lockPath)
	}
	if err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(file, "pid=%d\n", os.Getpid()); err != nil {
		file.Close()
		_ = os.Remove(lockPath)
		return nil, err
	}
	return &Lock{path: lockPath, file: file, createdDir: createdDir}, nil
}

func (l *Lock) Release() error {
	if l == nil {
		return nil
	}
	err := l.file.Close()
	if removeErr := os.Remove(l.path); err == nil {
		err = removeErr
	}
	if l.createdDir {
		lockDir := filepath.Dir(l.path)
		if removeErr := os.Remove(lockDir); err == nil && !ignorableRemoveDirError(removeErr) {
			err = removeErr
		}
		if removeErr := os.Remove(filepath.Dir(lockDir)); err == nil && !ignorableRemoveDirError(removeErr) {
			err = removeErr
		}
	}
	return err
}

func ignorableRemoveDirError(err error) bool {
	return err == nil || os.IsNotExist(err) || errors.Is(err, syscall.ENOTEMPTY)
}
