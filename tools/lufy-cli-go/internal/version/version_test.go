package version

import (
	"strings"
	"testing"
)

func TestCurrentNormalizesEmptyBuildMetadata(t *testing.T) {
	oldVersion, oldCommit, oldBuildDate := Version, Commit, BuildDate
	t.Cleanup(func() {
		Version, Commit, BuildDate = oldVersion, oldCommit, oldBuildDate
	})

	Version, Commit, BuildDate = "", "", ""
	info := Current()
	if info.Version != "dev" || info.Commit != "unknown" || info.BuildDate != "unknown" || !info.DevBuild {
		t.Fatalf("unexpected current info: %+v", info)
	}
	if info.GOOS == "" || info.GOARCH == "" {
		t.Fatalf("platform fields should be populated: %+v", info)
	}
}

func TestCurrentReleaseBuildAndString(t *testing.T) {
	oldVersion, oldCommit, oldBuildDate := Version, Commit, BuildDate
	t.Cleanup(func() {
		Version, Commit, BuildDate = oldVersion, oldCommit, oldBuildDate
	})

	Version, Commit, BuildDate = "v1.2.3", "abc123", "2026-05-24T00:00:00Z"
	info := Current()
	if info.DevBuild {
		t.Fatalf("release metadata should not be marked dev: %+v", info)
	}
	out := info.String()
	for _, want := range []string{"lufy-ai v1.2.3", "commit: abc123", "buildDate: 2026-05-24T00:00:00Z", "goos:", "goarch:"} {
		if !strings.Contains(out, want) {
			t.Fatalf("version string missing %q: %s", want, out)
		}
	}
	if strings.Contains(out, "development build") {
		t.Fatalf("release string should not include development label: %s", out)
	}
}

func TestInfoStringMarksDevBuild(t *testing.T) {
	info := Info{Version: "dev", Commit: "unknown", BuildDate: "unknown", GOOS: "testos", GOARCH: "testarch", DevBuild: true}
	out := info.String()
	if !strings.Contains(out, "dev (development build)") || !strings.Contains(out, "testos") || !strings.Contains(out, "testarch") {
		t.Fatalf("unexpected dev string: %s", out)
	}
}
