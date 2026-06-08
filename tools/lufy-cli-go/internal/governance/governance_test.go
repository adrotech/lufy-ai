package governance

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/assets"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/installer"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

func TestInfoAndDoctorForInstalledTarget(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "go.mod"), []byte("module example.com/app\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := installer.NewService().Run(installer.Options{Target: target, Yes: true, NoEngram: true, Scope: assets.ScopeProject}, &bytes.Buffer{}); err != nil {
		t.Fatalf("install fixture: %v", err)
	}

	svc := NewService()
	info, err := svc.BuildInfo(Options{Target: target, Scope: assets.ScopeProject})
	if err != nil {
		t.Fatalf("BuildInfo() error = %v", err)
	}
	if !info.Installed || info.CatalogAssets == 0 || info.ManifestAssets == 0 {
		t.Fatalf("info missing installed/catalog/manifest data: %#v", info)
	}
	if !contains(info.Stacks, "go") || !containsPrefix(info.Surfaces, "go-library:") {
		t.Fatalf("info missing stack/surface context: %#v", info)
	}

	doctor, err := svc.BuildDoctor(Options{Target: target, Scope: assets.ScopeProject})
	if err != nil {
		t.Fatalf("BuildDoctor() error = %v", err)
	}
	if !doctor.OK {
		t.Fatalf("doctor should be ok: %#v", doctor)
	}

	var infoOut bytes.Buffer
	if err := svc.Info(Options{Target: target, Scope: assets.ScopeProject}, &infoOut); err != nil {
		t.Fatalf("Info() error = %v", err)
	}
	for _, want := range []string{"Instalado: sí", "Assets catalogo efectivo", "Fingerprint catalogo", "Stacks: go"} {
		if !strings.Contains(infoOut.String(), want) {
			t.Fatalf("Info() output missing %q: %s", want, infoOut.String())
		}
	}

	var doctorOut bytes.Buffer
	if err := svc.Doctor(Options{Target: target, Scope: assets.ScopeProject}, &doctorOut); err != nil {
		t.Fatalf("Doctor() error = %v", err)
	}
	if !strings.Contains(doctorOut.String(), "Doctor OK") {
		t.Fatalf("Doctor() output unexpected: %s", doctorOut.String())
	}
}

func TestPinAndUnpinManagedAsset(t *testing.T) {
	target := t.TempDir()
	if err := installer.NewService().Run(installer.Options{Target: target, Yes: true, NoEngram: true, Scope: assets.ScopeProject}, &bytes.Buffer{}); err != nil {
		t.Fatalf("install fixture: %v", err)
	}
	svc := NewService()
	var out bytes.Buffer
	if err := svc.Pin(PinOptions{Target: target, Path: "lufy-ia.harness.md", Reason: "manual override"}, &out); err != nil {
		t.Fatalf("Pin() error = %v", err)
	}
	pinned := stateMustLoadForGovernance(t, target).AssetMap()["lufy-ia.harness.md"]
	if !pinned.Pinned || pinned.PinnedAt == "" || pinned.PinnedReason != "manual override" || pinned.LastAction != "pin" {
		t.Fatalf("asset not pinned: %#v", pinned)
	}
	info, err := svc.BuildInfo(Options{Target: target, Scope: assets.ScopeProject})
	if err != nil {
		t.Fatalf("BuildInfo() error = %v", err)
	}
	if info.Pinned != 1 {
		t.Fatalf("info pinned count unexpected: %#v", info)
	}
	doctor, err := svc.BuildDoctor(Options{Target: target, Scope: assets.ScopeProject})
	if err != nil {
		t.Fatalf("BuildDoctor() error = %v", err)
	}
	if !doctor.OK || !hasDoctorCheck(doctor.Checks, "info", "pinned/frozen=1") {
		t.Fatalf("doctor should expose pinned info without failing: %#v", doctor)
	}

	out.Reset()
	if err := svc.Unpin(PinOptions{Target: target, Path: "lufy-ia.harness.md"}, &out); err != nil {
		t.Fatalf("Unpin() error = %v", err)
	}
	unpinned := stateMustLoadForGovernance(t, target).AssetMap()["lufy-ia.harness.md"]
	if unpinned.Pinned || unpinned.PinnedAt != "" || unpinned.PinnedReason != "" || unpinned.LastAction != "unpin" {
		t.Fatalf("asset not unpinned: %#v", unpinned)
	}
}

func TestDoctorFailsOnPendingLufyNewConflict(t *testing.T) {
	target := t.TempDir()
	if err := installer.NewService().Run(installer.Options{Target: target, Yes: true, NoEngram: true, Scope: assets.ScopeProject}, &bytes.Buffer{}); err != nil {
		t.Fatalf("install fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "tui.json"), []byte("{\"user\":true}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "tui.json.lufy-new"), []byte("{\"new\":true}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := NewService().BuildDoctor(Options{Target: target, Scope: assets.ScopeProject})
	if err != nil {
		t.Fatalf("BuildDoctor() error = %v", err)
	}
	if report.OK || !hasDoctorCheck(report.Checks, "fail", ".lufy-new=1") {
		t.Fatalf("doctor should fail on pending lufy-new: %#v", report)
	}
}

func TestDoctorReportsMissingManifest(t *testing.T) {
	target := t.TempDir()
	report, err := NewService().BuildDoctor(Options{Target: target, Scope: assets.ScopeProject})
	if err != nil {
		t.Fatalf("BuildDoctor() error = %v", err)
	}
	if report.OK {
		t.Fatalf("doctor should fail without manifest")
	}
	found := false
	for _, check := range report.Checks {
		if check.Level == "fail" && strings.Contains(check.Message, "falta manifest") {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing manifest check not found: %#v", report.Checks)
	}
}

func TestInfoJSONOutput(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	if err := NewService().Info(Options{Target: target, JSON: true, Scope: assets.ScopeProject}, &out); err != nil {
		t.Fatalf("Info(JSON) error = %v", err)
	}
	if !strings.Contains(out.String(), `"installed": false`) || !strings.Contains(out.String(), `"catalogAssets"`) {
		t.Fatalf("Info(JSON) output unexpected: %s", out.String())
	}
}

func TestDoctorJSONOutputReturnsErrorWhenBlocked(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	err := NewService().Doctor(Options{Target: target, JSON: true, Scope: assets.ScopeProject}, &out)
	if err == nil || !strings.Contains(err.Error(), "doctor detectó problemas") {
		t.Fatalf("Doctor(JSON) expected blocked error, got %v", err)
	}
	if !strings.Contains(out.String(), `"ok": false`) || !strings.Contains(out.String(), `"level": "fail"`) {
		t.Fatalf("Doctor(JSON) output unexpected: %s", out.String())
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsPrefix(values []string, prefix string) bool {
	for _, value := range values {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func hasDoctorCheck(checks []DoctorCheck, level, text string) bool {
	for _, check := range checks {
		if check.Level == level && strings.Contains(check.Message, text) {
			return true
		}
	}
	return false
}

func stateMustLoadForGovernance(t *testing.T, target string) *state.InstallState {
	t.Helper()
	st, err := state.Load(target)
	if err != nil {
		t.Fatal(err)
	}
	if st == nil {
		t.Fatal("missing install-state")
	}
	return st
}
