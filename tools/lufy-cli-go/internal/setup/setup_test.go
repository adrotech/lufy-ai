package setup

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/versioncheck"
)

func TestSetupDryRunDoesNotWrite(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	err := NewService().Run(Options{Target: target, DryRun: true, SkipVersionCheck: true}, &out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Version: check omitido") || !strings.Contains(out.String(), "[apply] install") {
		t.Fatalf("unexpected output:\n%s", out.String())
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy")); !os.IsNotExist(err) {
		t.Fatalf("dry-run wrote .lufy: %v", err)
	}
}

func TestSetupJSONIncludesVersionAndFeatures(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	err := NewService().Run(Options{
		Target: target,
		DryRun: true,
		JSON:   true,
		VersionCheck: func() versioncheck.Result {
			return versioncheck.Result{Checked: true, CurrentVersion: "v1.0.0", LatestVersion: "v1.0.1", UpdateAvailable: true, Recommendation: "upgrade"}
		},
	}, &out)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Report
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid JSON %q: %v", out.String(), err)
	}
	if decoded.Version == nil || !decoded.Version.UpdateAvailable {
		t.Fatalf("missing update in JSON: %#v", decoded.Version)
	}
	if len(decoded.Features) == 0 {
		t.Fatalf("missing features: %#v", decoded)
	}
}

func TestSetupJSONWithoutYesFailsWhenMutating(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	err := NewService().Run(Options{Target: target, JSON: true, SkipVersionCheck: true}, &out)
	if err == nil || !strings.Contains(err.Error(), "requiere --yes") {
		t.Fatalf("expected --yes error, got %v", err)
	}
	var decoded Report
	if jsonErr := json.Unmarshal(out.Bytes(), &decoded); jsonErr != nil {
		t.Fatalf("expected JSON report before error, got %q: %v", out.String(), jsonErr)
	}
	if len(decoded.Features) == 0 {
		t.Fatalf("missing feature plan: %#v", decoded)
	}
}

func TestSetupRequireLatestBlocks(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	err := NewService().Run(Options{
		Target:        target,
		RequireLatest: true,
		VersionCheck: func() versioncheck.Result {
			return versioncheck.Result{Checked: true, CurrentVersion: "v1.0.0", LatestVersion: "v1.0.1", UpdateAvailable: true}
		},
	}, &out)
	if err == nil || !strings.Contains(err.Error(), "requiere ultima version") {
		t.Fatalf("expected require latest error, got %v", err)
	}
}

func TestSetupRequireLatestBlocksWhenCheckFails(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	err := NewService().Run(Options{
		Target:        target,
		RequireLatest: true,
		VersionCheck: func() versioncheck.Result {
			return versioncheck.Result{Checked: true, CurrentVersion: "v1.0.0", Error: "network unavailable"}
		},
	}, &out)
	if err == nil || !strings.Contains(err.Error(), "no pudo verificar") {
		t.Fatalf("expected verification error, got %v", err)
	}
}

func TestSetupApplyEndToEnd(t *testing.T) {
	target := t.TempDir()
	err := NewService().Run(Options{Target: target, Yes: true, SkipVersionCheck: true}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{
		filepath.Join(".lufy", "README.md"),
		filepath.Join(".lufy", "config", "project.yaml"),
		filepath.Join(".lufy", "managed-state", "install-state.json"),
		filepath.Join(".lufy", "memory", "MEMORY.md"),
		filepath.Join(".lufy", "context", "graph.json"),
	} {
		if _, statErr := os.Stat(filepath.Join(target, rel)); statErr != nil {
			t.Fatalf("expected %s after setup: %v", rel, statErr)
		}
	}
	report, err := NewService().BuildReport(Options{Target: target, SkipVersionCheck: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, feature := range report.Features {
		if feature.ID == "layout" || feature.ID == "install" || feature.ID == "memory" || feature.ID == "context-graph" {
			if feature.Status != "skip" {
				t.Fatalf("expected %s to skip after setup, got %#v", feature.ID, feature)
			}
		}
	}
}
