package openspec

import (
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

type Adapter struct{}

func New() Adapter {
	return Adapter{}
}

func (Adapter) ID() domain.MethodologyID {
	return domain.MethodologySpecWorkflow
}

func (Adapter) SupportedModes() []domain.MethodologyMode {
	return []domain.MethodologyMode{domain.MethodologyModeFull, domain.MethodologyModeLite}
}

func (Adapter) RenderWorkflow(model ports.WorkflowModel) ([]ports.AssetSpec, error) {
	if model.Selection.ID != domain.MethodologySpecWorkflow {
		return nil, nil
	}
	return []ports.AssetSpec{
		{ID: "methodology.spec.config", TargetRel: "openspec/config.yaml", Policy: "managed", Scope: "project"},
		{ID: "methodology.spec.readme", TargetRel: "openspec/README.md", Policy: "managed", Scope: "project"},
		{ID: "methodology.spec.upstream", TargetRel: "openspec/UPSTREAM.json", Policy: "managed", Scope: "project"},
		{ID: "methodology.spec.specs", TargetRel: "openspec/specs", Policy: "managed", Scope: "project"},
	}, nil
}

func (Adapter) VerifyWorkflow(ports.Target, domain.Tier) ([]ports.Check, error) {
	return []ports.Check{
		{Level: "info", Message: "openspec workflow verification is delegated to methodology commands"},
	}, nil
}
