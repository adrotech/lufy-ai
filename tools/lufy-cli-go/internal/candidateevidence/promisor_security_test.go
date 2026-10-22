package candidateevidence_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/candidateevidence"
)

func localPromisorFixture(t *testing.T) (clone, base string) {
	t.Helper()
	origin, base := candidateFixture(t)
	fixtureGit(t, origin, "config", "uploadpack.allowFilter", "true")
	fixtureGit(t, origin, "config", "uploadpack.allowAnySHA1InWant", "true")
	clone = t.TempDir()
	fixtureGit(t, origin, "clone", "--local", "--no-hardlinks", origin, clone)
	fixtureGit(t, clone, "config", "core.repositoryformatversion", "1")
	fixtureGit(t, clone, "config", "extensions.partialClone", "origin")
	fixtureGit(t, clone, "config", "remote.origin.promisor", "true")
	fixtureGit(t, clone, "config", "remote.origin.partialCloneFilter", "blob:none")
	object := filepath.Join(clone, ".git", "objects", base[:2], base[2:])
	if _, err := os.Stat(object); os.IsNotExist(err) {
		t.Skip("fixture requires local clone with loose base commit")
	}
	if err := os.Remove(object); err != nil {
		t.Fatal(err)
	}
	// El origen es otro directorio temporal, nunca un servidor de red.
	return clone, base
}

func TestLocalPromisorControlReallyMaterializesMissingBaseObject(t *testing.T) {
	clone, base := localPromisorFixture(t)
	objects := filepath.Join(clone, ".git", "objects")
	before := fingerprintFixtureTree(t, objects)
	if actual := fixtureGit(t, clone, "rev-parse", "--verify", base+"^{commit}"); actual != base {
		t.Fatal("local promisor control did not resolve the requested commit")
	}
	if reflect.DeepEqual(before, fingerprintFixtureTree(t, objects)) {
		t.Fatal("control did not exercise lazy fetch into local object database")
	}
}

func TestCaptureRejectsPromisorBeforeResolvingMissingObjects(t *testing.T) {
	clone, base := localPromisorFixture(t)
	objects := filepath.Join(clone, ".git", "objects")
	before := fingerprintFixtureTree(t, objects)
	_, err := candidateevidence.Capture(context.Background(), clone, base, fixtureEvidenceSHA, "reviewer")
	if err == nil {
		t.Error("partial clone/promisor must reject before resolving missing objects")
	}
	if !reflect.DeepEqual(before, fingerprintFixtureTree(t, objects)) {
		t.Fatal("Capture materialized Git objects via local lazy fetch before rejecting")
	}
}
