package harnessconfig

import (
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/projectconfig"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/state"
	"strings"
	"testing"
)

func TestResolvePrecedence(t *testing.T) {
	target := t.TempDir()
	writeInstallHarness(t, target, domain.DelegationOff)
	got, err := Resolve(Options{Target: target})
	if err != nil {
		t.Fatal(err)
	}
	if got.Config.Execution.Delegation != domain.DelegationOff || got.Config.Provenance.Execution != domain.HarnessSourceInstallState {
		t.Fatalf("state: %#v", got)
	}
	writeProjectHarness(t, target, domain.DelegationOnRequest)
	got, err = Resolve(Options{Target: target})
	if err != nil {
		t.Fatal(err)
	}
	if got.Config.Execution.Delegation != domain.DelegationOnRequest || got.Config.Provenance.Execution != domain.HarnessSourceProjectConfig {
		t.Fatalf("project: %#v", got)
	}
	got, err = Resolve(Options{Target: target, Requested: domain.HarnessConfig{Execution: domain.DefaultExecutionConfig()}, BlockOnMismatch: true})
	if err != nil {
		t.Fatal(err)
	}
	if got.Config.Execution.Delegation != domain.DelegationAuto || got.Config.Provenance.Execution != domain.HarnessSourceExplicit {
		t.Fatalf("explicit: %#v", got)
	}
	got, err = Resolve(Options{Target: t.TempDir()})
	if err != nil || got.Config.Tool != domain.ToolCodex || got.Config.Provenance.Execution != domain.HarnessSourceDefault {
		t.Fatalf("default: %#v %v", got, err)
	}
}
func TestResolveBlocksExecutionDrift(t *testing.T) {
	target := t.TempDir()
	writeProjectHarness(t, target, domain.DelegationOff)
	writeInstallHarness(t, target, domain.DelegationAuto)
	_, err := Resolve(Options{Target: target, BlockOnMismatch: true})
	if err == nil || !strings.Contains(err.Error(), "--delegation") {
		t.Fatalf("missing recovery: %v", err)
	}
}
func writeProjectHarness(t *testing.T, target string, d domain.DelegationPolicy) {
	t.Helper()
	c := projectconfig.ProjectConfig{SchemaVersion: projectconfig.SchemaVersion, Tool: domain.ToolCodex, Execution: domain.ExecutionConfig{Mode: domain.ExecutionInline, Delegation: d}}
	if err := (projectconfig.ConfigStore{}).Write(projectconfig.Path(target), c); err != nil {
		t.Fatal(err)
	}
}
func writeInstallHarness(t *testing.T, target string, d domain.DelegationPolicy) {
	t.Helper()
	s := state.InstallState{SchemaVersion: state.SchemaVersion, Tool: domain.ToolCodex, Execution: domain.ExecutionConfig{Mode: domain.ExecutionInline, Delegation: d}, TargetRoot: target}
	if err := state.WriteAtomic(target, s); err != nil {
		t.Fatal(err)
	}
}
