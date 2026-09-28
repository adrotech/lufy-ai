package adapters

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectReviewStatsParsesNumstatWithoutRenameHeuristics(t *testing.T) {
	repo := t.TempDir()
	gitStoreTest(t, repo, "init")
	gitStoreTest(t, repo, "config", "user.email", "test@example.com")
	gitStoreTest(t, repo, "config", "user.name", "Test User")
	writeStoreFile(t, filepath.Join(repo, "alpha.txt"), "one\ntwo\nthree\n")
	writeStoreFile(t, filepath.Join(repo, "space dir", "file name.txt"), "base\r\n")
	writeStoreFile(t, filepath.Join(repo, "rename-me.txt"), "rename\n")
	if err := os.WriteFile(filepath.Join(repo, "binary.dat"), []byte{'a', 0, 'b'}, 0o644); err != nil {
		t.Fatal(err)
	}
	gitStoreTest(t, repo, "add", ".")
	gitStoreTest(t, repo, "commit", "-m", "base")
	base := strings.TrimSpace(gitOutputStoreTest(t, repo, "rev-parse", "HEAD"))

	writeStoreFile(t, filepath.Join(repo, "alpha.txt"), "one\nchanged\nthree\nfour\n")
	writeStoreFile(t, filepath.Join(repo, "space dir", "file name.txt"), "base\r\nnext\r\n")
	if err := os.WriteFile(filepath.Join(repo, "binary.dat"), []byte{'x', 0, 'y'}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(repo, "rename-me.txt"), filepath.Join(repo, "renamed file.txt")); err != nil {
		t.Fatal(err)
	}
	gitStoreTest(t, repo, "add", "-A")

	got, err := CollectReviewStats(repo, base)
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalFiles != 5 || got.Additions != 4 || got.Deletions != 2 || got.Churn != 6 || got.Truncated {
		t.Fatalf("aggregate stats = %#v", got)
	}
	wantPaths := []string{"alpha.txt", "binary.dat", "rename-me.txt", "renamed file.txt", "space dir/file name.txt"}
	if len(got.Files) != len(wantPaths) {
		t.Fatalf("files = %#v", got.Files)
	}
	for i, want := range wantPaths {
		if got.Files[i].Path != want {
			t.Fatalf("files not deterministic: %#v", got.Files)
		}
	}
	if !got.Files[1].Binary || got.Files[1].Additions != 0 || got.Files[1].Deletions != 0 {
		t.Fatalf("binary stats not observable: %#v", got.Files[1])
	}
}

func TestCollectReviewStatsIsBoundedAndRejectsInvalidBase(t *testing.T) {
	repo := t.TempDir()
	gitStoreTest(t, repo, "init")
	gitStoreTest(t, repo, "config", "user.email", "test@example.com")
	gitStoreTest(t, repo, "config", "user.name", "Test User")
	writeStoreFile(t, filepath.Join(repo, "README.md"), "base\n")
	gitStoreTest(t, repo, "add", ".")
	gitStoreTest(t, repo, "commit", "-m", "base")
	base := strings.TrimSpace(gitOutputStoreTest(t, repo, "rev-parse", "HEAD"))
	for i := 0; i < 140; i++ {
		writeStoreFile(t, filepath.Join(repo, "many", fmt.Sprintf("file-%03d.txt", i)), "x\n")
	}
	gitStoreTest(t, repo, "add", ".")
	got, err := CollectReviewStats(repo, base)
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalFiles != 140 || len(got.Files) > 128 || !got.Truncated {
		t.Fatalf("bounded stats = total:%d listed:%d truncated:%t", got.TotalFiles, len(got.Files), got.Truncated)
	}
	if _, err := CollectReviewStats(repo, "missing-base"); err == nil || !strings.Contains(err.Error(), "missing-base") {
		t.Fatalf("invalid base error = %v", err)
	}
}
