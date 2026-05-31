package none

import (
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/core/domain"
	"github.com/adrotech/lufy-ai/tools/lufy-cli-go/internal/ports"
)

type Adapter struct{}

func New() Adapter {
	return Adapter{}
}

func (Adapter) ID() domain.MethodologyID {
	return domain.MethodologyNone
}

func (Adapter) SupportedModes() []domain.MethodologyMode {
	return []domain.MethodologyMode{domain.MethodologyModeNone}
}

func (Adapter) RenderWorkflow(ports.WorkflowModel) ([]ports.AssetSpec, error) {
	return nil, nil
}

func (Adapter) VerifyWorkflow(ports.Target, domain.Tier) ([]ports.Check, error) {
	return []ports.Check{
		{Level: "info", Message: "no formal methodology artifacts required"},
	}, nil
}
