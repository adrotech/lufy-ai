package harnessconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
)

func TestResolveUsesExplicitThenProjectThenInstallStateThenDefaults(t *testing.T) {
	target := t.TempDir()
	writeProjectHarness(t, target, domain.ToolCodex, domain.DefaultMethodologyByTier())

	requested := domain.HarnessConfig{Tool: domain.ToolInitialDefault}
	result, err := Resolve(Options{Target: target, Requested: requested})
	if err != nil {
		t.Fatal(err)
	}
	if result.Config.Tool != domain.ToolInitialDefault || result.Config.Provenance.Tool != domain.HarnessSourceExplicit {
		t.Fatalf("explicit tool did not win: %#v", result.Config)
	}
	if result.Config.Provenance.MethodologyByTier[domain.TierT1] != domain.HarnessSourceProjectConfig {
		t.Fatalf("project methodology provenance missing: %#v", result.Config.Provenance)
	}

	stateOnly := t.TempDir()
	writeInstallHarness(t, stateOnly, domain.ToolCodex, domain.DefaultMethodologyByTier())
	result, err = Resolve(Options{Target: stateOnly, Requested: domain.DefaultHarnessConfig()})
	if err != nil {
		t.Fatal(err)
	}
	if result.Config.Tool != domain.ToolCodex || result.Config.Provenance.Tool != domain.HarnessSourceInstallState {
		t.Fatalf("install state fallback missing: %#v", result.Config)
	}

	empty := t.TempDir()
	result, err = Resolve(Options{Target: empty, Requested: domain.DefaultHarnessConfig()})
	if err != nil {
		t.Fatal(err)
	}
	if result.Config.Tool != domain.ToolInitialDefault || result.Config.Provenance.Tool != domain.HarnessSourceDefault {
		t.Fatalf("default fallback missing: %#v", result.Config)
	}
}

func TestResolveBlocksUnresolvedProjectAndInstallStateMismatch(t *testing.T) {
	target := t.TempDir()
	writeProjectHarness(t, target, domain.ToolInitialDefault, domain.DefaultMethodologyByTier())
	writeInstallHarness(t, target, domain.ToolCodex, domain.DefaultMethodologyByTier())

	_, err := Resolve(Options{Target: target, Requested: domain.DefaultHarnessConfig(), BlockOnMismatch: true})
	if err == nil || !strings.Contains(err.Error(), "tool project=opencode install-state=codex") || !strings.Contains(strings.ToLower(err.Error()), "recovery") {
		t.Fatalf("expected actionable mismatch, got %v", err)
	}
}

func TestResolveAllowsExplicitValueToReconcileMatchingField(t *testing.T) {
	target := t.TempDir()
	writeProjectHarness(t, target, domain.ToolInitialDefault, domain.DefaultMethodologyByTier())
	writeInstallHarness(t, target, domain.ToolCodex, domain.DefaultMethodologyByTier())

	result, err := Resolve(Options{Target: target, Requested: domain.HarnessConfig{Tool: domain.ToolCodex}, BlockOnMismatch: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.Config.Tool != domain.ToolCodex || result.Config.Provenance.Tool != domain.HarnessSourceExplicit {
		t.Fatalf("explicit reconciliation missing: %#v", result.Config)
	}
}

func TestResolveReportsMethodologyTierMismatch(t *testing.T) {
	target := t.TempDir()
	projectMethodologies := domain.DefaultMethodologyByTier()
	installedMethodologies := domain.DefaultMethodologyByTier()
	installedMethodologies[domain.TierT2] = domain.MethodologySelection{ID: domain.MethodologyLufyWorkflow, Mode: domain.MethodologyModeLite, Required: true}
	writeProjectHarness(t, target, domain.ToolCodex, projectMethodologies)
	writeInstallHarness(t, target, domain.ToolCodex, installedMethodologies)

	_, err := Resolve(Options{Target: target, Requested: domain.DefaultHarnessConfig(), BlockOnMismatch: true})
	if err == nil || !strings.Contains(err.Error(), "methodology T2") || !strings.Contains(err.Error(), "project=openspec/lite") || !strings.Contains(err.Error(), "install-state=lufy-sdd/lite") {
		t.Fatalf("expected tier-specific mismatch, got %v", err)
	}
}

func writeProjectHarness(t *testing.T, target string, tool domain.ToolID, methodologies domain.MethodologyByTier) {
	t.Helper()
	cfg := projectconfig.ProjectConfig{SchemaVersion: projectconfig.SchemaVersion, Tool: tool, MethodologyByTier: methodologies}
	if err := (projectconfig.ConfigStore{}).Write(projectconfig.Path(target), cfg); err != nil {
		t.Fatal(err)
	}
}

func writeInstallHarness(t *testing.T, target string, tool domain.ToolID, methodologies domain.MethodologyByTier) {
	t.Helper()
	st := state.InstallState{
		SchemaVersion:         state.SchemaVersion,
		Tool:                  tool,
		MethodologyByTier:     methodologies,
		TargetRoot:            target,
		SourceChangeID:        "test",
		SourceRootFingerprint: "test",
	}
	if err := state.WriteAtomic(target, st); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(target, ".lufy", "managed-state", "install-state.json")); err != nil {
		t.Fatal(err)
	}
}
