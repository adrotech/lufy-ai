package status

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

func TestStatusReportsMissingAndDrift(t *testing.T) {
	target := t.TempDir()
	writeStatusFile(t, filepath.Join(target, "ok.txt"), "ok\n")
	writeStatusFile(t, filepath.Join(target, "drift.txt"), "original\n")
	okHash := hashStatusFile(t, filepath.Join(target, "ok.txt"))
	driftHash := hashStatusFile(t, filepath.Join(target, "drift.txt"))
	st := state.New(target, nil, []state.AssetState{
		{ID: "ok", TargetRel: "ok.txt", TargetSHA256: okHash},
		{ID: "drift", TargetRel: "drift.txt", TargetSHA256: driftHash},
		{ID: "missing", TargetRel: "missing.txt", TargetSHA256: driftHash},
	}, "test-fingerprint")
	if err := state.WriteAtomic(target, st); err != nil {
		t.Fatal(err)
	}
	writeStatusFile(t, filepath.Join(target, "drift.txt"), "changed\n")

	report, err := NewService().Build(target, true, "")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if report.OK || report.Assets != 3 || report.Missing != 1 || report.Drifted != 1 || report.Errors != 0 {
		t.Fatalf("unexpected report: %#v", report)
	}
	if len(report.AssetDetails) != 3 {
		t.Fatalf("expected verbose asset details, got %#v", report.AssetDetails)
	}
}

func TestStatusJSONOutput(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, JSON: true}, &out); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	var report Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON output: %v body=%s", err, out.String())
	}
	if report.Installed || !report.OK || report.TargetRoot == "" {
		t.Fatalf("unexpected JSON report: %#v", report)
	}
}

func TestStatusVerboseOutput(t *testing.T) {
	target := t.TempDir()
	writeStatusFile(t, filepath.Join(target, "AGENTS.md"), "ok\n")
	hash := hashStatusFile(t, filepath.Join(target, "AGENTS.md"))
	if err := state.WriteAtomic(target, state.New(target, nil, []state.AssetState{{ID: "AGENTS.md", TargetRel: "AGENTS.md", TargetSHA256: hash}}, "test-fingerprint")); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Verbose: true}, &out); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("- [ok] AGENTS.md")) {
		t.Fatalf("verbose output missing asset detail: %s", out.String())
	}
}

func TestStatusReportsLufyNewForNoReplaceDrift(t *testing.T) {
	target := t.TempDir()
	writeStatusFile(t, filepath.Join(target, "tui.json"), "original\n")
	originalHash := hashStatusFile(t, filepath.Join(target, "tui.json"))
	st := state.New(target, nil, []state.AssetState{{ID: "tui.json", TargetRel: "tui.json", TargetSHA256: originalHash, Policy: "no-replace", Scope: "project"}}, "test-fingerprint")
	if err := state.WriteAtomic(target, st); err != nil {
		t.Fatal(err)
	}
	writeStatusFile(t, filepath.Join(target, "tui.json"), "user\n")
	writeStatusFile(t, filepath.Join(target, "tui.json.lufy-new"), "new\n")

	report, err := NewService().Build(target, true, "")
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if !report.OK || report.Drifted != 0 || len(report.AssetDetails) != 1 || report.AssetDetails[0].Status != "lufy-new" {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestStatusHumanOutputAndShortHash(t *testing.T) {
	target := t.TempDir()
	writeStatusFile(t, filepath.Join(target, "AGENTS.md"), "ok\n")
	hash := hashStatusFile(t, filepath.Join(target, "AGENTS.md"))
	if err := state.WriteAtomic(target, state.New(target, nil, []state.AssetState{{ID: "AGENTS.md", TargetRel: "AGENTS.md", TargetSHA256: hash}}, "test-fingerprint")); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := NewService().Run(Options{Target: target, Verbose: true, Scope: assets.ScopeBoth}, &out); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	for _, want := range []string{"Status para", "Scope: both", "globalRoot=", "Instalado: sí", "expected=" + hash[:12]} {
		if !bytes.Contains(out.Bytes(), []byte(want)) {
			t.Fatalf("human status output missing %q: %s", want, out.String())
		}
	}
	if shortHash("short") != "short" || shortHash("1234567890abcdef") != "1234567890ab" {
		t.Fatalf("shortHash unexpected")
	}
}

func writeStatusFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hashStatusFile(t *testing.T, path string) string {
	t.Helper()
	hash, err := assets.FileSHA256(path)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}
