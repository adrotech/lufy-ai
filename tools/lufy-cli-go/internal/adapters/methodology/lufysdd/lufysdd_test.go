package lufysdd

import (
	"os"
	"path/filepath"
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

	if !hasTarget(assets, ".lufy/workflows/sdd/specs") {
		t.Fatalf("full lufy-sdd assets missing specs: %#v", assets)
	}
	for _, target := range []string{".lufy/workflows/sdd/README.md", ".lufy/workflows/sdd/config.yaml", ".lufy/workflows/sdd/actions", ".lufy/workflows/sdd/templates", ".lufy/workflows/sdd/changes", ".lufy/workflows/sdd/decisions", ".lufy/workflows/sdd/verification", ".lufy/workflows/sdd/archive"} {
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

	if hasTarget(assets, ".lufy/workflows/sdd/specs") {
		t.Fatalf("lite lufy-sdd assets include specs: %#v", assets)
	}
	if !hasTarget(assets, ".lufy/workflows/sdd/verification") {
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

func TestVerifyWorkflowReportsReadyStatus(t *testing.T) {
	target := t.TempDir()
	root := filepath.Join(target, ".lufy", "workflows", "sdd")
	for _, rel := range []string{"actions", "templates", "changes", "decisions", "verification", "archive", "specs"} {
		if err := os.MkdirAll(filepath.Join(root, rel), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("ready\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config.yaml"), []byte("version: 1\nworkflow: lufy-sdd\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	checks, err := New().VerifyWorkflow(ports.Target{Root: target}, domain.TierT1)
	if err != nil {
		t.Fatalf("verify workflow: %v", err)
	}
	if len(checks) != 1 || checks[0].Level != "info" || checks[0].Message != "workflow Lufy SDD listo" {
		t.Fatalf("checks = %#v", checks)
	}
}

func TestVerifyWorkflowReportsMissingAsset(t *testing.T) {
	target := t.TempDir()
	checks, err := New().VerifyWorkflow(ports.Target{Root: target}, domain.TierT2)
	if err != nil {
		t.Fatalf("verify workflow: %v", err)
	}
	if len(checks) == 0 || checks[0].Level != "fail" {
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
