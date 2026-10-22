package candidateevidence_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/candidateevidence"
)

func TestCaptureRejectsUninitializedGitlinkWithoutChildRepository(t *testing.T) {
	root, base := candidateFixture(t)
	fixtureGit(t, root, "update-index", "--add", "--cacheinfo", "160000,"+base+",nested")
	fixtureGit(t, root, "-c", "core.hooksPath="+os.DevNull, "commit", "-m", "synthetic-gitlink")
	if err := os.Mkdir(filepath.Join(root, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if fixtureGit(t, root, "status", "--porcelain") != "" {
		t.Fatal("fixture must show a clean but uninitialized gitlink")
	}
	if _, err := candidateevidence.Capture(context.Background(), root, base, fixtureEvidenceSHA, "reviewer"); err == nil {
		t.Fatal("uninitialized gitlink must not be accepted as a fully observed clean candidate")
	}
}
