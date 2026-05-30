package lufysdd

import (
	"testing"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

func TestAdapterMetadata(t *testing.T) {
	adapter := New()
	if adapter.ID() != domain.MethodologyLufyWorkflow {
		t.Fatalf("id = %s", adapter.ID())
	}

	modes := adapter.SupportedModes()
	if len(modes) != 2 || modes[0] != domain.MethodologyModeFull || modes[1] != domain.MethodologyModeLite {
		t.Fatalf("modes = %#v", modes)
	}
}

func TestRenderWorkflowFullIncludesSpecs(t *testing.T) {
	assets, err := New().RenderWorkflow(ports.WorkflowModel{
		Tier:      domain.TierT1,
		Selection: domain.MethodologySelection{ID: domain.MethodologyLufyWorkflow, Mode: domain.MethodologyModeFull, Required: true},
	})
	if err != nil {
		t.Fatalf("render workflow: %v", err)
	}

	if !hasTarget(assets, ".lufy/sdd/specs") {
		t.Fatalf("full lufy-sdd assets missing specs: %#v", assets)
	}
	for _, target := range []string{".lufy/sdd/README.md", ".lufy/sdd/changes", ".lufy/sdd/decisions", ".lufy/sdd/verification"} {
		if !hasTarget(assets, target) {
			t.Fatalf("full lufy-sdd assets missing %s: %#v", target, assets)
		}
	}
	for _, asset := range assets {
		if asset.Policy != "managed" || asset.Scope != "project" {
			t.Fatalf("lufy-sdd asset should be installable: %+v", asset)
		}
	}
}

func TestRenderWorkflowLiteOmitsSpecs(t *testing.T) {
	assets, err := New().RenderWorkflow(ports.WorkflowModel{
		Tier:      domain.TierT2,
		Selection: domain.MethodologySelection{ID: domain.MethodologyLufyWorkflow, Mode: domain.MethodologyModeLite, Required: true},
	})
	if err != nil {
		t.Fatalf("render workflow: %v", err)
	}

	if hasTarget(assets, ".lufy/sdd/specs") {
		t.Fatalf("lite lufy-sdd assets include specs: %#v", assets)
	}
	if !hasTarget(assets, ".lufy/sdd/verification") {
		t.Fatalf("lite lufy-sdd assets missing verification: %#v", assets)
	}
}

func TestRenderWorkflowIgnoresOtherMethodologies(t *testing.T) {
	assets, err := New().RenderWorkflow(ports.WorkflowModel{
		Tier:      domain.TierT1,
		Selection: domain.MethodologySelection{ID: domain.MethodologySpecWorkflow, Mode: domain.MethodologyModeFull, Required: true},
	})
	if err != nil {
		t.Fatalf("render workflow: %v", err)
	}
	if len(assets) != 0 {
		t.Fatalf("other methodology assets = %#v, want empty", assets)
	}
}

func TestRenderWorkflowRejectsUnsupportedMode(t *testing.T) {
	_, err := New().RenderWorkflow(ports.WorkflowModel{
		Tier:      domain.TierT3,
		Selection: domain.MethodologySelection{ID: domain.MethodologyLufyWorkflow, Mode: domain.MethodologyModeNone},
	})
	if err == nil {
		t.Fatalf("expected unsupported mode error")
	}
}

func TestVerifyWorkflowReportsFoundationStatus(t *testing.T) {
	checks, err := New().VerifyWorkflow(ports.Target{Root: "."}, domain.TierT2)
	if err != nil {
		t.Fatalf("verify workflow: %v", err)
	}
	if len(checks) != 1 || checks[0].Level != "info" {
		t.Fatalf("checks = %#v", checks)
	}
}

func hasTarget(assets []ports.AssetSpec, target string) bool {
	for _, asset := range assets {
		if asset.TargetRel == target {
			return true
		}
	}
	return false
}
