package lufysdd

import (
	"fmt"

	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

type Adapter struct{}

func New() Adapter {
	return Adapter{}
}

func (Adapter) ID() domain.MethodologyID {
	return domain.MethodologyLufyWorkflow
}

func (Adapter) SupportedModes() []domain.MethodologyMode {
	return []domain.MethodologyMode{domain.MethodologyModeFull, domain.MethodologyModeLite}
}

func (Adapter) RenderWorkflow(model ports.WorkflowModel) ([]ports.AssetSpec, error) {
	if model.Selection.ID != domain.MethodologyLufyWorkflow {
		return nil, nil
	}

	switch model.Selection.Mode {
	case domain.MethodologyModeFull:
		return append(baseAssets(), ports.AssetSpec{
			ID:        "methodology.lufy-sdd.specs",
			TargetRel: ".lufy/sdd/specs",
			Policy:    "managed",
			Scope:     "project",
		}), nil
	case domain.MethodologyModeLite:
		return baseAssets(), nil
	default:
		return nil, fmt.Errorf("lufy-sdd no soporta mode %s", model.Selection.Mode)
	}
}

func (Adapter) VerifyWorkflow(ports.Target, domain.Tier) ([]ports.Check, error) {
	return []ports.Check{
		{Level: "info", Message: "lufy-sdd workflow verification is structural in the installer layer"},
	}, nil
}

func baseAssets() []ports.AssetSpec {
	return []ports.AssetSpec{
		{ID: "methodology.lufy-sdd.readme", TargetRel: ".lufy/sdd/README.md", Policy: "managed", Scope: "project"},
		{ID: "methodology.lufy-sdd.changes", TargetRel: ".lufy/sdd/changes", Policy: "managed", Scope: "project"},
		{ID: "methodology.lufy-sdd.decisions", TargetRel: ".lufy/sdd/decisions", Policy: "managed", Scope: "project"},
		{ID: "methodology.lufy-sdd.verification", TargetRel: ".lufy/sdd/verification", Policy: "managed", Scope: "project"},
	}
}
