package candidateevidence_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/candidateevidence"
)

func fixtureShellLiteral(value string) string {
	// Git invoca los filtros con su shell; sólo rutas de fixtures sintéticos.
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func configureFixtureFilter(t *testing.T, kind string, included bool) (root, base, marker string) {
	t.Helper()
	root, base = candidateFixture(t)
	fixtureWrite(t, root, ".gitattributes", "candidate.txt filter=canary\n")
	fixtureGit(t, root, "add", ".gitattributes")
	fixtureGit(t, root, "-c", "core.hooksPath="+os.DevNull, "commit", "-m", "synthetic-attributes")
	marker = filepath.Join(t.TempDir(), "filter-executed")
	command := "printf FILTER_EXECUTED > " + fixtureShellLiteral(filepath.ToSlash(marker)) + "; "
	if kind == "clean" {
		command += "cat"
	} else {
		command += "exit 1"
	}
	key := "filter.canary." + kind
	if included {
		include := filepath.Join(root, ".git", "synthetic-filter-config")
		fixtureGit(t, root, "config", "--file", include, key, command)
		fixtureGit(t, root, "config", "include.path", include)
	} else {
		fixtureGit(t, root, "config", key, command)
	}
	// Forzar lectura del contenido para que status no se limite al stat-cache.
	fixtureWrite(t, root, "candidate.txt", "CANDIDATE\n")
	stamp := time.Now().Add(time.Second)
	if err := os.Chtimes(filepath.Join(root, "candidate.txt"), stamp, stamp); err != nil {
		t.Fatal(err)
	}
	return root, base, marker
}

func TestSyntheticGitStatusControlReallyExecutesCleanFilter(t *testing.T) {
	root, _, marker := configureFixtureFilter(t, "clean", false)
	fixtureGit(t, root, "status", "--porcelain")
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("synthetic control did not exercise configured clean filter")
	}
}

func TestCaptureRejectsConfiguredFiltersBeforeAnyCallbackExecutes(t *testing.T) {
	for _, kind := range []string{"clean", "process"} {
		for _, included := range []bool{false, true} {
			name := kind + "-direct"
			if included {
				name = kind + "-included"
			}
			t.Run(name, func(t *testing.T) {
				root, base, marker := configureFixtureFilter(t, kind, included)
				_, err := candidateevidence.Capture(context.Background(), root, base, fixtureEvidenceSHA)
				if err == nil {
					t.Error("configured Git callbacks must be rejected before status refresh")
				}
				if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
					t.Fatal("Capture executed a configured filter before rejecting the candidate")
				}
				if err != nil && strings.Contains(err.Error(), marker) {
					t.Fatal("filter command/path must not be echoed in diagnostics")
				}
			})
		}
	}
}
