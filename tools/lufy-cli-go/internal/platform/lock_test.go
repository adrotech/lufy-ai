package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAcquireLockRejectsConcurrentLockAndReleases(t *testing.T) {
	target := t.TempDir()
	lock, err := AcquireLock(target)
	if err != nil {
		t.Fatalf("AcquireLock() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy-ai", ".lock")); err != nil {
		t.Fatalf("lock file missing: %v", err)
	}

	second, err := AcquireLock(target)
	if err == nil {
		second.Release()
		t.Fatal("second AcquireLock() expected error")
	}
	if !strings.Contains(err.Error(), "otra operación lufy-ai") {
		t.Fatalf("unexpected lock error: %v", err)
	}

	if err := lock.Release(); err != nil {
		t.Fatalf("Release() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy-ai", ".lock")); !os.IsNotExist(err) {
		t.Fatalf("lock file not removed, err=%v", err)
	}
}
